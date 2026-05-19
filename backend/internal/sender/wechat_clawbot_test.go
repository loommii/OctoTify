package sender

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"octotify/internal/client/ilink"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gorm.io/datatypes"
)

// setupTestSender 创建测试用的 WechatClawbotSender 和 mock 服务器
func setupTestSender(t *testing.T, handler http.HandlerFunc) (*WechatClawbotSender, *httptest.Server) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(func() { server.Close() })

	logger := zaptest.NewLogger(t)
	ilinkClient := ilink.NewClient(logger, ilink.WithBaseURL(server.URL))
	sender := NewWechatClawbotSender(logger, ilinkClient)

	return sender, server
}

// ============================================================================
// 构造函数测试
// ============================================================================

func TestWechatClawbotSender_New(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ilinkClient := ilink.NewClient(logger)
	sender := NewWechatClawbotSender(logger, ilinkClient)

	require.NotNil(t, sender, "expected sender instance, got nil")
	assert.NotNil(t, sender.logger, "expected logger to be set")
	assert.NotNil(t, sender.ilinkClient, "expected ilinkClient to be set")
}

// ============================================================================
// 配置校验测试
// ============================================================================

func TestWechatClawbotSender_ConfigValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ilinkClient := ilink.NewClient(logger)
	sender := NewWechatClawbotSender(logger, ilinkClient)

	tests := []struct {
		name        string
		config      datatypes.JSON
		wantErr     bool
		errContains string
	}{
		{
			name:        "BotToken为空",
			config:      datatypes.JSON(`{"ilink_bot_id":"bot123","ilink_user_id":"user123"}`),
			wantErr:     true,
			errContains: "bot_token 不能为空",
		},
		{
			name: "IlinkBotID为空",
			config: datatypes.JSON(`{
				"bot_token": "test-token",
				"ilink_user_id": "user123"
			}`),
			wantErr:     true,
			errContains: "IlinkBotID 不能为空",
		},
		{
			name: "IlinkUserID为空",
			config: datatypes.JSON(`{
				"bot_token": "test-token",
				"ilink_bot_id": "bot123"
			}`),
			wantErr:     true,
			errContains: "IlinkUserID 不能为空",
		},
		{
			name:        "无效JSON",
			config:      datatypes.JSON(`{invalid`),
			wantErr:     true,
			errContains: "解析微信ClawBot渠道配置失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.Send(context.Background(), tt.config, "title", "content")
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// 发送成功测试
// ============================================================================

func TestWechatClawbotSender_Send_Success(t *testing.T) {
	var mu sync.Mutex
	var receivedBody string
	var receivedAuth string
	var receivedAuthType string
	var receivedUIN string
	var requestMethod string
	var requestPath string
	var capturedClientIDs []string

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestMethod = r.Method
		requestPath = r.URL.Path
		receivedAuth = r.Header.Get("Authorization")
		receivedAuthType = r.Header.Get("AuthorizationType")
		receivedUIN = r.Header.Get("X-WECHAT-UIN")
		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBody = string(bodyBytes)
		mu.Unlock()

		// 提取 ClientID 用于唯一性验证
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
			if msg, ok := bodyMap["msg"].(map[string]interface{}); ok {
				if clientID, ok := msg["client_id"].(string); ok {
					mu.Lock()
					capturedClientIDs = append(capturedClientIDs, clientID)
					mu.Unlock()
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ret":0,"errmsg":"ok"}`))
	}

	sender, _ := setupTestSender(t, handler)

	config := datatypes.JSON(`{
		"bot_token": "test_token_12345",
		"ilink_bot_id": "bot123@im.bot",
		"ilink_user_id": "user123@im.wechat"
	}`)

	err := sender.Send(context.Background(), config, "Test Title", "Hello from OctoTify!")
	require.NoError(t, err, "expected no error")

	mu.Lock()
	defer mu.Unlock()

	// 验证请求方法和路径
	assert.Equal(t, http.MethodPost, requestMethod)
	assert.Equal(t, "/bot/sendmessage", requestPath)

	// 验证认证请求头
	assert.Equal(t, "Bearer test_token_12345", receivedAuth)
	assert.Equal(t, "ilink_bot_token", receivedAuthType)
	assert.NotEmpty(t, receivedUIN, "expected X-WECHAT-UIN header to be set")

	// 验证请求体
	var bodyMap map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(receivedBody), &bodyMap))

	msg, ok := bodyMap["msg"].(map[string]interface{})
	require.True(t, ok, "expected msg field in body")

	assert.Equal(t, "bot123@im.bot", msg["from_user_id"])
	assert.Equal(t, "user123@im.wechat", msg["to_user_id"])
	assert.Equal(t, float64(2), msg["message_type"])
	assert.Equal(t, float64(2), msg["message_state"])

	// 验证消息内容
	itemList := msg["item_list"].([]interface{})
	require.Len(t, itemList, 1, "expected item_list to have one item")
	item := itemList[0].(map[string]interface{})
	assert.Equal(t, float64(1), item["type"])
	textItem := item["text_item"].(map[string]interface{})
	expectedContent := "【Test Title】\nHello from OctoTify!"
	assert.Equal(t, expectedContent, textItem["text"])

	// 验证 base_info
	baseInfo := bodyMap["base_info"].(map[string]interface{})
	assert.Equal(t, "1.0.0", baseInfo["channel_version"])

	// 验证 ClientID 是有效的 UUID
	clientID := msg["client_id"].(string)
	_, err = uuid.Parse(clientID)
	assert.NoError(t, err, "expected client_id to be a valid UUID, got: %s", clientID)
}

// ============================================================================
// 超长内容截断测试
// ============================================================================

func TestWechatClawbotSender_Send_TruncateLongContent(t *testing.T) {
	var mu sync.Mutex
	var capturedContent string

	handler := func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()

		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			if msg, ok := bodyMap["msg"].(map[string]interface{}); ok {
				if itemList, ok := msg["item_list"].([]interface{}); ok && len(itemList) > 0 {
					if item, ok := itemList[0].(map[string]interface{}); ok {
						if textItem, ok := item["text_item"].(map[string]interface{}); ok {
							mu.Lock()
							capturedContent = textItem["text"].(string)
							mu.Unlock()
						}
					}
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ret":0,"errmsg":"ok"}`))
	}

	sender, _ := setupTestSender(t, handler)

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	longContent := strings.Repeat("a", 3000)
	err := sender.Send(context.Background(), config, "Title", longContent)
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()

	expectedLen := wechatClawbotMaxContentLen
	assert.Equal(t, expectedLen, utf8.RuneCountInString(capturedContent),
		"expected content length %d (chars), got %d (chars)", expectedLen, utf8.RuneCountInString(capturedContent))

	assert.True(t, strings.HasSuffix(capturedContent, wechatClawbotTruncateSuffix),
		"expected content to end with %q, got last 20 chars: %q",
		wechatClawbotTruncateSuffix, capturedContent[len(capturedContent)-20:])
}

// ============================================================================
// 边界值测试
// ============================================================================

func TestWechatClawbotSender_Send_BoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		content     string
		expectTrunc bool
		expectLen   int
	}{
		{
			name:        "空内容",
			title:       "Title",
			content:     "",
			expectTrunc: false,
			expectLen:   8,
		},
		{
			name:        "刚好达到限制",
			title:       "Title",
			content:     strings.Repeat("a", wechatClawbotMaxContentLen-8),
			expectTrunc: false,
			expectLen:   wechatClawbotMaxContentLen,
		},
		{
			name:        "刚好超过限制",
			title:       "Title",
			content:     strings.Repeat("a", wechatClawbotMaxContentLen-7),
			expectTrunc: true,
			expectLen:   wechatClawbotMaxContentLen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			var capturedContent string

			handler := func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				r.Body.Close()

				var bodyMap map[string]interface{}
				if err := json.Unmarshal(body, &bodyMap); err == nil {
					if msg, ok := bodyMap["msg"].(map[string]interface{}); ok {
						if itemList, ok := msg["item_list"].([]interface{}); ok && len(itemList) > 0 {
							if item, ok := itemList[0].(map[string]interface{}); ok {
								if textItem, ok := item["text_item"].(map[string]interface{}); ok {
									mu.Lock()
									capturedContent = textItem["text"].(string)
									mu.Unlock()
								}
							}
						}
					}
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ret":0,"errmsg":"ok"}`))
			}

			sender, _ := setupTestSender(t, handler)

			config := datatypes.JSON(`{
				"bot_token": "test-token",
				"ilink_bot_id": "bot123",
				"ilink_user_id": "user123"
			}`)

			err := sender.Send(context.Background(), config, tt.title, tt.content)
			require.NoError(t, err)

			mu.Lock()
			defer mu.Unlock()

			assert.Equal(t, tt.expectLen, utf8.RuneCountInString(capturedContent),
				"[%s] expected length %d (chars), got %d (chars)", tt.name, tt.expectLen, utf8.RuneCountInString(capturedContent))

			if tt.expectTrunc {
				assert.True(t, strings.HasSuffix(capturedContent, wechatClawbotTruncateSuffix),
					"[%s] expected truncated content", tt.name)
			}
		})
	}
}

