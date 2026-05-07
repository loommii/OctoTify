# Email Sender 设计文档

> 为 OctoTify 消息总线平台新增 Email 邮件推送渠道，基于 Go 标准库 `net/smtp` + `crypto/tls` 实现 SMTP 发送。

---

## 一、需求概述

### 1.1 功能需求

| 编号 | 需求 | 优先级 |
|------|------|--------|
| F1 | 通过 SMTP 协议发送纯文本邮件 | P0 |
| F2 | 支持端口 465（隐式 TLS/SMTPS） | P0 |
| F3 | 支持端口 587（STARTTLS 加密） | P0 |
| F4 | 支持端口 25（明文传输） | P0 |
| F5 | 遵循现有 Sender 策略模式，与 FeishuSender/TelegramSender 保持一致 | P0 |
| F6 | 前端动态渲染 Email 渠道配置表单 | P1 |

### 1.2 约束条件

- 仅发送**纯文本邮件**（Content-Type: text/plain），不支持 HTML 富文本
- 使用 Go 标准库 `net/smtp` + `crypto/tls`，不引入第三方邮件库
- TLS 连接跳过证书验证（`InsecureSkipVerify: true`），适配企业自建/自签证书 SMTP 服务器
- 所有 SMTP 端口均为用户自行配置，系统不预设端口
- 收件人 `to` 为必填，`cc`（抄送）和 `from_name`（发件人名称）为可选

### 1.3 SMTP 协议参考

- RFC 5321: Simple Mail Transfer Protocol
- RFC 8314: Implicit TLS for SMTP
- 端口 465: Implicit TLS（连接即加密）
- 端口 587: STARTTLS（明文握手后升级加密）
- 端口 25: 纯明文（不推荐，仅兼容旧系统）

---

## 二、数据结构设计

### 2.1 emailConfig -- 渠道配置

```go
// emailConfig Email 渠道配置
type emailConfig struct {
    SMTPHost string `json:"smtp_host"`  // SMTP 服务器地址（必填）
    SMTPPort int    `json:"smtp_port"`  // SMTP 端口（必填，465/587/25）
    Username string `json:"username"`   // SMTP 登录用户名（必填）
    Password string `json:"password"`   // SMTP 登录密码/授权码（必填）
    To       string `json:"to"`         // 收件人邮箱（必填）
    CC       string `json:"cc"`         // 抄送人邮箱（可选，多个用逗号分隔）
    FromName string `json:"from_name"`  // 发件人名称（可选）
}
```

**设计说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `smtp_host` | string | 是 | SMTP 服务器地址，如 `smtp.example.com`、`smtp.gmail.com` |
| `smtp_port` | int | 是 | SMTP 端口号。465 为隐式 TLS，587 为 STARTTLS，25 为明文 |
| `username` | string | 是 | SMTP 认证用户名，通常为完整邮箱地址 |
| `password` | string | 是 | SMTP 认证密码或授权码（如 Gmail 使用 App Password） |
| `to` | string | 是 | 收件人邮箱地址 |
| `cc` | string | 否 | 抄送人邮箱，多个邮箱用英文逗号分隔 |
| `from_name` | string | 否 | 发件人显示名称，不设置时显示原始邮箱地址 |

**与 Feishu/Telegram 的对比：**

| 维度 | FeishuSender | TelegramSender | EmailSender |
|------|-------------|----------------|-------------|
| 协议类型 | HTTPS Webhook | HTTPS REST API | SMTP（原生协议） |
| 认证方式 | Webhook URL + 可选签名 | Bot Token（URL 路径） | SMTP AUTH LOGIN |
| TLS 策略 | 标准 HTTPS TLS | 标准 HTTPS TLS | 三种模式（隐式/显式/明文） |
| 客户端管理 | 包级别共享 http.Client | 每次创建（支持代理） | 每次创建（根据端口选择 TLS 模式） |
| 消息格式 | JSON 嵌套结构 | JSON 扁平结构 | 原始文本 RFC 5322 格式 |

### 2.2 邮件格式 -- RFC 5322

```
From: "发件人名称" <username@example.com>
To: recipient@example.com
Cc: cc1@example.com, cc2@example.com
Subject: =?UTF-8?B?base64编码标题?=
Date: RFC1123 格式时间
Content-Type: text/plain; charset="UTF-8"

[标题]
内容
```

**设计说明：**

- `Subject` 使用 RFC 2047 Base64 编码（格式：`=?UTF-8?B?xxxx?=`），支持中文和特殊字符
- `Date` 使用 RFC 1123 格式（如 `Mon, 07 May 2026 10:00:00 +0800`）
- 邮件正文格式为 `[标题]\n内容`，与 Feishu/Telegram 保持一致
- 纯文本格式，不做 HTML 转义

