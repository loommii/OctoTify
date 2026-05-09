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

	"go.uber.org/zap/zaptest"
	"gorm.io/datatypes"
)

func TestWechatClawbotSender_New(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatClawbotSender(logger)

	if sender == nil {
		t.Fatal("expected sender instance, got nil")
	}
	if sender.logger == nil {
		t.Fatal("expected logger to be set")
	}
}

func TestWechatClawbotSender_ConfigValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatClawbotSender(logger)

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
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Send() error = %v, want error containing %q", err, tt.errContains)
			}
		})
	}
}

func TestWechatClawbotSender_Send_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatClawbotSender(logger)

	var mu sync.Mutex
	var receivedBody string
	var receivedAuth string
	var requestMethod string
	var requestPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestMethod = r.Method
		requestPath = r.URL.Path
		receivedAuth = r.Header.Get("Authorization")
		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBody = string(bodyBytes)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ret":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	oldBaseURL := wechatClawbotBaseURL
	wechatClawbotBaseURL = server.URL
	defer func() { wechatClawbotBaseURL = oldBaseURL }()

	config := datatypes.JSON(`{
		"bot_token": "test_token_12345",
		"ilink_bot_id": "bot123@im.bot",
		"ilink_user_id": "user123@im.wechat"
	}`)

	err := sender.Send(context.Background(), config, "Test Title", "Hello from OctoTify!")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if requestMethod != http.MethodPost {
		t.Errorf("expected method POST, got %q", requestMethod)
	}

	expectedPath := "/ilink/bot/sendmessage"
	if requestPath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, requestPath)
	}

	if receivedAuth != "Bearer test_token_12345" {
		t.Errorf("expected Authorization 'Bearer test_token_12345', got %q", receivedAuth)
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal([]byte(receivedBody), &bodyMap); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}

	msg, ok := bodyMap["msg"].(map[string]interface{})
	if !ok {
		t.Fatal("expected msg field in body")
	}
	if msg["from_user_id"] != "bot123@im.bot" {
		t.Errorf("expected from_user_id 'bot123@im.bot', got %q", msg["from_user_id"])
	}
	if msg["to_user_id"] != "user123@im.wechat" {
		t.Errorf("expected to_user_id 'user123@im.wechat', got %q", msg["to_user_id"])
	}

	itemList := msg["item_list"].([]interface{})
	if len(itemList) == 0 {
		t.Fatal("expected item_list to have at least one item")
	}
	item := itemList[0].(map[string]interface{})
	textItem := item["text_item"].(map[string]interface{})
	expectedContent := "【Test Title】\nHello from OctoTify!"
	if textItem["text"] != expectedContent {
		t.Errorf("expected text %q, got %q", expectedContent, textItem["text"])
	}

	if receivedAuth == "" {
		t.Error("expected Authorization header to be set")
	}
}

