package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestRequestLogger_BasicMode 测试基础日志模式（不输出请求体和响应体）
func TestRequestLogger_BasicMode(t *testing.T) {
	logger, capture := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, false))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	entry := entries[0]
	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "request completed", entry["msg"])
	assert.Equal(t, "GET", entry["method"])
	assert.Equal(t, "/test", entry["path"])
	assert.Equal(t, float64(200), entry["status"])
	assert.NotNil(t, entry["latency"])
}

// TestRequestLogger_FullDumpMode 测试完整日志模式（输出请求体、响应体、请求头、响应头）
func TestRequestLogger_FullDumpMode(t *testing.T) {
	logger, capture := newLoggerWithLevel(zap.DebugLevel)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, true))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})

	w := httptest.NewRecorder()
	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	entry := entries[0]
	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, "request completed (full dump)", entry["msg"])
	assert.Contains(t, entry["request_body"], "test")
	assert.Contains(t, entry["response_body"], "code")
	assert.NotEmpty(t, entry["request_headers"])
	assert.NotEmpty(t, entry["response_headers"])
}

// TestRequestLogger_DebugBodyButNotDebugLevel 测试日志级别不是 Debug 时不输出详细内容
func TestRequestLogger_DebugBodyButNotDebugLevel(t *testing.T) {
	logger, capture := newLoggerWithLevel(zap.InfoLevel)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, true))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	entry := entries[0]
	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "request completed", entry["msg"])
}

// TestRequestLogger_BasicLogFields 测试基础日志字段完整性
func TestRequestLogger_BasicLogFields(t *testing.T) {
	logger, capture := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, false))
	router.POST("/api/sources", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": 1})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/sources", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	entry := entries[0]
	assert.Equal(t, "POST", entry["method"])
	assert.Equal(t, "/api/sources", entry["path"])
	assert.Equal(t, float64(201), entry["status"])
	assert.NotNil(t, entry["latency"])
}

// TestRequestLogger_RequestIDInLog 测试日志中包含 request_id
func TestRequestLogger_RequestIDInLog(t *testing.T) {
	logger, capture := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, false))
	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-rid")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "test-rid", entries[0]["request_id"])
}

// TestRequestLogger_PathWithQueryString 测试日志中的路径包含查询字符串
func TestRequestLogger_PathWithQueryString(t *testing.T) {
	logger, capture := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, false))
	router.GET("/api/sources", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sources?page=1&size=10", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	path := entries[0]["path"].(string)
	assert.Contains(t, path, "/api/sources")
	assert.Contains(t, path, "page=1")
	assert.Contains(t, path, "size=10")
}

// TestRequestLogger_FullDumpRequestBody 测试完整模式下请求体被正确记录
func TestRequestLogger_FullDumpRequestBody(t *testing.T) {
	logger, capture := newLoggerWithLevel(zap.DebugLevel)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, true))
	router.POST("/test", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, string(body))
	})

	w := httptest.NewRecorder()
	reqBody := `{"name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	assert.Contains(t, entries[0]["request_body"], reqBody)
}

// TestRequestLogger_FullDumpResponseBody 测试完整模式下响应体被正确记录
func TestRequestLogger_FullDumpResponseBody(t *testing.T) {
	logger, capture := newLoggerWithLevel(zap.DebugLevel)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, true))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": "hello"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	respBody := entries[0]["response_body"].(string)
	assert.Contains(t, respBody, "code")
	assert.Contains(t, respBody, "hello")
}

// TestRequestLogger_BinaryResponseImageOmitted 测试二进制响应（image）被省略
func TestRequestLogger_BinaryResponseImageOmitted(t *testing.T) {
	logger, capture := newLoggerWithLevel(zap.DebugLevel)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, true))
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "image/png")
		c.Data(http.StatusOK, "image/png", []byte{0x89, 0x50, 0x4E, 0x47})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "[binary content omitted]", entries[0]["response_body"])
}

// TestRequestLogger_BinaryResponseOctetStreamOmitted 测试二进制响应（octet-stream）被省略
func TestRequestLogger_BinaryResponseOctetStreamOmitted(t *testing.T) {
	logger, capture := newLoggerWithLevel(zap.DebugLevel)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, true))
	router.GET("/test", func(c *gin.Context) {
		c.Header("Content-Type", "application/octet-stream")
		c.Data(http.StatusOK, "application/octet-stream", []byte{0x00, 0x01, 0x02})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "[binary content omitted]", entries[0]["response_body"])
}

// TestRequestLogger_FullDumpHeaders 测试完整模式下请求头和响应头被正确记录
func TestRequestLogger_FullDumpHeaders(t *testing.T) {
	logger, capture := newLoggerWithLevel(zap.DebugLevel)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, true))
	router.POST("/test", func(c *gin.Context) {
		c.Header("X-Custom-Header", "custom-value")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("X-Custom-Request-Header", "request-value")
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	reqHeaders := entries[0]["request_headers"].(string)
	respHeaders := entries[0]["response_headers"].(string)
	assert.Contains(t, reqHeaders, "X-Custom-Request-Header")
	assert.Contains(t, reqHeaders, "request-value")
	assert.Contains(t, respHeaders, "X-Custom-Header")
	assert.Contains(t, respHeaders, "custom-value")
}

// TestRequestLogger_FailedToReadRequestBody 测试读取请求体失败时的错误处理
func TestRequestLogger_FailedToReadRequestBody(t *testing.T) {
	logger, capture := newLoggerWithLevel(zap.DebugLevel)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(RequestLogger(logger, true))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	errReader := &errorReader{}
	req := httptest.NewRequest(http.MethodPost, "/test", errReader)
	req.ContentLength = 100
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	reqBody := entries[0]["request_body"].(string)
	assert.Contains(t, reqBody, "failed to read request body:")
}
