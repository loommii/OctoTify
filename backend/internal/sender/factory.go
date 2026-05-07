package sender

import (
	"fmt"
	"sync"

	"octotify/pkg/xerr"

	"go.uber.org/zap"
)

// SenderFactory 推送发送器工厂
// 根据渠道类型创建对应的 Sender 实例
type SenderFactory struct {
	mu      sync.RWMutex
	senders map[string]Sender
	logger  *zap.Logger
}

// NewSenderFactory 创建发送器工厂
func NewSenderFactory(logger *zap.Logger) *SenderFactory {
	return &SenderFactory{
		logger: logger,
		senders: map[string]Sender{
			"wechat":   NewWechatSender(logger),
			"telegram": NewTelegramSender(logger),
			"dingtalk": NewDingtalkSender(logger),
			"email":    NewEmailSender(logger),
			"webhook":  NewWebhookSender(logger),
			"feishu":   NewFeishuSender(logger),
		},
	}
}

// Create 根据渠道类型创建 Sender 实例
// 如果渠道类型无效，返回零值 Sender 和错误，调用方应先检查 error
func (f *SenderFactory) Create(channelType string) (Sender, error) {
	f.mu.RLock()
	snd, ok := f.senders[channelType]
	f.mu.RUnlock()

	if !ok {
		return nil, xerr.ErrChannelInvalidType.WithInternal(fmt.Errorf("不支持的渠道类型: %s", channelType))
	}
	return snd, nil
}

// Register 动态注册新的渠道发送器类型（支持运行时扩展第三方渠道）
func (f *SenderFactory) Register(channelType string, snd Sender) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.senders[channelType] = snd
}
