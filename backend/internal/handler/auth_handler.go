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

// ChangePassword godoc
// @Summary      修改密码
// @Description  修改当前用户的登录密码，修改成功后所有 Refresh Token 将被撤销
// @Description  新密码要求：8-128 个字符，必须包含小写字母、大写字母和数字
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      dto.ChangePasswordReq  true  "修改密码请求"
// @Success      200   {object}  response.Response  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/user/password [put]
// @Security     BearerAuth
func (h *AuthHandler) ChangePassword(c *gin.Context) {
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

	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			response.Fail(c, xerr.ErrBadRequest.Code, "请求参数格式错误")
			return
		}
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			response.Fail(c, xerr.ErrBadRequest.Code, "请求参数类型错误")
			return
		}
		response.Fail(c, xerr.ErrBadRequest.Code, "请求参数格式错误")
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil)
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

// Register godoc
// @Summary      用户注册
// @Description  注册新用户，返回 Access Token 和 Refresh Token
// @Description  密码要求：8-128 个字符，必须包含小写字母、大写字母和数字
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterReq  true  "注册请求"
// @Success      200   {object}  response.Response{data=dto.AuthResp}  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/user/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			response.Fail(c, xerr.ErrBadRequest.Code, "请求参数格式错误")
			return
		}
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			response.Fail(c, xerr.ErrBadRequest.Code, "请求参数类型错误")
			return
		}
		response.Fail(c, xerr.ErrBadRequest.Code, "请求参数格式错误")
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, resp)
}
