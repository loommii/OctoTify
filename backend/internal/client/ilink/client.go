package ilink

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL = "https://ilinkai.weixin.qq.com/ilink"
	defaultTimeout = 30 * time.Second
)

// Client iLink 平台 API 客户端
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// Option 客户端配置选项
type Option func(*Client)

// WithHTTPClient 设置自定义 HTTP 客户端
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithBaseURL 设置自定义基础 URL（用于测试）
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// NewClient 创建 iLink API 客户端
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    defaultBaseURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GetQRCode 调用 iLink API 获取二维码
func (c *Client) GetQRCode(ctx context.Context) (*QRCodeResponse, error) {
	const path = "/bot/get_bot_qrcode?bot_type=3"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result QRCodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// GetQRCodeStatus 调用 iLink API 查询二维码状态
// iLink 的 get_qrcode_status 接口本身就是长轮询设计，需要 35-40 秒才返回
func (c *Client) GetQRCodeStatus(ctx context.Context, qrcode string) (*QRStatusResponse, error) {
	pollCtx, cancel := context.WithTimeout(ctx, PollAPITimeout)
	defer cancel()

	path := fmt.Sprintf("/bot/get_qrcode_status?qrcode=%s", url.QueryEscape(qrcode))

	req, err := http.NewRequestWithContext(pollCtx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result QRStatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}
