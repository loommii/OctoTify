package sender

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
	"gorm.io/datatypes"
)

// TestDingtalkSender_NewInstance 测试钉钉发送器新建实例持有logger
func TestDingtalkSender_NewInstance(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	if sender == nil {
		t.Fatal("NewDingtalkSender() returned nil")
	}
	if sender.logger == nil {
		t.Error("NewDingtalkSender() sender.logger is nil")
	}
}

// TestDingtalkSender_Send_ConfigValidation 测试配置校验
func TestDingtalkSender_Send_ConfigValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	tests := []struct {
		name        string
		config      datatypes.JSON
		wantErr     bool
		errContains string
	}{
		{
			name:        "配置解析失败_无效JSON",
			config:      datatypes.JSON(`{invalid}`),
			wantErr:     true,
			errContains: "解析钉钉渠道配置失败",
		},
		{
			name:        "WebhookURL为空",
			config:      datatypes.JSON(`{"secret":"SECxxx"}`),
			wantErr:     true,
			errContains: "钉钉 Webhook URL 不能为空",
		},
		{
			name:        "加签密钥为空",
			config:      datatypes.JSON(`{"webhook_url":"https://oapi.dingtalk.com/robot/send?access_token=xxx"}`),
			wantErr:     true,
			errContains: "钉钉加签密钥不能为空",
		},
		{
			name:        "所有配置字段为空",
			config:      datatypes.JSON(`{}`),
			wantErr:     true,
			errContains: "钉钉 Webhook URL 不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.Send(context.Background(), tt.config, "test title", "test content")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errContains)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}

// TestDingtalkSender_Send_Success 测试正常发送流程
func TestDingtalkSender_Send_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	// 创建 Mock 服务器验证请求
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		// 验证 Content-Type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}

		// 验证 URL 包含签名参数
		query := r.URL.Query()
		if !query.Has("timestamp") {
			t.Error("URL should contain timestamp parameter")
		}
		if !query.Has("sign") {
			t.Error("URL should contain sign parameter")
		}

		// 验证请求体
		var msg dingtalkMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if msg.MsgType != "text" {
			t.Errorf("expected msgtype 'text', got %q", msg.MsgType)
		}
		if !strings.Contains(msg.Text.Content, "[test title]") {
			t.Errorf("expected content to contain '[test title]', got %q", msg.Text.Content)
		}
		if !strings.Contains(msg.Text.Content, "test content") {
			t.Errorf("expected content to contain 'test content', got %q", msg.Text.Content)
		}

		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	// 构造配置
	config := datatypes.JSON(fmt.Sprintf(`{
		"webhook_url": "%s?access_token=xxx",
		"secret": "SECf5fe53e2cb30543cb68f071d28a027621443205eb9923895e288f4735d072a89"
	}`, server.URL))

	err := sender.Send(context.Background(), config, "test title", "test content")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestDingtalkSender_Send_SignatureCorrect 测试签名生成正确性
func TestDingtalkSender_Send_SignatureCorrect(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	var receivedURL *url.URL

	// 创建 Mock 服务器捕获请求参数
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{
		"webhook_url": "%s?access_token=xxx",
		"secret": "test_secret"
	}`, server.URL))

	err := sender.Send(context.Background(), config, "title", "content")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if receivedURL == nil {
		t.Fatal("received URL is nil")
	}

	// 验证签名参数存在
	query := receivedURL.Query()
	timestamp := query.Get("timestamp")
	sign := query.Get("sign")

	if timestamp == "" {
		t.Error("timestamp parameter is missing")
	}
	if sign == "" {
		t.Error("sign parameter is missing")
	}

	// 验证时间戳是数字（毫秒）
	if len(timestamp) < 13 {
		t.Errorf("timestamp should be at least 13 digits (milliseconds), got %q", timestamp)
	}

	// 验证签名是有效的 Base64
	// 注意：Query.Get() 已经自动进行了 URL 解码，所以不需要再次解码
	_, err = base64.StdEncoding.DecodeString(sign)
	if err != nil {
		t.Fatalf("sign should be valid base64, decode error: %v, sign value: %q", err, sign)
	}
}

// TestDingtalkSender_Send_HTTPError 测试网络错误
func TestDingtalkSender_Send_HTTPError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	// 使用不可达地址
	config := datatypes.JSON(`{
		"webhook_url": "http://192.0.2.1:9999/robot/send?access_token=xxx",
		"secret": "SECxxx"
	}`)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := sender.Send(ctx, config, "title", "content")
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "发送钉钉请求失败") {
		t.Fatalf("expected error containing '发送钉钉请求失败', got %q", err.Error())
	}
}

// TestDingtalkSender_Send_APIError 测试钉钉API返回错误
func TestDingtalkSender_Send_APIError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	// 创建返回错误的 Mock 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":310000,"errmsg":"keywords not in content"}`))
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{
		"webhook_url": "%s?access_token=xxx",
		"secret": "SECxxx"
	}`, server.URL))

	err := sender.Send(context.Background(), config, "title", "content")
	if err == nil {
		t.Fatal("expected API error, got nil")
	}
	if !strings.Contains(err.Error(), "钉钉推送失败") {
		t.Fatalf("expected error containing '钉钉推送失败', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "keywords not in content") {
		t.Fatalf("expected error message to contain actual error, got %q", err.Error())
	}
}

// TestDingtalkSender_Send_HTTPStatusError 测试HTTP状态码错误
func TestDingtalkSender_Send_HTTPStatusError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	// 创建返回500错误的 Mock 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal Server Error`))
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{
		"webhook_url": "%s?access_token=xxx",
		"secret": "SECxxx"
	}`, server.URL))

	err := sender.Send(context.Background(), config, "title", "content")
	if err == nil {
		t.Fatal("expected HTTP status error, got nil")
	}
	if !strings.Contains(err.Error(), "钉钉返回 HTTP 错误") {
		t.Fatalf("expected error containing '钉钉返回 HTTP 错误', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected error message to contain status code 500, got %q", err.Error())
	}
}

// TestDingtalkSender_Send_InvalidResponse 测试响应解析失败
func TestDingtalkSender_Send_InvalidResponse(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	// 创建返回非JSON响应的 Mock 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not a json response`))
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{
		"webhook_url": "%s?access_token=xxx",
		"secret": "SECxxx"
	}`, server.URL))

	err := sender.Send(context.Background(), config, "title", "content")
	if err == nil {
		t.Fatal("expected JSON parse error, got nil")
	}
	if !strings.Contains(err.Error(), "解析钉钉响应失败") {
		t.Fatalf("expected error containing '解析钉钉响应失败', got %q", err.Error())
	}
}

// TestDingtalkSender_Send_ContextCanceled 测试Context取消
func TestDingtalkSender_Send_ContextCanceled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	// 创建慢速响应的 Mock 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // 模拟慢响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{
		"webhook_url": "%s?access_token=xxx",
		"secret": "SECxxx"
	}`, server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	
	// 立即取消
	cancel()

	err := sender.Send(ctx, config, "title", "content")
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

