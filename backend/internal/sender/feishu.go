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

	"gorm.io/datatypes"
)

type FeishuSender struct{}

type feishuConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
}

type feishuMessage struct {
	MsgType string               `json:"msg_type"`
	Content feishuMessageContent `json:"content"`
}

type feishuMessageContent struct {
	Text string `json:"text"`
}

type feishuResponse struct {
	Code int    `json:"StatusCode"`
	Msg  string `json:"msg"`
}

var feishuHTTPClient = &http.Client{Timeout: 30 * time.Second}

func (s *FeishuSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	var cfg feishuConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("解析飞书渠道配置失败: %w", err)
	}

	if cfg.WebhookURL == "" {
		return fmt.Errorf("飞书 Webhook URL 不能为空")
	}

	msg := feishuMessage{
		MsgType: "text",
		Content: feishuMessageContent{
			Text: fmt.Sprintf("[%s]\n%s", title, content),
		},
	}

	bodyBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("构造飞书消息体失败: %w", err)
	}

	webhookURL := cfg.WebhookURL
	if cfg.Secret != "" {
		timestamp := time.Now().Unix()
		stringToSign := fmt.Sprintf("%d\n%s", timestamp, cfg.Secret)

		mac := hmac.New(sha256.New, []byte(cfg.Secret))
		mac.Write([]byte(stringToSign))
		sign := base64.URLEncoding.EncodeToString(mac.Sum(nil))

		webhookURL = fmt.Sprintf("%s?timestamp=%d&sign=%s", cfg.WebhookURL, timestamp, sign)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := feishuHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送飞书请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取飞书响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("飞书返回 HTTP 错误: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var feishuResp feishuResponse
	if err := json.Unmarshal(respBody, &feishuResp); err != nil {
		return fmt.Errorf("解析飞书响应失败: %w", err)
	}

	if feishuResp.Code != 0 {
		return fmt.Errorf("飞书推送失败: %s", feishuResp.Msg)
	}

	return nil
}
