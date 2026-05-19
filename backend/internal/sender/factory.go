package sender

import (
	"fmt"
	"sync"

	"octotify/internal/client/ilink"
	"octotify/pkg/xerr"

	"go.uber.org/zap"
)

// SenderFactory 推送发送器工厂
// 根据渠道类型创建对应的 Sender 实例
// 支持内置渠道（wechat、telegram、dingtalk 等）和运行时动态注册
type SenderFactory struct {
	mu      sync.RWMutex      // 读写锁，保护 senders map 的并发安全
	senders map[string]Sender // 渠道类型 → 发送器实例映射
	logger  *zap.Logger       // 日志记录器
}

// NewSenderFactory 创建发送器工厂
// 初始化所有内置渠道的发送器实例
// 参数:
//   - logger: 日志记录器
//   - ilinkClient: iLink API 客户端（用于微信ClawBot渠道）
func NewSenderFactory(logger *zap.Logger, ilinkClient *ilink.Client) *SenderFactory {
	return &SenderFactory{
		logger: logger,
		senders: map[string]Sender{
			"wechat":         NewWechatSender(logger),
			"wechat_clawbot": NewWechatClawbotSender(logger, ilinkClient),
			"telegram":       NewTelegramSender(logger),
			"dingtalk":       NewDingtalkSender(logger),
			"email":          NewEmailSender(logger),
			"webhook":        NewWebhookSender(logger),
			"feishu":         NewFeishuSender(logger),
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
