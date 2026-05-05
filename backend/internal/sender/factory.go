package sender

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/datatypes"

	"octotify/pkg/xerr"
)

type Sender interface {
	Send(ctx context.Context, config datatypes.JSON, title string, content string) error
}

type SenderFactory struct {
	senders map[string]Sender
	logger  *zap.Logger
}

func NewSenderFactory(logger *zap.Logger) *SenderFactory {
	return &SenderFactory{
		logger: logger,
		senders: map[string]Sender{
			"wechat":   &WechatSender{},
			"telegram": &TelegramSender{},
			"dingtalk": &DingtalkSender{},
			"email":    &EmailSender{},
			"webhook":  &WebhookSender{},
			"feishu":   NewFeishuSender(logger),
		},
	}
}

func (f *SenderFactory) Create(channelType string) (Sender, error) {
	snd, ok := f.senders[channelType]
	if !ok {
		return nil, xerr.ErrChannelInvalidType.WithInternal(fmt.Errorf("不支持的渠道类型: %s", channelType))
	}
	return snd, nil
}

type WechatSender struct{}

func (s *WechatSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	return nil
}

type TelegramSender struct{}

func (s *TelegramSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	return nil
}

type DingtalkSender struct{}

func (s *DingtalkSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	return nil
}

type EmailSender struct{}

func (s *EmailSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	return nil
}

type WebhookSender struct{}

func (s *WebhookSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	return nil
}
