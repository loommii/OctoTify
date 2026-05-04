package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"octotify/internal/handler/dto"
	"octotify/internal/middleware"
	"octotify/internal/service"
	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

// MessageHandler 消息管理 HTTP 处理器
type MessageHandler struct {
	messageService *service.MessageService
}

// NewMessageHandler 创建消息管理处理器实例
func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// ListMessages godoc
// @Summary      查看消息列表
// @Description  查看当前用户的消息记录列表，按创建时间倒序排列
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        page      query     int  false  "页码" minimum(1) default(1)
// @Param        page_size query     int  false  "每页条数" minimum(1) maximum(100) default(20)
// @Success      200       {object}  response.Response{data=[]dto.MessageDTO}  "成功"
// @Failure      200       {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/messages [get]
// @Security     BearerAuth
func (h *MessageHandler) ListMessages(c *gin.Context) {
	// 从上下文获取用户 ID
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

	// 绑定并校验分页参数
	var pageReq dto.PageReq
	if err := c.ShouldBindQuery(&pageReq); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	// 调用服务层查询消息列表
	list, total, err := h.messageService.ListMessages(c.Request.Context(), userID, &pageReq)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithPage(c, list, total, pageReq.Page, pageReq.PageSize)
}

// FilterMessages godoc
// @Summary      筛选消息
// @Description  根据来源、渠道、状态、时间范围等条件筛选消息
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        source_id   query     int64  false  "来源 ID"
// @Param        channel_id  query     int64  false  "渠道 ID"
// @Param        status      query     int    false  "推送状态"
// @Param        start_date  query     int64  false  "开始时间（Unix 毫秒）"
// @Param        end_date    query     int64  false  "结束时间（Unix 毫秒）"
// @Param        keyword     query     string false  "关键词（搜索标题和内容）"
// @Param        page        query     int    false  "页码" minimum(1) default(1)
// @Param        page_size   query     int    false  "每页条数" minimum(1) maximum(100) default(20)
// @Success      200         {object}  response.Response{data=[]dto.MessageDTO}  "成功"
// @Failure      200         {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/messages/filter [get]
// @Security     BearerAuth
func (h *MessageHandler) FilterMessages(c *gin.Context) {
	// 从上下文获取用户 ID
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

	// 初始化筛选条件默认值
	filter := &dto.MessageFilterReq{
		PageReq: dto.PageReq{
			Page:     1,
			PageSize: 20,
		},
	}
	// 绑定并校验筛选参数
	if err := c.ShouldBindQuery(filter); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	// 调用服务层筛选消息
	list, total, err := h.messageService.FilterMessages(c.Request.Context(), userID, filter)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithPage(c, list, total, filter.Page, filter.PageSize)
}

// GetMessageDetail godoc
// @Summary      查看消息详情
// @Description  查看单条消息的详细信息，包含来源和渠道信息
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        id  path      int64  true  "消息 ID"
// @Success      200 {object}  response.Response{data=dto.MessageDetailDTO}  "成功"
// @Failure      200 {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/messages/{id} [get]
// @Security     BearerAuth
func (h *MessageHandler) GetMessageDetail(c *gin.Context) {
	// 从上下文获取用户 ID
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

	// 解析消息 ID
	messageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "无效的消息 ID")
		return
	}

	// 调用服务层查询消息详情
	message, err := h.messageService.GetMessageByID(c.Request.Context(), userID, messageID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, message)
}

// DeleteMessage godoc
// @Summary      删除消息
// @Description  删除单条消息记录（软删除）
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        id  path      int64  true  "消息 ID"
// @Success      200 {object}  response.Response  "成功"
// @Failure      200 {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/messages/{id} [delete]
// @Security     BearerAuth
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	// 从上下文获取用户 ID
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

	// 解析消息 ID
	messageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "无效的消息 ID")
		return
	}

	// 调用服务层删除消息
	err = h.messageService.DeleteMessage(c.Request.Context(), userID, messageID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "已删除", nil)
}
