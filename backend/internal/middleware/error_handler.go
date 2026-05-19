package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"octotify/pkg/ctxutil"
	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

// ErrorHandler 统一错误处理中间件，捕获 c.Error() 挂起的错误并返回统一格式
func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *xerr.AppError
		if errors.As(err, &appErr) {
			if appErr.Internal != "" {
				rid := ctxutil.GetRequestID(c.Request.Context())
				logger.Error("[BizError]",
					zap.String("request_id", rid),
					zap.Int("error_code", appErr.Code),
					zap.String("msg", appErr.Msg),
					zap.String("internal", appErr.Internal),
					zap.String("path", c.Request.URL.Path),
				)
			}
			response.Fail(c, appErr.Code, appErr.Msg)
			return
		}

		rid := ctxutil.GetRequestID(c.Request.Context())
		logger.Error("[UnknownError]",
			zap.String("request_id", rid),
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
		)
		response.Fail(c, xerr.ErrInternalServer.Code, xerr.ErrInternalServer.Msg)
	}
}
