package ilink

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/url"
	"octotify/pkg/ctxutil"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"
	"resty.dev/v3"
)

const (
	defaultBaseURL = "https://ilinkai.weixin.qq.com/ilink" // iLink 平台 API 基础地址
	defaultTimeout = 60 * time.Second                       // 默认 HTTP 请求超时时间
)

// Client iLink 平台 API 客户端
// 统一管理所有 iLink 协议的 HTTP 请求，包括认证、消息发送等
// 设计为全局共享实例，多 bot 场景下通过动态凭证区分身份
type Client struct {
	restyClient *resty.Client // resty HTTP 客户端，自动管理连接池
	logger      *zap.Logger   // 日志记录器
	wechatUIN   string        // 微信用户标识号，代表整个 OctoTify 服务实例
	uinOnce     sync.Once     // 保证 UIN 只生成一次，线程安全
}

// Option 客户端配置选项
// 使用函数选项模式，支持灵活配置客户端行为
type Option func(*Client)

// WithBaseURL 设置自定义基础 URL
// 主要用于单元测试，通过 httptest.Server 注入 mock 地址
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.restyClient.SetBaseURL(baseURL)
	}
}

// NewClient 创建 iLink API 客户端
// 使用 resty 作为底层 HTTP 客户端，自动管理连接池和超时
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

// generateWechatUIN 生成随机 UIN（User Identification Number）
// UIN 是 iLink 协议要求的客户端身份标识，用于防伪造和会话追踪
// 生成方式：32 位随机整数 → 转字符串 → Base64 编码
func (c *Client) generateWechatUIN() string {
	var n uint32
	_ = binary.Read(rand.Reader, binary.LittleEndian, &n)
	s := fmt.Sprintf("%d", n)
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// getWechatUIN 获取 UIN（线程安全，进程内只生成一次）
// UIN 代表整个 OctoTify 服务实例，所有 bot 共享同一个 UIN
func (c *Client) getWechatUIN() string {
	c.uinOnce.Do(func() {
		c.wechatUIN = c.generateWechatUIN()
	})
	return c.wechatUIN
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

// SendMessage 调用 iLink API 发送消息（支持动态凭证）
// 适用于多 bot 场景，每次发送时动态指定 botToken
// 对应 iLink API: POST /ilink/bot/sendmessage
// 必需请求头:
//   - Authorization: Bearer {botToken}
//   - AuthorizationType: ilink_bot_token
//   - X-WECHAT-UIN: 客户端身份标识
func (c *Client) SendMessage(ctx context.Context, req *SendMessageRequest, botToken string) (*SendMessageResponse, error) {
	const path = "/bot/sendmessage"

	// 设置请求超时（15 秒，与 weclaw 保持一致）
	sendCtx, cancel := context.WithTimeout(ctx, SendAPITimeout)
	defer cancel()

	logger := ctxutil.LoggerWithRequestID(ctx, c.logger)
	if logger != nil {
		logger.Debug("iLink SendMessage 请求",
			zap.String("method", "POST"),
			zap.String("path", path),
			zap.String("from_user_id", req.Msg.FromUserID),
			zap.String("to_user_id", req.Msg.ToUserID),
		)
	}

	bodyBytes, err := sonic.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 发送 POST 请求，设置 iLink 协议必需的请求头
	resp, err := c.restyClient.R().
		SetContext(sendCtx).
		SetHeader("Content-Type", "application/json").
		SetHeader("AuthorizationType", "ilink_bot_token").
		SetHeader("Authorization", "Bearer "+botToken).
		SetHeader("X-WECHAT-UIN", c.getWechatUIN()).
		SetBody(bodyBytes).
		Post(path)

	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(resp.Bytes()))
	}

	if logger != nil {
		logger.Debug("iLink SendMessage 响应",
			zap.Int("status_code", resp.StatusCode()),
			zap.String("raw_response", string(resp.Bytes())),
		)
	}

	var result SendMessageResponse
	if err := sonic.Unmarshal(resp.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if logger != nil {
		logger.Debug("iLink SendMessage 解析结果",
			zap.Int("ret", result.Ret),
			zap.String("errmsg", result.ErrMsg),
		)
	}

	// 检查业务状态码，ret != 0 表示 iLink 服务端处理失败
	if result.Ret != 0 {
		return &result, fmt.Errorf("iLink 推送失败: %s (ret: %d)", result.ErrMsg, result.Ret)
	}

	return &result, nil
}
