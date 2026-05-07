package sender

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// DingtalkSender 钉钉消息发送器
type DingtalkSender struct {
	logger *zap.Logger
}

// NewDingtalkSender 创建钉钉发送器实例
func NewDingtalkSender(logger *zap.Logger) *DingtalkSender {
	return &DingtalkSender{logger: logger}
}

// Send 发送钉钉消息
func (s *DingtalkSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	// TODO: 实现钉钉推送逻辑
	s.logger.Warn("Dingtalk 渠道尚未实现")
	return fmt.Errorf("渠道 dingtalk 尚未实现")
}
