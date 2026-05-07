package middleware

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// setupTestContext 创建一个基础的测试上下文和响应记录器
func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	return c, w
}

// buildTestRequest 构建一个包含指定方法、路径、请求头和请求体的测试请求
func buildTestRequest(method, path string, headers map[string]string, body *bytes.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c.Request = req
	return c, w
}

// generateTestKeyPair 生成用于 JWT 测试的 RSA 密钥对
func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(nil, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA private key: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}

// logCapture 用于捕获日志条目的结构体，支持并发安全
type logCapture struct {
	mu      sync.Mutex
	entries []map[string]interface{}
}

// logWriter 自定义的 io.Writer，用于将日志内容解析并捕获到 logCapture 中
type logWriter struct {
	capture *logCapture
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	var entry map[string]interface{}
	if err := json.Unmarshal(p, &entry); err != nil {
		return len(p), nil
	}
	w.capture.mu.Lock()
	w.capture.entries = append(w.capture.entries, entry)
	w.capture.mu.Unlock()
	return len(p), nil
}

// GetEntries 获取所有捕获的日志条目（副本）
func (c *logCapture) GetEntries() []map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]map[string]interface{}, len(c.entries))
	copy(result, c.entries)
	return result
}

// GetLastEntry 获取最后一条日志条目
func (c *logCapture) GetLastEntry() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) == 0 {
		return nil
	}
	return c.entries[len(c.entries)-1]
}

// newTestLogger 创建一个 Debug 级别的测试日志器和对应的捕获器
func newTestLogger() (*zap.Logger, *logCapture) {
	capture := &logCapture{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&logWriter{capture: capture}),
		zap.DebugLevel,
	)
	return zap.New(core), capture
}

// newLoggerWithLevel 创建一个指定日志级别的测试日志器和对应的捕获器
func newLoggerWithLevel(level zapcore.LevelEnabler) (*zap.Logger, *logCapture) {
	capture := &logCapture{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&logWriter{capture: capture}),
		level,
	)
	return zap.New(core), capture
}

// errorReader 模拟读取请求体时发生错误的 io.Reader
type errorReader struct{}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, &readError{}
}

// readError 模拟的读取错误
type readError struct{}

func (e *readError) Error() string {
	return "simulated read error"
}
