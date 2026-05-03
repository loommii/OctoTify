package validator

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// 用户名正则表达式：只允许字母、数字和下划线
var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")

// Init 初始化自定义验证器，注册到 Gin 的验证器引擎
func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", validateUsername)
		v.RegisterValidation("password", validatePassword)
	}
}

// validateUsername 验证用户名格式
// 要求：长度3-64个字符，只能包含字母、数字和下划线
func validateUsername(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return len(value) >= 3 && len(value) <= 64 && usernameRegex.MatchString(value)
}

// validatePassword 验证密码强度
// 要求：长度8-128个字符，必须同时包含大写字母、小写字母和数字
func validatePassword(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if len(value) < 8 || len(value) > 128 {
		return false
	}

	var hasLower, hasUpper, hasDigit bool
	for _, ch := range value {
		switch {
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		}
	}

	return hasLower && hasUpper && hasDigit
}
