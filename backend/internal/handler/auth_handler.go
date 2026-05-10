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
// @Summary      用户登录
// @Description  使用用户名和密码登录系统，登录成功后返回 Access Token 和 Refresh Token。
// @Description  Access Token 有效期 1 小时，用于调用需要认证的接口。
// @Description  Refresh Token 有效期 7 天，用于刷新 Access Token。
// @Description  ## 密码要求
// @Description  - 8-128 个字符
// @Description  - 必须包含小写字母
// @Description  - 必须包含大写字母
// @Description  - 必须包含数字
// @Description  ## 使用场景
// @Description  1. 用户首次访问系统时进行登录
// @Description  2. Refresh Token 过期后重新登录
// @Description  ## 注意事项
// @Description  - 登录成功后请妥善保存 Access Token 和 Refresh Token
// @Description  - Access Token 过期后请使用 Refresh Token 调用刷新接口
// @Description  - 连续多次登录失败不会锁定账户
// @Description  ## 错误码说明
// @Description  - 110200: 用户名或密码错误
// @Description  - 110201: 登录失败
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginReq  true  "登录请求参数"
// @Success      200   {object}  response.Response{data=dto.AuthResp}  "登录成功，返回认证信息"
// @Router       /auth/login [post]
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
// @Summary      退出登录
// @Description  撤销当前用户的所有 Refresh Token，退出登录后需要重新登录才能访问系统。
// @Description  ## 注意事项
// @Description  - Access Token 在过期前仍有效（JWT 无状态特性，最长 1 小时后自动失效）
// @Description  - 前端需主动清除本地存储的 Access Token 和 Refresh Token
// @Description  - 退出登录后，所有使用旧 Refresh Token 的刷新请求都会失败
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌或认证令牌无效或已过期
// @Description  - 110304: 退出登录失败
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Response  "退出成功"
// @Router       /auth/logout [post]
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
// @Summary      刷新令牌
// @Description  使用 Refresh Token 获取新的 Access Token 和 Refresh Token。
// @Description  每次刷新都会返回新的 Access Token 和 Refresh Token，旧 Refresh Token 继续有效。
// @Description  ## 使用场景
// @Description  1. Access Token 过期（1 小时后）时调用此接口
// @Description  2. 前端应在请求收到 401 响应时自动调用此接口
// @Description  ## 注意事项
// @Description  - Refresh Token 有效期 7 天，过期后需重新登录
// @Description  - 刷新成功后请更新本地存储的 Access Token 和 Refresh Token
// @Description  - 如果 Refresh Token 已被撤销（如退出登录、修改密码），则无法刷新
// @Description  ## 错误码说明
// @Description  - 110300: 刷新令牌无效
// @Description  - 110301: 刷新令牌已撤销
// @Description  - 110302: 刷新令牌已过期
// @Description  - 110303: 刷新令牌失败
// @Tags         用户认证
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshReq  true  "刷新令牌请求参数"
// @Success      200   {object}  response.Response{data=dto.AuthResp}  "刷新成功"
// @Router       /auth/refresh [post]
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
