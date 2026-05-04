package handler

import (
	"github.com/gin-gonic/gin"

	"octotify/internal/handler/dto"
	"octotify/internal/service"
	"octotify/pkg/response"
	"octotify/pkg/xerr"
)

type PushHandler struct {
	messageService *service.MessageService
}

func NewPushHandler(messageService *service.MessageService) *PushHandler {
	return &PushHandler{
		messageService: messageService,
	}
}

// PushMessage godoc
// @Summary      推送消息
// @Description  通过来源 Token 推送消息到所有绑定的渠道
// @Tags         消息推送
// @Accept       json
// @Produce      json
// @Param        token  path      string  true  "来源 Token"
// @Param        body   body      dto.PushMessageReq  true  "推送消息请求"
// @Success      200    {object}  response.Response{data=dto.PushResponse}  "成功"
// @Failure      200    {object}  response.Response  "业务错误（code != 0）"
// @Router       /push/{token} [post]
// @Security     BearerAuth
func (h *PushHandler) PushMessage(c *gin.Context) {
	sourceToken := c.Param("token")
	if sourceToken == "" {
		response.Fail(c, xerr.ErrBadRequest.Code, "来源 Token 不能为空")
		return
	}

	var req dto.PushMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	result, err := h.messageService.PushMessage(c.Request.Context(), sourceToken, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, result)
}
