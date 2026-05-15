package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"octotify/pkg/jwtx"
	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

// ContextKeyUserID Gin Context 中存储用户 ID 的键名
const (
	ContextKeyUserID = "user_id"
)

// JWTAuth JWT 认证中间件
// 功能：从请求头中提取并验证 Bearer Token，验证通过后将用户 ID 存入 Context
func JWTAuth(jwtHelper *jwtx.JWTHelper) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取 Authorization 请求头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, xerr.ErrJWTMissingToken.Code)
			c.Abort()
			return
		}

		// 2. 校验 Bearer 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, xerr.ErrJWTInvalidFormat.Code)
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// 3. 验证 JWT 令牌（签名 + 过期时间）
		_, claims, err := jwtHelper.ValidateToken(tokenStr)
		if err != nil {
			response.Unauthorized(c, xerr.ErrJWTInvalidToken.Code)
			c.Abort()
			return
		}

		// 4. 检查令牌类型是否为 Access Token
		if !claims.IsAccessToken() {
			response.Unauthorized(c, xerr.ErrJWTWrongTokenType.Code)
			c.Abort()
			return
		}

		// 5. 将用户 ID 存入 Context，供后续 Handler 使用
		c.Set(ContextKeyUserID, claims.UID)
		c.Next()
	}
}
