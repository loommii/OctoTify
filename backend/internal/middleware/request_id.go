package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求 ID 中间件，优先从请求头取，没有则生成 UUID v7（无连字符）
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			u, err := uuid.NewV7()
			if err != nil {
				u = uuid.New()
			}
			rid = strings.ReplaceAll(u.String(), "-", "")
		}
		c.Set("request_id", rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
