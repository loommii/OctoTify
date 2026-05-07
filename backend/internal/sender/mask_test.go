package sender

import (
	"testing"
)

// TestMaskToken 测试 Token 脱敏函数（从 telegram_test.go 迁移，因 maskToken 已提取到 mask.go）
func TestMaskToken(t *testing.T) {
	// 表驱动测试用例
	tests := []struct {
		name  string // 用例名称
		token string // 输入 token
		want  string // 期望输出
	}{
		{
			name:  "长token应保留前4位和后4位",
			token: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
			want:  "1234...wxyz",
		},
		{
			name:  "恰好8位应返回星号",
			token: "12345678",
			want:  "****",
		},
		{
			name:  "少于8位应返回星号",
			token: "short",
			want:  "****",
		},
		{
			name:  "空字符串应返回星号",
			token: "",
			want:  "****",
		},
		{
			name:  "9位token应脱敏为前4后4",
			token: "123456789",
			want:  "1234...6789",
		},
		{
			name:  "恰好超过8位应正确脱敏",
			token: "abcdefghij",
			want:  "abcd...ghij",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskToken(tt.token)
			if got != tt.want {
				t.Errorf("maskToken(%q) = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}

// TestMaskEmailUsername 测试邮箱用户名脱敏函数（从 email_test.go 迁移，因 maskEmailUsername 已提取到 mask.go）
func TestMaskEmailUsername(t *testing.T) {
	// 表驱动测试用例
	tests := []struct {
		name  string // 用例名称
		email string // 输入邮箱
		want  string // 期望输出
	}{
		{
			name:  "正常邮箱应脱敏用户名为星号",
			email: "user@example.com",
			want:  "***@example.com",
		},
		{
			name:  "QQ邮箱应正确脱敏",
			email: "123456789@qq.com",
			want:  "***@qq.com",
		},
		{
			name:  "无@符号应返回星号",
			email: "invalid-email",
			want:  "***",
		},
		{
			name:  "空字符串应返回星号",
			email: "",
			want:  "***",
		},
		{
			name:  "仅有@符号应返回星号",
			email: "@",
			want:  "***",
		},
		{
			name:  "@开头无用户名应返回星号",
			email: "@example.com",
			want:  "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskEmailUsername(tt.email)
			if got != tt.want {
				t.Errorf("maskEmailUsername(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}
