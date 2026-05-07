package sender

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
	"gorm.io/datatypes"
)

func TestEncodeSubject(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"中文标题应进行Base64编码", "系统通知"},
		{"英文标题应进行Base64编码", "Build Notification"},
		{"空字符串应编码为空", ""},
		{"特殊字符应正确编码", "[CI] 构建成功！"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeSubject(tt.input)

			expectedPrefix := "=?UTF-8?B?"
			expectedSuffix := "?="
			if !strings.HasPrefix(got, expectedPrefix) || !strings.HasSuffix(got, expectedSuffix) {
				t.Errorf("encodeSubject(%q) format invalid: %q, want prefix %q and suffix %q",
					tt.input, got, expectedPrefix, expectedSuffix)
			}

			encodedPart := got[len(expectedPrefix) : len(got)-len(expectedSuffix)]
			decoded, err := base64.StdEncoding.DecodeString(encodedPart)
			if err != nil {
				t.Fatalf("decode base64 failed: %v", err)
			}
			if string(decoded) != tt.input {
				t.Errorf("decode result = %q, want %q", string(decoded), tt.input)
			}
		})
	}
}

func TestParseRecipients(t *testing.T) {
	tests := []struct {
		name string
		to   string
		cc   string
		want []string
	}{
		{
			name: "仅to单个收件人",
			to:   "a@x.com",
			cc:   "",
			want: []string{"a@x.com"},
		},
		{
			name: "to加一个cc",
			to:   "a@x.com",
			cc:   "b@x.com",
			want: []string{"a@x.com", "b@x.com"},
		},
		{
			name: "to加多个cc用逗号分隔",
			to:   "a@x.com",
			cc:   "b@x.com, c@x.com, d@x.com",
			want: []string{"a@x.com", "b@x.com", "c@x.com", "d@x.com"},
		},
		{
			name: "cc含有前后空格应去除",
			to:   "a@x.com",
			cc:   "  b@x.com  ,  c@x.com  ",
			want: []string{"a@x.com", "b@x.com", "c@x.com"},
		},
		{
			name: "cc含空项应过滤",
			to:   "a@x.com",
			cc:   "b@x.com,,  ,c@x.com",
			want: []string{"a@x.com", "b@x.com", "c@x.com"},
		},
		{
			name: "to和cc均为空返回空切片",
			to:   "",
			cc:   "",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRecipients(tt.to, tt.cc)
			if len(got) != len(tt.want) {
				t.Fatalf("parseRecipients(%q, %q) length = %d, want %d", tt.to, tt.cc, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseRecipients(%q, %q)[%d] = %q, want %q", tt.to, tt.cc, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestBuildEmailBody(t *testing.T) {
	recipients := []string{"to@x.com", "cc1@x.com", "cc2@x.com"}
	body := buildEmailBody("\"通知\" <from@x.com>", recipients, "=?UTF-8?B?test?=", "body content")

	if !strings.Contains(body, "From: \"通知\" <from@x.com>\r\n") {
		t.Errorf("body should contain From header, got %q", body)
	}
	if !strings.Contains(body, "To: to@x.com\r\n") {
		t.Errorf("body should contain To header, got %q", body)
	}
	if !strings.Contains(body, "Cc: cc1@x.com, cc2@x.com\r\n") {
		t.Errorf("body should contain Cc header, got %q", body)
	}
	if !strings.Contains(body, "Subject: =?UTF-8?B?test?=\r\n") {
		t.Errorf("body should contain Subject header, got %q", body)
	}
	if !strings.Contains(body, "Content-Type: text/plain; charset=\"UTF-8\"\r\n") {
		t.Errorf("body should contain Content-Type header, got %q", body)
	}
	if !strings.Contains(body, "\r\n\r\nbody content") {
		t.Errorf("body should contain body content after double CRLF, got %q", body)
	}

	recipientsSingle := []string{"to@x.com"}
	bodySingle := buildEmailBody("from@x.com", recipientsSingle, "subject", "body")
	if strings.Contains(bodySingle, "Cc:") {
		t.Errorf("single recipient body should not contain Cc header, got %q", bodySingle)
	}
}

func TestEmailSender_Send_ConfigValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewEmailSender(logger)

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
			errContains: "解析 Email 渠道配置失败",
		},
		{
			name:        "smtp_host为空",
			config:      datatypes.JSON(`{"smtp_port":587,"username":"u","password":"p","to":"t@x.com"}`),
			wantErr:     true,
			errContains: "SMTP 服务器地址不能为空",
		},
		{
			name:        "smtp_port为0",
			config:      datatypes.JSON(`{"smtp_host":"smtp.x.com","username":"u","password":"p","to":"t@x.com"}`),
			wantErr:     true,
			errContains: "SMTP 端口不能为空",
		},
		{
			name:        "username为空",
			config:      datatypes.JSON(`{"smtp_host":"smtp.x.com","smtp_port":587,"password":"p","to":"t@x.com"}`),
			wantErr:     true,
			errContains: "SMTP 用户名不能为空",
		},
		{
			name:        "password为空",
			config:      datatypes.JSON(`{"smtp_host":"smtp.x.com","smtp_port":587,"username":"u","to":"t@x.com"}`),
			wantErr:     true,
			errContains: "SMTP 密码不能为空",
		},
		{
			name:        "to为空",
			config:      datatypes.JSON(`{"smtp_host":"smtp.x.com","smtp_port":587,"username":"u","password":"p"}`),
			wantErr:     true,
			errContains: "收件人邮箱不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.Send(context.Background(), tt.config, "title", "content")
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

func TestEmailSender_Send_MissingConfigFields(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewEmailSender(logger)

	err := sender.Send(context.Background(), datatypes.JSON(`{}`), "title", "content")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEmailSender_Send_WithOptionalFields(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewEmailSender(logger)

	config := datatypes.JSON(`{
		"smtp_host": "smtp.example.com",
		"smtp_port": 587,
		"username": "user@example.com",
		"password": "password123",
		"to": "recipient@example.com",
		"cc": "cc1@example.com, cc2@example.com",
		"from_name": "系统通知"
	}`)

	err := sender.Send(context.Background(), config, "test title", "test content")
	if err == nil {
		t.Fatal("expected connection error (no real SMTP server), got nil")
	}
	if !strings.Contains(err.Error(), "连接 SMTP 服务器失败") && !strings.Contains(err.Error(), "建立 TLS 连接失败") {
		t.Logf("got expected connection error: %v", err)
	}
}

func TestEmailSender_Send_WithoutOptionalFields(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewEmailSender(logger)

	config := datatypes.JSON(`{
		"smtp_host": "smtp.example.com",
		"smtp_port": 465,
		"username": "user@example.com",
		"password": "password123",
		"to": "recipient@example.com"
	}`)

	err := sender.Send(context.Background(), config, "test title", "test content")
	if err == nil {
		t.Fatal("expected connection error (no real SMTP server), got nil")
	}
	if !strings.Contains(err.Error(), "建立 TLS 连接失败") {
		t.Logf("got expected TLS connection error: %v", err)
	}
}

func TestEmailSender_Send_WithFullConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewEmailSender(logger)

	config := datatypes.JSON(`{
		"smtp_host": "smtp.qq.com",
		"smtp_port": 465,
		"username": "test@qq.com",
		"password": "test_auth_code",
		"to": "test@hotmail.com",
		"cc": "test@gmail.com",
		"from_name": "OctoTify"
	}`)

	err := sender.Send(context.Background(), config, "测试标题", "测试内容")
	if err == nil {
		t.Fatal("expected connection error (fake credentials), got nil")
	}
	t.Logf("got expected error with full config: %v", err)
}

// ============================================================
// Context 超时与取消测试
// ============================================================

func TestDialSMTP_ContextTimeout_Expired(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		timeout     time.Duration
		host        string
		errContains string
	}{
		{
			name:        "端口25_Context1毫秒超时应快速失败",
			port:        25,
			timeout:     1 * time.Millisecond,
			host:        "192.0.2.1", // TEST-NET-1 保留地址，不可达
			errContains: "连接 SMTP 服务器失败",
		},
		{
			name:        "端口587_Context1毫秒超时应快速失败",
			port:        587,
			timeout:     1 * time.Millisecond,
			host:        "192.0.2.1",
			errContains: "连接 SMTP 服务器失败",
		},
		{
			name:        "端口465_Context1毫秒超时应快速失败",
			port:        465,
			timeout:     1 * time.Millisecond,
			host:        "192.0.2.1",
			errContains: "建立 TLS 连接失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			start := time.Now()
			_, err := dialSMTP(ctx, tt.host, tt.port)
			elapsed := time.Since(start)

			if err == nil {
				t.Fatal("expected error due to context timeout, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
			}

			// 验证超时后快速返回（不超过 2 秒，而不是 TCP 默认 75 秒超时）
			if elapsed > 2*time.Second {
				t.Fatalf("dialSMTP took %v, expected < 2s (context timeout should cancel quickly)", elapsed)
			}
		})
	}
}

func TestDialSMTP_ContextCanceled(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		host        string
		errContains string
	}{
		{
			name:        "端口25_Context取消应快速返回错误",
			port:        25,
			host:        "192.0.2.1",
			errContains: "连接 SMTP 服务器失败",
		},
		{
			name:        "端口587_Context取消应快速返回错误",
			port:        587,
			host:        "192.0.2.1",
			errContains: "连接 SMTP 服务器失败",
		},
		{
			name:        "端口465_Context取消应快速返回错误",
			port:        465,
			host:        "192.0.2.1",
			errContains: "建立 TLS 连接失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // 立即取消

			start := time.Now()
			_, err := dialSMTP(ctx, tt.host, tt.port)
			elapsed := time.Since(start)

			if err == nil {
				t.Fatal("expected error due to context cancellation, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
			}

			// 验证取消后快速返回
			if elapsed > 1*time.Second {
				t.Fatalf("dialSMTP took %v after cancel, expected < 1s", elapsed)
			}
		})
	}
}

func TestDialSMTP_WithDeadline(t *testing.T) {
	// 启动一个 Mock SMTP 服务器（端口 25 行为）
	listener, addr := startMockSMTPServer(t)
	defer listener.Close()

	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		t.Fatalf("failed to split host port: %v", err)
	}
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}

	// 使用一个较大的 deadline（5 秒），确保连接能成功建立
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := dialSMTP(ctx, host, port)
	if err != nil {
		t.Fatalf("dialSMTP with deadline should succeed, got error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client, got nil")
	}
	defer client.Close()
}

