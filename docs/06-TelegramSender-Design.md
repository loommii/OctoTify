# Telegram Sender 设计文档

> 为 OctoTify 消息总线平台新增 Telegram Bot 推送渠道，支持可选 HTTP 代理。

---

## 一、需求概述

### 1.1 功能需求

| 编号 | 需求 | 优先级 |
|------|------|--------|
| F1 | 通过 Telegram Bot API 发送文本消息 | P0 |
| F2 | 支持可选 HTTP 代理（渠道级别配置） | P0 |
| F3 | 遵循现有 Sender 策略模式，与 FeishuSender 保持一致 | P0 |
| F4 | 前端动态渲染 Telegram 渠道配置表单 | P1 |

### 1.2 约束条件

- 代理配置为**渠道级别**（per-channel），非全局配置，不同 Telegram 渠道可使用不同代理
- Telegram Bot API 单条消息长度上限 4096 字符
- 中国大陆用户无法直接访问 `api.telegram.org`，需通过代理中转

### 1.3 Telegram Bot API 参考

- 端点：`POST https://api.telegram.org/bot{token}/sendMessage`
- 请求体：`{"chat_id": "...", "text": "...", "parse_mode": "HTML"}`
- 响应体：`{"ok": true/false, "error_code": ..., "description": ...}`

---

## 二、数据结构设计

### 2.1 telegramConfig -- 渠道配置

```go
// telegramConfig Telegram 渠道配置
type telegramConfig struct {
    BotToken string `json:"bot_token"` // Bot Token（必填）
    ChatID   string `json:"chat_id"`   // Chat ID（必填，支持数字和 @channel_name 格式）
    Proxy    string `json:"proxy"`     // HTTP 代理地址（可选，如 http://127.0.0.1:7890）
}
```

**设计说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `bot_token` | string | 是 | BotFather 创建 Bot 后获得的 Token，格式如 `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz` |
| `chat_id` | string | 是 | 目标聊天 ID。私有聊天为数字（如 `-1001234567890`），公开频道可为 `@channel_name` |
| `proxy` | string | 否 | HTTP/SOCKS5 代理地址。为空时直连；格式支持 `http://host:port`、`https://host:port`、`socks5://host:port` |

**与 FeishuSender 的对比：**

| 维度 | FeishuSender | TelegramSender |
|------|-------------|----------------|
| 认证方式 | Webhook URL + 可选签名密钥 | Bot Token（嵌入 URL 路径） |
| 目标地址 | 固定 Webhook URL | 由 Bot Token 拼接 API URL |
| 代理需求 | 无（飞书国内可达） | 可选（Telegram 国内不可达） |
| 消息格式 | `msg_type` + `content` 嵌套 | 扁平 `text` + `parse_mode` |

### 2.2 telegramMessage -- 请求消息体

```go
// telegramMessage Telegram sendMessage API 请求体
type telegramMessage struct {
    ChatID    string `json:"chat_id"`              // 目标聊天 ID
    Text      string `json:"text"`                 // 消息文本内容
    ParseMode string `json:"parse_mode,omitempty"` // 解析模式：HTML / Markdown
}
```

**设计说明：**

- `ParseMode` 使用 `HTML` 格式，因为 HTML 比 Markdown 在 Telegram 中解析更稳定，不易因用户消息中的特殊字符导致解析错误
- 标题通过 HTML `<b>` 标签加粗显示，格式为 `<b>[标题]</b>\n内容`
- `ParseMode` 设为 `omitempty`，未来可扩展为可配置项

### 2.3 telegramResponse -- API 响应体

```go
// telegramResponse Telegram API 响应体
type telegramResponse struct {
    Ok          bool   `json:"ok"`                    // 请求是否成功
    ErrorCode   int    `json:"error_code,omitempty"`   // 错误码（ok=false 时存在）
    Description string `json:"description,omitempty"`  // 错误描述（ok=false 时存在）
}
```

**Telegram API 常见错误码：**

