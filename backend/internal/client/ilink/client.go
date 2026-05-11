package ilink

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"resty.dev/v3"
)

const (
	defaultBaseURL = "https://ilinkai.weixin.qq.com/ilink"
	defaultTimeout = 60 * time.Second
)

// Client iLink 平台 API 客户端
type Client struct {
	restyClient *resty.Client
}

// Option 客户端配置选项
type Option func(*Client)

// WithBaseURL 设置自定义基础 URL（用于测试）
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.restyClient.SetBaseURL(baseURL)
	}
}

// NewClient 创建 iLink API 客户端
func NewClient(opts ...Option) *Client {
	restyClient := resty.New().
		SetTimeout(defaultTimeout).
		SetBaseURL(defaultBaseURL)

	c := &Client{
		restyClient: restyClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GetQRCode 调用 iLink API 获取二维码
func (c *Client) GetQRCode(ctx context.Context) (*QRCodeResponse, error) {
	const path = "/bot/get_bot_qrcode"

	var result QRCodeResponse
	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetQueryParam("bot_type", "3").
		SetResult(&result).
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(resp.Bytes()))
	}

	return &result, nil
}

// GetQRCodeStatus 调用 iLink API 查询二维码状态
// iLink 的 get_qrcode_status 接口本身就是长轮询设计，需要 35-40 秒才返回
func (c *Client) GetQRCodeStatus(ctx context.Context, qrcode string) (*QRStatusResponse, error) {
	pollCtx, cancel := context.WithTimeout(ctx, PollAPITimeout)
	defer cancel()

	var result QRStatusResponse
	resp, err := c.restyClient.R().
		SetContext(pollCtx).
		SetQueryParam("qrcode", url.QueryEscape(qrcode)).
		SetResult(&result).
		Get("/bot/get_qrcode_status")

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(resp.Bytes()))
	}

	return &result, nil
}
