package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"octotify/internal/handler/dto"
	"octotify/internal/middleware"
	"octotify/internal/service"
	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

type ChannelHandler struct {
	channelService *service.ChannelService
	logger         *zap.Logger
}

func NewChannelHandler(channelService *service.ChannelService, logger *zap.Logger) *ChannelHandler {
	return &ChannelHandler{channelService: channelService, logger: logger}
}

// getUserID 从 Gin 上下文中提取当前用户 ID
// 提取失败时返回错误响应并返回 false
func (h *ChannelHandler) getUserID(c *gin.Context) (int64, bool) {
	userIDStr, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Fail(c, xerr.ErrUnauthorized.Code, "未提供认证令牌")
		return 0, false
	}

	userIDStrTyped, ok := userIDStr.(string)
	if !ok {
		response.Fail(c, xerr.ErrUnauthorized.Code, "无效的认证信息")
		return 0, false
	}

	userID, err := strconv.ParseInt(userIDStrTyped, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrUnauthorized.Code, "无效的认证信息")
		return 0, false
	}

	return userID, true
}

// GetChannelTypes godoc
// @Security     BearerAuth
func (h *ChannelHandler) GetChannelTypes(c *gin.Context) {
	if _, exists := h.getUserID(c); !exists {
		return
	}

	metas := h.channelService.GetChannelTypes()
	response.Success(c, metas)
}

// CreateChannel godoc
// @Security     BearerAuth
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req dto.CreateChannelReq
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

	channel, err := h.channelService.CreateChannel(c.Request.Context(), userID, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, channel)
}

// UpdateChannel godoc
// @Security     BearerAuth
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	channelIDStr := c.Param("id")
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "渠道ID格式错误")
		return
	}

	var req dto.UpdateChannelReq
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

	err = h.channelService.UpdateChannel(c.Request.Context(), userID, channelID, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}

// ListChannels godoc
// @Security     BearerAuth
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var pageReq dto.PageReq
	if err := c.ShouldBindQuery(&pageReq); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	list, total, err := h.channelService.ListChannels(c.Request.Context(), userID, &pageReq)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithPage(c, list, total, pageReq.Page, pageReq.PageSize)
}

// GetChannelDetail godoc
// @Security     BearerAuth
func (h *ChannelHandler) GetChannelDetail(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	channelIDStr := c.Param("id")
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "渠道ID格式错误")
		return
	}

	channel, err := h.channelService.GetChannelByID(c.Request.Context(), userID, channelID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, channel)
}

// TestChannel godoc
// @Security     BearerAuth
func (h *ChannelHandler) TestChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	channelIDStr := c.Param("id")
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "渠道ID格式错误")
		return
	}

	err = h.channelService.TestChannel(c.Request.Context(), userID, channelID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "测试成功", nil)
}

// DisableChannel godoc
// @Security     BearerAuth
func (h *ChannelHandler) DisableChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	channelIDStr := c.Param("id")
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "渠道ID格式错误")
		return
	}

	err = h.channelService.DisableChannel(c.Request.Context(), userID, channelID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "已停用", nil)
}

// EnableChannel godoc
// @Security     BearerAuth
func (h *ChannelHandler) EnableChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	channelIDStr := c.Param("id")
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "渠道ID格式错误")
		return
	}

	err = h.channelService.EnableChannel(c.Request.Context(), userID, channelID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "已启用", nil)
}

// DeleteChannel godoc
// @Security     BearerAuth
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	channelIDStr := c.Param("id")
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "渠道ID格式错误")
		return
	}

	err = h.channelService.DeleteChannel(c.Request.Context(), userID, channelID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "已删除", nil)
}

// StartBind godoc
