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

type MessageHandler struct {
	messageService *service.MessageService
}

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

	pageReq := &dto.PageReq{
		Page:     1,
		PageSize: 20,
	}
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			pageReq.Page = page
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			pageReq.PageSize = pageSize
		}
	}

	list, total, err := h.messageService.ListMessages(c.Request.Context(), userID, pageReq)
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

	filter := &dto.MessageFilterReq{
		PageReq: dto.PageReq{
			Page:     1,
			PageSize: 20,
		},
	}
	if err := c.ShouldBindQuery(filter); err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "请求参数格式错误")
		return
	}

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

	messageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "无效的消息 ID")
		return
	}

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

	messageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "无效的消息 ID")
		return
	}

	err = h.messageService.DeleteMessage(c.Request.Context(), userID, messageID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil)
}
