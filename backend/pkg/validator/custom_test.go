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

// TestValidateChannelConfig 测试渠道配置必填字段校验
func TestValidateChannelConfig(t *testing.T) {
	tests := []struct {
		name        string
		channelType string
		config      map[string]any
		wantErr     bool
		errContains string
	}{
		// === 合法配置 ===
		{
			name:        "telegram_合法_必填字段完整",
			channelType: "telegram",
			config:      map[string]any{"bot_token": "123456:ABC", "chat_id": "12345"},
			wantErr:     false,
		},
		{
			name:        "telegram_合法_包含可选字段",
			channelType: "telegram",
			config:      map[string]any{"bot_token": "123456:ABC", "chat_id": "12345", "proxy": "http://127.0.0.1:7890"},
			wantErr:     false,
		},
		{
			name:        "feishu_合法_只提供必填字段",
			channelType: "feishu",
			config:      map[string]any{"webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"},
			wantErr:     false,
		},
		{
			name:        "feishu_合法_包含可选字段",
			channelType: "feishu",
			config:      map[string]any{"webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx", "secret": "xxx"},
			wantErr:     false,
		},
		{
			name:        "dingtalk_合法_必填字段完整",
			channelType: "dingtalk",
			config:      map[string]any{"webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx", "secret": "SECxxx"},
			wantErr:     false,
		},
		{
			name:        "email_合法_必填字段完整",
			channelType: "email",
			config: map[string]any{
				"smtp_host": "smtp.example.com",
				"smtp_port": 587,
				"username":  "user@example.com",
				"password":  "pass123",
				"to":        "recipient@example.com",
			},
			wantErr: false,
		},

		// === 空配置 ===
		{
			name:        "空配置_telegram",
			channelType: "telegram",
			config:      map[string]any{},
			wantErr:     true,
			errContains: "缺少必填配置字段",
		},
		{
			name:        "空配置_feishu",
			channelType: "feishu",
			config:      map[string]any{},
			wantErr:     true,
			errContains: "缺少必填配置字段",
		},
		{
			name:        "空配置_email",
			channelType: "email",
			config:      map[string]any{},
			wantErr:     true,
			errContains: "缺少必填配置字段",
		},

		// === 缺失必填字段 ===
		{
			name:        "telegram_缺少bot_token",
			channelType: "telegram",
			config:      map[string]any{"chat_id": "12345"},
			wantErr:     true,
			errContains: "bot_token",
		},
		{
			name:        "telegram_缺少chat_id",
			channelType: "telegram",
			config:      map[string]any{"bot_token": "123456:ABC"},
			wantErr:     true,
			errContains: "chat_id",
		},
		{
			name:        "telegram_缺少多个必填字段",
			channelType: "telegram",
			config:      map[string]any{},
			wantErr:     true,
			errContains: "bot_token",
		},
		{
			name:        "dingtalk_缺少secret",
			channelType: "dingtalk",
			config:      map[string]any{"webhook_url": "https://xxx"},
			wantErr:     true,
			errContains: "secret",
		},
		{
			name:        "email_缺少多个必填字段",
			channelType: "email",
			config:      map[string]any{"smtp_host": "smtp.example.com"},
			wantErr:     true,
			errContains: "smtp_port",
		},


		// === 字段值为空字符串 ===
		{
			name:        "telegram_bot_token为空字符串",
			channelType: "telegram",
			config:      map[string]any{"bot_token": "", "chat_id": "12345"},
			wantErr:     true,
			errContains: "bot_token",
		},
		{
			name:        "telegram_chat_id为空字符串",
			channelType: "telegram",
			config:      map[string]any{"bot_token": "123456:ABC", "chat_id": ""},
			wantErr:     true,
			errContains: "chat_id",
		},
		{
			name:        "dingtalk_webhook_url为空字符串",
			channelType: "dingtalk",
			config:      map[string]any{"webhook_url": "", "secret": "SECxxx"},
			wantErr:     true,
			errContains: "webhook_url",
		},

		// === 字段值为 nil ===
		{
			name:        "telegram_bot_token为nil",
			channelType: "telegram",
			config:      map[string]any{"bot_token": nil, "chat_id": "12345"},
			wantErr:     true,
			errContains: "bot_token",
		},
		{
			name:        "telegram_chat_id为nil",
			channelType: "telegram",
			config:      map[string]any{"bot_token": "123456:ABC", "chat_id": nil},
			wantErr:     true,
			errContains: "chat_id",
		},

		// === 未知渠道类型 ===
		{
			name:        "未知渠道类型",
			channelType: "unknown_channel",
			config:      map[string]any{"some_key": "some_value"},
			wantErr:     true,
			errContains: "未知的渠道类型",
		},
		{
			name:        "空字符串渠道类型",
			channelType: "",
			config:      map[string]any{},
			wantErr:     true,
			errContains: "未知的渠道类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChannelConfig(tt.channelType, tt.config)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, but got none")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, but got: %v", err)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil {
					t.Errorf("expected error containing '%s', but got nil", tt.errContains)
					return
				}
				if !containsStr(err.Error(), tt.errContains) {
					t.Errorf("expected error containing '%s', but got: %v", tt.errContains, err)
				}
			}
		})
	}
}

// containsStr 检查字符串是否包含子串
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