| error_code | 含义 | 处理策略 |
|------------|------|----------|
| 400 | Bad Request（参数错误） | 检查 chat_id 格式 |
| 401 | Unauthorized（Token 无效） | 提示用户检查 Bot Token |
| 403 | Forbidden（Bot 被封禁或无权限） | 提示用户检查 Bot 权限 |
| 404 | Not Found（chat_id 不存在） | 提示用户检查 Chat ID |
| 429 | Too Many Requests（频率限制） | 记录日志，不重试 |
| 500 | Internal Server Error | 记录日志，可重试 |

---

## 三、Send() 方法流程设计

### 3.1 流程图

```
Send(ctx, config, title, content)
  │
  ├─ 1. 解析渠道配置 JSON → telegramConfig
  │     └─ 失败 → 返回 "解析 Telegram 渠道配置失败"
  │
  ├─ 2. 校验必填字段
  │     ├─ bot_token 为空 → 返回 "Bot Token 不能为空"
  │     └─ chat_id 为空 → 返回 "Chat ID 不能为空"
  │
  ├─ 3. 记录调试日志（脱敏：仅记录 token 前后各 4 位）
  │
  ├─ 4. 构造消息文本
  │     ├─ 拼接格式：<b>[title]</b>\ncontent
  │     └─ 超过 4096 字符 → 截断并追加 "\n\n[消息已截断]"
  │
  ├─ 5. 构造 telegramMessage 请求体
  │     └─ {chat_id, text, parse_mode: "HTML"}
  │
  ├─ 6. 序列化请求体为 JSON
  │
  ├─ 7. 创建 HTTP 客户端
  │     ├─ proxy 不为空 → 通过代理创建 Client
  │     └─ proxy 为空 → 创建直连 Client（30s 超时）
  │
  ├─ 8. 创建 HTTP POST 请求
  │     └─ URL: https://api.telegram.org/bot{token}/sendMessage
  │     └─ Header: Content-Type: application/json
  │
  ├─ 9. 发送请求
  │     └─ 网络错误 → 返回 "发送 Telegram 请求失败"
  │
  ├─ 10. 读取响应体
  │
  ├─ 11. 检查 HTTP 状态码
  │     └─ 非 200 → 返回 "Telegram 返回 HTTP 错误"
  │
  ├─ 12. 解析响应 JSON → telegramResponse
  │
  ├─ 13. 检查业务状态 ok 字段
  │     ├─ ok=true → 推送成功，返回 nil
  │     └─ ok=false → 返回 "Telegram 推送失败: {description}"
  │
  └─ 完成
```

### 3.2 消息长度截断策略

Telegram `sendMessage` API 单条消息上限为 **4096** 个 UTF-8 字符。截断策略如下：

```
原始文本 = "<b>[title]</b>\n" + content
if len(原始文本) > 4096:
    截断后缀 = "\n\n[消息已截断]"
    可用长度 = 4096 - len(截断后缀)
    最终文本 = 原始文本[:可用长度] + 截断后缀
```

**设计决策：**

- 截断发生在最终拼接后，而非仅截断 content 部分，因为标题 + 格式标签也占用字符数
- 追加 `[消息已截断]` 提示，让接收方知晓内容不完整
- 不采用自动分片发送（多消息），原因：
  1. 分片逻辑复杂，需处理 HTML 标签闭合
  2. 分片顺序无法保证
  3. 当前业务场景（通知推送）单条消息通常不会超限

---

## 四、HTTP 客户端创建策略

### 4.1 核心设计

与 FeishuSender 使用包级别共享 `feishuHTTPClient` 不同，TelegramSender 需要根据渠道配置动态决定是否使用代理，因此**不能使用共享的包级别客户端**。

**策略：每次 Send() 调用时根据 proxy 配置创建 HTTP 客户端。**

### 4.2 实现方案

```go
const telegramHTTPTimeout = 30 * time.Second

// newHTTPClient 根据代理配置创建 HTTP 客户端
// proxy 为空时直连，非空时通过代理连接
func newHTTPClient(proxy string) (*http.Client, error) {
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
    }

    if proxy != "" {
        proxyURL, err := url.Parse(proxy)
        if err != nil {
            return nil, fmt.Errorf("解析代理地址失败: %w", err)
        }
        transport.Proxy = http.ProxyURL(proxyURL)
    }

    return &http.Client{
        Timeout:   telegramHTTPTimeout,
        Transport: transport,
    }, nil
}
```

