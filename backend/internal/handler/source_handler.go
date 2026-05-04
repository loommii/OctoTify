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
// @Description  创建一个新的消息来源，系统自动生成 Source Token，用于外部系统推送消息。
// @Description  ## 使用场景
// @Description  1. CI/CD 流水线创建为消息来源
// @Description  2. 监控系统创建为消息来源
// @Description  3. 业务系统创建为消息来源
// @Description  ## 注意事项
// @Description  - Source Token 仅在创建时返回一次，请妥善保存
// @Description  - 如忘记 Token，可通过查看令牌接口获取（需密码验证）
// @Description  - 来源创建后默认启用状态
// @Description  - 每个来源可绑定多个推送渠道
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120100: 来源名称已存在
// @Description  - 120101: 创建来源失败
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateSourceReq  true  "创建消息来源请求参数"
// @Success      200   {object}  response.Response{data=dto.SourceDTO}  "创建成功"
// @Router       /sources [post]
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

// ListSources godoc
// @Summary      查看消息来源列表
// @Description  分页查询当前用户的所有消息来源列表，返回来源基本信息（不包含 Token）。
// @Description  ## 使用场景
// @Description  1. 来源管理页面展示来源列表
// @Description  2. 下拉选择器获取来源选项
// @Description  ## 注意事项
// @Description  - 列表按创建时间倒序排列
// @Description  - 返回的来源不包含 Token 字段
// @Description  - 已删除的来源不会出现在列表中
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "页码，从 1 开始，默认 1"  default(1)  minimum(1)
// @Param        page_size  query     int  false  "每页条数，默认 20，最大 100"  default(20)  minimum(1)  maximum(100)
// @Success      200   {object}  response.Response{data=response.PageResult{list=[]dto.SourceDTO}}  "查询成功"
// @Router       /sources [get]
// @Security     BearerAuth
func (h *SourceHandler) ListSources(c *gin.Context) {
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

	list, total, err := h.sourceService.ListSources(c.Request.Context(), userID, &pageReq)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithPage(c, list, total, pageReq.Page, pageReq.PageSize)
}

// UpdateSource godoc
// @Summary      编辑消息来源
// @Description  编辑消息来源的名称和描述，不影响已绑定的渠道和推送功能。
// @Description  ## 使用场景
// @Description  1. 修改来源名称使其更易识别
// @Description  2. 更新来源描述信息
// @Description  ## 注意事项
// @Description  - 只能编辑自己创建的来源
// @Description  - 名称不能与其他来源重复
// @Description  - 编辑后不影响 Source Token
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120100: 来源名称已存在
// @Description  - 120102: 来源不存在
// @Description  - 120103: 无权限编辑
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64              true  "来源ID，在路径中传递"  minimum(1)
// @Param        body  body      dto.UpdateSourceReq  true  "编辑消息来源请求参数"
// @Success      200   {object}  response.Response  "编辑成功"
// @Router       /sources/{id} [put]
// @Security     BearerAuth
func (h *SourceHandler) UpdateSource(c *gin.Context) {
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

	sourceIDStr := c.Param("id")
	sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "来源ID格式错误")
		return
	}

	var req dto.UpdateSourceReq
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

	err = h.sourceService.UpdateSource(c.Request.Context(), sourceID, userID, &req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "更新成功", nil)
}

// GetSourceDetail godoc
// @Summary      查看来源详情
// @Description  查询指定消息来源的详细信息，包含来源 Token、使用时间和已绑定的有效渠道列表。
// @Description  ## 使用场景
// @Description  1. 来源详情页展示
// @Description  2. 查看来源绑定的渠道
// @Description  3. 查看来源的推送 Token
// @Description  ## 注意事项
// @Description  - 返回的 Token 可用于推送消息接口
// @Description  - 渠道列表只包含启用状态的渠道
// @Description  - 只能查看自己创建的来源
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 120102: 来源不存在
// @Description  - 120103: 无权限查看
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64  true  "来源ID，在路径中传递"  minimum(1)
// @Success      200   {object}  response.Response{data=dto.SourceDetailResponse}  "查询成功"
// @Router       /sources/{id} [get]
// @Security     BearerAuth
func (h *SourceHandler) GetSourceDetail(c *gin.Context) {
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

	sourceIDStr := c.Param("id")
	sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "来源ID格式错误")
		return
	}

	detail, err := h.sourceService.GetSourceDetail(c.Request.Context(), sourceID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, detail)
}

