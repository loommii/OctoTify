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

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type GotifySender struct {
	logger *zap.Logger
}

func NewGotifySender(logger *zap.Logger) *GotifySender {
	return &GotifySender{logger: logger}
}

type gotifyConfig struct {
	ServerURL string `json:"server_url"`
	AppToken  string `json:"app_token"`
	Priority  int    `json:"priority"`
}

type gotifyMessage struct {
	Title    string       `json:"title"`
	Message  string       `json:"message"`
	Priority int          `json:"priority"`
	Extras   gotifyExtras `json:"extras"`
}

type gotifyExtras struct {
	ClientDisplay gotifyClientDisplay `json:"client::display"`
}

type gotifyClientDisplay struct {
	ContentType string `json:"contentType"`
}

type gotifyResponse struct {
	ID int `json:"id"`
}

type gotifyErrorResponse struct {
	Error            string `json:"error"`
	ErrorCode        int    `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
}

const (
	gotifyHTTPTimeout    = 30 * time.Second
	gotifyMaxMessageLen  = 4096
	gotifyTruncateSuffix = "\n\n[消息已截断]"
)

func (s *GotifySender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	var cfg gotifyConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("解析 Gotify 渠道配置失败: %w", err)
	}

	if cfg.ServerURL == "" {
		return fmt.Errorf("Gotify 服务器地址不能为空")
	}
	if cfg.AppToken == "" {
		return fmt.Errorf("Gotify App Token 不能为空")
	}

	serverURL := strings.TrimSuffix(cfg.ServerURL, "/")
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "https://" + serverURL
	}

	parsedURL, err := url.Parse(serverURL)
	if err != nil || parsedURL.Host == "" {
		return fmt.Errorf("Gotify 服务器地址格式错误")
	}

	priority := cfg.Priority
	if priority < 0 || priority > 10 {
		priority = 5
	}

	s.logger.Debug("Gotify 渠道配置",
		zap.String("server_url", serverURL),
		zap.String("app_token", maskToken(cfg.AppToken)),
		zap.Int("priority", priority),
	)

	escapedContent := escapeMarkdown(content)
	text := fmt.Sprintf("**[%s]**\n\n%s", escapeMarkdown(title), escapedContent)
	text = truncateMessage(text, gotifyMaxMessageLen, gotifyTruncateSuffix)

	msg := gotifyMessage{
		Title:    title,
		Message:  text,
		Priority: priority,
		Extras: gotifyExtras{
			ClientDisplay: gotifyClientDisplay{
				ContentType: "text/markdown",
			},
		},
	}

	bodyBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("构造 Gotify 消息体失败: %w", err)
	}

	s.logger.Debug("Gotify 请求体",
		zap.String("body", string(bodyBytes)),
	)

	client := &http.Client{
		Timeout: gotifyHTTPTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/") + "/message"
	apiURL := parsedURL.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", cfg.AppToken)

	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Gotify 网络请求失败",
			zap.String("host", parsedURL.Host),
			zap.Error(err),
		)
		return fmt.Errorf("发送 Gotify 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("读取 Gotify 响应失败: %w", err)
	}

	s.logger.Debug("Gotify 原始响应",
		zap.Int("http_status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	if resp.StatusCode != http.StatusOK {
		var errResp gotifyErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return fmt.Errorf("Gotify 返回 HTTP 错误: %d, error: %s, description: %s", resp.StatusCode, errResp.Error, errResp.ErrorDescription)
		}
		return fmt.Errorf("Gotify 返回 HTTP 错误: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var successResp gotifyResponse
	if err := json.Unmarshal(respBody, &successResp); err != nil {
		s.logger.Error("解析 Gotify 响应失败",
			zap.String("raw_body", string(respBody)),
			zap.Error(err),
		)
		return fmt.Errorf("解析 Gotify 响应失败: %w", err)
	}

	s.logger.Debug("Gotify 响应解析",
		zap.Int("message_id", successResp.ID),
	)

	s.logger.Info("Gotify 推送成功")
	return nil
}

// escapeMarkdown 转义 Markdown 行内格式字符，防止用户内容破坏 Gotify 的 Markdown 渲染。
// 仅转义会影响行内格式的字符（加粗、斜体、链接、代码等），
// 保留 #（标题）和 -（列表）等块级语法，因为 Gotify 使用 text/markdown 渲染消息。
func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}
