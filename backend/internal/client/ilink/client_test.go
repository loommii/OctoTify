package ilink

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// ============================================================================
// 测试辅助函数
// ============================================================================

// setupTestClient 创建测试用的 iLink 客户端和日志观察者
func setupTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *observer.ObservedLogs) {
	t.Helper()

	iLinkServer := httptest.NewServer(handler)
	t.Cleanup(func() { iLinkServer.Close() })

	observedCore, logs := observer.New(zap.DebugLevel)
	logger := zap.New(observedCore)

	client := NewClient(logger, WithBaseURL(iLinkServer.URL))
	return client, logs
}

// ============================================================================
// TestClient_GetQRCode 测试获取二维码功能
// 基于真实 iLink API 响应数据（已脱敏）
// ============================================================================

func TestClient_GetQRCode(t *testing.T) {
	t.Run("成功：返回二维码", func(t *testing.T) {
		// 真实响应脱敏：{"qrcode":"test-qrcode-hash","qrcode_img_content":"https://liteapp.weixin.qq.com/q/7GiQu1?qrcode=test-qrcode-hash&bot_type=3","ret":0}
		handler := func(w http.ResponseWriter, r *http.Request) {
			// 验证请求参数
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/bot/get_bot_qrcode", r.URL.Path)
			assert.Equal(t, "3", r.URL.Query().Get("bot_type"))

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"qrcode":"test-qrcode-hash","qrcode_img_content":"https://liteapp.weixin.qq.com/q/7GiQu1?qrcode=test-qrcode-hash&bot_type=3","ret":0}`))
		}

		client, logs := setupTestClient(t, handler)
		ctx := context.Background()

		result, err := client.GetQRCode(ctx)

		require.NoError(t, err)
		assert.Equal(t, "test-qrcode-hash", result.QRCode)
		assert.Equal(t, "https://liteapp.weixin.qq.com/q/7GiQu1?qrcode=test-qrcode-hash&bot_type=3", result.QRCodeImgContent)
		assert.Equal(t, 0, result.RetCode)

		// 验证日志输出
		assert.Equal(t, 1, logs.FilterMessage("iLink GetQRCode 请求").Len())
		assert.Equal(t, 1, logs.FilterMessage("iLink GetQRCode 响应").Len())
		assert.Equal(t, 1, logs.FilterMessage("iLink GetQRCode 解析结果").Len())
	})

	t.Run("成功：verify request params", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/bot/get_bot_qrcode", r.URL.Path)
			assert.Equal(t, "3", r.URL.Query().Get("bot_type"))

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"qrcode":"verify-test","qrcode_img_content":"https://example.com/qr.png","ret":0}`))
		}

		client, _ := setupTestClient(t, handler)
		result, err := client.GetQRCode(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "verify-test", result.QRCode)
	})

	t.Run("失败：HTTP 500 错误", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal server error"}`))
		}

		client, _ := setupTestClient(t, handler)
		_, err := client.GetQRCode(context.Background())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 500")
	})

	t.Run("失败：网络请求失败", func(t *testing.T) {
		// 创建一个立即关闭的服务器来模拟网络错误
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		server.Close()

		observedCore, _ := observer.New(zap.DebugLevel)
		logger := zap.New(observedCore)
		client := NewClient(logger, WithBaseURL(server.URL))

		_, err := client.GetQRCode(context.Background())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "请求失败")
	})

	t.Run("失败：JSON 解析失败", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`not valid json`))
		}

		client, _ := setupTestClient(t, handler)
		_, err := client.GetQRCode(context.Background())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "解析响应失败")
	})
}

// ============================================================================
// TestClient_GetQRCodeStatus 测试查询二维码状态功能
// 基于真实 iLink API 响应数据（已脱敏）
// ============================================================================