### 4.3 方案对比

| 方案 | 描述 | 优点 | 缺点 | 结论 |
|------|------|------|------|------|
| A. 每次创建 | 每次 Send() 创建新 Client | 简单直接，代理配置隔离 | 每次创建有微小开销 | **采用** |
| B. 缓存池 | 按 proxy 地址缓存 Client | 避免重复创建 | 需要管理缓存生命周期、并发安全 | 过度设计 |
| C. 包级别共享 | 与 FeishuSender 相同 | 零创建开销 | 不支持 per-channel 代理 | 不满足需求 |

**选择方案 A 的理由：**

1. `http.Client` 创建开销极小（仅初始化 Transport 结构体），不涉及网络连接
2. 实际网络连接由 `Transport` 内部连接池管理，创建新 Client 不会导致连接浪费
3. 代码简洁，无需管理缓存和并发安全
4. 与现有 FeishuSender 的 30 秒超时策略保持一致

### 4.4 代理协议支持

| 协议 | 格式 | 说明 |
|------|------|------|
| HTTP | `http://host:port` | 最常见，如 Clash 默认 |
| HTTPS | `https://host:port` | 加密代理 |
| SOCKS5 | `socks5://host:port` | 需引入 `golang.org/x/net/proxy` |

**首期实现仅支持 HTTP/HTTPS 代理**，原因：

1. Go 标准库 `net/http` 原生支持 HTTP 代理，零额外依赖
2. SOCKS5 需要引入 `golang.org/x/net/proxy`，增加依赖
3. 大多数用户代理工具（Clash、V2Ray）均提供 HTTP 代理端口
4. 后续如需 SOCKS5 支持可增量添加，不影响现有接口

---

## 五、错误处理与日志策略

### 5.1 错误处理层级

```
层级 1：配置解析错误
  → fmt.Errorf("解析 Telegram 渠道配置失败: %w", err)

层级 2：必填字段校验
  → fmt.Errorf("Telegram Bot Token 不能为空")
  → fmt.Errorf("Telegram Chat ID 不能为空")

层级 3：代理配置错误
  → fmt.Errorf("解析代理地址失败: %w", err)

层级 4：网络请求错误
  → fmt.Errorf("发送 Telegram 请求失败: %w", err)

层级 5：HTTP 状态码错误
  → fmt.Errorf("Telegram 返回 HTTP 错误: %d, body: %s", statusCode, body)

层级 6：业务逻辑错误（ok=false）
  → fmt.Errorf("Telegram 推送失败: %s (error_code: %d)", description, errorCode)
```

### 5.2 日志策略

| 级别 | 场景 | 示例 |
|------|------|------|
| Debug | 配置解析结果 | `zap.String("bot_token", "1234...wxyz"), zap.String("chat_id", "-100xxx"), zap.Bool("has_proxy", true)` |
| Debug | 请求体 | `zap.String("body", string(bodyBytes))` |
| Debug | 原始响应 | `zap.Int("http_status", 200), zap.String("body", string(respBody))` |
| Error | 网络请求失败 | `zap.String("url", apiURL), zap.Error(err)` |
| Error | 响应解析失败 | `zap.String("raw_body", string(respBody)), zap.Error(err)` |
| Warn | 业务推送失败 | `zap.Int("error_code", 403), zap.String("description", "Forbidden")` |
| Info | 推送成功 | 无额外字段 |

### 5.3 安全脱敏规则

**严格禁止在日志中记录完整 Bot Token 和代理认证信息。**

| 字段 | 日志输出规则 | 示例 |
|------|-------------|------|
| `bot_token` | 仅显示前 4 位 + `...` + 后 4 位 | `1234...wxyz` |
| `chat_id` | 完整记录（非敏感信息） | `-1001234567890` |
| `proxy` | 记录 host:port，隐藏认证信息 | `http://127.0.0.1:7890`（不含 user:pass） |
| `secret` | 仅记录 `has_secret: true/false` | 同 FeishuSender 策略 |

**Bot Token 脱敏函数：**

```go
// maskToken 对 Bot Token 进行脱敏处理
// "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz" → "1234...wxyz"
func maskToken(token string) string {
    if len(token) <= 8 {
        return "****"
    }
    return token[:4] + "..." + token[len(token)-4:]
}
```

