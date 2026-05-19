package validator

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"octotify/internal/handler/dto"
)

// 用户名正则表达式：只允许字母、数字和下划线
var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")

// Init 初始化自定义验证器，注册到 Gin 的验证器引擎
func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", validateUsername)
		v.RegisterValidation("password", validatePassword)
		v.RegisterValidation("channel_type", validateChannelType)
	}
}

// validateChannelType 验证渠道类型
func validateChannelType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return dto.ValidChannelTypes[value]
}

// validateUsername 验证用户名格式
// 要求：长度3-64个字符，只能包含字母、数字和下划线
func validateUsername(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return len(value) >= 3 && len(value) <= 64 && usernameRegex.MatchString(value)
}

// validatePassword 验证密码强度
// 要求：长度8-64个字符，必须同时包含大写字母、小写字母和数字
func validatePassword(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if len(value) < 8 || len(value) > 64 {
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

// ValidateChannelConfig 根据渠道类型元数据校验配置必填字段
func ValidateChannelConfig(channelType string, config map[string]any) error {
	var meta *dto.ChannelTypeMeta
	for i := range dto.ChannelTypeMetas {
		if dto.ChannelTypeMetas[i].Type == channelType {
			meta = &dto.ChannelTypeMetas[i]
			break
		}
	}

	if meta == nil {
		return fmt.Errorf("未知的渠道类型: %s", channelType)
	}

	var missingFields []string
	for _, field := range meta.ConfigFields {
		if !field.Required {
			continue
		}

		val, exists := config[field.Name]
		if !exists {
			missingFields = append(missingFields, field.Name)
			continue
		}

		switch v := val.(type) {
		case string:
			if v == "" {
				missingFields = append(missingFields, field.Name)
			}
		case nil:
			missingFields = append(missingFields, field.Name)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("缺少必填配置字段: %v", missingFields)
	}

	return nil
}
