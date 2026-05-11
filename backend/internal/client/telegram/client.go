package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"resty.dev/v3"
)

const (
	DefaultBaseURL = "https://api.telegram.org/bot"
	defaultTimeout = 60 * time.Second
)

// Client Telegram Bot API 客户端
type Client struct {
	restyClient *resty.Client
	baseURL     string
}

// Option 客户端配置选项
type Option func(*Client)

// WithBaseURL 设置自定义基础 URL（用于测试）
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithProxy 设置 HTTP 代理
func WithProxy(proxy string) Option {
	return func(c *Client) {
		c.restyClient.SetProxy(proxy)
	}
}

// NewClient 创建 Telegram API 客户端
func NewClient(opts ...Option) *Client {
	restyClient := resty.New().
		SetTimeout(defaultTimeout)

	c := &Client{
		restyClient: restyClient,
		baseURL:     DefaultBaseURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// MessageRequest Telegram sendMessage API 请求体
type MessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// Response Telegram API 响应体
type Response struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

// SendMessage 发送 Telegram 消息
func (c *Client) SendMessage(ctx context.Context, botToken string, msg MessageRequest) (*Response, error) {
	apiURL := c.baseURL + botToken + "/sendMessage"

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(msg).
		Post(apiURL)

	if err != nil {
		return nil, fmt.Errorf("发送 Telegram 请求失败: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("Telegram 返回 HTTP 错误: %d, body: %s", resp.StatusCode(), string(resp.Bytes()))
	}

	var result Response
	if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("解析 Telegram 响应失败: %w", err)
	}

	return &result, nil
}
