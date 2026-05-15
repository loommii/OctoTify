package response

import (
	"errors"
	"net/http"
	"octotify/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Response 统一响应结构
type Response struct {
	Code int    `json:"code"`           // 业务状态码，0 表示成功，非 0 表示失败
	Msg  string `json:"msg"`            // 提示信息，成功时为"请求成功"，失败时为错误描述
	Data any    `json:"data,omitempty"` // 业务数据，成功时返回业务数据，失败时可能为空或错误详情
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
	lang := c.GetHeader("Accept-Language")
	translatedMsg := xerr.TranslateMsg(code, lang)
	if translatedMsg != "" {
		msg = translatedMsg
	}
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

// Unauthorized JWT 鉴权失败响应，返回 HTTP 401 + 业务 code
func Unauthorized(c *gin.Context, code int) {
	lang := c.GetHeader("Accept-Language")
	msg := xerr.TranslateMsg(code, lang)
	if msg == "" {
		msg = "未登录或Token已过期"
	}
	c.JSON(http.StatusUnauthorized, Response{
		Code: code,
		Msg:  msg,
	})
}

// FailWithData 失败响应，携带附加数据
func FailWithData(c *gin.Context, code int, msg string, data any) {
	lang := c.GetHeader("Accept-Language")
	translatedMsg := xerr.TranslateMsg(code, lang)
	if translatedMsg != "" {
		msg = translatedMsg
	}
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

// PageResult 分页响应数据
type PageResult struct {
	List     any   `json:"list"`      // 数据列表
	Total    int64 `json:"total"`     // 总记录数
	Page     int   `json:"page"`      // 当前页码
	PageSize int   `json:"page_size"` // 每页条数
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
	Field   string `json:"field"`   // 字段名称
	Message string `json:"message"` // 错误描述
}

// HandleValidationError 处理参数校验错误，返回字段级错误信息
func HandleValidationError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if ok := errors.As(err, &validationErrors); !ok {
		c.Error(xerr.ErrBadRequest.WithInternal(err))
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
		Code: xerr.ErrBadRequest.Code,
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
