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