---

## 六、代码变更清单

### 6.1 新增文件

| 文件 | 说明 |
|------|------|
| `backend/internal/sender/telegram.go` | TelegramSender 完整实现 |

### 6.2 修改文件

| 文件 | 变更内容 |
|------|----------|
| `backend/internal/sender/factory.go` | 将 `TelegramSender` 从空壳替换为 `NewTelegramSender(logger)` |
| `backend/internal/handler/dto/channel_dto.go` | 在 `ChannelTypeMetas` 中添加 Telegram 渠道元数据 |
| `backend/pkg/xerr/codes.go` | 新增 Telegram 相关第三方服务错误码（可选） |

### 6.3 无需修改的文件

| 文件 | 原因 |
|------|------|
| `message_service.go` | 通过工厂模式调用，无需感知具体 Sender 实现 |
| `channel_handler.go` | 通过 `GetChannelTypes()` 动态获取，无需修改 |
| 前端组件 | 通过 `config_fields` 动态渲染，无需修改 |

---

## 七、详细代码设计

### 7.1 telegram.go 完整结构

```go
package sender

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"

    "go.uber.org/zap"
    "gorm.io/datatypes"
)

// TelegramSender Telegram 消息发送器
// 官方文档：https://core.telegram.org/bots/api#sendmessage
type TelegramSender struct {
    logger *zap.Logger
}

// NewTelegramSender 创建 Telegram 发送器实例
func NewTelegramSender(logger *zap.Logger) *TelegramSender {
    return &TelegramSender{logger: logger}
}

// telegramConfig Telegram 渠道配置
type telegramConfig struct {
    BotToken string `json:"bot_token"` // Bot Token
    ChatID   string `json:"chat_id"`   // Chat ID
    Proxy    string `json:"proxy"`     // HTTP 代理地址（可选）
}

// telegramMessage Telegram sendMessage API 请求体
type telegramMessage struct {
    ChatID    string `json:"chat_id"`
    Text      string `json:"text"`
    ParseMode string `json:"parse_mode,omitempty"`
}

// telegramResponse Telegram API 响应体
type telegramResponse struct {
    Ok          bool   `json:"ok"`
    ErrorCode   int    `json:"error_code,omitempty"`
    Description string `json:"description,omitempty"`
}

const (
    telegramAPIBaseURL = "https://api.telegram.org/bot"
    telegramHTTPTimeout = 30 * time.Second
    telegramMaxMessageLen = 4096
    telegramTruncateSuffix = "\n\n[消息已截断]"
)

// Send 发送 Telegram 消息
func (s *TelegramSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
    // 1. 解析渠道配置
    var cfg telegramConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return fmt.Errorf("解析 Telegram 渠道配置失败: %w", err)
    }

    s.logger.Debug("Telegram 渠道配置",
        zap.String("bot_token", maskToken(cfg.BotToken)),
        zap.String("chat_id", cfg.ChatID),
        zap.Bool("has_proxy", cfg.Proxy != ""),
    )

    // 2. 校验必填字段
    if cfg.BotToken == "" {
        return fmt.Errorf("Telegram Bot Token 不能为空")
    }
    if cfg.ChatID == "" {
        return fmt.Errorf("Telegram Chat ID 不能为空")
    }

    // 3. 构造消息文本
    text := fmt.Sprintf("<b>[%s]</b>\n%s", title, content)
    text = truncateMessage(text, telegramMaxMessageLen, telegramTruncateSuffix)

    // 4. 构造请求体
    msg := telegramMessage{
        ChatID:    cfg.ChatID,
        Text:      text,
        ParseMode: "HTML",
    }

    bodyBytes, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("构造 Telegram 消息体失败: %w", err)
    }

    s.logger.Debug("Telegram 请求体",
        zap.String("body", string(bodyBytes)),
    )

    // 5. 创建 HTTP 客户端（根据代理配置）
    client, err := newHTTPClient(cfg.Proxy)
    if err != nil {
        return fmt.Errorf("创建 HTTP 客户端失败: %w", err)
    }

    // 6. 创建 HTTP 请求
    apiURL := telegramAPIBaseURL + cfg.BotToken + "/sendMessage"
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
    if err != nil {
        return fmt.Errorf("创建 HTTP 请求失败: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    // 7. 发送请求
    resp, err := client.Do(req)
    if err != nil {
        s.logger.Error("Telegram 网络请求失败",
            zap.String("url", telegramAPIBaseURL+maskToken(cfg.BotToken)+"/sendMessage"),
            zap.Bool("use_proxy", cfg.Proxy != ""),
            zap.Error(err),
        )
        return fmt.Errorf("发送 Telegram 请求失败: %w", err)
    }
    defer resp.Body.Close()

    // 8. 读取响应体
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("读取 Telegram 响应失败: %w", err)
    }

    s.logger.Debug("Telegram 原始响应",
        zap.Int("http_status", resp.StatusCode),
        zap.String("body", string(respBody)),
    )

    // 9. 检查 HTTP 状态码
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("Telegram 返回 HTTP 错误: %d, body: %s", resp.StatusCode, string(respBody))
    }

    // 10. 解析响应 JSON
    var tgResp telegramResponse
    if err := json.Unmarshal(respBody, &tgResp); err != nil {
        s.logger.Error("解析 Telegram 响应失败",
            zap.String("raw_body", string(respBody)),
            zap.Error(err),
        )
        return fmt.Errorf("解析 Telegram 响应失败: %w", err)
    }

    s.logger.Debug("Telegram 响应解析",
        zap.Bool("ok", tgResp.Ok),
        zap.Int("error_code", tgResp.ErrorCode),
        zap.String("description", tgResp.Description),
    )

    // 11. 检查业务状态
    if !tgResp.Ok {
        s.logger.Warn("Telegram 推送业务失败",
            zap.Int("error_code", tgResp.ErrorCode),
            zap.String("description", tgResp.Description),
        )
        return fmt.Errorf("Telegram 推送失败: %s (error_code: %d)", tgResp.Description, tgResp.ErrorCode)
    }

    s.logger.Info("Telegram 推送成功")
    return nil
}
```