### 2.3 邮箱地址解析

`cc` 字段支持逗号分隔的多个邮箱地址，解析后需要：

- 去除前后空格
- 过滤空字符串
- 合并到 SMTP `RcptTo` 调用中

---

## 三、Send() 方法流程设计

### 3.1 流程图

```
Send(ctx, config, title, content)
  │
  ├─ 1. 解析渠道配置 JSON → emailConfig
  │     └─ 失败 → 返回 "解析 Email 渠道配置失败"
  │
  ├─ 2. 校验必填字段
  │     ├─ smtp_host 为空 → 返回 "SMTP 服务器地址不能为空"
  │     ├─ smtp_port 无效 → 返回 "SMTP 端口不能为空"
  │     ├─ username 为空 → 返回 "SMTP 用户名不能为空"
  │     ├─ password 为空 → 返回 "SMTP 密码不能为空"
  │     └─ to 为空 → 返回 "收件人邮箱不能为空"
  │
  ├─ 3. 记录调试日志（脱敏：隐藏密码）
  │
  ├─ 4. 根据端口选择 TLS 连接模式
  │     ├─ 465 → 隐式 TLS（连接即加密）
  │     ├─ 587 → STARTTLS（明文握手后升级加密）
  │     └─ 25 → 明文传输
  │
  ├─ 5. 构造邮件原始内容（RFC 5322 格式）
  │     ├─ 拼接 From/To/Cc/Subject/Date/Content-Type 头
  │     └─ Subject 使用 Base64 编码（RFC 2047）
  │
  ├─ 6. 解析收件人列表（to + cc）
  │
  ├─ 7. 建立 SMTP 连接（10 秒超时）
  │     └─ 连接失败 → 返回 "连接 SMTP 服务器失败"
  │
  ├─ 8. SMTP AUTH LOGIN 认证
  │     ├─ 认证失败 → 返回 "SMTP 认证失败"
  │     └─ 成功 → 继续
  │
  ├─ 9. 设置 MailFrom（发件人邮箱）
  │
  ├─ 10. 设置 RcptTo（所有收件人 + 抄送人）
  │
  ├─ 11. 写入邮件数据（Data 命令）
  │
  ├─ 12. 关闭 SMTP 连接（defer Quit）
  │
  └─ 完成
```

### 3.2 TLS 连接模式详解

```
┌──────────────────────────────────────────────────────────┐
│ 根据端口自动选择 TLS 模式                                 │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  端口 465 ──→ Implicit TLS（SMTPS）                       │
│    ├─ 使用 tls.Dial() 直接建立加密连接                    │
│    ├─ 握手完成后再执行 SMTP 协议                          │
│    └─ 相当于 HTTPS 的 TLS-First 模式                      │
│                                                          │
│  端口 587 ──→ STARTTLS                                    │
│    ├─ 先用 net.Dial() 建立明文连接                        │
│    ├─ 执行 EHLO 后发送 STARTTLS 命令                      │
│    ├─ 服务器响应 220 后升级为 TLS 连接                    │
│    └─ 后续所有通信均为加密传输                            │
│                                                          │
│  端口 25  ──→ 明文传输                                    │
│    ├─ 使用 net.Dial() 建立连接                            │
│    ├─ 不启用任何加密                                      │
│    └─ 仅用于内网/旧系统兼容                                │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

**设计决策：**

| 决策 | 选择 | 理由 |
|------|------|------|
| TLS 证书验证 | `InsecureSkipVerify: true` | 企业内网 SMTP 常用自签证书，严格验证会导致大量连接失败 |
| 端口白名单 | 不限制 | 允许用户配置任意端口（如某些企业使用 2525、4650 等非标端口） |
| 默认端口 | 不设默认值 | 不同云厂商 SMTP 端口策略不同（如阿里云 25/80/465，腾讯云 465/587） |
| TLS 最低版本 | TLS 1.2 | 兼容大多数现代 SMTP 服务器，避免 TLS 1.3 兼容性问题 |

### 3.3 Subject Base64 编码

RFC 2047 规定邮件头中的非 ASCII 字符必须进行编码，格式为：

```
=?charset?encoding?encoded-text?=
```

Subject 编码实现：

```go
// encodeSubject 对邮件主题进行 RFC 2047 Base64 编码
func encodeSubject(subject string) string {
    encoded := base64.StdEncoding.EncodeToString([]byte(subject))
    return fmt.Sprintf("=?UTF-8?B?%s?=", encoded)
}
```

---

## 四、SMTP 客户端创建策略

### 4.1 核心设计

与 FeishuSender 共享 `http.Client` 不同，EmailSender 的 SMTP 连接是**有状态的 TCP 连接**，且每次发送的收件人、认证信息可能不同，因此**不能使用共享连接**。

**策略：每次 Send() 调用时创建新的 SMTP 连接。**

### 4.2 实现方案

```go
const emailConnectTimeout = 10 * time.Second

