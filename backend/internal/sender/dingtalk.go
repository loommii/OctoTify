package sender

import (
	"context"
	"encoding/json"
	"fmt"

	"octotify/internal/client/dingtalk"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// DingtalkSender 钉钉消息发送器
// 官方文档：https://open.dingtalk.com/document/orgapp/custom-robots-send-group-messages
type DingtalkSender struct {
	logger *zap.Logger
	client *dingtalk.Client
}

// NewDingtalkSender 创建钉钉发送器实例
func NewDingtalkSender(logger *zap.Logger) *DingtalkSender {
	return &DingtalkSender{
		logger: logger,
		client: dingtalk.NewClient(),
	}
}

// dingtalkConfig 钉钉渠道配置
type dingtalkConfig struct {
	WebhookURL string `json:"webhook_url"` // 钉钉机器人 Webhook 地址
	Secret     string `json:"secret"`      // 加签密钥（必填）
}

// Send 发送钉钉消息
func (s *DingtalkSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	// 解析渠道配置
	var cfg dingtalkConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("解析钉钉渠道配置失败: %w", err)
	}

	s.logger.Debug("钉钉渠道配置",
		zap.String("webhook_url", maskURL(cfg.WebhookURL)),
		zap.Bool("has_secret", cfg.Secret != ""),
	)

	// 校验 Webhook URL
	if cfg.WebhookURL == "" {
		return fmt.Errorf("钉钉 Webhook URL 不能为空")
	}

	// 校验加签密钥（必填）
	if cfg.Secret == "" {
		return fmt.Errorf("钉钉加签密钥不能为空")
	}

	// 构造钉钉消息体（文本类型）
	msg := dingtalk.MessageRequest{
		MsgType: "text",
		Text: dingtalk.TextContent{
			Content: fmt.Sprintf("[%s]\n%s", title, content),
		},
	}

	s.logger.Debug("钉钉请求体",
		zap.String("msg_type", msg.MsgType),
		zap.String("content", msg.Text.Content),
	)

	// 发送请求
	resp, err := s.client.SendMessage(ctx, cfg.WebhookURL, cfg.Secret, msg)
	if err != nil {
		s.logger.Error("钉钉网络请求失败",
			zap.String("webhook_url", maskURL(cfg.WebhookURL)),
			zap.Error(err),
		)
		return fmt.Errorf("发送钉钉请求失败: %w", err)
	}

	s.logger.Debug("钉钉响应解析",
		zap.Int("errcode", resp.ErrCode),
		zap.String("errmsg", resp.ErrMsg),
	)

	// 检查业务状态码：成功时 errcode=0
	if resp.ErrCode != 0 {
		s.logger.Warn("钉钉推送业务失败",
			zap.Int("errcode", resp.ErrCode),
			zap.String("errmsg", resp.ErrMsg),
		)
		return fmt.Errorf("钉钉推送失败: %s (errcode: %d)", resp.ErrMsg, resp.ErrCode)
	}

	s.logger.Info("钉钉推送成功")
	return nil
}
