package errors

// AppError 应用级业务错误
type AppError struct {
	Code     int    `json:"code"` // 业务错误码
	Msg      string `json:"msg"`  // 面向用户的提示
	Internal string `json:"-"`    // 内部错误信息，仅日志使用
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Internal != "" {
		return e.Internal
	}
	return e.Msg
}

// New 创建 AppError
func New(code int, msg string) *AppError {
	return &AppError{
		Code: code,
		Msg:  msg,
	}
}

// WithInternal 附加内部错误信息，返回新对象不修改原始定义
func (e *AppError) WithInternal(err error) *AppError {
	newErr := *e
	newErr.Internal = err.Error()
	return &newErr
}
