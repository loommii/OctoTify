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
// @Description  查看当前用户的消息记录列表，按创建时间倒序排列。
// @Description  ## 使用场景
// @Description  1. 消息管理页面展示消息列表
// @Description  2. 查看最近推送的消息
// @Description  ## 注意事项
// @Description  - 列表按创建时间倒序排列
// @Description  - 已删除的消息不会出现在列表中
// @Description  - 返回的消息不包含完整的内容详情
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        page      query     int  false  "页码，从 1 开始，默认 1" minimum(1) default(1)
// @Param        page_size query     int  false  "每页条数，默认 20，最大 100" minimum(1) maximum(100) default(20)
// @Success      200       {object}  response.Response{data=response.PageResult{list=[]dto.MessageDTO}}  "查询成功"
// @Router       /messages [get]
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
// @Description  根据来源、渠道、状态、时间范围等条件筛选消息，支持多条件组合查询。
// @Description  ## 筛选条件说明
// @Description  - **source_id**: 按消息来源筛选
// @Description  - **channel_id**: 按推送渠道筛选
// @Description  - **status**: 按推送状态筛选（100-待推送 200-成功 300-失败）
// @Description  - **start_date/end_date**: 按时间范围筛选（Unix 毫秒时间戳）
// @Description  - **keyword**: 按关键词搜索消息标题和内容
// @Description  ## 使用场景
// @Description  1. 查看某个来源的所有消息
// @Description  2. 查看某个渠道的所有消息
// @Description  3. 查看推送失败的消息
// @Description  4. 查看某个时间段内的消息
// @Description  ## 注意事项
// @Description  - 所有筛选条件都是可选的，可以组合使用
// @Description  - 多个条件之间是 AND 关系
// @Description  - 结果按创建时间倒序排列
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        source_id   query     int64  false  "来源 ID（可选）"
// @Param        channel_id  query     int64  false  "渠道 ID（可选）"
// @Param        status      query     int    false  "推送状态（可选）：100-待推送 200-成功 300-失败"
// @Param        start_date  query     int64  false  "开始时间（可选，Unix 毫秒时间戳）"
// @Param        end_date    query     int64  false  "结束时间（可选，Unix 毫秒时间戳）"
// @Param        keyword     query     string false  "关键词（可选，搜索标题和内容）"
// @Param        page        query     int    false  "页码，从 1 开始，默认 1" minimum(1) default(1)
// @Param        page_size   query     int    false  "每页条数，默认 20，最大 100" minimum(1) maximum(100) default(20)
// @Success      200         {object}  response.Response{data=response.PageResult{list=[]dto.MessageDTO}}  "查询成功"
// @Router       /messages/filter [get]
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
// @Description  查看单条消息的详细信息，包含消息内容、来源名称、渠道名称和推送状态。
// @Description  ## 使用场景
// @Description  1. 消息详情页展示
// @Description  2. 查看消息的完整内容
// @Description  3. 查看消息的推送状态
// @Description  ## 注意事项
// @Description  - 只能查看自己用户下的消息
// @Description  - 包含完整的消息内容
// @Description  - 包含来源和渠道的详细信息
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120300: 消息不存在
// @Description  - 120301: 无权限查看
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        id  path      int64  true  "消息ID，在路径中传递" minimum(1)
// @Success      200 {object}  response.Response{data=dto.MessageDetailDTO}  "查询成功"
// @Router       /messages/{id} [get]
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
// @Description  删除单条消息记录（软删除），删除后消息不再显示在列表中。
// @Description  ## 使用场景
// @Description  1. 清理不需要的消息记录
// @Description  2. 删除敏感消息
// @Description  ## 注意事项
// @Description  - 删除是软删除，数据库中标记为已删除状态
// @Description  - 只能删除自己用户下的消息
// @Description  - 删除后无法恢复
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120300: 消息不存在
// @Description  - 120302: 无权限删除
// @Description  - 120303: 删除消息失败
// @Tags         消息管理
// @Accept       json
// @Produce      json
// @Param        id  path      int64  true  "消息ID，在路径中传递" minimum(1)
// @Success      200 {object}  response.Response  "删除成功"
// @Router       /messages/{id} [delete]
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