### 7.2 辅助函数

```go
// newHTTPClient 根据代理配置创建 HTTP 客户端
func newHTTPClient(proxy string) (*http.Client, error) {
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
    }

    if proxy != "" {
        proxyURL, err := url.Parse(proxy)
        if err != nil {
            return nil, fmt.Errorf("解析代理地址失败: %w", err)
        }
        transport.Proxy = http.ProxyURL(proxyURL)
    }

    return &http.Client{
        Timeout:   telegramHTTPTimeout,
        Transport: transport,
    }, nil
}

// maskToken 对 Bot Token 进行脱敏处理
func maskToken(token string) string {
    if len(token) <= 8 {
        return "****"
    }
    return token[:4] + "..." + token[len(token)-4:]
}

// truncateMessage 截断超长消息
func truncateMessage(text string, maxLen int, suffix string) string {
    if len(text) <= maxLen {
        return text
    }
    availableLen := maxLen - len(suffix)
    if availableLen <= 0 {
        return text[:maxLen]
    }
    return text[:availableLen] + suffix
}
```

### 7.3 factory.go 变更

```go
// 变更前
"telegram": &TelegramSender{},

// 变更后
"telegram": NewTelegramSender(logger),
```

同时删除文件底部的空壳 `TelegramSender` 结构体定义：

```go
// 删除以下代码
type TelegramSender struct{}

func (s *TelegramSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
    return nil
}
```

### 7.4 channel_dto.go 变更

在 `ChannelTypeMetas` 切片中追加 Telegram 渠道元数据：

