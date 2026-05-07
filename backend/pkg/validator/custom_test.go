// Package validator 提供自定义校验器的单元测试
package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"

	"octotify/internal/handler/dto"
)

// setupValidator 创建并配置用于测试的校验器实例
func setupValidator(t *testing.T) *validator.Validate {
	t.Helper()
	v := validator.New()

	// 注册自定义校验规则
	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("password", validatePassword)

	// 注册渠道类型校验（需要访问 dto.ValidChannelTypes）
	v.RegisterValidation("channel_type", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return dto.ValidChannelTypes[value]
	})

	return v
}

// generateString 生成指定长度的字符串
func generateString(n int, ch string) string {
	result := ""
	for i := 0; i < n; i++ {
		result += ch
	}
	return result
}

// TestInit 测试自定义校验器初始化，验证校验器引擎类型正确
func TestInit(t *testing.T) {
	v := validator.New()
	if v == nil {
		t.Fatal("expected validator instance to be created")
	}
}

// TestValidateUsername 测试用户名格式校验规则
// 规则：长度3-64个字符，只能包含字母、数字和下划线
func TestValidateUsername(t *testing.T) {
	v := setupValidator(t)

	// 定义测试用例：包含合法和非法的用户名
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid_simple", "user123", false},                                                          // 字母数字组合
		{"valid_with_underscore", "user_name", false},                                               // 包含下划线
		{"valid_mixed", "User_123_Test", false},                                                     // 大小写字母、数字、下划线混合
		{"valid_min_length", "abc", false},                                                          // 最小长度3
		{"valid_max_length", generateString(64, "a"), false}, // 最大长度64
		{"too_short", "ab", true},                            // 长度不足3
		{"too_long", generateString(65, "a"), true},          // 长度65（超过64）
		{"contains_hyphen", "user-name", true},                                                      // 包含连字符（非法）
		{"contains_space", "user name", true},                                                       // 包含空格（非法）
		{"contains_dot", "user.name", true},                                                         // 包含点（非法）
		{"contains_special", "user@name", true},                                                     // 包含特殊字符（非法）
		{"empty", "", true},                                                                         // 空字符串
		{"only_numbers", "12345", false},                                                            // 纯数字
		{"only_letters", "username", false},                                                         // 纯字母
		{"only_underscores", "___", false},                                                          // 纯下划线
	}

	type TestStruct struct {
		Username string `validate:"username"`
	}

	// 遍历测试用例，验证每个用户名的校验结果
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := TestStruct{Username: tt.value}
			err := v.Struct(obj)

			if tt.wantErr && err == nil {
				t.Errorf("expected error for username '%s', but got none", tt.value)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error for username '%s', but got: %v", tt.value, err)
			}
		})
	}
}

// TestValidatePassword 测试密码强度校验规则
// 规则：长度8-128个字符，必须同时包含大写字母、小写字母和数字
func TestValidatePassword(t *testing.T) {
	v := setupValidator(t)

	// 定义测试用例：包含各种密码强度场景
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid_simple", "Password1", false},      // 简单合法密码
		{"valid_complex", "Str0ng!@#Pass", false}, // 复杂密码（含特殊字符）
		{"valid_min_length", "Abcdefg1", false},   // 最小长度8
		{"valid_max_length", "A" + generateString(126, "b") + "1", false}, // 最大长度128
		{"too_short", "Pass1", true},              // 长度不足8
		{"too_long", "A" + generateString(127, "b") + "1", true}, // 长度129（超过128）
		{"no_uppercase", "password1", true},           // 缺少大写字母
		{"no_lowercase", "PASSWORD1", true},           // 缺少小写字母
		{"no_digit", "Passwordabc", true},             // 缺少数字
		{"empty", "", true},                           // 空密码
		{"only_letters", "Abcdefghijk", true},         // 只有字母，缺少数字
		{"only_numbers", "12345678", true},            // 只有数字，缺少字母
		{"special_chars_allowed", "Pass!@#$1", false}, // 包含特殊字符（允许）
		{"underscore_allowed", "Pass_word1", false},   // 包含下划线（允许）
	}

	type TestStruct struct {
		Password string `validate:"password"`
	}

	// 遍历测试用例，验证每个密码的校验结果
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := TestStruct{Password: tt.value}
			err := v.Struct(obj)

			if tt.wantErr && err == nil {
				t.Errorf("expected error for password '%s', but got none", tt.value)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error for password '%s', but got: %v", tt.value, err)
			}
		})
	}
}

