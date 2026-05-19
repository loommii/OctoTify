package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"

	"octotify/internal/handler/dto"
	"octotify/internal/middleware"
	"octotify/internal/service"
	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

// AuthHandler 用户认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login godoc
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			response.Fail(c, xerr.ErrLoginInvalidCredentials.Code, "用户名或密码错误")
			return
		}
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			response.Fail(c, xerr.ErrLoginInvalidCredentials.Code, "用户名或密码错误")
			return
		}
		response.Fail(c, xerr.ErrLoginInvalidCredentials.Code, "用户名或密码错误")
		return
	}

	// 调用认证服务进行登录
	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, resp)
}

// Logout godoc
// @Security     BearerAuth
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从上下文中获取当前用户 ID
	userIDStr, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Fail(c, xerr.ErrUnauthorized.Code, "未提供认证令牌")
		return
	}

	userID, err := strconv.ParseInt(userIDStr.(string), 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrUnauthorized.Code, "无效的认证信息")
		return
	}

	// 调用认证服务撤销 Refresh Token
	if err := h.authService.Logout(c.Request.Context(), userID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil)
}

// RefreshToken godoc
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			response.Fail(c, xerr.ErrRefreshTokenInvalid.Code, "刷新令牌无效")
			return
		}
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			response.Fail(c, xerr.ErrRefreshTokenInvalid.Code, "刷新令牌无效")
			return
		}
		response.Fail(c, xerr.ErrRefreshTokenInvalid.Code, "刷新令牌无效")
		return
	}

	// 调用认证服务刷新令牌
	resp, err := h.authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, resp)
}