```go
var ChannelTypeMetas = []ChannelTypeMeta{
    {
        Type:        ChannelTypeFeishu,
        Name:        "飞书",
        Description: "飞书自定义机器人",
        ConfigFields: []ConfigField{
            {
                Name:        "webhook_url",
                Label:       "Webhook 地址",
                Type:        "url",
                Required:    true,
                Placeholder: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
            },
            {
                Name:        "secret",
                Label:       "签名密钥（可选）",
                Type:        "password",
                Required:    false,
                Placeholder: "用于签名校验的密钥",
            },
        },
    },
    // ---- 新增 Telegram 渠道 ----
    {
        Type:        ChannelTypeTelegram,
        Name:        "Telegram",
        Description: "Telegram Bot 推送",
        ConfigFields: []ConfigField{
            {
                Name:        "bot_token",
                Label:       "Bot Token",
                Type:        "password",
                Required:    true,
                Placeholder: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
            },
            {
                Name:        "chat_id",
                Label:       "Chat ID",
                Type:        "string",
                Required:    true,
                Placeholder: "-1001234567890 或 @channel_name",
            },
            {
                Name:        "proxy",
                Label:       "HTTP 代理（可选）",
                Type:        "url",
                Required:    false,
                Placeholder: "http://127.0.0.1:7890",
            },
        },
    },
}
```

同时删除文件中 TODO 注释里的 Telegram 占位代码（第 76-83 行）。

---

## 八、前端配置字段定义

### 8.1 前端无需代码修改

前端通过 `GET /channel-types` API 动态获取 `config_fields`，使用 `v-for` 渲染表单字段。新增 Telegram 渠道类型后，前端会自动渲染以下表单：

| 字段名 | 标签 | 输入类型 | 必填 | 占位符 |
|--------|------|----------|------|--------|
| `bot_token` | Bot Token | password | 是 | 1234567890:ABCdefGHIjklMNOpqrsTUVwxyz |
| `chat_id` | Chat ID | text | 是 | -1001234567890 或 @channel_name |
| `proxy` | HTTP 代理（可选） | text | 否 | http://127.0.0.1:7890 |

### 8.2 前端渲染效果

创建渠道页面选择 Telegram 后，表单区域将自动展示：

```
┌─────────────────────────────────────────────┐
│ 渠道名称                                     │
│ [                                    ]       │
├─────────────────────────────────────────────┤
│ Bot Token *                                  │
│ [************************************]  (password 输入框) │
│                                              │
│ Chat ID *                                    │
│ [-1001234567890 或 @channel_name  ]          │
│                                              │
│ HTTP 代理（可选）                              │
│ [http://127.0.0.1:7890            ]          │
└─────────────────────────────────────────────┘
```

### 8.3 ConfigField type 映射

后端 `ConfigField.Type` 与前端输入控件的映射关系（已有逻辑，无需修改）：

| 后端 type | 前端控件 | 说明 |
|-----------|---------|------|
| `string` | `<input type="text">` | 普通文本 |
| `password` | `<input type="password">` | 密码输入 |
| `url` | `<input type="text">` | URL 输入 |
| `number` | `<input type="text">` | 数字输入 |

> 注意：`proxy` 字段后端 type 设为 `url`（而非 `string`），语义上更准确，前端渲染为普通文本输入框。

---

## 九、安全考量

### 9.1 Bot Token 保护

| 风险 | 措施 |
|------|------|
| 日志泄露 | `maskToken()` 脱敏，仅显示 `1234...wxyz` |
| API 响应泄露 | 错误消息中不包含完整 Token |
| 数据库存储 | Token 存储在 channels.config JSON 字段中，与飞书 webhook_url 同级，遵循现有安全策略 |
| 前端展示 | 使用 `password` 类型输入框，编辑时回显为掩码 |

### 9.2 代理地址保护

| 风险 | 措施 |
|------|------|
| 代理认证信息泄露 | 日志中仅记录 `has_proxy: true/false`，不记录代理地址详情 |
| 代理 URL 中的用户名密码 | `http://user:pass@host:port` 格式中的认证信息不应出现在日志中 |
| SSRF 风险 | 代理地址由用户自己配置，指向用户自己的代理服务，不存在 SSRF 风险 |

### 9.3 请求 URL 安全

Telegram API URL 中包含 Bot Token（`https://api.telegram.org/bot{token}/sendMessage`），在日志中记录请求 URL 时必须使用脱敏后的 Token：

```go
// 正确
zap.String("url", telegramAPIBaseURL+maskToken(cfg.BotToken)+"/sendMessage")

// 错误 - 严禁
zap.String("url", apiURL)  // apiURL 包含完整 Token
```

---

## 十、边界情况处理

### 10.1 消息长度超限

