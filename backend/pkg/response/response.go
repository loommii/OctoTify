package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	apperrors "octotify/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Code int    `json:"code"`           // 业务状态码，0 表示成功
	Msg  string `json:"msg"`            // 提示信息
	Data any    `json:"data,omitempty"` // 业务数据，无数据时省略
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "请求成功",
		Data: data,
	})
}

// SuccessWithMsg 成功响应，自定义消息
func SuccessWithMsg(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  msg,
		Data: data,
	})
}

// Fail 失败响应，统一返回 HTTP 200 + 业务 code
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

// Unauthorized JWT 鉴权失败响应，返回 HTTP 401 + 业务 code
func Unauthorized(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code: code,
		Msg:  msg,
	})
}

// FailWithData 失败响应，携带附加数据
func FailWithData(c *gin.Context, code int, msg string, data any) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

// PageResult 分页响应数据
type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// SuccessWithPage 分页成功响应
func SuccessWithPage(c *gin.Context, list any, total int64, page, pageSize int) {
	Success(c, PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// FieldError 字段级校验错误
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// HandleValidationError 处理参数校验错误，返回字段级错误信息
func HandleValidationError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); !ok {
		c.Error(apperrors.ErrBadRequest.WithInternal(err))
		return
	}

	fieldErrors := make([]FieldError, 0, len(validationErrors))
	for _, e := range validationErrors {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   e.Field(),
			Message: msgForTag(e),
		})
	}

	c.AbortWithStatusJSON(http.StatusOK, Response{
		Code: apperrors.ErrBadRequest.Code,
		Msg:  "请求参数校验失败",
		Data: fieldErrors,
	})
}

// msgForTag 将 validator 标签翻译为中文提示
func msgForTag(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "不能为空"
	case "email":
		return "邮箱格式不正确"
	case "min":
		return "长度不能小于 " + e.Param()
	case "max":
		return "长度不能大于 " + e.Param()
	case "oneof":
		return "值必须是以下之一: " + e.Param()
	default:
		return "校验失败: " + e.Tag()
	}
}
