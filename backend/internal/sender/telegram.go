package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"octotify/internal/client/telegram"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// TelegramSender Telegram 消息发送器
// 官方文档：https://core.telegram.org/bots/api#sendmessage
type TelegramSender struct {
	logger  *zap.Logger
	baseURL string
}

// NewTelegramSender 创建 Telegram 发送器实例
func NewTelegramSender(logger *zap.Logger) *TelegramSender {
	return &TelegramSender{
		logger:  logger,
		baseURL: telegram.DefaultBaseURL,
	}
}

// telegramConfig Telegram 渠道配置
type telegramConfig struct {
	BotToken string `json:"bot_token"` // Bot Token
	ChatID   string `json:"chat_id"`   // Chat ID
	Proxy    string `json:"proxy"`     // HTTP 代理地址（可选）
}

const (
	telegramMaxMessageLen  = 4096
	telegramTruncateSuffix = "\n\n[消息已截断]"
)

// Send 发送 Telegram 消息
func (s *TelegramSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	// 1. 解析渠道配置
	var cfg telegramConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("解析 Telegram 渠道配置失败: %w", err)
	}

	s.logger.Debug("Telegram 渠道配置",
		zap.String("bot_token", maskToken(cfg.BotToken)),
		zap.String("chat_id", cfg.ChatID),
		zap.Bool("has_proxy", cfg.Proxy != ""),
	)

	// 2. 校验必填字段
	if cfg.BotToken == "" {
		return fmt.Errorf("Telegram Bot Token 不能为空")
	}
	if cfg.ChatID == "" {
		return fmt.Errorf("Telegram Chat ID 不能为空")
	}

	// 3. 构造消息文本
	text := fmt.Sprintf("<b>[%s]</b>\n%s", escapeHTML(title), escapeHTML(content))
	text = truncateMessage(text, telegramMaxMessageLen, telegramTruncateSuffix)

	// 4. 构造请求体
	msg := telegram.MessageRequest{
		ChatID:    cfg.ChatID,
		Text:      text,
		ParseMode: "HTML",
	}

	s.logger.Debug("Telegram 请求体",
		zap.String("chat_id", msg.ChatID),
		zap.String("text", msg.Text),
		zap.String("parse_mode", msg.ParseMode),
	)

	// 5. 创建客户端（根据代理配置）
	opts := []telegram.Option{telegram.WithBaseURL(s.baseURL)}
	if cfg.Proxy != "" {
		opts = append(opts, telegram.WithProxy(cfg.Proxy))
	}
	client := telegram.NewClient(opts...)

	// 6. 发送请求
	resp, err := client.SendMessage(ctx, cfg.BotToken, msg)
	if err != nil {
		s.logger.Error("Telegram 网络请求失败",
			zap.String("bot_token", maskToken(cfg.BotToken)),
			zap.Bool("use_proxy", cfg.Proxy != ""),
			zap.Error(err),
		)
		return fmt.Errorf("发送 Telegram 请求失败: %w", err)
	}

	s.logger.Debug("Telegram 响应解析",
		zap.Bool("ok", resp.Ok),
		zap.Int("error_code", resp.ErrorCode),
		zap.String("description", resp.Description),
	)

	// 7. 检查业务状态
	if !resp.Ok {
		s.logger.Warn("Telegram 推送业务失败",
			zap.Int("error_code", resp.ErrorCode),
			zap.String("description", resp.Description),
		)
		return fmt.Errorf("Telegram 推送失败: %s (error_code: %d)", resp.Description, resp.ErrorCode)
	}

	s.logger.Info("Telegram 推送成功")
	return nil
}

// truncateMessage 截断超长消息
// 注意：Telegram 的 4096 限制是字符数（rune count），不是字节数
func truncateMessage(text string, maxLen int, suffix string) string {
	charCount := utf8.RuneCountInString(text)
	if charCount <= maxLen {
		return text
	}
	suffixCharCount := utf8.RuneCountInString(suffix)
	availableLen := maxLen - suffixCharCount
	if availableLen <= 0 {
		availableLen = maxLen
	}
	byteIndex := 0
	charIndex := 0
	for _, r := range text {
		if charIndex >= availableLen {
			break
		}
		byteIndex += utf8.RuneLen(r)
		charIndex++
	}
	return text[:byteIndex] + suffix
}

// escapeHTML 转义 HTML 特殊字符，防止 Telegram 解析错误
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
