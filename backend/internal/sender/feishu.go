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
	"time"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// FeishuSender 飞书消息发送器
// 官方文档：https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot
type FeishuSender struct {
	logger *zap.Logger
}

// NewFeishuSender 创建飞书发送器实例
func NewFeishuSender(logger *zap.Logger) *FeishuSender {
	return &FeishuSender{logger: logger}
}

// feishuConfig 飞书渠道配置
type feishuConfig struct {
	WebhookURL string `json:"webhook_url"` // 飞书机器人 Webhook 地址
	Secret     string `json:"secret"`      // 签名密钥（可选）
}

// feishuMessage 飞书消息体
type feishuMessage struct {
	MsgType   string               `json:"msg_type"`            // 消息类型：text
	Content   feishuMessageContent `json:"content"`             // 消息内容
	Timestamp int64                `json:"timestamp,omitempty"` // 时间戳（签名校验时使用）
	Sign      string               `json:"sign,omitempty"`      // 签名（签名校验时使用）
}

// feishuMessageContent 飞书文本消息内容
type feishuMessageContent struct {
	Text string `json:"text"` // 文本内容
}

// feishuResponse 飞书 API 响应
type feishuResponse struct {
	Code int    `json:"code"` // 业务状态码：0 表示成功
	Msg  string `json:"msg"`  // 提示信息
	Raw  string // 原始响应体（用于调试）
}

// feishuHTTPClient 飞书 HTTP 客户端（30 秒超时）
var feishuHTTPClient = &http.Client{Timeout: 30 * time.Second}

// Send 发送飞书消息
func (s *FeishuSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	// 解析渠道配置
	var cfg feishuConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("解析飞书渠道配置失败: %w", err)
	}

	// 脱敏处理：Webhook URL 中包含认证凭据，仅记录 host
	s.logger.Debug("飞书渠道配置",
		zap.String("host", maskHost(cfg.WebhookURL)),
		zap.Bool("has_secret", cfg.Secret != ""),
	)

	// 校验 Webhook URL
	if cfg.WebhookURL == "" {
		return fmt.Errorf("飞书 Webhook URL 不能为空")
	}

	// 构造飞书消息体（文本类型）
	msg := feishuMessage{
		MsgType: "text",
		Content: feishuMessageContent{
			Text: fmt.Sprintf("[%s]\n%s", title, content),
		},
	}

	// 如果配置了签名密钥，生成签名并附加到请求体
	if cfg.Secret != "" {
		timestamp := time.Now().Unix()
		// 飞书签名计算方式（特殊）：
		// 1. 拼接 stringToSign = timestamp + "\n" + secret
		// 2. 使用 HMAC-SHA256 算法，以 stringToSign 作为 key 对空数据进行签名
		// 3. 将结果进行 Base64 编码
		// 注意：飞书的 HMAC 签名是将 timestamp+secret 作为 HMAC 的 key，
		// 而不是作为 data。这与常见的签名方式不同。
		stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + cfg.Secret

		var data []byte
		h := hmac.New(sha256.New, []byte(stringToSign))
		_, err := h.Write(data)
		if err != nil {
			return fmt.Errorf("计算飞书签名失败: %w", err)
		}

		sign := base64.StdEncoding.EncodeToString(h.Sum(nil))

		msg.Timestamp = timestamp
		msg.Sign = sign

		s.logger.Debug("飞书签名参数",
			zap.Int64("timestamp", timestamp),
			zap.String("sign", sign),
		)
	}

	bodyBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("构造飞书消息体失败: %w", err)
	}

	s.logger.Debug("飞书请求体",
		zap.String("body", string(bodyBytes)),
	)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.WebhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := feishuHTTPClient.Do(req)
	if err != nil {
		s.logger.Error("飞书网络请求失败",
			zap.String("host", maskHost(cfg.WebhookURL)),
			zap.Error(err),
		)
		return fmt.Errorf("发送飞书请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取飞书响应失败: %w", err)
	}

	s.logger.Debug("飞书原始响应",
		zap.Int("http_status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("飞书返回 HTTP 错误: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应 JSON
	var feishuResp feishuResponse
	if err := json.Unmarshal(respBody, &feishuResp); err != nil {
		s.logger.Error("解析飞书响应失败",
			zap.String("raw_body", string(respBody)),
			zap.Error(err),
		)
		return fmt.Errorf("解析飞书响应失败: %w", err)
	}

	s.logger.Debug("飞书响应解析",
		zap.Int("status_code", feishuResp.Code),
		zap.String("msg", feishuResp.Msg),
	)

	// 检查业务状态码：成功时 code=0 且 msg="success"
	if feishuResp.Code != 0 || feishuResp.Msg != "success" {
		s.logger.Warn("飞书推送业务失败",
			zap.Int("status_code", feishuResp.Code),
			zap.String("msg", feishuResp.Msg),
		)
		return fmt.Errorf("飞书推送失败: %s", feishuResp.Msg)
	}

	s.logger.Info("飞书推送成功")
	return nil
}