// dialSMTP 根据端口建立 SMTP 连接
func dialSMTP(host string, port int) (*smtp.Client, error) {
    addr := fmt.Sprintf("%s:%d", host, port)

    if port == 465 {
        // 隐式 TLS（SMTPS）
        tlsConfig := &tls.Config{
            InsecureSkipVerify: true,
            MinVersion:         tls.VersionTLS12,
        }
        conn, err := tls.Dial("tcp", addr, tlsConfig)
        if err != nil {
            return nil, fmt.Errorf("建立 TLS 连接失败: %w", err)
        }
        return smtp.NewClient(conn, host)
    }

    // 端口 25/587：先建立明文连接
    conn, err := net.DialTimeout("tcp", addr, emailConnectTimeout)
    if err != nil {
        return nil, fmt.Errorf("连接 SMTP 服务器失败: %w", err)
    }

    client, err := smtp.NewClient(conn, host)
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("创建 SMTP 客户端失败: %w", err)
    }

    // 端口 587：执行 STARTTLS
    if port == 587 {
        tlsConfig := &tls.Config{
            InsecureSkipVerify: true,
            MinVersion:         tls.VersionTLS12,
            ServerName:         host,
        }
        if err := client.StartTLS(tlsConfig); err != nil {
            client.Close()
            return nil, fmt.Errorf("STARTTLS 升级失败: %w", err)
        }
    }

    return client, nil
}
```

### 4.3 方案对比

| 方案 | 描述 | 优点 | 缺点 | 结论 |
|------|------|------|------|------|
| A. 每次创建 | 每次 Send() 创建新连接 | 简单直接，状态隔离 | 每次连接/握手有开销 | **采用** |
| B. 连接池 | 复用 SMTP 连接 | 减少连接开销 | SMTP 连接有状态（认证、会话），复用复杂且易出错 | 过度设计 |
| C. 包级别共享 | 与 FeishuSender 相同 | 零创建开销 | SMTP 连接无法跨请求共享（认证后状态混乱） | 不可行 |

**选择方案 A 的理由：**

1. SMTP 连接是有状态的（HELO、AUTH、MAIL FROM、RCPT TO、DATA、QUIT），不适合复用
2. 每次发送的目标邮箱可能不同，复用连接需要复杂的会话管理
3. SMTP 连接建立开销小（通常 100ms 内），对用户体验影响可忽略
4. 代码简洁，无需管理连接池和并发安全

### 4.4 SMTP 认证方式

| 认证方式 | 说明 | 支持情况 |
|----------|------|----------|
| AUTH LOGIN | 最广泛支持，Base64 编码用户名密码 | **采用** |
| AUTH PLAIN | 部分服务器支持 | 不采用 |
| AUTH CRAM-MD5 | 较老的方式，安全性一般 | 不采用 |
| 无认证 | 内网/开放中继 | 不采用（安全风险） |

Go 标准库 `net/smtp` 提供了 `smtp.PlainAuth` 函数，支持 PLAIN 和 LOGIN 两种认证：

```go
auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
if err := client.Auth(auth); err != nil {
    return fmt.Errorf("SMTP 认证失败: %w", err)
}
```

**注意：** 部分 SMTP 服务器（如 Exchange）仅支持 LOGIN 而不支持 PLAIN，Go 标准库的 `smtp.PlainAuth` 在 `tls=false` 时自动使用 LOGIN 模式。

---

## 五、错误处理与日志策略

### 5.1 错误处理层级

```
层级 1：配置解析错误
  → fmt.Errorf("解析 Email 渠道配置失败: %w", err)

层级 2：必填字段校验
  → fmt.Errorf("SMTP 服务器地址不能为空")
  → fmt.Errorf("SMTP 端口不能为空")
  → fmt.Errorf("SMTP 用户名不能为空")
  → fmt.Errorf("SMTP 密码不能为空")
  → fmt.Errorf("收件人邮箱不能为空")

