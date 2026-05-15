package ilink

import (
	"context"
	"fmt"
	"net/url"
	"octotify/pkg/ctxutil"
	"time"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"
	"resty.dev/v3"
)

const (
	defaultBaseURL = "https://ilinkai.weixin.qq.com/ilink"
	defaultTimeout = 60 * time.Second
)

// Client iLink 平台 API 客户端
type Client struct {
	restyClient *resty.Client
	logger      *zap.Logger
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
func NewClient(logger *zap.Logger, opts ...Option) *Client {
	restyClient := resty.New().
		SetTimeout(defaultTimeout).
		SetBaseURL(defaultBaseURL)

	c := &Client{
		restyClient: restyClient,
		logger:      logger.Named("ilink_client"),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GetQRCode 调用 iLink API 获取二维码
func (c *Client) GetQRCode(ctx context.Context) (*QRCodeResponse, error) {
	const path = "/bot/get_bot_qrcode"

	logger := ctxutil.LoggerWithRequestID(ctx, c.logger)
	if logger != nil {
		logger.Debug("iLink GetQRCode 请求",
			zap.String("method", "GET"),
			zap.String("path", path),
			zap.String("bot_type", "3"),
		)
	}

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetQueryParam("bot_type", "3").
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(resp.Bytes()))
	}

	if logger != nil {
		logger.Debug("iLink GetQRCode 响应",
			zap.Int("status_code", resp.StatusCode()),
			zap.String("raw_response", string(resp.Bytes())),
		)
	}

	var result QRCodeResponse
	if err := sonic.Unmarshal(resp.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if logger != nil {
		logger.Debug("iLink GetQRCode 解析结果",
			zap.String("qrcode", result.QRCode),
			zap.String("qrcode_img_content", result.QRCodeImgContent),
			zap.Int("ret", result.RetCode),
		)
	}

	return &result, nil
}

// GetQRCodeStatus 调用 iLink API 查询二维码状态
// iLink 的 get_qrcode_status 接口本身就是长轮询设计，需要 35-40 秒才返回
func (c *Client) GetQRCodeStatus(ctx context.Context, qrcode string) (*QRStatusResponse, error) {
	pollCtx, cancel := context.WithTimeout(ctx, PollAPITimeout)
	defer cancel()

	const path = "/bot/get_qrcode_status"

	logger := ctxutil.LoggerWithRequestID(ctx, c.logger)
	if logger != nil {
		logger.Debug("iLink GetQRCodeStatus 请求",
			zap.String("method", "GET"),
			zap.String("path", path),
			zap.String("qrcode", qrcode),
		)
	}

	resp, err := c.restyClient.R().
		SetContext(pollCtx).
		SetQueryParam("qrcode", url.QueryEscape(qrcode)).
		Get(path)

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(resp.Bytes()))
	}

	var result QRStatusResponse
	if err := sonic.Unmarshal(resp.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if logger != nil {
		logger.Debug("iLink GetQRCodeStatus 响应",
			zap.Int("status_code", resp.StatusCode()),
			zap.String("raw_response", string(resp.Bytes())),
			zap.String("status", result.Status),
			zap.String("bot_token", result.BotToken),
			zap.String("ilink_bot_id", result.ILinkBotID),
			zap.String("ilink_user_id", result.ILinkUserID),
		)
	}

	return &result, nil
}
