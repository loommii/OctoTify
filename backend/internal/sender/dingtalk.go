package sender

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// DingtalkSender 钉钉消息发送器
// 官方文档：https://open.dingtalk.com/document/orgapp/custom-robots-send-group-messages
type DingtalkSender struct {
	logger *zap.Logger
}

// NewDingtalkSender 创建钉钉发送器实例
func NewDingtalkSender(logger *zap.Logger) *DingtalkSender {
	return &DingtalkSender{logger: logger}
}

// dingtalkConfig 钉钉渠道配置
type dingtalkConfig struct {
	WebhookURL string `json:"webhook_url"` // 钉钉机器人 Webhook 地址
	Secret     string `json:"secret"`      // 加签密钥（必填）
}

// dingtalkMessage 钉钉消息体
type dingtalkMessage struct {
	MsgType string               `json:"msgtype"` // 消息类型：text
	Text    dingtalkTextContent  `json:"text"`    // 文本内容
}

// dingtalkTextContent 钉钉文本消息内容
type dingtalkTextContent struct {
	Content string `json:"content"` // 文本内容
}

// dingtalkResponse 钉钉 API 响应
type dingtalkResponse struct {
	ErrCode int    `json:"errcode"` // 错误码：0 表示成功
	ErrMsg  string `json:"errmsg"`  // 错误信息
}

// dingtalkHTTPClient 钉钉 HTTP 客户端（30 秒超时）
var dingtalkHTTPClient = &http.Client{Timeout: 30 * time.Second}

// Send 发送钉钉消息
func (s *DingtalkSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	// 解析渠道配置
	var cfg dingtalkConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("解析钉钉渠道配置失败: %w", err)
	}

	s.logger.Debug("钉钉渠道配置",
		zap.String("webhook_url", cfg.WebhookURL),
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
	msg := dingtalkMessage{
		MsgType: "text",
		Text: dingtalkTextContent{
			Content: fmt.Sprintf("[%s]\n%s", title, content),
		},
	}

	// 生成签名并附加到 URL
	timestamp := time.Now().UnixMilli()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, cfg.Secret)

	h := hmac.New(sha256.New, []byte(cfg.Secret))
	h.Write([]byte(stringToSign))
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	requestURL := fmt.Sprintf("%s&timestamp=%d&sign=%s", cfg.WebhookURL, timestamp, sign)

	s.logger.Debug("钉钉签名参数",
		zap.Int64("timestamp", timestamp),
		zap.String("sign", sign),
	)

	// 序列化消息体
	bodyBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("构造钉钉消息体失败: %w", err)
	}

	s.logger.Debug("钉钉请求体",
		zap.String("body", string(bodyBytes)),
	)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建钉钉 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := dingtalkHTTPClient.Do(req)
	if err != nil {
		s.logger.Error("钉钉网络请求失败",
			zap.String("url", requestURL),
			zap.Error(err),
		)
		return fmt.Errorf("发送钉钉请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取钉钉响应失败: %w", err)
	}

	s.logger.Debug("钉钉原始响应",
		zap.Int("http_status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("钉钉返回 HTTP 错误: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应 JSON
	var dingtalkResp dingtalkResponse
	if err := json.Unmarshal(respBody, &dingtalkResp); err != nil {
		s.logger.Error("解析钉钉响应失败",
			zap.String("raw_body", string(respBody)),
			zap.Error(err),
		)
		return fmt.Errorf("解析钉钉响应失败: %w", err)
	}

	s.logger.Debug("钉钉响应解析",
		zap.Int("errcode", dingtalkResp.ErrCode),
		zap.String("errmsg", dingtalkResp.ErrMsg),
	)

	// 检查业务状态码：成功时 errcode=0
	if dingtalkResp.ErrCode != 0 {
		s.logger.Warn("钉钉推送业务失败",
			zap.Int("errcode", dingtalkResp.ErrCode),
			zap.String("errmsg", dingtalkResp.ErrMsg),
		)
		return fmt.Errorf("钉钉推送失败: %s (errcode: %d)", dingtalkResp.ErrMsg, dingtalkResp.ErrCode)
	}

	s.logger.Info("钉钉推送成功")
	return nil
}
