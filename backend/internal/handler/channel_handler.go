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

type ChannelHandler struct {
	channelService *service.ChannelService
}

func NewChannelHandler(channelService *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{channelService: channelService}
}

// CreateChannel godoc
// @Summary      创建推送渠道
// @Description  创建一个新的推送渠道，配置渠道类型、名称和凭证信息
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateChannelReq  true  "创建推送渠道请求"
// @Success      200   {object}  response.Response{data=dto.ChannelDTO}  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/channels [post]
// @Security     BearerAuth
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
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
// @Summary      编辑推送渠道
// @Description  编辑推送渠道的名称和配置信息
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64              true  "渠道ID"
// @Param        body  body      dto.UpdateChannelReq  true  "编辑推送渠道请求"
// @Success      200   {object}  response.Response  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/channels/{id} [put]
// @Security     BearerAuth
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
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
// @Summary      查看推送渠道列表
// @Description  分页查询当前用户的所有推送渠道列表
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "页码，从 1 开始"  default(1)  minimum(1)
// @Param        page_size  query     int  false  "每页条数，最大 100"  default(20)  minimum(1)  maximum(100)
// @Success      200   {object}  response.Response{data=response.PageResult{list=[]dto.ChannelDTO}}  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/channels [get]
// @Security     BearerAuth
func (h *ChannelHandler) ListChannels(c *gin.Context) {
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
// @Summary      查看渠道详情
// @Description  查询指定推送渠道的详细信息
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID"
// @Success      200   {object}  response.Response{data=dto.ChannelDTO}  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/channels/{id} [get]
// @Security     BearerAuth
func (h *ChannelHandler) GetChannelDetail(c *gin.Context) {
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
// @Summary      测试渠道连接
// @Description  发送测试消息到指定渠道，验证渠道配置是否正确
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID"
// @Success      200   {object}  response.Response  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/channels/{id}/test [post]
// @Security     BearerAuth
func (h *ChannelHandler) TestChannel(c *gin.Context) {
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
// @Summary      停用推送渠道
// @Description  停用指定推送渠道，停用后该渠道不再接收消息
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID"
// @Success      200   {object}  response.Response  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/channels/{id}/disable [put]
// @Security     BearerAuth
func (h *ChannelHandler) DisableChannel(c *gin.Context) {
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
// @Summary      启用推送渠道
// @Description  启用指定推送渠道，恢复消息推送功能
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID"
// @Success      200   {object}  response.Response  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/channels/{id}/enable [put]
// @Security     BearerAuth
func (h *ChannelHandler) EnableChannel(c *gin.Context) {
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
// @Summary      删除推送渠道
// @Description  软删除指定推送渠道，同时解除与所有来源的关联关系
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID"
// @Success      200   {object}  response.Response  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/channels/{id} [delete]
// @Security     BearerAuth
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
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
