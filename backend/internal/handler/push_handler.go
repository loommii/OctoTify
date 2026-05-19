package handler

import (
	"github.com/gin-gonic/gin"

	"octotify/internal/handler/dto"
	"octotify/internal/service"
	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

// PushHandler 消息推送 HTTP 处理器
type PushHandler struct {
	messageService *service.MessageService
}

// NewPushHandler 创建消息推送处理器实例
func NewPushHandler(messageService *service.MessageService) *PushHandler {
	return &PushHandler{
		messageService: messageService,
	}
}

// PushMessage godoc
// @Security     SourceTokenAuth
func (h *PushHandler) PushMessage(c *gin.Context) {
	// 从 URL 路径参数获取来源 Token
	sourceToken := c.Param("token")
	if sourceToken == "" {
		response.Fail(c, xerr.ErrBadRequest.Code, "来源 Token 不能为空")
		return
	}

	// 绑定并校验请求体参数
	var req dto.PushMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	// 调用服务层推送消息
	result, err := h.messageService.PushMessage(c.Request.Context(), sourceToken, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, result)
}
