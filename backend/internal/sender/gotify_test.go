package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"go.uber.org/zap/zaptest"
	"gorm.io/datatypes"
)

func TestGotifySender_Send(t *testing.T) {
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gotifyResponse{ID: 42})
	})

	tests := []struct {
		name        string
		config      datatypes.JSON
		title       string
		content     string
		handler     http.HandlerFunc
		wantErr     bool
		errContains string
	}{
		{
			name:        "配置解析失败_无效JSON",
			config:      datatypes.JSON(`{invalid}`),
			title:       "test",
			content:     "test",
			handler:     nil,
			wantErr:     true,
			errContains: "解析 Gotify 渠道配置失败",
		},
		{
			name:        "server_url为空",
			config:      datatypes.JSON(`{"app_token":"testtoken"}`),
			title:       "test",
			content:     "test",
			handler:     nil,
			wantErr:     true,
			errContains: "Gotify 服务器地址不能为空",
		},
		{
			name:        "app_token为空",
			config:      datatypes.JSON(`{"server_url":"https://example.com"}`),
			title:       "test",
			content:     "test",
			handler:     nil,
			wantErr:     true,
			errContains: "Gotify App Token 不能为空",
		},
		{
			name:        "server_url格式错误",
			config:      datatypes.JSON(`{"server_url":"://invalid","app_token":"testtoken"}`),
			title:       "test",
			content:     "test",
			handler:     nil,
			wantErr:     true,
			errContains: "发送 Gotify 请求失败",
		},
		{
			name:        "priority越界_默认值5",
			config:      datatypes.JSON(`{"server_url":"https://example.com","app_token":"testtoken","priority":15}`),
			title:       "test",
			content:     "test",
			handler:     successHandler,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "server_url尾部斜杠_应被去除",
			config:      datatypes.JSON(`{"server_url":"https://example.com/","app_token":"testtoken"}`),
			title:       "test",
			content:     "test",
			handler:     successHandler,
			wantErr:     false,
			errContains: "",
		},
		{
			name:    "推送成功",
			config:  datatypes.JSON(`{"server_url":"https://example.com","app_token":"testtoken"}`),
			title:   "通知标题",
			content: "通知内容",
			handler: successHandler,
			wantErr: false,
		},
		{
			name:    "推送失败_HTTP401未授权",
			config:  datatypes.JSON(`{"server_url":"https://example.com","app_token":"testtoken"}`),
			title:   "test",
			content: "test",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(gotifyErrorResponse{
					Error:            "unauthorized",
					ErrorCode:        401,
					ErrorDescription: "you need a valid access token",
				})
			}),
			wantErr:     true,
			errContains: "Gotify 返回 HTTP 错误",
		},
		{
			name:    "推送失败_业务错误",
			config:  datatypes.JSON(`{"server_url":"https://example.com","app_token":"testtoken"}`),
			title:   "test",
			content: "test",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(gotifyErrorResponse{
					Error:            "invalid token",
					ErrorCode:        400,
					ErrorDescription: "token is not valid",
				})
			}),
			wantErr:     true,
			errContains: "description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			sender := NewGotifySender(logger)

			var cleanup func()
			if tt.handler != nil {
				server := httptest.NewServer(tt.handler)
				config := tt.config
				cfg, _ := config.MarshalJSON()
				var m map[string]interface{}
				json.Unmarshal(cfg, &m)

				testURL := server.URL
				if serverURL, ok := m["server_url"].(string); ok {
					parsed, err := url.Parse(serverURL)
					if err == nil {
						testURL += parsed.Path
					}
				}
				m["server_url"] = testURL
				newCfg, _ := json.Marshal(m)
				tt.config = datatypes.JSON(newCfg)

				cleanup = func() { server.Close() }
			} else {
				cleanup = func() {}
			}
			defer cleanup()

			err := sender.Send(context.Background(), tt.config, tt.title, tt.content)

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