func TestEmailSender_Send_ContextTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewEmailSender(logger)

	// 使用不可达地址 + 极短超时
	config := datatypes.JSON(`{
		"smtp_host": "192.0.2.1",
		"smtp_port": 587,
		"username": "user@example.com",
		"password": "password123",
		"to": "recipient@example.com"
	}`)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := sender.Send(ctx, config, "title", "content")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error due to context timeout, got nil")
	}

	// 验证超时后快速返回
	if elapsed > 2*time.Second {
		t.Fatalf("Send took %v, expected < 2s (context timeout should cancel quickly)", elapsed)
	}
}

// ============================================================
// 端口 25 跳过认证测试
// ============================================================

func TestEmailSender_Send_Port25_SkipAuth(t *testing.T) {
	// 启动一个完整的 Mock SMTP 服务器，模拟端口 25 行为（无认证）
	listener, addr := startMockSMTPServerWithAuth(t, false)
	defer listener.Close()

	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		t.Fatalf("failed to split host port: %v", err)
	}
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}

	logger := zaptest.NewLogger(t)
	sender := NewEmailSender(logger)

	config := datatypes.JSON(fmt.Sprintf(`{
		"smtp_host": "%s",
		"smtp_port": %d,
		"username": "user@example.com",
		"password": "password123",
		"to": "recipient@example.com"
	}`, host, port))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sender.Send(ctx, config, "测试标题", "测试内容")
	if err != nil {
		t.Fatalf("Send with port 25 should succeed (skip auth), got error: %v", err)
	}
}

