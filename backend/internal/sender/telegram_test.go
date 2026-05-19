package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"octotify/internal/client/telegram"

	"go.uber.org/zap/zaptest"
	"gorm.io/datatypes"
)

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		maxLen int
		suffix string
		want   string
	}{
		{"短消息不截断", "hello", 10, "...", "hello"},
		{"恰好等于maxLen不截断", "12345", 5, "...", "12345"},
		{"超过maxLen应截断并追加后缀", "1234567890", 8, "...", "12345..."},
		{"后缀长度超过maxLen回退到maxLen加后缀", "1234567890", 2, "...", "12..."},
		{"空文本不截断", "", 10, "...", ""},
		{"后缀为空直接截断到maxLen", "1234567890", 5, "", "12345"},
		{"availableLen恰好为0回退到maxLen加后缀", "1234567890", 3, "abc", "123abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateMessage(tt.text, tt.maxLen, tt.suffix)
			if got != tt.want {
				t.Errorf("truncateMessage(%q, %d, %q) = %q, want %q",
					tt.text, tt.maxLen, tt.suffix, got, tt.want)
			}
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"无特殊字符不转换", "hello world", "hello world"},
		{"小于号转义", "a < b", "a &lt; b"},
		{"大于号转义", "a > b", "a &gt; b"},
		{"and符号转义", "a & b", "a &amp; b"},
		{"所有特殊字符同时转义", "<script>&alert()</script>",
			"&lt;script&gt;&amp;alert()&lt;/script&gt;"},
		{"空字符串", "", ""},
		{"已转义and符号会二次转义", "a &amp; b", "a &amp;amp; b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeHTML(tt.input)
			if got != tt.want {
				t.Errorf("escapeHTML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func newTestSender(t *testing.T, serverURL string) *TelegramSender {
	t.Helper()
	logger := zaptest.NewLogger(t)
	sender := NewTelegramSender(logger)
	sender.baseURL = serverURL + "/bot"
	return sender
}

func TestTelegramSender_Send(t *testing.T) {
	validConfig := datatypes.JSON(`{"bot_token":"testtoken","chat_id":"-1001234567890"}`)

	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
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
			errContains: "解析 Telegram 渠道配置失败",
		},
		{
			name:        "Bot Token为空",
			config:      datatypes.JSON(`{"chat_id":"-1001234567890"}`),
			title:       "test",
			content:     "test",
			handler:     nil,
			wantErr:     true,
			errContains: "Telegram Bot Token 不能为空",
		},
		{
			name:        "Chat ID为空",
			config:      datatypes.JSON(`{"bot_token":"testtoken"}`),
			title:       "test",
			content:     "test",
			handler:     nil,
			wantErr:     true,
			errContains: "Telegram Chat ID 不能为空",
		},
		{
			name:        "代理地址格式错误",
			config:      datatypes.JSON(`{"bot_token":"testtoken","chat_id":"-1001234567890","proxy":"://invalid"}`),
			title:       "test",
			content:     "test",
			handler:     nil,
			wantErr:     true,
			errContains: "发送 Telegram 请求失败",
		},
		{
			name:        "推送成功",
			config:      validConfig,
			title:       "通知标题",
			content:     "通知内容",
			handler:     successHandler,
			wantErr:     false,
			errContains: "",
		},
		{
			name:    "推送失败_业务错误chat_not_found",
			config:  validConfig,
			title:   "test",
			content: "test",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"ok":          false,
					"error_code":  400,
					"description": "Bad Request: chat not found",
				})
			}),
			wantErr:     true,
			errContains: "Telegram 推送失败",
		},
		{
			name:    "推送失败_HTTP500错误",
			config:  validConfig,
			title:   "test",
			content: "test",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "Internal Server Error")
			}),
			wantErr:     true,
			errContains: "Telegram 返回 HTTP 错误",
		},
		{
			name:    "推送失败_响应体非JSON",
			config:  validConfig,
			title:   "test",
			content: "test",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "this is not json")
			}),
			wantErr:     true,
			errContains: "解析 Telegram 响应失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			var sender *TelegramSender

			if tt.handler != nil {
				server := httptest.NewServer(tt.handler)
				sender = newTestSender(t, server.URL)
				cleanup = server.Close
			} else {
				sender = newTestSender(t, "http://127.0.0.1:1")
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

func TestTelegramSender_Send_NetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	serverURL := server.URL
	server.Close()

	sender := newTestSender(t, serverURL)

	config := datatypes.JSON(`{"bot_token":"testtoken","chat_id":"-1001234567890"}`)
	err := sender.Send(context.Background(), config, "test", "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "发送 Telegram 请求失败") {
		t.Fatalf("expected error containing '发送 Telegram 请求失败', got %q", err.Error())
	}
}

func TestTelegramSender_Send_MessageTruncation(t *testing.T) {
	var capturedBody telegram.MessageRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	sender := newTestSender(t, server.URL)

	longContent := strings.Repeat("a", 5000)
	config := datatypes.JSON(`{"bot_token":"testtoken","chat_id":"-1001234567890"}`)
	err := sender.Send(context.Background(), config, "test", longContent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(capturedBody.Text, "[消息已截断]") {
		t.Fatalf("expected truncated message to contain '[消息已截断]', got %q", capturedBody.Text)
	}
	if utf8.RuneCountInString(capturedBody.Text) > telegramMaxMessageLen {
		t.Fatalf("expected message length <= %d (chars), got %d (chars)", telegramMaxMessageLen, utf8.RuneCountInString(capturedBody.Text))
	}
	if capturedBody.ChatID != "-1001234567890" {
		t.Fatalf("expected chat_id '-1001234567890', got %q", capturedBody.ChatID)
	}
	if capturedBody.ParseMode != "HTML" {
		t.Fatalf("expected parse_mode 'HTML', got %q", capturedBody.ParseMode)
	}
}

func TestTelegramSender_Send_HTMLEscaping(t *testing.T) {
	var capturedBody telegram.MessageRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		json.Unmarshal(body, &capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	sender := newTestSender(t, server.URL)

	config := datatypes.JSON(`{"bot_token":"testtoken","chat_id":"-1001234567890"}`)
	err := sender.Send(context.Background(), config, "<script>", "a & b > c")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(capturedBody.Text, "&lt;script&gt;") {
		t.Fatalf("expected title to be HTML-escaped in text, got %q", capturedBody.Text)
	}
	if !strings.Contains(capturedBody.Text, "a &amp; b &gt; c") {
		t.Fatalf("expected content to be HTML-escaped in text, got %q", capturedBody.Text)
	}
	if capturedBody.ParseMode != "HTML" {
		t.Fatalf("expected parse_mode 'HTML', got %q", capturedBody.ParseMode)
	}
}
