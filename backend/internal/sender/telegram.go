package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// TelegramSender Telegram 消息发送器
// 官方文档：https://core.telegram.org/bots/api#sendmessage
type TelegramSender struct {
	logger *zap.Logger
}

// NewTelegramSender 创建 Telegram 发送器实例
func NewTelegramSender(logger *zap.Logger) *TelegramSender {
	return &TelegramSender{logger: logger}
}

// telegramConfig Telegram 渠道配置
type telegramConfig struct {
	BotToken string `json:"bot_token"` // Bot Token
	ChatID   string `json:"chat_id"`   // Chat ID
	Proxy    string `json:"proxy"`     // HTTP 代理地址（可选）
}

// telegramMessage Telegram sendMessage API 请求体
type telegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// telegramResponse Telegram API 响应体
type telegramResponse struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

var telegramAPIBaseURL = "https://api.telegram.org/bot"

const (
	telegramHTTPTimeout      = 30 * time.Second
	telegramMaxMessageLen    = 4096
	telegramTruncateSuffix   = "\n\n[消息已截断]"
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
	msg := telegramMessage{
		ChatID:    cfg.ChatID,
		Text:      text,
		ParseMode: "HTML",
	}

	bodyBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("构造 Telegram 消息体失败: %w", err)
	}

	s.logger.Debug("Telegram 请求体",
		zap.String("body", string(bodyBytes)),
	)

	// 5. 创建 HTTP 客户端（根据代理配置）
	client, err := newHTTPClient(cfg.Proxy)
	if err != nil {
		return fmt.Errorf("创建 HTTP 客户端失败: %w", err)
	}

	// 6. 创建 HTTP 请求
	apiURL := telegramAPIBaseURL + cfg.BotToken + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 7. 发送请求
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Telegram 网络请求失败",
			zap.String("url", telegramAPIBaseURL+maskToken(cfg.BotToken)+"/sendMessage"),
			zap.Bool("use_proxy", cfg.Proxy != ""),
			zap.Error(err),
		)
		return fmt.Errorf("发送 Telegram 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 8. 读取响应体
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("读取 Telegram 响应失败: %w", err)
	}

	s.logger.Debug("Telegram 原始响应",
		zap.Int("http_status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	// 9. 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram 返回 HTTP 错误: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 10. 解析响应 JSON
	var tgResp telegramResponse
	if err := json.Unmarshal(respBody, &tgResp); err != nil {
		s.logger.Error("解析 Telegram 响应失败",
			zap.String("raw_body", string(respBody)),
			zap.Error(err),
		)
		return fmt.Errorf("解析 Telegram 响应失败: %w", err)
	}

	s.logger.Debug("Telegram 响应解析",
		zap.Bool("ok", tgResp.Ok),
		zap.Int("error_code", tgResp.ErrorCode),
		zap.String("description", tgResp.Description),
	)

	// 11. 检查业务状态
	if !tgResp.Ok {
		s.logger.Warn("Telegram 推送业务失败",
			zap.Int("error_code", tgResp.ErrorCode),
			zap.String("description", tgResp.Description),
		)
		return fmt.Errorf("Telegram 推送失败: %s (error_code: %d)", tgResp.Description, tgResp.ErrorCode)
	}

	s.logger.Info("Telegram 推送成功")
	return nil
}

// newHTTPClient 根据代理配置创建 HTTP 客户端
func newHTTPClient(proxy string) (*http.Client, error) {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("解析代理地址失败: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &http.Client{
		Timeout:   telegramHTTPTimeout,
		Transport: transport,
	}, nil
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
	// 将字符数转换为字节索引，确保不截断多字节字符
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
