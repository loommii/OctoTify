// errors.go 定义 AppError 核心结构体和方法
//
// 提供以下功能：
//   - AppError 结构体：包含 Code、Msg、Internal、Cause 字段
//   - New(): 创建业务错误
//   - WithInternal(): 附加内部错误信息
//   - WithCause(): 仅附加原始错误
//   - Wrap(): 格式化包装错误
//   - HTTPStatus(): 获取 HTTP 状态码
//   - IsBadRequest()/IsServerError(): 错误类型判断工具
package xerr

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError 应用级业务错误
type AppError struct {
	Code       int    `json:"code"` // 业务错误码
	Msg        string `json:"msg"`  // 面向用户的提示
	httpStatus int    `json:"-"`    // HTTP 状态码
	Internal   string `json:"-"`    // 内部错误信息，仅日志使用
	Cause      error  `json:"-"`    // 原始错误，支持错误链
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Cause != nil && e.Internal != "" {
		return fmt.Sprintf("%s: %s", e.Msg, e.Internal)
	}
	if e.Internal != "" {
		return e.Internal
	}
	return e.Msg
}

// Unwrap 实现 Go 1.13+ 错误链接口，支持 errors.Is/As
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is 支持错误等价判断，相同错误码视为等价
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// New 创建 AppError
func New(code int, msg ...string) *AppError {
	m := ""
	if len(msg) > 0 {
		m = msg[0]
	}
	return &AppError{
		Code:       code,
		Msg:        m,
		httpStatus: HTTPStatusFromCode(code),
	}
}

// WithInternal 附加内部错误信息，返回新对象不修改原始定义
func (e *AppError) WithInternal(err error) *AppError {
	if err == nil {
		return e
	}
	newErr := *e
	newErr.Internal = err.Error()
	newErr.Cause = err
	return &newErr
}

// WithCause 仅附加原始错误，不修改 Internal 字符串
func (e *AppError) WithCause(err error) *AppError {
	if err == nil {
		return e
	}
	newErr := *e
	newErr.Cause = err
	return &newErr
}

// Wrap 包装错误，添加上下文信息，保留原始错误链
func (e *AppError) Wrap(format string, args ...any) *AppError {
	newErr := *e
	newErr.Internal = fmt.Sprintf(format, args...)
	return &newErr
}

// HTTPStatus 返回对应的 HTTP 状态码
func (e *AppError) HTTPStatus() int {
	if e.httpStatus != 0 {
		return e.httpStatus
	}
	return http.StatusOK
}

// IsBadRequest 判断是否为客户端请求错误（4xx）
func IsBadRequest(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		status := appErr.HTTPStatus()
		return status >= 400 && status < 500
	}
	return false
}

// IsServerError 判断是否为服务器内部错误（5xx）
func IsServerError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		status := appErr.HTTPStatus()
		return status >= 500
	}
	return false
}