// ============================================================================
// ClientID 唯一性测试
// ============================================================================

func TestWechatClawbotSender_Send_ClientIDUnique(t *testing.T) {
	var mu sync.Mutex
	var capturedClientIDs []string

	handler := func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()

		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			if msg, ok := bodyMap["msg"].(map[string]interface{}); ok {
				if clientID, ok := msg["client_id"].(string); ok {
					mu.Lock()
					capturedClientIDs = append(capturedClientIDs, clientID)
					mu.Unlock()
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ret":0,"errmsg":"ok"}`))
	}

	sender, _ := setupTestSender(t, handler)

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	// 连续发送两次
	err := sender.Send(context.Background(), config, "Title1", "Content1")
	require.NoError(t, err)

	err = sender.Send(context.Background(), config, "Title2", "Content2")
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, capturedClientIDs, 2, "expected two ClientIDs")
	assert.NotEqual(t, capturedClientIDs[0], capturedClientIDs[1],
		"expected ClientIDs to be unique, got: %s and %s", capturedClientIDs[0], capturedClientIDs[1])

	// 验证两个都是有效的 UUID
	_, err = uuid.Parse(capturedClientIDs[0])
	assert.NoError(t, err, "first ClientID should be a valid UUID")
	_, err = uuid.Parse(capturedClientIDs[1])
	assert.NoError(t, err, "second ClientID should be a valid UUID")
}