| 场景 | 处理 |
|------|------|
| 标题 + 内容 > 4096 字符 | 截断并追加 `[消息已截断]` 提示 |
| 标题本身 > 4096 字符 | 极端情况，截断标题部分 |
| HTML 标签占用字符 | `<b>[]</b>\n` 共 11 字符，在可用长度内扣除 |

### 10.2 Chat ID 格式

| 格式 | 示例 | 说明 |
|------|------|------|
| 数字（私有聊天） | `-1001234567890` | 负数表示群组/频道 |
| 数字（用户聊天） | `123456789` | 正数表示私有聊天 |
| @用户名（公开频道） | `@my_channel` | 仅公开频道可用 |

`chat_id` 字段使用 `string` 类型（而非 `int64`），因为：
1. Telegram API 接受字符串格式的 chat_id
2. 支持 `@channel_name` 格式
3. 避免大数字精度丢失（JavaScript 中 Number 精度问题）

### 10.3 Bot 未启动对话

| 场景 | Telegram 响应 | 处理 |
|------|-------------|------|
| 用户未向 Bot 发送过 /start | `{"ok": false, "error_code": 403, "description": "Forbidden: bot can't initiate conversation with a user"}` | 返回错误，提示用户先向 Bot 发送消息 |
| Bot 被移出群组 | `{"ok": false, "error_code": 403, "description": "Forbidden: bot was kicked from the supergroup chat"}` | 返回错误，提示用户重新添加 Bot |
| Chat ID 不存在 | `{"ok": false, "error_code": 400, "description": "Bad Request: chat not found"}` | 返回错误，提示检查 Chat ID |

### 10.4 代理连接失败

| 场景 | 处理 |
|------|------|
| 代理地址格式错误 | `newHTTPClient()` 中 `url.Parse()` 失败，返回 "解析代理地址失败" |
| 代理服务不可达 | `client.Do()` 返回网络错误，返回 "发送 Telegram 请求失败" |
| 代理认证失败 | 代理返回 407，`client.Do()` 返回错误 |

### 10.5 网络超时

| 场景 | 处理 |
|------|------|
| 直连超时（中国大陆无代理） | 30 秒超时后返回错误，用户应配置代理 |
| 代理连接超时 | 30 秒超时后返回错误 |
| Telegram API 响应慢 | 30 秒超时后返回错误 |

### 10.6 HTML 解析错误

| 场景 | 处理 |
|------|------|
| 消息内容包含未转义的 `<` `>` `&` | Telegram API 返回 `Bad Request: can't parse entities` |
| 预防措施 | 对用户内容中的 HTML 特殊字符进行转义 |

**HTML 转义函数：**

```go
// escapeHTML 转义 HTML 特殊字符，防止 Telegram 解析错误
func escapeHTML(s string) string {
    s = strings.ReplaceAll(s, "&", "&amp;")
    s = strings.ReplaceAll(s, "<", "&lt;")
    s = strings.ReplaceAll(s, ">", "&gt;")
    return s
}
```

消息构造时对 title 和 content 进行转义：

```go
text := fmt.Sprintf("<b>[%s]</b>\n%s", escapeHTML(title), escapeHTML(content))
```

---

## 十一、UML 活动图

### 11.1 Telegram 推送活动图

