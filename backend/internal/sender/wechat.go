package sender

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// WechatSender 微信消息发送器
type WechatSender struct {
	logger *zap.Logger
}

// NewWechatSender 创建微信发送器实例
func NewWechatSender(logger *zap.Logger) *WechatSender {
	return &WechatSender{logger: logger}
}

// Send 发送微信消息
func (s *WechatSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	// TODO: 实现微信推送逻辑
	s.logger.Warn("Wechat 渠道尚未实现")
	return fmt.Errorf("渠道 wechat 尚未实现")
}