// TestValidateChannelType 测试渠道类型校验规则
// 规则：必须是预定义的合法渠道类型（大小写敏感）
func TestValidateChannelType(t *testing.T) {
	v := setupValidator(t)

	// 定义测试用例：包含合法和非法的渠道类型
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid_feishu", "feishu", false},     // 飞书
		{"valid_wechat", "wechat", false},     // 企业微信
		{"valid_telegram", "telegram", false}, // Telegram
		{"valid_dingtalk", "dingtalk", false}, // 钉钉
		{"valid_email", "email", false},       // 邮件
		{"valid_webhook", "webhook", false},   // Webhook
		{"invalid_type", "invalid", true},     // 非法类型
		{"empty", "", true},                   // 空字符串
		{"uppercase", "FEISHU", true},         // 大写（大小写敏感，应为非法）
		{"mixed_case", "Feishu", true},        // 混合大小写（应为非法）
	}

	type TestStruct struct {
		ChannelType string `validate:"channel_type"`
	}

	// 遍历测试用例，验证每个渠道类型的校验结果
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := TestStruct{ChannelType: tt.value}
			err := v.Struct(obj)

			if tt.wantErr && err == nil {
				t.Errorf("expected error for channel_type '%s', but got none", tt.value)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error for channel_type '%s', but got: %v", tt.value, err)
			}
		})
	}
}

// TestValidateUsername_BoundaryLength 测试用户名长度的边界条件（3字符和64字符）
func TestValidateUsername_BoundaryLength(t *testing.T) {
	v := setupValidator(t)

	type TestStruct struct {
		Username string `validate:"username"`
	}

	// 测试最小边界：3字符（合法）
	exact3 := "abc"
	obj1 := TestStruct{Username: exact3}
	err1 := v.Struct(obj1)
	if err1 != nil {
		t.Errorf("expected no error for 3-char username, but got: %v", err1)
	}

	// 测试最大边界：64字符（合法）
	exact64 := generateString(64, "a")
	obj2 := TestStruct{Username: exact64}
	err2 := v.Struct(obj2)
	if err2 != nil {
		t.Errorf("expected no error for 64-char username, but got: %v", err2)
	}

	// 测试超出最大边界：65字符（非法）
	exact65 := generateString(65, "a")
	obj3 := TestStruct{Username: exact65}
	err3 := v.Struct(obj3)
	if err3 == nil {
		t.Error("expected error for 65-char username, but got none")
	}
}

// TestValidatePassword_BoundaryLength 测试密码长度的边界条件（8字符和128字符）
func TestValidatePassword_BoundaryLength(t *testing.T) {
	v := setupValidator(t)

	type TestStruct struct {
		Password string `validate:"password"`
	}

	// 测试最小边界：8字符（合法）
	exact8 := "Abcdefg1"
	obj1 := TestStruct{Password: exact8}
	err1 := v.Struct(obj1)
	if err1 != nil {
		t.Errorf("expected no error for 8-char password, but got: %v", err1)
	}

	// 测试最大边界：128字符（合法）
	exact128 := "A" + generateString(126, "b") + "1"
	obj2 := TestStruct{Password: exact128}
	err2 := v.Struct(obj2)
	if err2 != nil {
		t.Errorf("expected no error for 128-char password, but got: %v", err2)
	}

	// 测试超出最大边界：129字符（非法）
	exact129 := "A" + generateString(127, "b") + "1"
	obj3 := TestStruct{Password: exact129}
	err3 := v.Struct(obj3)
	if err3 == nil {
		t.Error("expected error for 129-char password, but got none")
	}
}

// makeString 生成指定长度的字符串，用于测试边界条件（已废弃，使用 generateString 替代）
func makeString(n int, ch string) string {
	return generateString(n, ch)
}
