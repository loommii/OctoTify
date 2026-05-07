package sender

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap/zaptest"
	"gorm.io/datatypes"
)

// TestWechatSender_NewInstance 测试微信发送器新建实例持有logger
func TestWechatSender_NewInstance(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatSender(logger)

	if sender == nil {
		t.Fatal("NewWechatSender() returned nil")
	}
	if sender.logger == nil {
		t.Error("NewWechatSender() sender.logger is nil")
	}
}

// TestWechatSender_Send 测试微信发送器Send方法返回未实现错误
func TestWechatSender_Send(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewWechatSender(logger)

	// 表驱动测试用例
	tests := []struct {
		name        string
		config      datatypes.JSON
		title       string
		content     string
		errContains string // 期望错误包含的字符串
	}{
		{
			name:        "空配置应返回未实现错误",
			config:      nil,
			title:       "",
			content:     "",
			errContains: "尚未实现",
		},
		{
			name:        "有效配置应返回未实现错误",
			config:      datatypes.JSON(`{"key":"value"}`),
			title:       "test title",
			content:     "test content",
			errContains: "尚未实现",
		},
		{
			name:        "任意输入应返回未实现错误",
			config:      datatypes.JSON(`{}`),
			title:       "通知",
			content:     "这是一条测试消息",
			errContains: "尚未实现",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.Send(context.Background(), tt.config, tt.title, tt.content)
			if err == nil {
				t.Fatalf("WechatSender.Send() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("WechatSender.Send() error = %q, want containing %q", err.Error(), tt.errContains)
			}
		})
	}
}