// ============================================================================
// 业务错误测试
// ============================================================================

func TestWechatClawbotSender_Send_BusinessError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":10001,"errmsg":"quota exceeded"}`))
	}

	sender, _ := setupTestSender(t, handler)

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	err := sender.Send(context.Background(), config, "Title", "Content")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "发送微信ClawBot消息失败")
	assert.Contains(t, err.Error(), "iLink 推送失败")
	assert.Contains(t, err.Error(), "quota exceeded")
}

// ============================================================================
// HTTP 错误测试
// ============================================================================

func TestWechatClawbotSender_Send_HttpError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal Server Error`))
	}

	sender, _ := setupTestSender(t, handler)

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	err := sender.Send(context.Background(), config, "Title", "Content")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 500")
}

// ============================================================================
// 空响应体测试（现在视为成功，因为 {} 解析后 ret=0）
// ============================================================================

func TestWechatClawbotSender_Send_EmptyResponse(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}

	sender, _ := setupTestSender(t, handler)

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	err := sender.Send(context.Background(), config, "Title", "Content")
	// 空 JSON {} 解析后 ret 默认为 0，视为成功
	require.NoError(t, err, "empty JSON response should be treated as success (ret=0)")
}

// ============================================================================
// 无效响应测试
// ============================================================================

func TestWechatClawbotSender_Send_InvalidResponse(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not a json response`))
	}

	sender, _ := setupTestSender(t, handler)

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	err := sender.Send(context.Background(), config, "Title", "Content")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "发送微信ClawBot消息失败")
	assert.Contains(t, err.Error(), "解析响应失败")
}
