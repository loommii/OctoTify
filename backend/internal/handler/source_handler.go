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

type SourceHandler struct {
	sourceService *service.SourceService
}

func NewSourceHandler(sourceService *service.SourceService) *SourceHandler {
	return &SourceHandler{sourceService: sourceService}
}

// CreateSource godoc
// @Summary      创建消息来源
// @Description  创建一个新的消息来源，系统自动生成 Source Token
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateSourceReq  true  "创建消息来源请求"
// @Success      200   {object}  response.Response{data=dto.SourceDTO}  "成功"
// @Failure      200   {object}  response.Response  "业务错误（code != 0）"
// @Router       /api/sources [post]
// @Security     BearerAuth
func (h *SourceHandler) CreateSource(c *gin.Context) {
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

	var req dto.CreateSourceReq
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

	source, err := h.sourceService.CreateSource(c.Request.Context(), userID, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, source)
}
