package sender

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// WebhookSender Webhook 消息发送器
type WebhookSender struct {
	logger *zap.Logger
}

// NewWebhookSender 创建 Webhook 发送器实例
func NewWebhookSender(logger *zap.Logger) *WebhookSender {
	return &WebhookSender{logger: logger}
}

// Send 发送 Webhook 消息
func (s *WebhookSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	// TODO: 实现 Webhook 推送逻辑
	s.logger.Warn("Webhook 渠道尚未实现")
	return fmt.Errorf("渠道 webhook 尚未实现")
}
