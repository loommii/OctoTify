package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"octotify/pkg/response"
	"octotify/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestErrorHandler_NoError 测试没有错误时中间件不干预正常流程
func TestErrorHandler_NoError(t *testing.T) {
	logger, capture := newTestLogger()
	c, w := buildTestRequest(http.MethodGet, "/test", nil, nil)

	middleware := ErrorHandler(logger)
	middleware(c)

	assert.Equal(t, 0, w.Body.Len())
	assert.Equal(t, 0, len(capture.GetEntries()))
}

// TestErrorHandler_WithAppErrorAndInternal 测试带有内部错误详情的 AppError
func TestErrorHandler_WithAppErrorAndInternal(t *testing.T) {
	logger, capture := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(ErrorHandler(logger))
	router.GET("/test", func(c *gin.Context) {
		c.Error(xerr.ErrSourceNotFound.WithInternal(errors.New("db timeout")))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(xerr.ErrSourceNotFound.Code), resp["code"])
	assert.Equal(t, xerr.TranslateMsg(xerr.ErrSourceNotFound.Code, ""), resp["msg"])

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "db timeout", entries[0]["internal"])
}

// TestErrorHandler_WithAppErrorNoInternal 测试不带内部错误详情的 AppError（不应记录日志）
func TestErrorHandler_WithAppErrorNoInternal(t *testing.T) {
	logger, capture := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(ErrorHandler(logger))
	router.GET("/test", func(c *gin.Context) {
		c.Error(xerr.ErrSourceNotFound)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(xerr.ErrSourceNotFound.Code), resp["code"])
	assert.Equal(t, xerr.TranslateMsg(xerr.CodeSourceNotFound, ""), resp["msg"])

	assert.Equal(t, 0, len(capture.GetEntries()))
}

// TestErrorHandler_WithNonAppError 测试非 AppError 类型的错误（降级为内部服务器错误）
func TestErrorHandler_WithNonAppError(t *testing.T) {
	logger, _ := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(ErrorHandler(logger))
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("unknown error"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(xerr.ErrInternalServer.Code), resp["code"])
	assert.Equal(t, xerr.TranslateMsg(xerr.CodeInternalServer, ""), resp["msg"])
}

// TestErrorHandler_MultipleErrors 测试多个错误时取最后一个错误
func TestErrorHandler_MultipleErrors(t *testing.T) {
	logger, _ := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(ErrorHandler(logger))
	router.GET("/test", func(c *gin.Context) {
		c.Error(xerr.ErrSourceNotFound)
		c.Error(xerr.ErrChannelNotFound)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(xerr.ErrChannelNotFound.Code), resp["code"])
	assert.Equal(t, xerr.TranslateMsg(xerr.CodeChannelNotFound, ""), resp["msg"])
}

// TestErrorHandler_AppErrorResponseFormat 测试自定义 AppError 的响应格式
func TestErrorHandler_AppErrorResponseFormat(t *testing.T) {
	logger, _ := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(ErrorHandler(logger))
	router.POST("/test", func(c *gin.Context) {
		c.Error(&xerr.AppError{Code: 110600, Msg: "渠道名称不能为空"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	router.ServeHTTP(w, req)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 110600, resp.Code)
	assert.Equal(t, "渠道名称不能为空", resp.Msg)
}

// TestErrorHandler_UnknownErrorResponseFormat 测试未知错误的响应格式
func TestErrorHandler_UnknownErrorResponseFormat(t *testing.T) {
	logger, _ := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(ErrorHandler(logger))
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("unexpected"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrInternalServer.Code, resp.Code)
	assert.Equal(t, xerr.TranslateMsg(xerr.CodeInternalServer, ""), resp.Msg)
}

// TestErrorHandler_LogFields 测试错误日志中的字段完整性
func TestErrorHandler_LogFields(t *testing.T) {
	logger, capture := newTestLogger()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(ErrorHandler(logger))
	router.POST("/api/test", func(c *gin.Context) {
		c.Error(xerr.ErrChannelParamNameEmpty.WithInternal(errors.New("internal detail")))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	router.ServeHTTP(w, req)

	entries := capture.GetEntries()
	assert.Equal(t, 1, len(entries))
	entry := entries[0]
	assert.NotNil(t, entry["request_id"])
	assert.Equal(t, float64(xerr.ErrChannelParamNameEmpty.Code), entry["error_code"])
	assert.Equal(t, xerr.ErrChannelParamNameEmpty.Msg, entry["msg"])
	assert.Equal(t, "internal detail", entry["internal"])
	assert.Equal(t, "/api/test", entry["path"])
}