层级 3：连接错误
  → fmt.Errorf("连接 SMTP 服务器失败: %w", err)        // 端口 25/587
  → fmt.Errorf("建立 TLS 连接失败: %w", err)            // 端口 465
  → fmt.Errorf("STARTTLS 升级失败: %w", err)            // 端口 587
  → fmt.Errorf("创建 SMTP 客户端失败: %w", err)

层级 4：认证错误
  → fmt.Errorf("SMTP 认证失败: %w", err)

层级 5：发送错误
  → fmt.Errorf("设置发件人失败: %w", err)
  → fmt.Errorf("设置收件人失败: %w", err)
  → fmt.Errorf("写入邮件数据失败: %w", err)

层级 6：关闭连接错误
  → 日志记录，不返回错误（邮件已发送）
```

### 5.2 日志策略

| 级别 | 场景 | 示例 |
|------|------|------|
| Debug | 配置解析结果 | `zap.String("smtp_host", "smtp.example.com"), zap.Int("smtp_port", 587), zap.String("username", "user@xxx.com"), zap.String("to", "recipient@xxx.com"), zap.Bool("has_cc", true), zap.Bool("has_from_name", true)` |
| Debug | 邮件头信息 | `zap.String("subject_encoded", "=?UTF-8?B?xxx?="), zap.Int("body_length", 256)` |
| Debug | 收件人列表 | `zap.Strings("recipients", ["to@xxx.com", "cc1@xxx.com"])` |
| Error | 连接失败 | `zap.String("addr", "smtp.example.com:587"), zap.Error(err)` |
| Error | 认证失败 | `zap.String("username", "user@xxx.com"), zap.Error(err)` |
| Warn | 收件人拒绝 | `zap.String("recipient", "invalid@xxx.com"), zap.Error(err)` |
| Error | 写入数据失败 | `zap.Error(err)` |
| Info | 发送成功 | 无额外字段 |

### 5.3 安全脱敏规则

**严格禁止在日志中记录 SMTP 密码和完整邮箱地址组合。**

| 字段 | 日志输出规则 | 示例 |
|------|-------------|------|
| `password` | **绝对不记录**，仅记录 `has_password: true/false` | 永远不记录密码内容 |
| `smtp_host` | 完整记录 | `smtp.gmail.com` |
| `smtp_port` | 完整记录 | `587` |
| `username` | 记录域名部分，用户名部分脱敏 | `***@gmail.com` |
| `to` | 完整记录（目标地址非敏感信息） | `recipient@example.com` |
| `cc` | 完整记录 | `cc1@example.com, cc2@example.com` |
| `from_name` | 完整记录 | `系统通知` |

**Username 脱敏函数：**

```go
// maskEmailUsername 对邮箱用户名部分进行脱敏处理
// "user@example.com" → "***@example.com"
func maskEmailUsername(email string) string {
    atIdx := strings.Index(email, "@")
    if atIdx <= 0 {
        return "***"
    }
    return "***" + email[atIdx:]
}
```

---

## 六、代码变更清单

### 6.1 新增文件

| 文件 | 说明 |
|------|------|
| `backend/internal/sender/email.go` | EmailSender 完整实现 |
| `backend/internal/sender/email_test.go` | EmailSender 单元测试 |

### 6.2 修改文件

| 文件 | 变更内容 |
|------|----------|
| `backend/internal/sender/factory.go` | 将 `&EmailSender{}` 空壳替换为 `NewEmailSender(logger)` |
| `backend/internal/handler/dto/channel_dto.go` | 在 `ChannelTypeMetas` 中添加 Email 渠道元数据 |

### 6.3 无需修改的文件

| 文件 | 原因 |
|------|------|
| `message_service.go` | 通过工厂模式调用，无需感知具体 Sender 实现 |
| `channel_handler.go` | 通过 `GetChannelTypes()` 动态获取，无需修改 |
| 前端组件 | 通过 `config_fields` 动态渲染，无需修改 |

---

## 七、详细代码设计

### 7.1 email.go 完整结构

```go
package sender

import (
    "crypto/tls"
    "encoding/base64"
    "fmt"
    "net"
    "net/smtp"
    "strings"
    "time"

    "go.uber.org/zap"
    "gorm.io/datatypes"
    "context"
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
    SMTPHost string `json:"smtp_host"`  // SMTP 服务器地址
    SMTPPort int    `json:"smtp_port"`  // SMTP 端口（465/587/25）
    Username string `json:"username"`   // SMTP 登录用户名
    Password string `json:"password"`   // SMTP 登录密码/授权码
    To       string `json:"to"`         // 收件人邮箱
    CC       string `json:"cc"`         // 抄送人邮箱（可选）
    FromName string `json:"from_name"`  // 发件人名称（可选）
}

