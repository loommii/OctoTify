package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"octotify/internal/client/ilink"
	"octotify/internal/middleware"
	"octotify/internal/sender"
	"octotify/internal/service"
	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

// handlerRedirectTransport 将 HTTP 请求重定向到测试服务器的 RoundTripper
// 用于模拟 iLink API，无需修改生产代码中的硬编码 URL
type handlerRedirectTransport struct {
	targetURL string
}

func (t *handlerRedirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, _ := url.Parse(t.targetURL)
	newURL := &url.URL{
		Scheme:   target.Scheme,
		Host:     target.Host,
		Path:     req.URL.Path,
		RawQuery: req.URL.RawQuery,
	}
	newReq := req.Clone(req.Context())
	newReq.URL = newURL
	return http.DefaultTransport.RoundTrip(newReq)
}

// int64Ptr 返回 int64 的指针，用于 table-driven test 中的可选 userID
func int64Ptr(v int64) *int64 {
	return &v
}

// setupBindHandlerTest 设置绑定流程 Handler 测试环境
// 创建模拟 iLink 服务器、ChannelService、ChannelHandler 和 Gin 路由
func setupBindHandlerTest(t *testing.T, iLinkHandler http.HandlerFunc, userID *int64) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// 创建模拟 iLink 服务器
	iLinkServer := httptest.NewServer(iLinkHandler)
	t.Cleanup(func() { iLinkServer.Close() })

	// 创建测试 DB 和 Logger
	db := service.SetupTestDB(t)
	logger := service.SetupTestLogger(t)
	factory := sender.NewSenderFactory(logger)

	// 创建 ChannelService，HTTP 请求重定向到模拟 iLink 服务器
	transport := &handlerRedirectTransport{targetURL: iLinkServer.URL}
	ilinkClient := ilink.NewClient(
		ilink.WithHTTPClient(&http.Client{Transport: transport}),
		ilink.WithBaseURL(iLinkServer.URL),
	)
	svc := service.NewChannelServiceForTest(
		db, logger, factory,
		ilinkClient,
	)

	// 创建 ChannelHandler
	handler := NewChannelHandler(svc, logger)

	// 创建 Gin 路由
	r := gin.New()
	r.Use(middleware.ErrorHandler(logger))

	// 如果提供了 userID，添加模拟认证中间件
	if userID != nil {
		r.Use(func(c *gin.Context) {
			c.Set(middleware.ContextKeyUserID, strconv.FormatInt(*userID, 10))
			c.Next()
		})
	}

	// 注册绑定相关路由
	r.POST("/channels/wechat-clawbot/bind", handler.StartBind)
	r.POST("/channels/wechat-clawbot/bind/status", handler.GetBindStatus)

	return r
}

// parseResponse 解析 HTTP 响应为 response.Response
func parseResponse(t *testing.T, w *httptest.ResponseRecorder) response.Response {
	t.Helper()
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "解析响应体失败")
	return resp
}

// TestChannelHandler_StartBind 测试发起微信ClawBot扫码绑定的 Handler 层
// 覆盖场景：成功获取二维码、未认证
func TestChannelHandler_StartBind(t *testing.T) {
	tests := []struct {
		name         string
		userID       *int64 // nil 表示未认证
		iLinkHandler http.HandlerFunc
		wantHTTPCode int
		wantBizCode  int
		wantQRCode   string
		wantQRURL    string
	}{
		{
			name:   "成功：返回 qrcode_url 和 qrcode",
			userID: int64Ptr(1),
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"qrcode":             "handler-test-qrcode",
					"qrcode_img_content": "data:image/png;base64,TEST",
				})
			},
			wantHTTPCode: http.StatusOK,
			wantBizCode:  0,
			wantQRCode:   "handler-test-qrcode",
			wantQRURL:    "data:image/png;base64,TEST",
		},
		{
			name:         "失败：未认证",
			userID:       nil,
			iLinkHandler: nil, // 不需要 iLink 服务器
			wantHTTPCode: http.StatusOK,
			wantBizCode:  xerr.ErrUnauthorized.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为不需要 iLink 的用例提供默认 handler
			iLinkHandler := tt.iLinkHandler
			if iLinkHandler == nil {
				iLinkHandler = func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}
			}

			r := setupBindHandlerTest(t, iLinkHandler, tt.userID)

			req := httptest.NewRequest(http.MethodPost, "/channels/wechat-clawbot/bind", nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantHTTPCode, w.Code)

			resp := parseResponse(t, w)
			assert.Equal(t, tt.wantBizCode, resp.Code)

			if tt.wantBizCode == 0 {
				dataMap, ok := resp.Data.(map[string]interface{})
				require.True(t, ok, "响应 data 应为 map")
				assert.Equal(t, tt.wantQRCode, dataMap["qrcode"])
				assert.Equal(t, tt.wantQRURL, dataMap["qrcode_url"])
			}
		})
	}
}