func TestGotifySender_Send_NetworkError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewGotifySender(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{"server_url":"%s","app_token":"testtoken"}`, server.URL))
	err := sender.Send(context.Background(), config, "test", "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "发送 Gotify 请求失败") {
		t.Fatalf("expected error containing '发送 Gotify 请求失败', got %q", err.Error())
	}
}

func TestGotifySender_Send_MessageTruncation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewGotifySender(logger)

	var capturedBody gotifyMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gotifyResponse{ID: 1})
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{"server_url":"%s","app_token":"testtoken"}`, server.URL))
	longContent := strings.Repeat("a", 5000)
	err := sender.Send(context.Background(), config, "test", longContent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(capturedBody.Message, "[消息已截断]") {
		t.Fatalf("expected truncated message to contain '[消息已截断]', got %q", capturedBody.Message)
	}
	if len(capturedBody.Message) > gotifyMaxMessageLen {
		t.Fatalf("expected message length <= %d, got %d", gotifyMaxMessageLen, len(capturedBody.Message))
	}
	if capturedBody.Extras.ClientDisplay.ContentType != "text/markdown" {
		t.Fatalf("expected contentType 'text/markdown', got %q", capturedBody.Extras.ClientDisplay.ContentType)
	}
}

func TestGotifySender_Send_MarkdownEscaping(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewGotifySender(logger)

	var capturedBody gotifyMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gotifyResponse{ID: 1})
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{"server_url":"%s","app_token":"testtoken"}`, server.URL))
	err := sender.Send(context.Background(), config, "*test*", "_foo_ [bar]")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(capturedBody.Message, "\\*test\\*") {
		t.Fatalf("expected escaped asterisks in message, got %q", capturedBody.Message)
	}
	if !strings.Contains(capturedBody.Message, "\\_foo\\_") {
		t.Fatalf("expected escaped underscores in message, got %q", capturedBody.Message)
	}
	if !strings.Contains(capturedBody.Message, "\\[bar\\]") {
		t.Fatalf("expected escaped brackets in message, got %q", capturedBody.Message)
	}
}

func TestGotifySender_Send_RequestPath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewGotifySender(logger)

	var capturedPath string
	var capturedToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedToken = r.URL.Query().Get("token")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gotifyResponse{ID: 1})
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{"server_url":"%s","app_token":"mySecretToken"}`, server.URL))
	err := sender.Send(context.Background(), config, "test", "test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedPath != "/message" {
		t.Fatalf("expected request path '/message', got %q", capturedPath)
	}
	if capturedToken != "mySecretToken" {
		t.Fatalf("expected token 'mySecretToken', got %q", capturedToken)
	}
}

func TestGotifySender_Send_HTTPMethod(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewGotifySender(logger)

	var capturedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gotifyResponse{ID: 1})
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{"server_url":"%s","app_token":"testtoken"}`, server.URL))
	err := sender.Send(context.Background(), config, "test", "test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedMethod != http.MethodPost {
		t.Fatalf("expected HTTP method POST, got %q", capturedMethod)
	}
}

func TestGotifySender_Send_ContentTypeHeader(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewGotifySender(logger)

	var capturedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(gotifyResponse{ID: 1})
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{"server_url":"%s","app_token":"testtoken"}`, server.URL))
	err := sender.Send(context.Background(), config, "test", "test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedContentType != "application/json" {
		t.Fatalf("expected Content-Type 'application/json', got %q", capturedContentType)
	}
}

func TestGotifySender_Send_ResponseNonJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewGotifySender(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "this is not json")
	}))
	defer server.Close()

	config := datatypes.JSON(fmt.Sprintf(`{"server_url":"%s","app_token":"testtoken"}`, server.URL))
	err := sender.Send(context.Background(), config, "test", "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "解析 Gotify 响应失败") {
		t.Fatalf("expected error containing '解析 Gotify 响应失败', got %q", err.Error())
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"无特殊字符不转义", "hello world", "hello world"},
		{"星号转义", "*test*", "\\*test\\*"},
		{"下划线转义", "_foo_", "\\_foo\\_"},
		{"方括号转义", "[bar]", "\\[bar\\]"},
		{"圆括号转义", "(parens)", "\\(parens\\)"},
		{"井号转义", "#heading", "\\#heading"},
		{"加号转义", "+item", "\\+item"},
		{"减号转义", "-item", "\\-item"},
		{"感叹号转义", "!important", "\\!important"},
		{"反引号转义", "`code`", "\\`code\\`"},
		{"管道符转义", "a|b", "a\\|b"},
		{"反斜杠转义", "path\\file", "path\\\\file"},
		{"所有特殊字符同时转义", "*test* _foo_ [bar]", "\\*test\\* \\_foo\\_ \\[bar\\]"},
		{"空字符串", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("escapeMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
