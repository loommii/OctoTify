package middleware

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

// ContextKeyStepUpVerified 二次验证通过标记的键名
const ContextKeyStepUpVerified = "stepup_verified"

// PasswordVerifyFunc 密码验证函数签名
type PasswordVerifyFunc func(ctx context.Context, userID int64, password string) error

// StepUpAuth 密码二次验证中间件（Step-up Authentication）
// 用于敏感操作（查看/重置令牌、启用/停用/删除来源）前的密码验证
// 要求请求体中包含 {"password": "用户密码"}
func StepUpAuth(verifyFn PasswordVerifyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取当前用户 ID（由 JWTAuth 中间件注入）
		userIDStr, exists := c.Get(ContextKeyUserID)
		if !exists {
			response.Unauthorized(c, xerr.ErrUnauthorized.Code, "未提供认证令牌")
			c.Abort()
			return
		}

		userIDStrVal, ok := userIDStr.(string)
		if !ok {
			response.Unauthorized(c, xerr.ErrUnauthorized.Code, "无效的认证信息")
			c.Abort()
			return
		}

		userID, err := strconv.ParseInt(userIDStrVal, 10, 64)
		if err != nil {
			response.Unauthorized(c, xerr.ErrUnauthorized.Code, "无效的认证信息")
			c.Abort()
			return
		}

		// 2. 解析请求体中的密码
		var req struct {
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, xerr.ErrBadRequest.Code, "请求参数格式错误")
			c.Abort()
			return
		}

		// 3. 验证密码
		if err := verifyFn(c.Request.Context(), userID, req.Password); err != nil {
			response.Fail(c, xerr.CodeLoginInvalidCredentials, "密码错误")
			c.Abort()
			return
		}

		// 4. 验证通过，标记注入上下文
		c.Set(ContextKeyStepUpVerified, true)
		c.Next()
	}
}
