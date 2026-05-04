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
// @Summary      推送消息
// @Description  外部系统通过 Source Token 推送消息到平台，消息会被并发推送到所有绑定的有效渠道。
// @Description  ## 认证方式
// @Description  使用 Source Token 进行认证，而非用户 Access Token。
// @Description  Token 格式：src{uuid}，例如：src0196a3b2c4d50000a1b2c3d4e5f67890
// @Description  ## 推送流程
// @Description  1. 系统验证 Source Token 有效性
// @Description  2. 查找该来源绑定的所有启用状态的渠道
// @Description  3. 并发推送到所有渠道
// @Description  4. 返回各渠道的推送结果
// @Description  ## 使用场景
// @Description  1. CI/CD 流水线推送构建结果
// @Description  2. 监控系统推送告警信息
// @Description  3. 业务系统推送通知消息
// @Description  ## 注意事项
// @Description  - Source Token 通过 URL 路径参数传递
// @Description  - 消息会并发推送到所有绑定的有效渠道
// @Description  - 返回结果包含每个渠道的推送状态
// @Description  - 推送失败不影响其他渠道的推送
// @Description  - 来源被停用后无法推送消息
// @Description  ## 错误码说明
// @Description  - 120400: 来源 Token 无效
// @Description  - 120401: 来源已停用
// @Description  - 120402: 来源不存在
// @Description  - 120403: 没有可用的推送渠道
// @Description  - 120404: 推送消息失败
// @Tags         消息推送
// @Accept       json
// @Produce      json
// @Param        token  path      string  true  "来源 Token，格式为 src{uuid}"
// @Param        body   body      dto.PushMessageReq  true  "推送消息请求参数"
// @Success      200    {object}  response.Response{data=dto.PushResponse}  "推送成功"
// @Router       /push/{token} [post]
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