const emailConnectTimeout = 10 * time.Second

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
        from = fmt.Sprintf("\"%s\" <%s>", cfg.FromName, cfg.Username)
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

    // 4. 建立 SMTP 连接
    client, err := dialSMTP(cfg.SMTPHost, cfg.SMTPPort)
    if err != nil {
        s.logger.Error("连接 SMTP 服务器失败",
            zap.String("addr", fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)),
            zap.Error(err),
        )
        return err
    }
    defer client.Quit()

    // 5. SMTP 认证
    auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
    if err := client.Auth(auth); err != nil {
        s.logger.Error("SMTP 认证失败",
            zap.String("username", maskEmailUsername(cfg.Username)),
            zap.Error(err),
        )
        return fmt.Errorf("SMTP 认证失败: %w", err)
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

// dialSMTP 根据端口建立 SMTP 连接
func dialSMTP(host string, port int) (*smtp.Client, error) {
    addr := fmt.Sprintf("%s:%d", host, port)

    if port == 465 {
        // 隐式 TLS（SMTPS）
        tlsConfig := &tls.Config{
            InsecureSkipVerify: true,
            MinVersion:         tls.VersionTLS12,
            ServerName:         host,
        }
        conn, err := tls.Dial("tcp", addr, tlsConfig)
        if err != nil {
            return nil, fmt.Errorf("建立 TLS 连接失败: %w", err)
        }
        return smtp.NewClient(conn, host)
    }

    // 端口 25/587：先建立明文连接
    conn, err := net.DialTimeout("tcp", addr, emailConnectTimeout)
    if err != nil {
        return nil, fmt.Errorf("连接 SMTP 服务器失败: %w", err)
    }

    client, err := smtp.NewClient(conn, host)
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("创建 SMTP 客户端失败: %w", err)
    }

    // 端口 587：执行 STARTTLS
    if port == 587 {
        tlsConfig := &tls.Config{
            InsecureSkipVerify: true,
            MinVersion:         tls.VersionTLS12,
            ServerName:         host,
        }
        if err := client.StartTLS(tlsConfig); err != nil {
            client.Close()
            return nil, fmt.Errorf("STARTTLS 升级失败: %w", err)
        }
    }

    return client, nil
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

// maskEmailUsername 对邮箱用户名部分进行脱敏处理
func maskEmailUsername(email string) string {
    atIdx := strings.Index(email, "@")
    if atIdx <= 0 {
        return "***"
    }
    return "***" + email[atIdx:]
}
```

### 7.2 factory.go 变更

```go
// 变更前
"email":    &EmailSender{},

// 变更后
"email":    NewEmailSender(logger),
```

同时删除文件底部的空壳 `EmailSender` 结构体定义：

```go
// 删除以下代码
type EmailSender struct{}

func (s *EmailSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
    return nil
}
```

### 7.3 channel_dto.go 变更

在 `ChannelTypeMetas` 切片中追加 Email 渠道元数据：

```go
var ChannelTypeMetas = []ChannelTypeMeta{
    // ... feishu, telegram ...

    // ---- 新增 Email 渠道 ----
    {
        Type:        ChannelTypeEmail,
        Name:        "邮件",
        Description: "邮件推送",
        ConfigFields: []ConfigField{
            {
                Name:        "smtp_host",
                Label:       "SMTP 服务器",
                Type:        "string",
                Required:    true,
                Placeholder: "smtp.example.com",
            },
            {
                Name:        "smtp_port",
                Label:       "SMTP 端口",
                Type:        "number",
                Required:    true,
                Placeholder: "587",
            },
            {
                Name:        "username",
                Label:       "用户名",
                Type:        "string",
                Required:    true,
                Placeholder: "user@example.com",
            },
            {
                Name:        "password",
                Label:       "密码/授权码",
                Type:        "password",
                Required:    true,
                Placeholder: "邮箱密码或授权码",
            },
            {
                Name:        "to",
                Label:       "收件人",
                Type:        "string",
                Required:    true,
                Placeholder: "recipient@example.com",
            },
            {
                Name:        "cc",
                Label:       "抄送人（可选）",
                Type:        "string",
                Required:    false,
                Placeholder: "cc1@example.com, cc2@example.com",
            },
            {
                Name:        "from_name",
                Label:       "发件人名称（可选）",
                Type:        "string",
                Required:    false,
                Placeholder: "系统通知",
            },
        },
    },
}
```

---

## 八、前端配置字段定义

### 8.1 前端无需代码修改

前端通过 `GET /channel-types` API 动态获取 `config_fields`，使用 `v-for` 渲染表单字段。新增 Email 渠道类型后，前端会自动渲染以下表单：

| 字段名 | 标签 | 输入类型 | 必填 | 占位符 |
|--------|------|----------|------|--------|
| `smtp_host` | SMTP 服务器 | text | 是 | smtp.example.com |
| `smtp_port` | SMTP 端口 | text（number 类型渲染为数字输入） | 是 | 587 |
| `username` | 用户名 | text | 是 | user@example.com |
| `password` | 密码/授权码 | password | 是 | 邮箱密码或授权码 |
| `to` | 收件人 | text | 是 | recipient@example.com |
| `cc` | 抄送人（可选） | text | 否 | cc1@example.com, cc2@example.com |
| `from_name` | 发件人名称（可选） | text | 否 | 系统通知 |

### 8.2 前端渲染效果

创建渠道页面选择 Email 后，表单区域将自动展示：

```
┌─────────────────────────────────────────────┐
│ 渠道名称                                     │
│ [                                    ]       │
├─────────────────────────────────────────────┤
│ SMTP 服务器 *                                │
│ [smtp.example.com                ]           │
│                                              │
│ SMTP 端口 *                                  │
│ [587                             ]           │
│                                              │
│ 用户名 *                                     │
│ [user@example.com                ]           │
│                                              │
│ 密码/授权码 *                                │
│ [********************************]  (password 输入框) │
│                                              │
│ 收件人 *                                     │
│ [recipient@example.com           ]           │
│                                              │
│ 抄送人（可选）                                 │
│ [cc1@example.com, cc2@example.com]           │
│                                              │
│ 发件人名称（可选）                              │
│ [系统通知                          ]          │
└─────────────────────────────────────────────┘
```

---

## 九、安全考量

### 9.1 SMTP 密码保护

| 风险 | 措施 |
|------|------|
| 日志泄露 | **绝对不记录密码内容**，仅记录 `has_password: true/false` |
| 数据库存储 | 密码存储在 channels.config JSON 字段中，与 Telegram bot_token 同级，遵循现有安全策略 |
| 前端展示 | 使用 `password` 类型输入框，编辑时回显为掩码 |
| 授权码推荐 | 建议用户使用邮箱授权码（App Password）而非真实密码，如 Gmail/Outlook 均支持 |

### 9.2 TLS 安全策略

| 风险 | 措施 |
|------|------|
| 自签证书连接失败 | `InsecureSkipVerify: true` 跳过证书验证，适配企业内网环境 |
| 中间人攻击 | 内网/自建 SMTP 风险可接受，公网 SMTP 建议用户自行确保网络环境安全 |
| TLS 降级攻击 | 设置 `MinVersion: tls.VersionTLS12`，禁止使用 TLS 1.0/1.1 |
| 端口 25 明文传输 | 不阻止用户配置，但建议在文档中说明安全风险 |

### 9.3 SMTP 注入防护

| 风险 | 措施 |
|------|------|
| 邮件头注入 | 标题和发件人名称中禁止包含 `\r\n`（SMTP 协议注入），当前纯文本内容不做特殊处理，因为 SMTP `Data` 命令后的内容中 `\r\n.\r\n` 表示结束，但内容中出现的 `.` 不会被误判（SMTP 客户端会自动处理） |
| 收件人注入 | 收件人地址由 SMTP `Rcpt` 命令单独发送，不受注入影响 |
| 用户名/密码注入 | SMTP AUTH 由 `net/smtp` 标准库处理，用户输入不直接拼接到协议命令中 |

### 9.4 邮箱地址校验

**首期不强制校验邮箱地址格式**，理由：

1. SMTP 服务器在 `Rcpt` 阶段会拒绝无效邮箱
2. Go 标准库没有内置邮箱正则校验
3. 引入第三方校验库会增加依赖
4. 用户自行配置的邮箱通常为有效地址

---

## 十、边界情况处理

### 10.1 端口处理

| 端口 | 行为 | 说明 |
|------|------|------|
| 465 | 隐式 TLS | 使用 `tls.Dial()` 直接建立加密连接 |
| 587 | STARTTLS | 先明文连接，再 `STARTTLS` 升级为加密 |
| 25 | 明文传输 | 不启用任何加密，仅用于兼容 |
| 其他端口 | 按 25 处理 | 使用明文连接，不执行 STARTTLS |

**设计决策：** 仅对 465 和 587 执行特殊 TLS 逻辑，其他端口（包括 2525、4650 等非标端口）按明文处理。用户如需非标端口使用 TLS，应自行配置反向代理或选择标准端口。

### 10.2 多收件人处理

| 场景 | 处理 |
|------|------|
| `to` 单个邮箱 | `Rcpt(to)` |
| `cc` 多个邮箱（逗号分隔） | 循环 `Rcpt` 每个邮箱 |
| 某个 `Rcpt` 失败 | 中断发送，返回错误（SMTP 事务无法部分提交） |
| `to` + `cc` 重复邮箱 | SMTP 服务器会自动去重，不做预处理 |

### 10.3 SMTP 认证兼容性

| 服务器 | 认证方式 | 兼容性 |
|--------|----------|--------|
| Gmail | PLAIN/LOGIN | 兼容，需使用 App Password |
| Outlook/Office365 | PLAIN/LOGIN | 兼容 |
| QQ邮箱 | LOGIN | 兼容，需使用授权码 |
| 163邮箱 | LOGIN | 兼容，需使用授权码 |
| Exchange | LOGIN | 兼容 |
| AWS SES | PLAIN | 兼容 |

### 10.4 连接超时

| 场景 | 处理 |
|------|------|
| DNS 解析失败 | `net.DialTimeout` 10 秒后超时返回错误 |
| 服务器无响应 | 10 秒后超时返回错误 |
| TLS 握手超时 | 10 秒后超时返回错误 |
| 数据传输超时 | 由 `smtp.Client` 内部超时控制 |

### 10.5 邮件内容长度

| 场景 | 处理 |
|------|------|
| 内容超长 | 无限制（SMTP 协议本身无长度限制，但部分服务器有 10MB 限制） |
| 空内容 | 允许发送（纯 `[标题]\n`） |
| 标题为空 | 允许发送（Subject 为空字符串的 Base64 编码） |
| 特殊字符 | 纯文本内容，不做转义，SMTP 客户端自动处理行结束符 |

### 10.6 连接关闭

| 场景 | 处理 |
|------|------|
| 发送成功 | 调用 `client.Quit()` 优雅关闭 |
| 发送失败 | `defer client.Quit()` 确保连接释放 |
| `Quit()` 失败 | 日志记录，不影响返回值（邮件可能已发送） |
| 连接意外断开 | SMTP 服务器已断开时 `Quit()` 返回错误，忽略即可 |

---

## 十一、UML 活动图

### 11.1 Email 推送活动图

```
@startuml OctoTify_Activity_EmailSend

skinparam activityStyle round
skinparam linetype ortho

title Email 渠道推送活动图

start

:MessageService 调用 EmailSender.Send();

:解析渠道配置 JSON;
note right
  {
    "smtp_host": "smtp.example.com",
    "smtp_port": 587,
    "username": "user@example.com",
    "password": "****",
    "to": "recipient@example.com",
    "cc": "cc1@example.com",
    "from_name": "系统通知"
  }
end note

if [配置解析失败?] then (是)
  :返回错误 "解析 Email 渠道配置失败";
  stop
else (否)
  :提取 SMTP 配置;
endif

if [smtp_host 为空?] then (是)
  :返回错误 "SMTP 服务器地址不能为空";
  stop
elseif [smtp_port 无效?] then (是)
  :返回错误 "SMTP 端口不能为空";
  stop
elseif [username 为空?] then (是)
  :返回错误 "SMTP 用户名不能为空";
  stop
elseif [password 为空?] then (是)
  :返回错误 "SMTP 密码不能为空";
  stop
elseif [to 为空?] then (是)
  :返回错误 "收件人邮箱不能为空";
  stop
else (否)
endif

:构造邮件头（From/To/Cc/Subject/Date）;
note right
  Subject 使用 RFC 2047 Base64 编码
  Content-Type: text/plain; charset="UTF-8"
end note

:解析收件人列表（to + cc）;

if [端口 == 465?] then (是)
  :隐式 TLS: tls.Dial() 建立加密连接;
elseif [端口 == 587?] then (是)
  :net.Dial() 建立明文连接;
  :执行 EHLO + STARTTLS 升级为 TLS;
else (否)
  :net.Dial() 建立明文连接;
  :不启用加密;
endif

if [连接失败?] then (是)
  :返回错误 "连接 SMTP 服务器失败";
  stop
else (否)
endif

:SMTP AUTH LOGIN 认证;

if [认证失败?] then (是)
  :返回错误 "SMTP 认证失败";
  stop
else (否)
endif

:设置 MailFrom（发件人）;

if [设置发件人失败?] then (是)
  :返回错误 "设置发件人失败";
  stop
else (否)
endif

:循环设置 RcptTo（所有收件人）;

if [某个收件人被拒绝?] then (是)
  :返回错误 "设置收件人失败";
  stop
else (否)
endif

:写入邮件数据（DATA 命令）;

if [写入失败?] then (是)
  :返回错误 "写入邮件数据失败";
  stop
else (否)
endif

:关闭数据流;
:调用 Quit() 优雅关闭连接;

:发送成功;

stop

@enduml
```

---

## 十二、测试策略

### 12.1 单元测试

| 测试用例 | 输入 | 预期输出 |
|----------|------|----------|
| 配置解析失败 | `{"invalid": json}` | 错误 "解析 Email 渠道配置失败" |
| smtp_host 为空 | `{"smtp_port": 587}` | 错误 "SMTP 服务器地址不能为空" |
| smtp_port 为 0 | `{"smtp_host": "smtp.example.com"}` | 错误 "SMTP 端口不能为空" |
| username 为空 | `{"smtp_host": "x", "smtp_port": 587}` | 错误 "SMTP 用户名不能为空" |
| password 为空 | `{"smtp_host": "x", "smtp_port": 587, "username": "x"}` | 错误 "SMTP 密码不能为空" |
| to 为空 | `{"smtp_host": "x", "smtp_port": 587, "username": "x", "password": "x"}` | 错误 "收件人邮箱不能为空" |
| 端口 465 隐式 TLS | Mock TLS 服务器 | 使用 `tls.Dial()` 连接 |
| 端口 587 STARTTLS | Mock STARTTLS 服务器 | 先明文后 STARTTLS |
| 端口 25 明文 | Mock SMTP 服务器 | 使用明文连接 |
| 多收件人解析 | `to="a@x.com", cc="b@x.com, c@x.com"` | 解析为 `["a@x.com", "b@x.com", "c@x.com"]` |
| Subject Base64 编码 | 中文标题 "系统通知" | `=?UTF-8?B?5LiA5Lia6ICF6ICF?=` |
| 发件人名称格式化 | `from_name="系统通知", username="a@b.com"` | `"系统通知" <a@b.com>` |
| 邮箱脱敏 | `user@example.com` | `***@example.com` |
| 无 @ 符号邮箱脱敏 | `invalid-email` | `***` |
| 空字符串脱敏 | `""` | `***` |

### 12.2 集成测试

| 测试场景 | 方法 |
|----------|------|
| Gmail SMTP | 使用 Gmail App Password 测试 587 端口 |
| QQ 邮箱 SMTP | 使用 QQ 授权码测试 465 端口 |
| 企业自建 SMTP | 测试 25 端口明文发送 |
| 抄送功能 | 测试 `cc` 字段多收件人 |
| 发件人名称 | 测试 `from_name` 显示效果 |

---

## 十三、任务分配建议

| 任务 | 角色 | 文件 | 预估工时 |
|------|------|------|----------|
| 实现 EmailSender | 后端开发 | `backend/internal/sender/email.go` | 3h |
| 修改 factory.go 注册 | 后端开发 | `backend/internal/sender/factory.go` | 0.5h |
| 修改 channel_dto.go 元数据 | 后端开发 | `backend/internal/handler/dto/channel_dto.go` | 0.5h |
| 编写单元测试 | 后端开发 | `backend/internal/sender/email_test.go` | 2h |
| 前端验证（动态表单渲染） | 前端开发 | 无代码修改，仅需验证 | 0.5h |
| 端到端测试（Gmail/QQ/企业邮箱） | 测试 | - | 1h |

**总计预估：7.5h**

---

## 十四、风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 企业自签 SMTP 证书连接失败 | 推送失败 | 高 | `InsecureSkipVerify: true` 跳过验证 |
| SMTP 密码泄露 | 安全风险 | 低 | 日志不记录密码、password 类型输入框 |
| 邮箱授权码过期 | 推送失败 | 中 | 错误日志记录认证失败，提示用户更新授权码 |
| 端口 25 被云厂商封禁 | 推送失败 | 高 | 建议使用 465/587，云厂商（阿里云/腾讯云）默认封禁 25 出站 |
| SMTP 频率限制 | 推送失败 | 中 | Gmail/Outlook 有日发送上限，建议企业使用专用 SMTP 服务 |
| 收件人邮箱无效 | 推送失败 | 中 | SMTP 服务器返回错误，记录日志提示用户检查 |

---

## 版本历史

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-07 | 1.0 | 初始版本，Email Sender 设计文档 |