func TestWechatClawbotSender_Send_TruncateLongContent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatClawbotSender(logger)

	var mu sync.Mutex
	var capturedContent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		var msg ilinkSendMessageRequest
		json.Unmarshal(body, &msg)
		mu.Lock()
		if len(msg.Msg.ItemList) > 0 && msg.Msg.ItemList[0].TextItem != nil {
			capturedContent = msg.Msg.ItemList[0].TextItem.Text
		}
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ret":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	oldBaseURL := wechatClawbotBaseURL
	wechatClawbotBaseURL = server.URL
	defer func() { wechatClawbotBaseURL = oldBaseURL }()

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	longContent := strings.Repeat("a", 3000)

	err := sender.Send(context.Background(), config, "Title", longContent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	expectedLen := wechatClawbotMaxContentLen
	if utf8.RuneCountInString(capturedContent) != expectedLen {
		t.Errorf("expected content length %d (chars), got %d (chars)", expectedLen, utf8.RuneCountInString(capturedContent))
	}

	expectedSuffix := wechatClawbotTruncateSuffix
	if !strings.HasSuffix(capturedContent, expectedSuffix) {
		t.Errorf("expected content to end with %q, got last 20 chars: %q", expectedSuffix, capturedContent[len(capturedContent)-20:])
	}
}

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
			logger := zaptest.NewLogger(t)
			sender := NewWechatClawbotSender(logger)

			var mu sync.Mutex
			var capturedContent string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				r.Body.Close()
				var msg ilinkSendMessageRequest
				json.Unmarshal(body, &msg)
				mu.Lock()
				if len(msg.Msg.ItemList) > 0 && msg.Msg.ItemList[0].TextItem != nil {
					capturedContent = msg.Msg.ItemList[0].TextItem.Text
				}
				mu.Unlock()
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"ret":0,"errmsg":"ok"}`))
			}))
			defer server.Close()

			oldBaseURL := wechatClawbotBaseURL
			wechatClawbotBaseURL = server.URL
			defer func() { wechatClawbotBaseURL = oldBaseURL }()

			config := datatypes.JSON(`{
				"bot_token": "test-token",
				"ilink_bot_id": "bot123",
				"ilink_user_id": "user123"
			}`)

			err := sender.Send(context.Background(), config, tt.title, tt.content)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			mu.Lock()
			defer mu.Unlock()

			if utf8.RuneCountInString(capturedContent) != tt.expectLen {
				t.Errorf("[%s] expected length %d (chars), got %d (chars)", tt.name, tt.expectLen, utf8.RuneCountInString(capturedContent))
			}

			if tt.expectTrunc && !strings.HasSuffix(capturedContent, wechatClawbotTruncateSuffix) {
				t.Errorf("[%s] expected truncated content", tt.name)
			}
		})
	}
}

func TestWechatClawbotSender_Send_BusinessError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatClawbotSender(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ret":10001,"errmsg":"quota exceeded"}`))
	}))
	defer server.Close()

	oldBaseURL := wechatClawbotBaseURL
	wechatClawbotBaseURL = server.URL
	defer func() { wechatClawbotBaseURL = oldBaseURL }()

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	err := sender.Send(context.Background(), config, "Title", "Content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErr := "微信ClawBot推送失败: quota exceeded (ret: 10001)"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestWechatClawbotSender_Send_HttpError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatClawbotSender(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal Server Error`))
	}))
	defer server.Close()

	oldBaseURL := wechatClawbotBaseURL
	wechatClawbotBaseURL = server.URL
	defer func() { wechatClawbotBaseURL = oldBaseURL }()

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	err := sender.Send(context.Background(), config, "Title", "Content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "HTTP错误") {
		t.Errorf("expected error containing 'HTTP错误', got %q", err.Error())
	}
}

func TestWechatClawbotSender_Send_EmptyResponse(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatClawbotSender(logger)

	tests := []struct {
		name         string
		responseBody string
	}{
		{
			name:         "空响应体",
			responseBody: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			oldBaseURL := wechatClawbotBaseURL
			wechatClawbotBaseURL = server.URL
			defer func() { wechatClawbotBaseURL = oldBaseURL }()

			config := datatypes.JSON(`{
				"bot_token": "test-token",
				"ilink_bot_id": "bot123",
				"ilink_user_id": "user123"
			}`)

			err := sender.Send(context.Background(), config, "Title", "Content")
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tt.name)
			}

			if !strings.Contains(err.Error(), "空响应") {
				t.Errorf("[%s] expected error containing '空响应', got %q", tt.name, err.Error())
			}
		})
	}
}

func TestWechatClawbotSender_Send_InvalidResponse(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatClawbotSender(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not a json response`))
	}))
	defer server.Close()

	oldBaseURL := wechatClawbotBaseURL
	wechatClawbotBaseURL = server.URL
	defer func() { wechatClawbotBaseURL = oldBaseURL }()

	config := datatypes.JSON(`{
		"bot_token": "test-token",
		"ilink_bot_id": "bot123",
		"ilink_user_id": "user123"
	}`)

	err := sender.Send(context.Background(), config, "Title", "Content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "解析微信ClawBot响应失败") {
		t.Errorf("expected error containing '解析微信ClawBot响应失败', got %q", err.Error())
	}
}