// GetSourceToken godoc
// @Summary      查看来源令牌
// @Description  查询指定消息来源的 Token，需要密码二次验证以确保安全性。
// @Description  ## 使用场景
// @Description  1. 忘记 Token 时查看
// @Description  2. 配置外部系统推送时获取 Token
// @Description  ## 注意事项
// @Description  - 必须提供正确的用户密码才能查看
// @Description  - Token 格式为 src{uuid}，例如：src0196a3b2c4d50000a1b2c3d4e5f67890
// @Description  - 只能查看自己创建的来源
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 110200: 密码错误
// @Description  - 120102: 来源不存在
// @Description  - 120103: 无权限查看
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64              true  "来源ID，在路径中传递"  minimum(1)
// @Param        body  body      dto.VerifyPasswordReq  true  "密码验证请求参数"
// @Success      200   {object}  response.Response{data=dto.SourceTokenResponse}  "查看成功"
// @Router       /sources/{id}/token [post]
// @Security     BearerAuth
func (h *SourceHandler) GetSourceToken(c *gin.Context) {
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

	sourceIDStr := c.Param("id")
	sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "来源ID格式错误")
		return
	}

	var req dto.VerifyPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.sourceService.VerifyPassword(c.Request.Context(), userID, req.Password); err != nil {
		c.Error(err)
		return
	}

	token, err := h.sourceService.GetSourceToken(c.Request.Context(), sourceID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, dto.SourceTokenResponse{Token: token})
}

// ResetSourceToken godoc
// @Summary      重置来源令牌
// @Description  重置指定消息来源的 Token，需要密码二次验证，旧 Token 立即失效。
// @Description  ## 使用场景
// @Description  1. Token 泄露时紧急重置
// @Description  2. 定期更换 Token 以提高安全性
// @Description  ## 注意事项
// @Description  - 重置后旧 Token 立即失效，使用旧 Token 推送会失败
// @Description  - 必须更新所有使用旧 Token 的外部系统配置
// @Description  - 新 Token 仅在重置时返回一次，请妥善保存
// @Description  - 只能重置自己创建的来源
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 110200: 密码错误
// @Description  - 120102: 来源不存在
// @Description  - 120103: 无权限操作
// @Description  - 120104: 重置令牌失败
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64              true  "来源ID，在路径中传递"  minimum(1)
// @Param        body  body      dto.VerifyPasswordReq  true  "密码验证请求参数"
// @Success      200   {object}  response.Response{data=dto.SourceTokenResponse}  "重置成功"
// @Router       /sources/{id}/token [put]
// @Security     BearerAuth
func (h *SourceHandler) ResetSourceToken(c *gin.Context) {
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

	sourceIDStr := c.Param("id")
	sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "来源ID格式错误")
		return
	}

	var req dto.VerifyPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.sourceService.VerifyPassword(c.Request.Context(), userID, req.Password); err != nil {
		c.Error(err)
		return
	}

	newToken, err := h.sourceService.ResetSourceToken(c.Request.Context(), sourceID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, dto.SourceTokenResponse{Token: newToken})
}