func TestClient_GetQRCodeStatus(t *testing.T) {
	t.Run("成功：wait 状态", func(t *testing.T) {
		// 真实响应脱敏：{"ret":0,"status":"wait"}
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/bot/get_qrcode_status", r.URL.Path)
			assert.NotEmpty(t, r.URL.Query().Get("qrcode"))

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ret":0,"status":"wait"}`))
		}

		client, logs := setupTestClient(t, handler)
		ctx := context.Background()

		result, err := client.GetQRCodeStatus(ctx, "test-qrcode")

		require.NoError(t, err)
		assert.Equal(t, "wait", result.Status)
		assert.Empty(t, result.BotToken)
		assert.Empty(t, result.ILinkBotID)
		assert.Empty(t, result.ILinkUserID)

		// 验证日志输出
		assert.Equal(t, 1, logs.FilterMessage("iLink GetQRCodeStatus 请求").Len())
		assert.Equal(t, 1, logs.FilterMessage("iLink GetQRCodeStatus 响应").Len())
	})

	t.Run("成功：confirmed 状态", func(t *testing.T) {
		// 真实响应脱敏：{"baseurl":"https://ilinkai.weixin.qq.com","bot_token":"test-bot@im.bot:test-cred","ilink_bot_id":"test-bot@im.bot","ilink_user_id":"test-user@im.wechat","ret":0,"status":"confirmed"}
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"baseurl":"https://ilinkai.weixin.qq.com","bot_token":"test-bot@im.bot:test-cred","ilink_bot_id":"test-bot@im.bot","ilink_user_id":"test-user@im.wechat","ret":0,"status":"confirmed"}`))
		}

		client, _ := setupTestClient(t, handler)
		ctx := context.Background()

		result, err := client.GetQRCodeStatus(ctx, "test-qrcode")

		require.NoError(t, err)
		assert.Equal(t, "confirmed", result.Status)
		assert.Equal(t, "test-bot@im.bot:test-cred", result.BotToken)
		assert.Equal(t, "test-bot@im.bot", result.ILinkBotID)
		assert.Equal(t, "test-user@im.wechat", result.ILinkUserID)
	})

	t.Run("成功：expired 状态", func(t *testing.T) {
		// 真实响应脱敏：{"ret":0,"status":"expired"}
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ret":0,"status":"expired"}`))
		}

		client, _ := setupTestClient(t, handler)
		ctx := context.Background()

		result, err := client.GetQRCodeStatus(ctx, "test-qrcode")

		require.NoError(t, err)
		assert.Equal(t, "expired", result.Status)
		assert.Empty(t, result.BotToken)
		assert.Empty(t, result.ILinkBotID)
		assert.Empty(t, result.ILinkUserID)
	})

	t.Run("失败：HTTP 500 错误", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}

		client, _ := setupTestClient(t, handler)
		_, err := client.GetQRCodeStatus(context.Background(), "test-qrcode")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 500")
	})

	t.Run("失败：网络请求失败", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		server.Close()

		observedCore, _ := observer.New(zap.DebugLevel)
		logger := zap.New(observedCore)
		client := NewClient(logger, WithBaseURL(server.URL))

		_, err := client.GetQRCodeStatus(context.Background(), "test-qrcode")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "请求失败")
	})

	t.Run("失败：JSON 解析失败", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`not valid json`))
		}

		client, _ := setupTestClient(t, handler)
		_, err := client.GetQRCodeStatus(context.Background(), "test-qrcode")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "解析响应失败")
	})

	t.Run("失败：context 超时", func(t *testing.T) {
		// 模拟一个延迟响应，但 context 会先超时
		handler := func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Write([]byte(`{"ret":0,"status":"wait"}`))
		}

		client, _ := setupTestClient(t, handler)

		// 使用极短的超时时间来测试 context 取消
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := client.GetQRCodeStatus(ctx, "test-qrcode")

		require.Error(t, err)
		// context 超时或请求失败都是可接受的
		assert.True(t, err != nil)
	})
}

// ============================================================================
// TestClient_Options 测试客户端配置选项
// ============================================================================

