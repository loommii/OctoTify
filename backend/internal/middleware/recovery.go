package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

// CustomRecovery Panic 恢复中间件，捕获运行时 panic，记录包含请求 ID 的堆栈日志，
// 并返回统一的错误响应格式。
func CustomRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				rid, _ := c.Get("request_id")
				logger.Error("[PANIC]",
					zap.Any("request_id", rid),
					zap.Any("recover", r),
					zap.String("stack", string(debug.Stack())),
				)

				response.Fail(c,
					xerr.ErrInternalServer.Code,
					xerr.ErrInternalServer.Msg,
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}
