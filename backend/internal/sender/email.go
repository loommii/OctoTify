package sender

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// EmailSender 邮件消息发送器
// 基于 Go 标准库 net/smtp 实现，支持端口 465（隐式 TLS）、587（STARTTLS）、25（明文）
type EmailSender struct {
	logger *zap.Logger
}

// NewEmailSender 创建 Email 发送器实例
func NewEmailSender(logger *zap.Logger) *EmailSender {
	return &EmailSender{logger: logger}
}

// emailConfig Email 渠道配置
type emailConfig struct {
	SMTPHost      string `json:"smtp_host"`       // SMTP 服务器地址
	SMTPPort      int    `json:"smtp_port"`       // SMTP 端口（465/587/25）
	Username      string `json:"username"`        // SMTP 登录用户名
	Password      string `json:"password"`        // SMTP 登录密码/授权码
	To            string `json:"to"`              // 收件人邮箱
	CC            string `json:"cc"`              // 抄送人邮箱（可选）
	FromName      string `json:"from_name"`       // 发件人名称（可选）
	SkipTLSVerify bool   `json:"skip_tls_verify"` // 跳过 TLS 证书校验（仅自签名证书时启用）
}

const emailDefaultTimeout = 30 * time.Second

// Send 发送邮件消息
func (s *EmailSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	// 1. 解析渠道配置
	var cfg emailConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("解析 Email 渠道配置失败: %w", err)
	}

	s.logger.Debug("Email 渠道配置",
		zap.String("smtp_host", cfg.SMTPHost),
		zap.Int("smtp_port", cfg.SMTPPort),
		zap.String("username", maskEmailUsername(cfg.Username)),
		zap.Bool("has_password", cfg.Password != ""),
		zap.String("to", cfg.To),
		zap.Bool("has_cc", cfg.CC != ""),
		zap.Bool("has_from_name", cfg.FromName != ""),
	)

	// 2. 校验必填字段
	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP 服务器地址不能为空")
	}
	if cfg.SMTPPort == 0 {
		return fmt.Errorf("SMTP 端口不能为空")
	}
	if cfg.Username == "" {
		return fmt.Errorf("SMTP 用户名不能为空")
	}
	if cfg.Password == "" {
		return fmt.Errorf("SMTP 密码不能为空")
	}
	if cfg.To == "" {
		return fmt.Errorf("收件人邮箱不能为空")
	}

	// 3. 构造邮件内容
	subject := encodeSubject(title)
	from := cfg.Username
	if cfg.FromName != "" {
		encodedName := mime.QEncoding.Encode("UTF-8", cfg.FromName)
		from = fmt.Sprintf("%s <%s>", encodedName, cfg.Username)
	}

	// 拼接收件人列表（to + cc）
	recipients := parseRecipients(cfg.To, cfg.CC)

	body := buildEmailBody(from, recipients, subject, content)

	s.logger.Debug("邮件头信息",
		zap.String("subject_encoded", subject),
		zap.Int("body_length", len(body)),
	)

	s.logger.Debug("收件人列表",
		zap.Strings("recipients", recipients),
	)

	// 4. 建立 SMTP 连接（支持 context 超时控制）
	client, err := dialSMTP(ctx, cfg.SMTPHost, cfg.SMTPPort, cfg.SkipTLSVerify)
	if err != nil {
		s.logger.Error("连接 SMTP 服务器失败",
			zap.String("addr", fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)),
			zap.Error(err),
		)
		return err
	}
	defer client.Quit()

	// 5. SMTP 认证（端口 25 跳过认证，因为 PlainAuth 拒绝在明文连接上发送凭证）
	if cfg.SMTPPort != 25 {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
		if err := client.Auth(auth); err != nil {
			s.logger.Error("SMTP 认证失败",
				zap.String("username", maskEmailUsername(cfg.Username)),
				zap.Error(err),
			)
			return fmt.Errorf("SMTP 认证失败: %w", err)
		}
	} else {
		s.logger.Warn("端口 25 为明文传输，跳过 SMTP 认证（存在安全风险）")
	}

	// 6. 设置发件人
	if err := client.Mail(cfg.Username); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 7. 设置收件人
	for _, addr := range recipients {
		if err := client.Rcpt(addr); err != nil {
			s.logger.Warn("收件人被服务器拒绝",
				zap.String("recipient", addr),
				zap.Error(err),
			)
			return fmt.Errorf("设置收件人失败 (%s): %w", addr, err)
		}
	}

	// 8. 写入邮件数据
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始写入邮件数据失败: %w", err)
	}

	_, err = w.Write([]byte(body))
	if err != nil {
		w.Close()
		return fmt.Errorf("写入邮件数据失败: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭邮件数据流失败: %w", err)
	}

	s.logger.Info("Email 发送成功")
	return nil
}

// dialSMTP 根据端口建立 SMTP 连接（支持 context 超时控制）
func dialSMTP(ctx context.Context, host string, port int, skipTLSVerify bool) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	if port == 465 {
		// 隐式 TLS（SMTPS）
		tlsConfig := &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
			MinVersion:         tls.VersionTLS12,
			ServerName:         host,
		}
		// 使用 DialContext 建立 TLS 连接（受 context 控制）
		dialer := &tls.Dialer{
			NetDialer: &net.Dialer{},
			Config:    tlsConfig,
		}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("建立 TLS 连接失败: %w", err)
		}
		// 设置连接 deadline（控制后续 SMTP 操作）
		applyConnDeadline(conn, ctx)
		// 创建 SMTP 客户端（失败时需关闭连接，避免资源泄漏）
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("创建 SMTP 客户端失败: %w", err)
		}
		return client, nil
	}

	// 端口 25/587：先建立明文连接（受 context 控制）
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("连接 SMTP 服务器失败: %w", err)
	}

	// 设置连接 deadline（控制后续 SMTP 操作）
	applyConnDeadline(conn, ctx)

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}

	// 端口 587：执行 STARTTLS
	if port == 587 {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
			MinVersion:         tls.VersionTLS12,
			ServerName:         host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return nil, fmt.Errorf("STARTTLS 升级失败: %w", err)
		}
		// StartTLS 内部已替换底层连接为 TLS 连接，无需再次设置 deadline
		// 原始连接设置的 deadline 已由 StartTLS 内部继承到新连接
	}

	return client, nil
}

// applyConnDeadline 根据 context 的 deadline 设置连接超时
func applyConnDeadline(conn net.Conn, ctx context.Context) {
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	} else {
		conn.SetDeadline(time.Now().Add(emailDefaultTimeout))
	}
}

// buildEmailBody 构造邮件原始内容（RFC 5322 格式）
func buildEmailBody(from string, recipients []string, subject string, content string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("From: %s\r\n", from))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(recipients[:1], ", ")))
	if len(recipients) > 1 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(recipients[1:], ", ")))
	}
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	builder.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	builder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	builder.WriteString("\r\n")
	builder.WriteString(content)

	return builder.String()
}

// encodeSubject 对邮件主题进行 RFC 2047 Base64 编码
func encodeSubject(subject string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(subject))
	return fmt.Sprintf("=?UTF-8?B?%s?=", encoded)
}

// parseRecipients 解析收件人列表（to + cc）
func parseRecipients(to string, cc string) []string {
	var recipients []string
	if to != "" {
		recipients = append(recipients, strings.TrimSpace(to))
	}
	if cc != "" {
		for _, addr := range strings.Split(cc, ",") {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				recipients = append(recipients, addr)
			}
		}
	}
	return recipients
}