func TestClient_WithBaseURL(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"qrcode":"baseurl-test","qrcode_img_content":"https://example.com","ret":0}`))
	}

	client, _ := setupTestClient(t, handler)
	result, err := client.GetQRCode(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "baseurl-test", result.QRCode)
}

// ============================================================================
// TestClient_SendMessage 测试发送消息功能
// 基于真实 iLink API 响应数据（已脱敏）
// ============================================================================

func TestClient_SendMessage(t *testing.T) {
	t.Run("成功：发送文本消息", func(t *testing.T) {
		var receivedAuth string
		var receivedAuthType string
		var receivedUIN string
		var receivedBody []byte

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/bot/sendmessage", r.URL.Path)

			receivedAuth = r.Header.Get("Authorization")
			receivedAuthType = r.Header.Get("AuthorizationType")
			receivedUIN = r.Header.Get("X-WECHAT-UIN")
			receivedBody, _ = io.ReadAll(r.Body)

			// 真实响应：成功时返回空 JSON {}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
		}

		client, logs := setupTestClient(t, handler)
		ctx := context.Background()

		req := &SendMessageRequest{
			Msg: SendMsg{
				FromUserID:   "bot123@im.bot",
				ToUserID:     "user123@im.wechat",
				ClientID:     "test-client-id-001",
				MessageType:  2,
				MessageState: 2,
				ItemList: []MessageItem{
					{
						Type: 1,
						TextItem: &TextItem{
							Text: "【测试消息】这是一条测试消息",
						},
					},
				},
			},
			BaseInfo: BaseInfo{
				ChannelVersion: "1.0.0",
			},
		}

		result, err := client.SendMessage(ctx, req, "test-bot-token")

		require.NoError(t, err)
		assert.Equal(t, 0, result.Ret)
		assert.Empty(t, result.ErrMsg)

		// 验证请求头
		assert.Equal(t, "Bearer test-bot-token", receivedAuth)
		assert.Equal(t, "ilink_bot_token", receivedAuthType)
		assert.NotEmpty(t, receivedUIN, "expected X-WECHAT-UIN header to be set")

		// 验证请求体
		var bodyMap map[string]interface{}
		require.NoError(t, json.Unmarshal(receivedBody, &bodyMap))

		msg := bodyMap["msg"].(map[string]interface{})
		assert.Equal(t, "bot123@im.bot", msg["from_user_id"])
		assert.Equal(t, "user123@im.wechat", msg["to_user_id"])
		assert.Equal(t, "test-client-id-001", msg["client_id"])

		// 验证日志
		assert.Equal(t, 1, logs.FilterMessage("iLink SendMessage 请求").Len())
		assert.Equal(t, 1, logs.FilterMessage("iLink SendMessage 响应").Len())
		assert.Equal(t, 1, logs.FilterMessage("iLink SendMessage 解析结果").Len())
	})

	t.Run("成功：验证请求体结构（基于真实数据脱敏）", func(t *testing.T) {
		var receivedBody []byte

		handler := func(w http.ResponseWriter, r *http.Request) {
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ret":0,"errmsg":""}`))
		}

		client, _ := setupTestClient(t, handler)

		req := &SendMessageRequest{
			Msg: SendMsg{
				FromUserID:   "bot123@im.bot",
				ToUserID:     "user123@im.wechat",
				ClientID:     "unique-client-id",
				MessageType:  2,
				MessageState: 2,
				ItemList: []MessageItem{
					{
						Type: 1,
						TextItem: &TextItem{
							Text: "【OctoTify 测试消息】\n这是一条测试消息，用于验证渠道配置是否正确。",
						},
					},
				},
			},
			BaseInfo: BaseInfo{
				ChannelVersion: "1.0.0",
			},
		}

		_, err := client.SendMessage(context.Background(), req, "test-token")
		require.NoError(t, err)

		// 验证请求体与真实数据结构一致
		var bodyMap map[string]interface{}
		require.NoError(t, json.Unmarshal(receivedBody, &bodyMap))

		// 验证顶层结构
		assert.Contains(t, bodyMap, "msg")
		assert.Contains(t, bodyMap, "base_info")

		// 验证 msg 结构
		msg := bodyMap["msg"].(map[string]interface{})
		assert.Contains(t, msg, "from_user_id")
		assert.Contains(t, msg, "to_user_id")
		assert.Contains(t, msg, "client_id")
		assert.Contains(t, msg, "message_type")
		assert.Contains(t, msg, "message_state")
		assert.Contains(t, msg, "item_list")

		// 验证 base_info 结构
		baseInfo := bodyMap["base_info"].(map[string]interface{})
		assert.Equal(t, "1.0.0", baseInfo["channel_version"])
	})

	t.Run("失败：HTTP 500 错误", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal server error"}`))
		}

		client, _ := setupTestClient(t, handler)
		req := &SendMessageRequest{
			Msg: SendMsg{
				FromUserID:   "bot123@im.bot",
				ToUserID:     "user123@im.wechat",
				ClientID:     "test-id",
				MessageType:  2,
				MessageState: 2,
				ItemList:     []MessageItem{},
			},
		}

		_, err := client.SendMessage(context.Background(), req, "test-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 500")
	})

	t.Run("失败：网络请求失败", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		server.Close()

		observedCore, _ := observer.New(zap.DebugLevel)
		logger := zap.New(observedCore)
		client := NewClient(logger, WithBaseURL(server.URL))

		req := &SendMessageRequest{
			Msg: SendMsg{
				FromUserID:   "bot123@im.bot",
				ToUserID:     "user123@im.wechat",
				ClientID:     "test-id",
				MessageType:  2,
				MessageState: 2,
				ItemList:     []MessageItem{},
			},
		}

		_, err := client.SendMessage(context.Background(), req, "test-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "请求失败")
	})

	t.Run("失败：JSON 解析失败", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`not valid json`))
		}

		client, _ := setupTestClient(t, handler)
		req := &SendMessageRequest{
			Msg: SendMsg{
				FromUserID:   "bot123@im.bot",
				ToUserID:     "user123@im.wechat",
				ClientID:     "test-id",
				MessageType:  2,
				MessageState: 2,
				ItemList:     []MessageItem{},
			},
		}

		_, err := client.SendMessage(context.Background(), req, "test-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "解析响应失败")
	})

	t.Run("失败：业务错误（ret != 0）", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ret":10001,"errmsg":"quota exceeded"}`))
		}

		client, _ := setupTestClient(t, handler)
		req := &SendMessageRequest{
			Msg: SendMsg{
				FromUserID:   "bot123@im.bot",
				ToUserID:     "user123@im.wechat",
				ClientID:     "test-id",
				MessageType:  2,
				MessageState: 2,
				ItemList:     []MessageItem{},
			},
		}

		result, err := client.SendMessage(context.Background(), req, "test-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "iLink 推送失败")
		assert.Contains(t, err.Error(), "quota exceeded")
		assert.Equal(t, 10001, result.Ret)
		assert.Equal(t, "quota exceeded", result.ErrMsg)
	})

	t.Run("成功：空响应体 {} 视为成功", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
		}

		client, _ := setupTestClient(t, handler)
		req := &SendMessageRequest{
			Msg: SendMsg{
				FromUserID:   "bot123@im.bot",
				ToUserID:     "user123@im.wechat",
				ClientID:     "test-id",
				MessageType:  2,
				MessageState: 2,
				ItemList:     []MessageItem{},
			},
		}

		result, err := client.SendMessage(context.Background(), req, "test-token")
		require.NoError(t, err)
		assert.Equal(t, 0, result.Ret)
	})

	t.Run("失败：context 超时", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Write([]byte(`{"ret":0}`))
		}

		client, _ := setupTestClient(t, handler)
		req := &SendMessageRequest{
			Msg: SendMsg{
				FromUserID:   "bot123@im.bot",
				ToUserID:     "user123@im.wechat",
				ClientID:     "test-id",
				MessageType:  2,
				MessageState: 2,
				ItemList:     []MessageItem{},
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := client.SendMessage(ctx, req, "test-token")
		require.Error(t, err)
	})
}
