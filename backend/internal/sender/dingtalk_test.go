package sender

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap/zaptest"
	"gorm.io/datatypes"
)

// TestDingtalkSender_NewInstance 测试钉钉发送器新建实例持有logger
func TestDingtalkSender_NewInstance(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	if sender == nil {
		t.Fatal("NewDingtalkSender() returned nil")
	}
	if sender.logger == nil {
		t.Error("NewDingtalkSender() sender.logger is nil")
	}
}

// TestDingtalkSender_Send 测试钉钉发送器Send方法返回未实现错误
func TestDingtalkSender_Send(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sender := NewDingtalkSender(logger)

	tests := []struct {
		name        string
		config      datatypes.JSON
		title       string
		content     string
		errContains string
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
				t.Fatalf("DingtalkSender.Send() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("DingtalkSender.Send() error = %q, want containing %q", err.Error(), tt.errContains)
			}
		})
	}
}