// TestChannelHandler_GetBindStatus 测试查询微信ClawBot绑定状态的 Handler 层
// 覆盖场景：pending、confirmed（含凭证）、未认证、缺少 qrcode 参数、iLink API 失败降级
func TestChannelHandler_GetBindStatus(t *testing.T) {
	tests := []struct {
		name         string
		userID       *int64 // nil 表示未认证
		requestBody  string
		iLinkHandler http.HandlerFunc
		wantHTTPCode int
		wantBizCode  int
		wantStatus   string // 期望的绑定状态
		wantCreds    bool   // 是否期望返回凭证
	}{
		{
			name:        "成功：返回 pending 状态",
			userID:      int64Ptr(1),
			requestBody: `{"qrcode": "test-qr"}`,
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"status": ilink.StatusWait,
				})
			},
			wantHTTPCode: http.StatusOK,
			wantBizCode:  0,
			wantStatus:   ilink.StatusWait,
		},
		{
			name:        "成功：返回 confirmed 状态和凭证",
			userID:      int64Ptr(1),
			requestBody: `{"qrcode": "test-qr"}`,
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"status":        ilink.StatusConfirmed,
					"bot_token":     "handler-bot-token",
					"ilink_bot_id":  "h-bot-1",
					"ilink_user_id": "h-user-1",
				})
			},
			wantHTTPCode: http.StatusOK,
			wantBizCode:  0,
			wantStatus:   ilink.StatusConfirmed,
			wantCreds:    true,
		},
		{
			name:         "失败：未认证",
			userID:       nil,
			requestBody:  `{"qrcode": "test-qr"}`,
			iLinkHandler: nil,
			wantHTTPCode: http.StatusOK,
			wantBizCode:  xerr.ErrUnauthorized.Code,
		},
		{
			name:         "失败：缺少 qrcode 参数",
			userID:       int64Ptr(1),
			requestBody:  `{}`,
			iLinkHandler: nil,
			wantHTTPCode: http.StatusOK,
			wantBizCode:  xerr.ErrBadRequest.Code,
		},
		{
			name:        "失败：iLink API 失败返回业务错误",
			userID:      int64Ptr(1),
			requestBody: `{"qrcode": "test-qr"}`,
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantHTTPCode: http.StatusOK,
			wantBizCode:  xerr.ErrBindStatusFailed.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为不需要 iLink 的用例提供默认 handler
			iLinkHandler := tt.iLinkHandler
			if iLinkHandler == nil {
				iLinkHandler = func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}
			}

			r := setupBindHandlerTest(t, iLinkHandler, tt.userID)

			req := httptest.NewRequest(http.MethodPost, "/channels/wechat-clawbot/bind/status", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantHTTPCode, w.Code)

			resp := parseResponse(t, w)
			assert.Equal(t, tt.wantBizCode, resp.Code)

			if tt.wantBizCode == 0 && tt.wantStatus != "" {
				dataMap, ok := resp.Data.(map[string]interface{})
				require.True(t, ok, "响应 data 应为 map")
				assert.Equal(t, tt.wantStatus, dataMap["status"])

				if tt.wantCreds {
					credsMap, ok := dataMap["credentials"].(map[string]interface{})
					require.True(t, ok, "confirmed 状态应包含 credentials")
					assert.NotEmpty(t, credsMap["bot_token_ciphertext"], "bot_token_ciphertext 不应为空")
					assert.NotEmpty(t, credsMap["bot_token_nonce"], "bot_token_nonce 不应为空")
					assert.Equal(t, "h-bot-1", credsMap["ilink_bot_id"])
					assert.Equal(t, "h-user-1", credsMap["ilink_user_id"])
				}
			}
		})
	}
}