// DisableSource godoc
// @Summary      停用消息来源
// @Description  停用指定消息来源，需要密码二次验证，停用后该来源无法推送消息。
// @Description  ## 使用场景
// @Description  1. 临时停止某个来源的消息推送
// @Description  2. 来源不再使用时停用而非删除
// @Description  ## 注意事项
// @Description  - 停用后使用该 Token 推送的消息会被拒绝
// @Description  - 已推送的历史消息不受影响
// @Description  - 可以随时重新启用
// @Description  - 只能停用自己创建的来源
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 110200: 密码错误
// @Description  - 120102: 来源不存在
// @Description  - 120103: 无权限操作
// @Description  - 120105: 来源已停用
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64              true  "来源ID，在路径中传递"  minimum(1)
// @Param        body  body      dto.VerifyPasswordReq  true  "密码验证请求参数"
// @Success      200   {object}  response.Response  "停用成功"
// @Router       /sources/{id}/disable [put]
// @Security     BearerAuth
func (h *SourceHandler) DisableSource(c *gin.Context) {
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

	sourceIDStr := c.Param("id")
	sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "来源ID格式错误")
		return
	}

	var req dto.VerifyPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.sourceService.VerifyPassword(c.Request.Context(), userID, req.Password); err != nil {
		c.Error(err)
		return
	}

	err = h.sourceService.DisableSource(c.Request.Context(), sourceID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "已停用", nil)
}

// EnableSource godoc
// @Summary      启用消息来源
// @Description  启用指定消息来源，需要密码二次验证，恢复消息推送功能。
// @Description  ## 使用场景
// @Description  1. 重新启用之前停用的来源
// @Description  2. 恢复某个来源的消息推送能力
// @Description  ## 注意事项
// @Description  - 只能启用已停用的来源
// @Description  - 启用后该来源的 Token 立即生效
// @Description  - 只能启用自己创建的来源
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 110200: 密码错误
// @Description  - 120102: 来源不存在
// @Description  - 120103: 无权限操作
// @Description  - 120106: 来源已启用
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64              true  "来源ID，在路径中传递"  minimum(1)
// @Param        body  body      dto.VerifyPasswordReq  true  "密码验证请求参数"
// @Success      200   {object}  response.Response  "启用成功"
// @Router       /sources/{id}/enable [put]
// @Security     BearerAuth
func (h *SourceHandler) EnableSource(c *gin.Context) {
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

	sourceIDStr := c.Param("id")
	sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "来源ID格式错误")
		return
	}

	var req dto.VerifyPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.sourceService.VerifyPassword(c.Request.Context(), userID, req.Password); err != nil {
		c.Error(err)
		return
	}

	err = h.sourceService.EnableSource(c.Request.Context(), sourceID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "已启用", nil)
}

// DeleteSource godoc
// @Summary      删除消息来源
// @Description  软删除指定消息来源及其关联渠道关系，需要密码二次验证。删除后来源和所有关联数据不可恢复。
// @Description  ## 使用场景
// @Description  1. 永久删除不再使用的来源
// @Description  2. 清理测试数据
// @Description  ## 注意事项
// @Description  - 删除是软删除，数据库中标记为已删除状态
// @Description  - 删除后该来源的 Token 立即失效
// @Description  - 关联的渠道关系也会被删除
// @Description  - 只能删除自己创建的来源
// @Description  - 注意：DELETE 请求需在 Body 中传递密码（{"password": "xxx"}），主流网关均支持此用法
// @Description  ## 错误码说明
// @Description  - 100001: 未提供认证令牌
// @Description  - 110200: 密码错误
// @Description  - 120102: 来源不存在
// @Description  - 120103: 无权限操作
// @Description  - 120107: 删除来源失败
// @Tags         消息来源管理
// @Accept       json
// @Produce      json
// @Param        id    path      int64              true  "来源ID，在路径中传递"  minimum(1)
// @Param        body  body      dto.VerifyPasswordReq  true  "密码验证请求参数"
// @Success      200   {object}  response.Response  "删除成功"
// @Router       /sources/{id} [delete]
// @Security     BearerAuth
func (h *SourceHandler) DeleteSource(c *gin.Context) {
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

	sourceIDStr := c.Param("id")
	sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
	if err != nil {
		response.Fail(c, xerr.ErrBadRequest.Code, "来源ID格式错误")
		return
	}

	var req dto.VerifyPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.sourceService.VerifyPassword(c.Request.Context(), userID, req.Password); err != nil {
		c.Error(err)
		return
	}

	err = h.sourceService.DeleteSource(c.Request.Context(), sourceID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SuccessWithMsg(c, "已删除", nil)
}
