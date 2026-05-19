package sender

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"octotify/internal/client/ilink"

	"go.uber.org/zap/zaptest"
	"gorm.io/datatypes"
)

// TestNewSenderFactory 测试工厂初始化是否包含所有内置渠道
func TestNewSenderFactory(t *testing.T) {
	// 准备测试日志和 ilink client
	logger := zaptest.NewLogger(t)
	ilinkClient := ilink.NewClient(logger)

	// 创建工厂实例
	factory := NewSenderFactory(logger, ilinkClient)

	// 验证工厂非空
	if factory == nil {
		t.Fatal("NewSenderFactory() returned nil")
	}

	// 表驱动测试：验证所有内置渠道类型均被注册
	builtinChannels := []string{"wechat", "wechat_clawbot", "telegram", "dingtalk", "email", "webhook", "feishu"}

	for _, channelType := range builtinChannels {
		t.Run(fmt.Sprintf("内置渠道_%s应被注册", channelType), func(t *testing.T) {
			snd, err := factory.Create(channelType)
			if err != nil {
				t.Fatalf("Create(%q) unexpected error: %v", channelType, err)
			}
			if snd == nil {
				t.Fatalf("Create(%q) returned nil Sender", channelType)
			}
		})
	}
}

// TestSenderFactory_Create 测试工厂创建 Sender 的行为
func TestSenderFactory_Create(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ilinkClient := ilink.NewClient(logger)
	factory := NewSenderFactory(logger, ilinkClient)

	// 表驱动测试用例
	tests := []struct {
		name        string // 用例名称
		channelType string // 渠道类型
		wantErr     bool   // 是否期望错误
		errContains string // 期望错误包含的字符串
	}{
		{
			name:        "有效渠道wechat应返回Sender",
			channelType: "wechat",
			wantErr:     false,
		},
		{
			name:        "有效渠道telegram应返回Sender",
			channelType: "telegram",
			wantErr:     false,
		},
		{
			name:        "有效渠道dingtalk应返回Sender",
			channelType: "dingtalk",
			wantErr:     false,
		},
		{
			name:        "有效渠道email应返回Sender",
			channelType: "email",
			wantErr:     false,
		},
		{
			name:        "有效渠道webhook应返回Sender",
			channelType: "webhook",
			wantErr:     false,
		},
		{
			name:        "有效渠道feishu应返回Sender",
			channelType: "feishu",
			wantErr:     false,
		},
		{
			name:        "无效渠道应返回错误",
			channelType: "invalid_channel",
			wantErr:     true,
			errContains: "不支持的渠道类型",
		},
		{
			name:        "空字符串渠道应返回错误",
			channelType: "",
			wantErr:     true,
			errContains: "不支持的渠道类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snd, err := factory.Create(tt.channelType)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Create(%q) expected error, got nil", tt.channelType)
				}
				if snd != nil {
					t.Fatalf("Create(%q) expected nil Sender on error, got %T", tt.channelType, snd)
				}
				// 验证错误信息包含期望字符串
				errStr := err.Error()
				if !strings.Contains(errStr, tt.errContains) {
					t.Errorf("Create(%q) error = %q, want containing %q", tt.channelType, errStr, tt.errContains)
				}
			} else {
				if err != nil {
					t.Fatalf("Create(%q) unexpected error: %v", tt.channelType, err)
				}
				if snd == nil {
					t.Fatalf("Create(%q) returned nil Sender", tt.channelType)
				}
			}
		})
	}
}

// TestSenderFactory_Register 测试动态注册渠道发送器
func TestSenderFactory_Register(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("注册新渠道类型后应能成功创建", func(t *testing.T) {
		factory := NewSenderFactory(logger, ilink.NewClient(logger))

		// 创建自定义 Sender
		customSender := &MockSender{
			SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
				return nil
			},
		}

		// 注册新渠道
		factory.Register("custom_channel", customSender)

		// 验证可以创建
		snd, err := factory.Create("custom_channel")
		if err != nil {
			t.Fatalf("Create('custom_channel') unexpected error: %v", err)
		}
		if snd == nil {
			t.Fatal("Create('custom_channel') returned nil Sender")
		}
	})

	t.Run("注册同名渠道应覆盖原有实例", func(t *testing.T) {
		factory := NewSenderFactory(logger, ilink.NewClient(logger))

		// 先获取原始 wechat Sender
		originalSnd, err := factory.Create("wechat")
		if err != nil {
			t.Fatalf("Create('wechat') unexpected error: %v", err)
		}

		// 创建自定义 Sender 并覆盖 wechat
		overrideSender := &MockSender{
			SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
				return fmt.Errorf("overridden")
			},
		}
		factory.Register("wechat", overrideSender)

		// 验证已覆盖
		newSnd, err := factory.Create("wechat")
		if err != nil {
			t.Fatalf("Create('wechat') unexpected error after override: %v", err)
		}
		if newSnd == originalSnd {
			t.Fatal("Register('wechat') did not override the original Sender")
		}

		// 验证覆盖后的 Sender 行为
		err = newSnd.Send(context.Background(), nil, "test", "test")
		if err == nil {
			t.Fatal("expected overridden Sender to return error, got nil")
		}
		if err.Error() != "overridden" {
			t.Errorf("expected error 'overridden', got %q", err.Error())
		}
	})
}