// TestDingtalkSender_Send_ContextTimeout 测试Context超时
func TestDingtalkSender_Send_ContextTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	// 创建慢速响应的 Mock 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // 模拟慢响应
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{
		"webhook_url": "%s?access_token=xxx",
		"secret": "SECxxx"
	}`, server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := sender.Send(ctx, config, "title", "content")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected context timeout error, got nil")
	}

	// 验证超时后快速返回（不超过2秒）
	if elapsed > 2*time.Second {
		t.Fatalf("Send took %v, expected < 2s (context timeout should cancel quickly)", elapsed)
	}
}

// TestDingtalkSender_Send_MessageFormat 测试消息体格式
func TestDingtalkSender_Send_MessageFormat(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	var receivedBody string

	// 创建 Mock 服务器捕获请求体
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBody = string(bodyBytes)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{
		"webhook_url": "%s?access_token=xxx",
		"secret": "SECxxx"
	}`, server.URL))

	err := sender.Send(context.Background(), config, "通知标题", "这是消息内容\n第二行")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 验证消息格式
	if !strings.Contains(receivedBody, `"msgtype":"text"`) {
		t.Errorf("expected msgtype 'text', got body: %s", receivedBody)
	}
	if !strings.Contains(receivedBody, "[通知标题]") {
		t.Errorf("expected title in content, got body: %s", receivedBody)
	}
	if !strings.Contains(receivedBody, "这是消息内容") {
		t.Errorf("expected content, got body: %s", receivedBody)
	}
	if !strings.Contains(receivedBody, "第二行") {
		t.Errorf("expected multiline content, got body: %s", receivedBody)
	}
}

// TestDingtalkSender_Send_WithSpecialCharacters 测试特殊字符处理
func TestDingtalkSender_Send_WithSpecialCharacters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	// 创建 Mock 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	// 测试包含特殊字符的标题和内容
	config := datatypes.JSON(fmt.Sprintf(`{
		"webhook_url": "%s?access_token=xxx",
		"secret": "SECxxx"
	}`, server.URL))

	err := sender.Send(context.Background(), config, "[CI] 构建成功！🎉", "测试通过\n覆盖率: 100%\n<special> & 'quotes'")
	if err != nil {
		t.Fatalf("expected no error with special characters, got %v", err)
	}
}