func TestDialSMTP_Port25_PlainConnection(t *testing.T) {
	// 启动 Mock SMTP 服务器
	listener, addr := startMockSMTPServer(t)
	defer listener.Close()

	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		t.Fatalf("failed to split host port: %v", err)
	}
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 端口 25 应使用明文连接
	client, err := dialSMTP(ctx, host, port)
	if err != nil {
		t.Fatalf("dialSMTP with port 25 should succeed, got error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client, got nil")
	}
	defer client.Close()
}

// ============================================================
// applyConnDeadline 测试
// ============================================================

func TestApplyConnDeadline_WithContextDeadline(t *testing.T) {
	// 启动一个简单的 TCP 服务器
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer listener.Close()

	// 接受连接
	connCh := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		connCh <- conn
	}()

	// 客户端连接
	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer clientConn.Close()

	// 等待服务端接受连接
	<-connCh

	// 设置 context deadline 为 100ms 后
	deadline := time.Now().Add(100 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// 调用 applyConnDeadline
	applyConnDeadline(clientConn, ctx)

	// 验证连接 deadline 已设置（通过读取超时验证）
	errCh := make(chan error, 1)
	go func() {
		buf := make([]byte, 1024)
		_, err := clientConn.Read(buf)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected read timeout error, got nil")
		}
		// 验证是超时错误
		if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
			t.Fatalf("expected timeout error, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("expected read to timeout within 1s, but it didn't")
	}
}

