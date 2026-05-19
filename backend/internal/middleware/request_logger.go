package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"octotify/pkg/ctxutil"
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

		rid := ctxutil.GetRequestID(c.Request.Context())

		if debugBody && logger.Core().Enabled(zap.DebugLevel) {
			logFullRequestResponse(c, logger, path, start)
			return
		}

		// 记录请求开始日志
		logger.Debug("请求开始",
			zap.String("request_id", rid),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("client_ip", c.ClientIP()),
		)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// 记录请求完成日志
		logger.Info("请求完成",
			zap.String("request_id", rid),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		)
	}
}

// logFullRequestResponse 打印完整的请求和响应信息（包含请求头、请求体、响应头、响应体）
func logFullRequestResponse(c *gin.Context, logger *zap.Logger, path string, start time.Time) {
	rid := ctxutil.GetRequestID(c.Request.Context())

	// 读取请求体
	reqBody, err := c.GetRawData()
	if err != nil {
		reqBody = []byte("读取请求体失败: " + err.Error())
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	// 收集请求头
	reqHeaders := make([]string, 0)
	for k, v := range c.Request.Header {
		reqHeaders = append(reqHeaders, k+": "+strings.Join(v, ", "))
	}

	// 记录请求开始日志（完整模式）
	logger.Debug("请求开始（完整模式）",
		zap.String("request_id", rid),
		zap.String("method", c.Request.Method),
		zap.String("path", path),
		zap.String("client_ip", c.ClientIP()),
		zap.String("request_headers", strings.Join(reqHeaders, "\n")),
		zap.String("request_body", string(reqBody)),
	)

	rw := &responseWriter{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
	}
	c.Writer = rw

	c.Next()

	latency := time.Since(start)
	status := c.Writer.Status()

	// 收集响应头
	respHeaders := make([]string, 0)
	for k, v := range rw.Header() {
		respHeaders = append(respHeaders, k+": "+strings.Join(v, ", "))
	}

	respBody := rw.body.String()

	// 检测二进制内容类型，避免打印不可读的二进制数据
	contentType := rw.Header().Get("Content-Type")
	isBinary := strings.Contains(contentType, "image") ||
		strings.Contains(contentType, "application/octet-stream") ||
		strings.Contains(contentType, "application/pdf")

	if isBinary {
		respBody = "[二进制内容已省略]"
	}

	// 记录请求完成日志（完整模式）
	logger.Debug("请求完成（完整模式）",
		zap.String("request_id", rid),
		zap.String("method", c.Request.Method),
		zap.String("path", path),
		zap.Int("status", status),
		zap.Duration("latency", latency),
		zap.String("response_headers", strings.Join(respHeaders, "\n")),
		zap.String("response_body", respBody),
	)
}
