package dingtalk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"resty.dev/v3"
)

const (
	defaultTimeout = 60 * time.Second
)

// Client 钉钉 Webhook API 客户端
type Client struct {
	restyClient *resty.Client
}

// Option 客户端配置选项
type Option func(*Client)

// NewClient 创建钉钉 API 客户端
func NewClient(opts ...Option) *Client {
	restyClient := resty.New().
		SetTimeout(defaultTimeout)

	c := &Client{
		restyClient: restyClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// MessageRequest 钉钉消息请求
type MessageRequest struct {
	MsgType string      `json:"msgtype"`
	Text    TextContent `json:"text"`
}

// TextContent 钉钉文本消息内容
type TextContent struct {
	Content string `json:"content"`
}

// Response 钉钉 API 响应
type Response struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// SendMessage 发送钉钉消息
func (c *Client) SendMessage(ctx context.Context, webhookURL, secret string, msg MessageRequest) (*Response, error) {
	requestURL := buildSignedURL(webhookURL, secret)

	resp, err := c.restyClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(msg).
		Post(requestURL)

	if err != nil {
		return nil, fmt.Errorf("钉钉请求失败: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("钉钉返回 HTTP 错误: %d, body: %s", resp.StatusCode(), string(resp.Bytes()))
	}

	var result Response
	if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("解析钉钉响应失败: %w", err)
	}

	return &result, nil
}

// buildSignedURL 生成带签名的钉钉 Webhook URL
func buildSignedURL(webhookURL, secret string) string {
	timestamp := time.Now().UnixMilli()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	return fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, sign)
}