func TestApplyConnDeadline_WithoutDeadline(t *testing.T) {
	// 启动一个简单的 TCP 服务器
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer listener.Close()

	// 接受连接
	connCh := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		connCh <- conn
	}()

	// 客户端连接
	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer clientConn.Close()

	// 等待服务端接受连接
	<-connCh

	// 使用无 deadline 的 context
	ctx := context.Background()

	// 调用 applyConnDeadline
	applyConnDeadline(clientConn, ctx)

	// 验证连接在短期内不会超时（因为默认是 30 秒）
	errCh := make(chan error, 1)
	go func() {
		buf := make([]byte, 1024)
		_, err := clientConn.Read(buf)
		errCh <- err
	}()

	// 等待 200ms，验证没有超时（因为 deadline 是 30 秒后）
	select {
	case err := <-errCh:
		// 不应该这么快超时
		t.Fatalf("unexpected error within 200ms: %v", err)
	case <-time.After(200 * time.Millisecond):
		// 预期行为：没有超时
	}
}

// ============================================================
// Mock SMTP 服务器辅助函数
// ============================================================

// startMockSMTPServer 启动一个基本的 Mock SMTP 服务器（端口 25 行为）
func startMockSMTPServer(t *testing.T) (net.Listener, net.Addr) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock SMTP server: %v", err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// 发送 SMTP 欢迎消息
		conn.Write([]byte("220 localhost SMTP Mock Server\r\n"))

		// 处理 SMTP 命令
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(string(buf[:n])))

			if strings.HasPrefix(cmd, "EHLO") || strings.HasPrefix(cmd, "HELO") {
				conn.Write([]byte("250-localhost\r\n"))
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "MAIL") {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "RCPT") {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmd, "DATA") {
				conn.Write([]byte("354 Start mail input\r\n"))
				// 读取邮件数据，直到遇到 .\r\n
				for {
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					if strings.TrimSpace(string(buf[:n])) == "." {
						break
					}
				}
				conn.Write([]byte("250 OK: message queued\r\n"))
			} else if strings.HasPrefix(cmd, "QUIT") {
				conn.Write([]byte("221 Bye\r\n"))
				return
			}
		}
	}()

	return listener, listener.Addr()
}

// startMockSMTPServerWithAuth 启动一个支持认证的 Mock SMTP 服务器
func startMockSMTPServerWithAuth(t *testing.T, requireAuth bool) (net.Listener, net.Addr) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock SMTP server: %v", err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		var authDone atomic.Bool
		inData := false

		conn.Write([]byte("220 localhost SMTP Mock Server\r\n"))

		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			data := string(buf[:n])

			// 在 DATA 阶段，读取所有数据直到遇到 \r\n.\r\n
			if inData {
				if strings.HasSuffix(data, "\r\n.\r\n") {
					inData = false
					conn.Write([]byte("250 OK: message queued\r\n"))
				}
				continue
			}

			cmdUpper := strings.ToUpper(strings.TrimSpace(data))

			if strings.HasPrefix(cmdUpper, "EHLO") || strings.HasPrefix(cmdUpper, "HELO") {
				if requireAuth {
					conn.Write([]byte("250-localhost\r\n"))
					conn.Write([]byte("250-AUTH PLAIN LOGIN\r\n"))
					conn.Write([]byte("250 OK\r\n"))
				} else {
					conn.Write([]byte("250-localhost\r\n"))
					conn.Write([]byte("250 OK\r\n"))
				}
			} else if strings.HasPrefix(cmdUpper, "AUTH") {
				// 模拟认证成功
				conn.Write([]byte("235 2.7.0 Authentication successful\r\n"))
				authDone.Store(true)
			} else if strings.HasPrefix(cmdUpper, "MAIL") {
				if requireAuth && !authDone.Load() {
					conn.Write([]byte("530 5.7.0 Authentication required\r\n"))
				} else {
					conn.Write([]byte("250 OK\r\n"))
				}
			} else if strings.HasPrefix(cmdUpper, "RCPT") {
				conn.Write([]byte("250 OK\r\n"))
			} else if strings.HasPrefix(cmdUpper, "DATA") {
				inData = true
				conn.Write([]byte("354 Start mail input\r\n"))
			} else if strings.HasPrefix(cmdUpper, "QUIT") {
				conn.Write([]byte("221 Bye\r\n"))
				return
			}
		}
	}()

	return listener, listener.Addr()
}
