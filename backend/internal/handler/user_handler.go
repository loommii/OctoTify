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

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register godoc
// @Summary      用户注册
// @Description  注册新用户，返回 Access Token 和 Refresh Token
// @Description  密码要求：8-128 个字符，必须包含小写字母、大写字母和数字
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterReq  true  "注册请求"
// @Success      200   {object}  response.Response{data=dto.AuthResp}  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/user/register [post]
func (h *UserHandler) Register(c *gin.Context) {
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

	resp, err := h.userService.Register(c.Request.Context(), &req)
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
func (h *UserHandler) ChangePassword(c *gin.Context) {
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

	if err := h.userService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil)
}

// GetUserProfile godoc
// @Summary      查询用户信息
// @Description  获取当前登录用户的个人信息，需要 JWT 认证
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Response{data=dto.UserDTO}  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/user/profile [get]
// @Security     BearerAuth
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userIDStr, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Fail(c, xerr.ErrUnauthorized.Code, "未提供认证令牌")
		return
	}

	userProfile, err := h.userService.GetUserProfileByID(c.Request.Context(), userIDStr.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, userProfile)
}
