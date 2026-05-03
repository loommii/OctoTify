package handler

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	"octotify/internal/handler/dto"
	"octotify/internal/service"
	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login godoc
// @Summary      用户登录
// @Description  使用用户名和密码登录，返回 Access Token 和 Refresh Token
// @Description  密码要求：8-128 个字符，必须包含小写字母、大写字母和数字
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginReq  true  "登录请求"
// @Success      200   {object}  response.Response{data=dto.AuthResp}  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/auth/login [post]
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

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, resp)
}

// RefreshToken godoc
// @Summary      刷新令牌
// @Description  使用 Refresh Token 获取新的 Access Token 和 Refresh Token
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshReq  true  "刷新令牌请求"
// @Success      200   {object}  response.Response{data=dto.AuthResp}  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/auth/refresh [post]
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

	resp, err := h.authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, resp)
}
