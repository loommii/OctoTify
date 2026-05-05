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

// GetChannelTypes godoc
// @Summary      获取渠道类型元数据
// @Description  获取系统支持的所有推送渠道类型及其配置字段定义，用于前端动态渲染创建渠道表单。
// @Description  ## 使用场景
// @Description  1. 打开创建渠道页面时获取支持的渠道类型列表
// @Description  2. 根据 config_fields 动态生成配置表单
// @Description  ## 返回数据说明
// @Description  每个渠道类型包含：
// @Description  - type: 渠道类型标识（如 feishu, dingtalk, wechat 等）
// @Description  - name: 渠道类型名称（如 飞书, 钉钉, 企业微信 等）
// @Description  - description: 渠道类型描述
// @Description  - config_fields: 配置字段定义列表，用于前端动态表单渲染
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Success      200   {object}  response.Response{data=[]dto.ChannelTypeMeta}  "获取成功"
// @Router       /channel-types [get]
// @Security     BearerAuth
func (h *ChannelHandler) GetChannelTypes(c *gin.Context) {
	_, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Fail(c, xerr.ErrUnauthorized.Code, "未提供认证令牌")
		return
	}

	metas := h.channelService.GetChannelTypes()
	response.Success(c, metas)
}

// CreateChannel godoc
// @Summary      创建推送渠道
// @Description  创建一个新的推送渠道，配置渠道类型、名称和凭证信息。
// @Description  ## 支持的渠道类型
// @Description  - **wechat**: 企业微信群机器人
// @Description  - **telegram**: Telegram Bot
// @Description  - **dingtalk**: 钉钉群机器人
// @Description  - **email**: 邮件推送
// @Description  - **webhook**: 自定义 Webhook
// @Description  - **feishu**: 飞书自定义机器人
// @Description  ## 使用场景
// @Description  1. 配置企业微信群机器人
// @Description  2. 配置 Telegram Bot
// @Description  3. 配置钉钉群机器人
// @Description  ## 注意事项
// @Description  - 不同渠道类型需要不同的 Config 配置
// @Description  - 创建后默认启用状态
// @Description  - 建议创建后使用测试接口验证配置
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120200: 渠道类型不支持
// @Description  - 120201: 渠道名称已存在
// @Description  - 120202: 创建渠道失败
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateChannelReq  true  "创建推送渠道请求参数"
// @Success      200   {object}  response.Response{data=dto.ChannelDTO}  "创建成功"
// @Router       /channels [post]
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
// @Description  编辑推送渠道的名称和配置信息，不影响已推送的消息。
// @Description  ## 使用场景
// @Description  1. 修改渠道名称使其更易识别
// @Description  2. 更新渠道配置（如更换 Webhook 地址）
// @Description  ## 注意事项
// @Description  - 只能编辑自己创建的渠道
// @Description  - 名称不能与其他渠道重复
// @Description  - 修改配置后建议重新测试
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120200: 渠道类型不支持
// @Description  - 120201: 渠道名称已存在
// @Description  - 120203: 渠道不存在
// @Description  - 120204: 无权限编辑
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64              true  "渠道ID，在路径中传递"  minimum(1)
// @Param        body  body      dto.UpdateChannelReq  true  "编辑推送渠道请求参数"
// @Success      200   {object}  response.Response  "编辑成功"
// @Router       /channels/{id} [put]
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
// @Description  分页查询当前用户的所有推送渠道列表，返回渠道基本信息。
// @Description  ## 使用场景
// @Description  1. 渠道管理页面展示渠道列表
// @Description  2. 下拉选择器获取渠道选项
// @Description  ## 注意事项
// @Description  - 列表按创建时间倒序排列
// @Description  - 已删除的渠道不会出现在列表中
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "页码，从 1 开始，默认 1"  default(1)  minimum(1)
// @Param        page_size  query     int  false  "每页条数，默认 20，最大 100"  default(20)  minimum(1)  maximum(100)
// @Success      200   {object}  response.Response{data=response.PageResult{list=[]dto.ChannelDTO}}  "查询成功"
// @Router       /channels [get]
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
// @Description  查询指定推送渠道的详细信息，包括渠道类型、名称、配置和状态。
// @Description  ## 使用场景
// @Description  1. 渠道详情页展示
// @Description  2. 查看渠道配置信息
// @Description  ## 注意事项
// @Description  - 只能查看自己创建的渠道
// @Description  - Config 字段包含渠道的完整配置信息
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120203: 渠道不存在
// @Description  - 120204: 无权限查看
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID，在路径中传递"  minimum(1)
// @Success      200   {object}  response.Response{data=dto.ChannelDTO}  "查询成功"
// @Router       /channels/{id} [get]
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
// @Description  发送测试消息到指定渠道，验证渠道配置是否正确。
// @Description  ## 使用场景
// @Description  1. 创建渠道后验证配置
// @Description  2. 修改渠道配置后验证连接
// @Description  ## 注意事项
// @Description  - 测试消息会真实推送到目标渠道
// @Description  - 测试失败时返回详细错误信息
// @Description  - 只能测试自己创建的渠道
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120203: 渠道不存在
// @Description  - 120204: 无权限测试
// @Description  - 120205: 渠道已停用
// @Description  - 120206: 测试推送失败
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID，在路径中传递"  minimum(1)
// @Success      200   {object}  response.Response  "测试成功"
// @Router       /channels/{id}/test [post]
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
// @Description  停用指定推送渠道，停用后该渠道不再接收消息推送。
// @Description  ## 使用场景
// @Description  1. 临时停止某个渠道的消息推送
// @Description  2. 渠道维护期间停用
// @Description  ## 注意事项
// @Description  - 停用后推送到该渠道的消息会失败
// @Description  - 已推送的历史消息不受影响
// @Description  - 可以随时重新启用
// @Description  - 只能停用自己创建的渠道
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120203: 渠道不存在
// @Description  - 120204: 无权限操作
// @Description  - 120205: 渠道已停用
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID，在路径中传递"  minimum(1)
// @Success      200   {object}  response.Response  "停用成功"
// @Router       /channels/{id}/disable [put]
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
// @Description  启用指定推送渠道，恢复消息推送功能。
// @Description  ## 使用场景
// @Description  1. 重新启用之前停用的渠道
// @Description  2. 恢复某个渠道的消息推送能力
// @Description  ## 注意事项
// @Description  - 只能启用已停用的渠道
// @Description  - 启用后该渠道立即可以接收消息
// @Description  - 只能启用自己创建的渠道
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120203: 渠道不存在
// @Description  - 120204: 无权限操作
// @Description  - 120207: 渠道已启用
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID，在路径中传递"  minimum(1)
// @Success      200   {object}  response.Response  "启用成功"
// @Router       /channels/{id}/enable [put]
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
// @Description  软删除指定推送渠道，同时解除与所有来源的关联关系。删除后渠道和所有关联数据不可恢复。
// @Description  ## 使用场景
// @Description  1. 永久删除不再使用的渠道
// @Description  2. 清理测试数据
// @Description  ## 注意事项
// @Description  - 删除是软删除，数据库中标记为已删除状态
// @Description  - 删除后该渠道与所有来源的绑定关系解除
// @Description  - 只能删除自己创建的渠道
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120203: 渠道不存在
// @Description  - 120204: 无权限操作
// @Description  - 120208: 删除渠道失败
// @Tags         推送渠道管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "渠道ID，在路径中传递"  minimum(1)
// @Success      200   {object}  response.Response  "删除成功"
// @Router       /channels/{id} [delete]
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
