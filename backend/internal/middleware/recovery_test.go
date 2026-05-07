package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCustomRecovery_NoPanic 测试没有 panic 时中间件不干预正常流程
func TestCustomRecovery_NoPanic(t *testing.T) {
	logger, capture := newTestLogger()
	c, w := buildTestRequest(http.MethodGet, "/test", nil, nil)

	middleware := CustomRecovery(logger)
	middleware(c)

	assert.True(t, w.Body.Len() == 0)
	assert.Equal(t, 0, len(capture.GetEntries()))
}

// TestCustomRecovery_CatchPanic 测试捕获 panic 并返回统一错误响应
func TestCustomRecovery_CatchPanic(t *testing.T) {
	logger, _ := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CustomRecovery(logger))
	router.GET("/test", func(c *gin.Context) {
		panic("something went wrong")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(100004), resp["code"])
	assert.Equal(t, "服务器内部错误", resp["msg"])
}

// TestCustomRecovery_CatchStringPanic 测试捕获字符串类型的 panic
func TestCustomRecovery_CatchStringPanic(t *testing.T) {
	logger, _ := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CustomRecovery(logger))
	router.GET("/test", func(c *gin.Context) {
		panic("something went wrong")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(100004), resp["code"])
	assert.Equal(t, "服务器内部错误", resp["msg"])
}

// TestCustomRecovery_CatchErrorPanic 测试捕获 error 类型的 panic
func TestCustomRecovery_CatchErrorPanic(t *testing.T) {
	logger, _ := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CustomRecovery(logger))
	router.GET("/test", func(c *gin.Context) {
		panic(errors.New("db error"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(100004), resp["code"])
	assert.Equal(t, "服务器内部错误", resp["msg"])
}

// TestCustomRecovery_CatchNilPanic 测试捕获 nil panic（Go 1.21+ 会转换为 PanicNilError）
func TestCustomRecovery_CatchNilPanic(t *testing.T) {
	logger, capture := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CustomRecovery(logger))
	router.GET("/test", func(c *gin.Context) {
		var x interface{}
		panic(x)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(100004), resp["code"])
	assert.Equal(t, "服务器内部错误", resp["msg"])

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
}

// TestCustomRecovery_ResponseBodyOnPanic 测试 panic 后响应体格式正确
func TestCustomRecovery_ResponseBodyOnPanic(t *testing.T) {
	logger, _ := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CustomRecovery(logger))
	router.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(100004), resp["code"])
	assert.Equal(t, "服务器内部错误", resp["msg"])
}

// TestCustomRecovery_AbortsOnPanic 测试 panic 后调用 c.Abort() 阻止后续中间件执行
func TestCustomRecovery_AbortsOnPanic(t *testing.T) {
	logger, _ := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CustomRecovery(logger))
	executed := false
	router.GET("/test", func(c *gin.Context) {
		panic("test")
	}, func(c *gin.Context) {
		executed = true
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.False(t, executed)
}

// TestCustomRecovery_LogContainsRequestID 测试 panic 日志中包含 request_id
func TestCustomRecovery_LogContainsRequestID(t *testing.T) {
	logger, capture := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-rid-123")
		c.Next()
	})
	router.Use(CustomRecovery(logger))
	router.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	assert.Contains(t, entries[0], "request_id")
	assert.Equal(t, "test-rid-123", entries[0]["request_id"])
}
