package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	apperrors "octotify/pkg/errors"
	"octotify/pkg/response"
)

// CustomRecovery Panic 恢复中间件，返回统一格式并记录堆栈
func CustomRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("[PANIC]",
					zap.Any("recover", r),
					zap.String("stack", string(debug.Stack())),
				)

				response.Fail(c,
					apperrors.ErrInternalServer.Code,
					apperrors.ErrInternalServer.Msg,
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}
