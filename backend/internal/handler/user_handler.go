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
// @Description  注册新用户账号，注册成功后自动登录并返回 Access Token 和 Refresh Token。
// @Description  ## 密码要求
// @Description  - 8-128 个字符
// @Description  - 必须包含小写字母
// @Description  - 必须包含大写字母
// @Description  - 必须包含数字
// @Description  ## 使用场景
// @Description  1. 新用户首次创建账号
// @Description  2. 管理员为用户创建账号
// @Description  ## 注意事项
// @Description  - 用户名唯一，重复注册会返回错误
// @Description  - 注册成功后请妥善保存返回的 Token
// @Description  - 密码加密存储，系统无法查看原始密码
// @Description  ## 错误码说明
// @Description  - 110100: 用户名已存在
// @Description  - 110101: 注册失败
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterReq  true  "注册请求参数"
// @Success      200   {object}  response.Response{data=dto.AuthResp}  "注册成功"
// @Router       /user/register [post]
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
// @Description  修改当前用户的登录密码，修改成功后所有 Refresh Token 将被撤销，需要重新登录。
// @Description  ## 密码要求
// @Description  - 8-128 个字符
// @Description  - 必须包含小写字母
// @Description  - 必须包含大写字母
// @Description  - 必须包含数字
// @Description  ## 使用场景
// @Description  1. 用户定期修改密码
// @Description  2. 用户怀疑密码泄露时修改密码
// @Description  ## 注意事项
// @Description  - 修改密码后所有已登录的设备都需要重新登录
// @Description  - 旧密码必须正确才能修改
// @Description  - 新密码不能与旧密码相同
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 110200: 旧密码错误
// @Description  - 110102: 修改密码失败
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      dto.ChangePasswordReq  true  "修改密码请求参数"
// @Success      200   {object}  response.Response  "修改成功"
// @Router       /user/password [put]
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
// @Description  获取当前登录用户的个人信息，包括用户 ID、用户名和创建时间。
// @Description  ## 使用场景
// @Description  1. 页面加载时获取用户信息展示在导航栏
// @Description  2. 用户中心页面展示个人信息
// @Description  ## 注意事项
// @Description  - 需要提供有效的 Access Token
// @Description  - 只能查询自己的信息，不能查询其他用户
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌或认证令牌无效或已过期
// @Description  - 110103: 用户不存在
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Response{data=dto.UserDTO}  "查询成功"
// @Router       /user/profile [get]
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
