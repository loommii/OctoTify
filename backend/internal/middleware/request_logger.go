package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// responseWriter 自定义响应写入器，用于捕获响应体内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 写入响应数据，同时保存到 body 缓冲区
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 写入字符串响应，同时保存到 body 缓冲区
func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// RequestLogger 请求日志中间件
// 当 debugBody 为 true 且日志级别为 DEBUG 时，打印完整的请求和响应信息（包括请求头、请求体、响应头、响应体）
// 否则仅打印基本请求信息（方法、路径、状态码、耗时）
func RequestLogger(logger *zap.Logger, debugBody bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		if debugBody && logger.Core().Enabled(zap.DebugLevel) {
			logFullRequestResponse(c, logger, path, start)
			return
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		rid, _ := c.Get("request_id")

		logger.Info("request completed",
			zap.Any("request_id", rid),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		)
	}
}

// logFullRequestResponse 打印完整的请求和响应信息
func logFullRequestResponse(c *gin.Context, logger *zap.Logger, path string, start time.Time) {
	rid, _ := c.Get("request_id")

	reqBody, err := c.GetRawData()
	if err != nil {
		reqBody = []byte("failed to read request body: " + err.Error())
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	reqHeaders := make([]string, 0)
	for k, v := range c.Request.Header {
		reqHeaders = append(reqHeaders, k+": "+strings.Join(v, ", "))
	}

	rw := &responseWriter{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
	}
	c.Writer = rw

	c.Next()

	latency := time.Since(start)
	status := c.Writer.Status()

	respHeaders := make([]string, 0)
	for k, v := range rw.Header() {
		respHeaders = append(respHeaders, k+": "+strings.Join(v, ", "))
	}

	respBody := rw.body.String()

	contentType := rw.Header().Get("Content-Type")
	isBinary := strings.Contains(contentType, "image") ||
		strings.Contains(contentType, "application/octet-stream") ||
		strings.Contains(contentType, "application/pdf")

	if isBinary {
		respBody = "[binary content omitted]"
	}

	logger.Debug("request completed (full dump)",
		zap.Any("request_id", rid),
		zap.String("method", c.Request.Method),
		zap.String("path", path),
		zap.Int("status", status),
		zap.Duration("latency", latency),
		zap.String("request_headers", strings.Join(reqHeaders, "\n")),
		zap.String("request_body", string(reqBody)),
		zap.String("response_headers", strings.Join(respHeaders, "\n")),
		zap.String("response_body", respBody),
	)
}