```
@startuml OctoTify_Activity_TelegramSend

skinparam activityStyle round
skinparam linetype ortho

title Telegram 渠道推送活动图

start

:MessageService 调用 TelegramSender.Send();

:解析渠道配置 JSON;
note right
  {
    "bot_token": "1234...wxyz",
    "chat_id": "-100xxx",
    "proxy": "http://127.0.0.1:7890"
  }
end note

if [配置解析失败?] then (是)
  :返回错误 "解析 Telegram 渠道配置失败";
  stop
else (否)
  :提取 botToken, chatID, proxy;
endif

if [bot_token 为空?] then (是)
  :返回错误 "Bot Token 不能为空";
  stop
elseif [chat_id 为空?] then (是)
  :返回错误 "Chat ID 不能为空";
  stop
else (否)
endif

:转义 HTML 特殊字符;
:构造消息文本: "<b>[title]</b>\ncontent";

if [消息长度 > 4096?] then (是)
  :截断消息并追加 "[消息已截断]";
else (否)
endif

:构造 telegramMessage 请求体;
note right
  {
    "chat_id": "-100xxx",
    "text": "<b>[title]</b>\ncontent",
    "parse_mode": "HTML"
  }
end note

if [proxy 不为空?] then (是)
  :解析代理 URL;
  if [代理 URL 格式错误?] then (是)
    :返回错误 "解析代理地址失败";
    stop
  else (否)
    :创建带代理的 HTTP Client;
  endif
else (否)
  :创建直连 HTTP Client;
endif

:构造 HTTP POST 请求;
note right
  URL: https://api.telegram.org/bot{token}/sendMessage
  Header: Content-Type: application/json
  Timeout: 30s
end note

:发送请求到 Telegram API;

if [网络错误?] then (是)
  :返回错误 "发送 Telegram 请求失败";
  stop
else (否)
endif

:读取响应体;

if [HTTP 状态码 != 200?] then (是)
  :返回错误 "Telegram 返回 HTTP 错误";
  stop
else (否)
endif

:解析响应 JSON;

if [ok == false?] then (是)
  :根据 error_code 记录日志;
  :返回错误 "Telegram 推送失败: description";
  stop
else (否)
  :推送成功;
endif

stop

@enduml
```

---

## 十二、测试策略

### 12.1 单元测试

| 测试用例 | 输入 | 预期输出 |
|----------|------|----------|
| 配置解析失败 | `{"invalid": json}` | 错误 "解析 Telegram 渠道配置失败" |
| Bot Token 为空 | `{"chat_id": "123"}` | 错误 "Bot Token 不能为空" |
| Chat ID 为空 | `{"bot_token": "xxx"}` | 错误 "Chat ID 不能为空" |
| 代理地址格式错误 | `{"bot_token": "x", "chat_id": "1", "proxy": "://invalid"}` | 错误 "解析代理地址失败" |
| 消息截断 | 超长 content | 截断后追加 "[消息已截断]" |
| HTML 转义 | `<script>alert(1)</script>` | `&lt;script&gt;alert(1)&lt;/script&gt;` |
| Token 脱敏 | `1234567890:ABCdef` | `1234...Cdef` |
| 推送成功 | 有效配置 + Mock 200 OK | nil |
| 推送失败（业务） | 有效配置 + Mock ok=false | 错误含 description |
| 推送失败（HTTP） | 有效配置 + Mock 500 | 错误含 HTTP 状态码 |

### 12.2 集成测试

| 测试场景 | 方法 |
|----------|------|
| 直连推送 | 使用真实 Bot Token 和 Chat ID |
| 代理推送 | 配置本地代理后推送 |
| 频率限制 | 短时间内多次推送，验证 429 处理 |

---

## 十三、任务分配建议

| 任务 | 角色 | 文件 | 预估工时 |
|------|------|------|----------|
| 实现 TelegramSender | 后端开发 | `backend/internal/sender/telegram.go` | 2h |
| 修改 factory.go 注册 | 后端开发 | `backend/internal/sender/factory.go` | 0.5h |
| 修改 channel_dto.go 元数据 | 后端开发 | `backend/internal/handler/dto/channel_dto.go` | 0.5h |
| 编写单元测试 | 后端开发 | `backend/internal/sender/telegram_test.go` | 2h |
| 前端验证（动态表单渲染） | 前端开发 | 无代码修改，仅需验证 | 0.5h |
| 端到端测试 | 测试 | - | 1h |

**总计预估：6.5h**

---

## 十四、风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 中国大陆直连 Telegram 超时 | 推送失败 | 高 | 提供 proxy 配置项，日志提示用户配置代理 |
| Bot Token 泄露 | 安全风险 | 低 | 日志脱敏、password 类型输入框 |
| 消息超 4096 字符 | 内容丢失 | 中 | 截断 + 提示，未来可考虑分片 |
| SOCKS5 代理需求 | 功能缺失 | 中 | 首期仅支持 HTTP 代理，后续迭代添加 |
| Telegram API 变更 | 推送失败 | 低 | 使用稳定的 Bot API v1，关注官方变更 |

---

## 版本历史

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-07 | 1.0 | 初始版本，Telegram Sender 设计文档 |
