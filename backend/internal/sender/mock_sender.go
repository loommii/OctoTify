// Package sender 提供测试用的 Mock 发送器
package sender

import (
	"context"

	"gorm.io/datatypes"
)

// MockSender 用于测试的 Mock 发送器
type MockSender struct {
	SendFunc func(ctx context.Context, config datatypes.JSON, title string, content string) error
}

// Send 模拟发送消息
func (m *MockSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, config, title, content)
	}
	return nil
}

// MockSenderFactory 用于测试的 Mock 工厂
type MockSenderFactory struct {
	Senders map[string]Sender
}

// NewMockSenderFactory 创建 Mock 工厂
func NewMockSenderFactory() *MockSenderFactory {
	return &MockSenderFactory{
		Senders: make(map[string]Sender),
	}
}

// Create 创建发送器（使用 Mock）
func (f *MockSenderFactory) Create(channelType string) (Sender, error) {
	if snd, ok := f.Senders[channelType]; ok {
		return snd, nil
	}
	return nil, nil
}
