# 微信 ClawBot 绑定凭证加密传输方案

> 文档版本: v1.0  
> 创建日期: 2026-05-08  
> 状态: 待评审  

---

## 目录

1. [背景与目标](#1-背景与目标)
2. [方案架构图](#2-方案架构图)
3. [加密算法和技术选型](#3-加密算法和技术选型)
4. [后端接口变更](#4-后端接口变更)
5. [前端交互流程](#5-前端交互流程)
6. [安全边界分析](#6-安全边界分析)
7. [异常场景处理](#7-异常场景处理)
8. [前后端开发任务清单](#8-前后端开发任务清单)
9. [数据库变更](#9-数据库变更)

---

## 1. 背景与目标

### 1.1 当前问题

在现有的微信 ClawBot 绑定流程中，BotToken 等敏感凭证的传输路径如下：

```
iLink API → 后端服务 → 明文返回给前端 → 前端存入 form.config → 提交创建渠道时再次明文传输 → 存入数据库
```

**安全风险：**

- BotToken 在后端返回给前端时为明文，可通过浏览器 DevTools 或网络抓包直接查看
- 前端内存中存储明文凭证，存在 XSS 攻击导致凭证泄漏的风险
- 前端提交创建渠道时，凭证再次以明文在网络中传输
- 数据库中存储明文凭证，若数据库被入侵则凭证直接暴露

### 1.2 设计目标

| 目标 | 说明 |
|------|------|
| 传输加密 | BotToken 从后端返回前端前完成加密，网络传输全程密文 |
| 前端不透出 | 前端仅负责透传密文，不解密、不展示明文 |
| 后端解密存储 | 后端收到密文后解密，以密文形式存入数据库 |
| 密钥安全 | AES 密钥仅驻留服务器内存，不写入磁盘，重启后自动重新生成 |
| 向后兼容 | 非 ClawBot 渠道保持原有逻辑不变 |

### 1.3 核心设计原则

- **前端零知识**：前端无法获知 BotToken 明文内容
- **内存密钥**：AES 密钥仅存在于服务端内存中，重启失效
- **对称加密 + 随机 Nonce**：每次加密使用独立的随机 Nonce
- **AEAD 模式**：使用 AES-GCM 同时提供加密和完整性校验

---

## 2. 方案架构图

### 2.1 整体数据流向

```
┌─────────────────────────────────────────────────────────────────────┐
│                        绑定流程数据流                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌──────────┐    ① GET /get_bot_qrcode    ┌──────────┐            │
│   │  iLink   │◄───────────────────────────►│  Backend │            │
│   │   API    │    ② GET /get_qrcode_status │  Server  │            │
│   └──────────┘   (含明文 BotToken)          └────┬─────┘            │
│                                                   │                 │
│                                     ③ AES-256-GCM 加密凭证          │
│                                     key = 内存中的 32 字节密钥      │
│                                                   │                 │
│                                                   ▼                 │
│                                          ┌────────────────┐         │
│                                   GET    │  加密后的凭证   │         │
│                              ┌──────────►│  {              │         │
│                              │           │   ciphertext,   │         │
│                              │           │   nonce,        │         │
│                              │           │   tag           │         │
│                              │           │  }              │         │
│                              │           └────────┬───────┘         │
│                              │                    │                 │
│                    ┌─────────┴─────┐              │ HTTPS           │
│                    │    Browser    │◄─────────────┘                 │
│                    │  (前端页面)   │                                │
│                    │              │                                │
│                    │  表单显示:    │                                │
│                    │  BotToken: ** │                                │
│                    │  提交时透传:  │                                │
│                    │  原样回传密文 │                                │
│                    └───────┬───────┘                                │
│                            │ POST /channels (创建渠道)              │
│                            │ config.bot_token_ciphertext = 密文     │
│                            │                                       │
│                            ▼                                       │
│                     ┌──────────────┐                               │
│                     │   Backend    │                               │
│                     │   Server     │                               │
│                     │              │                               │
│                     │ ④ 解密密文   │                               │
│                     │ ⑤ 密文存入DB │                               │
│                     └──────┬───────┘                               │
│                            │                                       │
│                            ▼                                       │
│                     ┌──────────────┐                               │
│                     │  Database    │                               │
│                     │  (密文存储)   │                               │
│                     └──────────────┘                               │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 加密解密时序图

```
┌─────────┐    ┌──────────────┐    ┌───────┐    ┌──────────┐
│  Frontend│    │ ChannelHandler│    │Channel │    │  iLink   │
│          │    │              │    │Service │    │   API    │
└────┬─────┘    └──────┬───────┘    └───┬───┘    └────┬─────┘
     │                 │                │              │
     │  POST /bind     │                │              │
     │────────────────►│                │              │
     │                 │  GET /get_bot_qrcode           │
     │                 │───────────────►│              │
     │                 │                │              │
     │                 │  QR Code       │              │
     │                 │◄───────────────│              │
     │  QR + bind_id   │                │              │
     │◄────────────────│                │              │
     │                 │                │              │
     │  GET /bind/:id/status            │              │
     │────────────────►│                │              │
     │                 │  GET /get_qrcode_status        │
     │                 │───────────────►│              │
     │                 │                │              │
     │                 │  Status + plaintext creds      │
     │                 │◄───────────────│              │
     │                 │                │              │
     │                 │ [AES-256-GCM 加密]             │
     │                 │ key = server AES key           │
     │                 │ nonce = random 12 bytes        │
     │                 │                │              │
     │  ciphertext + nonce             │              │
     │◄────────────────│                │              │
     │                 │                │              │
     │ [前端存储密文, 不解密]            │              │
     │                 │                │              │
     │  POST /channels (创建渠道)       │              │
     │  config: {                       │              │
     │    bot_token_ciphertext: "...",  │              │
     │    bot_token_nonce: "...",       │              │
     │    ilink_bot_id: "...",          │              │
     │    ilink_user_id: "..."          │              │
     │  }                               │              │
     │────────────────►│                │              │
     │                 │                │              │
     │                 │ [检测 ClawBot 类型]            │
     │                 │ [解密密文 → 明文 BotToken]     │
     │                 │ [密文存入数据库]               │
     │                 │                │              │
     │  Channel DTO    │                │              │
     │◄────────────────│                │              │
```

### 2.3 消息发送时的解密流程

```
┌─────────────┐    ┌──────────────┐    ┌──────────────┐
│  Push API   │    │ MessageService│    │ WechatClawbot │
│  (来源推送)  │    │              │    │   Sender      │
└──────┬──────┘    └──────┬───────┘    └──────┬───────┘
       │                  │                   │
       │ POST /push/:token│                   │
       │─────────────────►│                   │
       │                  │                   │
       │                  │ 查询渠道配置       │
       │                  │ (密文从 DB 读取)    │
       │                  │                   │
       │                  │ [AES-256-GCM 解密]  │
       │                  │ key = server AES key│
       │                  │ nonce = 存储的 nonce │
       │                  │                   │
       │                  │ 明文 BotToken      │
       │                  │───────────────────►│
       │                  │                   │
       │                  │  iLink API 调用    │
       │                  │───────────────────►│
       │                  │                   │
       │  Response        │                   │
       │◄─────────────────│                   │
```

---

## 3. 加密算法和技术选型

### 3.1 算法选择: AES-256-GCM

| 特性 | 说明 |
|------|------|
| 算法 | AES (Advanced Encryption Standard) |
| 密钥长度 | 256 bit (32 字节) |
| 工作模式 | GCM (Galois/Counter Mode) |
| Nonce 长度 | 12 字节 (96 bit，GCM 推荐长度) |
| Tag 长度 | 16 字节 (128 bit，默认) |
| 标准 | NIST SP 800-38D |

### 3.2 为什么选择 AES-256-GCM

**相比其他方案的对比：**

| 方案 | 加密 | 完整性校验 | 性能 | 适用性 |
|------|------|-----------|------|--------|
| AES-256-CBC | 是 | 否（需额外 HMAC） | 中 | 不推荐 |
| AES-256-CTR | 是 | 否（需额外 HMAC） | 高 | 不推荐 |
| **AES-256-GCM** | **是** | **是（内建）** | **高** | **推荐** |
| ChaCha20-Poly1305 | 是 | 是 | 高（ARM 优） | 备选 |
| RSA 非对称 | 是 | 否 | 低 | 不适合批量数据 |

**选择理由：**

1. **AEAD 特性**：GCM 模式同时提供加密和认证，无需额外的 HMAC 计算
2. **Go 标准库原生支持**：`crypto/cipher` 包提供 `NewGCM` 函数，无需第三方依赖
3. **性能优秀**：硬件加速支持（AES-NI），现代 CPU 上性能极佳
4. **行业标准**：TLS 1.3、Signal Protocol 等广泛使用
5. **随机 Nonce 安全**：2^64 次加密后才有碰撞风险，对本项目绰绰有余

### 3.3 密钥管理策略

```
┌─────────────────────────────────────────────┐
│              AES 密钥生命周期                │
├─────────────────────────────────────────────┤
│                                             │
│  Server Startup                             │
│       │                                     │
│       ▼                                     │
│  ┌──────────────┐                           │
│  │ crypto/rand   │                          │
│  │ Generate 32B  │                          │
│  │ AES Key       │                          │
│  └──────┬───────┘                           │
│         │                                   │
│         ▼                                   │
│  ┌──────────────┐                           │
│  │ Memory Only   │◄── 所有加密/解密操作使用  │
│  │ (no disk)     │                           │
│  └──────┬───────┘                           │
│         │                                   │
│         ▼                                   │
│  ┌──────────────┐                           │
│  │ Singleton     │                          │
│  │ Instance      │                          │
│  └──────┬───────┘                           │
│         │                                   │
│    Server Shutdown / Restart                │
│         │                                   │
│         ▼                                   │
│  ┌──────────────┐                           │
│  │ Key Destroyed │◄── GC 自动回收            │
│  │ (lost)        │                           │
│  └──────────────┘                           │
│                                             │
└─────────────────────────────────────────────┘
```

**关键设计决策：**

| 决策 | 选择 | 理由 |
|------|------|------|
| 密钥持久化 | 不持久化 | 即使密钥丢失，数据库中只有密文，攻击者无法解密 |
| 密钥来源 | `crypto/rand` | 密码学安全的随机数生成器 |
| 密钥生命周期 | 服务进程生命周期 | 重启后所有历史密文无法解密（可接受） |
| 多实例部署 | 需要密钥共享 | 生产环境需通过密钥管理服务同步（见下方说明） |

**多实例部署方案（未来扩展）：**

当系统需要水平扩展时，AES 密钥需要通过以下方式共享：

```
方案 A: 环境变量注入
  ┌─────────────┐
  │  KMS / Vault │
  └──────┬──────┘
         │ 启动时拉取
         ▼
  ┌─────────────┐     ┌─────────────┐
  │  Instance 1  │     │  Instance 2  │
  │  (same key)  │     │  (same key)  │
  └─────────────┘     └─────────────┘

方案 B: 数据库加密存储（密钥加密密钥）
  ┌─────────────┐    ┌──────────────┐
  │  Master Key  │───►│ AES Key (加密)│
  │  (HSM/Vault) │    │  in DB       │
  └─────────────┘    └──────────────┘
```

---

## 4. 后端接口变更

### 4.1 新增模块: `pkg/aescipher`

**文件路径**: `/backend/pkg/aescipher/aescipher.go`

```go
package aescipher

// AESEncryptor AES-256-GCM 加密器
type AESEncryptor struct {
    key []byte  // 32 字节 AES 密钥
}

// NewAESEncryptor 创建加密器实例
func NewAESEncryptor() (*AESEncryptor, error)

// Encrypt 加密数据，返回 (ciphertext, nonce, error)
func (e *AESEncryptor) Encrypt(plaintext []byte) ([]byte, []byte, error)

// Decrypt 解密数据
func (e *AESEncryptor) Decrypt(ciphertext, nonce []byte) ([]byte, error)
```

**加密输出格式**：

- `ciphertext`: AES-GCM 加密后的密文（不含 Nonce 和 Tag）
- `nonce`: 12 字节随机 Nonce
- 实际返回给前端时，将 `ciphertext + nonce` 组合为 Base64 字符串传输

### 4.2 修改 `BindCredentials` 结构

**文件**: `/backend/internal/service/bind_store.go`

```go
// 变更前
type BindCredentials struct {
    BotToken    string `json:"bot_token"`
    IlinkBotID  string `json:"ilink_bot_id"`
    IlinkUserID string `json:"ilink_user_id"`
}

// 变更后
type BindCredentials struct {
    // 加密后的 BotToken（Base64 编码的 ciphertext+nonce）
    BotTokenCiphertext string `json:"bot_token_ciphertext"`
    BotTokenNonce      string `json:"bot_token_nonce"`
    // IlinkBotID 和 IlinkUserID 不加密（非敏感标识符）
    IlinkBotID         string `json:"ilink_bot_id"`
    IlinkUserID        string `json:"ilink_user_id"`
}
```

### 4.3 修改 `GetBindStatus` 接口响应

**接口**: `GET /api/channels/wechat-clawbot/bind/:bindId/status`

**响应变更**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "confirmed",
    "credentials": {
      "bot_token_ciphertext": "YWJjZGVm...（Base64 编码的密文）",
      "bot_token_nonce": "eHl6MTIz...（Base64 编码的 Nonce）",
      "ilink_bot_id": "wx1234567890",
      "ilink_user_id": "user_abc123"
    }
  }
}
```

**变更说明**：

| 字段 | 变更前 | 变更后 | 说明 |
|------|--------|--------|------|
| `credentials.bot_token` | 明文字符串 | **移除** | 不再返回明文 |
| `credentials.bot_token_ciphertext` | 无 | **新增** | AES-256-GCM 加密后的密文（Base64） |
| `credentials.bot_token_nonce` | 无 | **新增** | 加密时使用的随机 Nonce（Base64） |
| `credentials.ilink_bot_id` | 明文 | 保持不变 | 非敏感标识符 |
| `credentials.ilink_user_id` | 明文 | 保持不变 | 非敏感标识符 |

### 4.4 修改 `CreateChannel` 接口请求

**接口**: `POST /api/channels`

**请求体变更（仅 wechat_clawbot 类型）**：

```json
{
  "type": "wechat_clawbot",
  "name": "我的微信机器人",
  "config": {
    "bot_token_ciphertext": "YWJjZGVm...（前端透传的密文）",
    "bot_token_nonce": "eHl6MTIz...（前端透传的 nonce）",
    "ilink_bot_id": "wx1234567890",
    "ilink_user_id": "user_abc123"
  }
}
```

**处理逻辑**：

```
1. 检测渠道类型是否为 wechat_clawbot
2. 如果是，从 config 中读取 bot_token_ciphertext 和 bot_token_nonce
3. 使用内存中的 AES 密钥解密密文，得到明文 BotToken
4. 用明文 BotToken 构造 wechatClawbotConfig 对象
5. 将密文（而非明文）存入数据库的 config 字段
6. 返回渠道信息
```

### 4.5 修改 `WechatClawbotSender.Send` 方法

**文件**: `/backend/internal/sender/wechat_clawbot.go`

```go
func (s *WechatClawbotSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
    // 1. 解析渠道配置
    var cfg wechatClawbotConfig
    if err := json.Unmarshal(config, &cfg); err != nil {
        return fmt.Errorf("解析微信ClawBot渠道配置失败: %w", err)
    }

    // 2. 检测配置格式：如果是加密格式，先解密
    if cfg.BotTokenCiphertext != "" && cfg.BotTokenNonce != "" {
        plaintext, err := aescipher.GlobalDecrypt(
            cfg.BotTokenCiphertext,
            cfg.BotTokenNonce,
        )
        if err != nil {
            return fmt.Errorf("解密BotToken失败: %w", err)
        }
        cfg.BotToken = string(plaintext)
    }

    // 3. 后续逻辑不变...
}
```

**配置结构变更**：

```go
// 变更前
type wechatClawbotConfig struct {
    BotToken    string `json:"bot_token"`
    ILinkBotID  string `json:"ilink_bot_id"`
    ILinkUserID string `json:"ilink_user_id"`
}

// 变更后（兼容两种格式）
type wechatClawbotConfig struct {
    // 明文格式（旧数据兼容）
    BotToken string `json:"bot_token"`
    // 加密格式（新数据）
    BotTokenCiphertext string `json:"bot_token_ciphertext"`
    BotTokenNonce      string `json:"bot_token_nonce"`
    // 共用字段
    ILinkBotID  string `json:"ilink_bot_id"`
    ILinkUserID string `json:"ilink_user_id"`
}
```

### 4.6 服务器启动时初始化 AES 密钥

**文件**: `/backend/internal/server/server.go`

```go
func (s *Server) initDependencies(cfg *config.Config, db *gorm.DB, logger *zap.Logger) {
    // ... 现有初始化逻辑 ...

    // 初始化 AES 加密器
    encryptor, err := aescipher.NewAESEncryptor()
    if err != nil {
        logger.Fatal("初始化 AES 加密器失败", zap.Error(err))
    }
    aescipher.SetGlobalEncryptor(encryptor)
    
    logger.Info("AES-256-GCM 加密器已初始化")

    // ... 后续初始化逻辑 ...
}
```

### 4.7 新增错误码

**文件**: `/backend/pkg/xerr/codes.go`

```go
// 微信ClawBot加密模块错误 1109XX (扩展)
var (
    // 现有错误码...
    ErrBindSessionNotFound = New(110900, "绑定会话不存在或已过期")
    ErrBindQRCodeFailed    = New(110901, "获取绑定二维码失败")
    ErrBindStatusFailed    = New(110902, "查询绑定状态失败")
    ErrBindExpired         = New(110903, "绑定二维码已过期")
    
    // 新增加密相关错误码
    ErrBindEncryptFailed   = New(110904, "凭证加密失败")
    ErrBindDecryptFailed   = New(110905, "凭证解密失败")
)
```

### 4.8 数据库存储格式

**渠道 config 字段（JSON 类型）存储内容**：

```json
{
  "bot_token_ciphertext": "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY=",
  "bot_token_nonce": "eHl6MTIzNDU2Nzg5MGFiYw==",
  "ilink_bot_id": "wx1234567890",
  "ilink_user_id": "user_abc123"
}
```

**注意**：数据库不存储明文 `bot_token`，仅存储加密后的密文。

---

## 5. 前端交互流程

### 5.1 前端绑定流程（用户视角）

```
┌─────────────────────────────────────────────┐
│              用户操作流程                     │
├─────────────────────────────────────────────┤
│                                             │
│  1. 用户进入"创建推送渠道"页面               │
│  2. 选择"微信ClawBot"渠道类型                │
│  3. 点击"开始绑定"按钮                       │
│  4. 页面显示二维码                           │
│  5. 用户使用微信扫描二维码                   │
│  6. 等待扫码确认（前端每3秒轮询状态）         │
│  7. 扫码成功后，显示"绑定成功"               │
│     - 表单中的 BotToken 字段显示为 ****      │
│     - 用户无法查看或修改明文                 │
│  8. 填写渠道名称，点击"创建"                │
│  9. 提交表单时，密文原样传回后端            │
│ 10. 创建成功，跳转到渠道列表                │
│                                             │
└─────────────────────────────────────────────┘
```

### 5.2 前端凭证处理流程

```
┌─────────────────────────────────────────────┐
│            前端数据处理流程                  │
├─────────────────────────────────────────────┤
│                                             │
│  接收绑定成功响应:                           │
│  ┌───────────────────────────────────────┐  │
│  │ {                                     │  │
│  │   "status": "confirmed",              │  │
│  │   "credentials": {                    │  │
│  │     "bot_token_ciphertext": "...",    │  │
│  │     "bot_token_nonce": "...",         │  │
│  │     "ilink_bot_id": "wx123",          │  │
│  │     "ilink_user_id": "user_abc"       │  │
│  │   }                                   │  │
│  │ }                                     │  │
│  └───────────────────────────────────────┘  │
│                    │                        │
│                    ▼                        │
│  填充到表单 config:                          │
│  ┌───────────────────────────────────────┐  │
│  │ form.config = {                       │  │
│  │   bot_token_ciphertext: "...",        │  │
│  │   bot_token_nonce: "...",             │  │
│  │   ilink_bot_id: "wx123",              │  │
│  │   ilink_user_id: "user_abc"           │  │
│  │ }                                     │  │
│  └───────────────────────────────────────┘  │
│                    │                        │
│                    ▼                        │
│  表单显示（BotToken 字段）:                  │
│  ┌───────────────────────────────────────┐  │
│  │ Bot Token: ********                   │  │
│  │ (不可编辑，仅展示为掩码)               │  │
│  └───────────────────────────────────────┘  │
│                    │                        │
│                    ▼                        │
│  提交表单时:                                 │
│  ┌───────────────────────────────────────┐  │
│  │ POST /api/channels                    │  │
│  │ {                                     │  │
│  │   "type": "wechat_clawbot",           │  │
│  │   "name": "我的微信机器人",            │  │
│  │   "config": {                         │  │
│  │     "bot_token_ciphertext": "...",    │  │
│  │     "bot_token_nonce": "...",         │  │
│  │     "ilink_bot_id": "wx123",          │  │
│  │     "ilink_user_id": "user_abc"       │  │
│  │   }                                   │  │
│  │ }                                     │  │
│  └───────────────────────────────────────┘  │
│                                             │
└─────────────────────────────────────────────┘
```

### 5.3 前端代码变更点

**文件**: `/frontend/src/api/channels.ts`

```typescript
// 变更前
export interface BindCredentials {
  bot_token: string
  ilink_bot_id: string
  ilink_user_id: string
}

// 变更后
export interface BindCredentials {
  bot_token_ciphertext: string  // 新增：加密后的密文
  bot_token_nonce: string       // 新增：加密 nonce
  ilink_bot_id: string
  ilink_user_id: string
}
```

**文件**: `/frontend/src/views/channels/ChannelCreateView.vue`

```vue
<!-- 变更前：绑定成功后填充明文 -->
<!-- 
if (res.data.credentials) {
  form.value.config = {
    bot_token: res.data.credentials.bot_token || '',
    ilink_bot_id: res.data.credentials.ilink_bot_id || '',
    ilink_user_id: res.data.credentials.ilink_user_id || '',
  }
}
-->

<!-- 变更后：绑定成功后填充密文 -->
<script setup lang="ts">
// 绑定成功后的处理逻辑
watch(bindStatus, (newStatus) => {
  if (newStatus === 'confirmed') {
    stopPolling()
    stopCountdown()
    showSuccess('绑定成功！请继续填写渠道信息', 2000)
  }
})

// 填充配置时，使用密文字段
async function startBind() {
  // ... 绑定逻辑不变 ...
}

// 轮询获取绑定状态
function startPolling() {
  stopPolling()
  bindPollingTimer.value = window.setInterval(async () => {
    try {
      const res = await getWechatClawbotBindStatus(bindID.value)
      if (res.data) {
        bindStatus.value = res.data.status
        // 绑定成功后，填充密文到表单配置
        if (res.data.credentials) {
          form.value.config = {
            bot_token_ciphertext: res.data.credentials.bot_token_ciphertext || '',
            bot_token_nonce: res.data.credentials.bot_token_nonce || '',
            ilink_bot_id: res.data.credentials.ilink_bot_id || '',
            ilink_user_id: res.data.credentials.ilink_user_id || '',
          }
        }
      }
    } catch (err) {
      console.error('查询绑定状态失败', err)
    }
  }, 3000)
}
</script>
```

**表单显示变更**：

```vue
<!-- 微信ClawBot 渠道的 BotToken 字段显示为掩码 -->
<div class="form-group">
  <label>Bot Token</label>
  <input 
    type="password" 
    value="********" 
    readonly 
    disabled
    class="credential-masked"
  />
  <span class="credential-hint">已加密保护，用户无法查看</span>
</div>
```

### 5.4 前端类型定义变更

**文件**: `/frontend/src/types/api.ts`

```typescript
// 渠道配置类型定义
export interface WechatClawbotConfig {
  bot_token_ciphertext: string  // 加密后的密文
  bot_token_nonce: string       // 加密 nonce
  ilink_bot_id: string
  ilink_user_id: string
}
```

---

## 6. 安全边界分析

### 6.1 信任边界

```
┌────────────────────────────────────────────────────────────┐
│                     信任边界图                              │
├────────────────────────────────────────────────────────────┤
│                                                            │
│   ┌─────────────────────────────────────────────────┐      │
│   │              不可信区域 (Untrusted)              │      │
│   │                                                 │      │
│   │  ┌──────────┐         ┌──────────┐             │      │
│   │  │  Browser │         │  Network │             │      │
│   │  │  (前端)  │◄───────►│  (HTTPS) │             │      │
│   │  │          │         │          │             │      │
│   │  │  仅持有:  │         │  传输:   │             │      │
│   │  │  - 密文   │         │  - 密文   │             │      │
│   │  │  - Nonce │         │  - Nonce │             │      │
│   │  │  - 标识符 │         │  - 标识符 │             │      │
│   │  └──────────┘         └──────────┘             │      │
│   └──────────────────────┬─────────────────────────┘      │
│                          │                                 │
│                    信任边界 (Trust Boundary)               │
│                          │                                 │
│   ┌──────────────────────┴─────────────────────────┐      │
│   │              可信区域 (Trusted)                 │      │
│   │                                                 │      │
│   │  ┌──────────────┐         ┌──────────────┐     │      │
│   │  │   Backend     │         │   Database   │     │      │
│   │  │   Server      │         │              │     │      │
│   │  │               │         │  存储:       │     │      │
│   │  │  持有:        │         │  - 密文      │     │      │
│   │  │  - AES Key   │────────►│  - Nonce     │     │      │
│   │  │  (内存)      │         │  - 标识符     │     │      │
│   │  │               │         │              │     │      │
│   │  │  可执行:      │         │  不存储:     │     │      │
│   │  │  - 加密      │         │  - 明文      │     │      │
│   │  │  - 解密      │         │  BotToken    │     │      │
│   │  └──────────────┘         └──────────────┘     │      │
│   │                                                 │      │
│   └─────────────────────────────────────────────────┘      │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### 6.2 安全分析矩阵

| 攻击场景 | 风险等级 | 防护措施 | 剩余风险 |
|---------|---------|---------|---------|
| 网络嗅探 | **低** | HTTPS 传输 + AES 加密 | 极低（双重保护） |
| XSS 攻击 | **低** | 前端不持有明文 | 中（可能获取密文但无法解密） |
| CSRF 攻击 | **低** | JWT 认证 + SameSite Cookie | 低 |
| 数据库泄漏 | **中** | 密文存储 + 密钥内存化 | 中（需配合密钥获取才能解密） |
| 服务器内存 dump | **高** | 密钥仅内存中，重启失效 | 高（运行时可被提取） |
| 中间人攻击 | **极低** | HTTPS + 证书校验 | 极低 |
| 重放攻击 | **低** | 每次加密使用随机 Nonce | 低（密文不可重用） |

### 6.3 安全增强建议

**当前方案已满足基本安全需求，以下为进一步增强建议（可选实施）：**

| 建议 | 优先级 | 说明 |
|------|--------|------|
| CSP 头 | 高 | 配置 Content-Security-Policy 减少 XSS 风险 |
| 密钥轮换 | 中 | 定期更换 AES 密钥，需重新加密历史数据 |
| 审计日志 | 中 | 记录所有加密/解密操作，便于安全审计 |
| HSM 集成 | 低 | 使用硬件安全模块管理密钥（适合企业级） |

---

## 7. 异常场景处理

### 7.1 服务器重启

**场景描述**：服务器重启后，内存中的 AES 密钥丢失。

**影响分析**：

| 影响范围 | 说明 |
|---------|------|
| 新绑定流程 | 无影响，新密钥生成后正常加密 |
| 已存储的渠道 | 无法解密历史密文，消息推送失败 |
| 前端缓存的密文 | 无法解密，需重新绑定 |

**处理策略**：

```
┌─────────────────────────────────────────────┐
│            服务器重启处理流程                │
├─────────────────────────────────────────────┤
│                                             │
│  1. 服务器启动                               │
│  2. 生成新的 32 字节 AES 密钥               │
│  3. 记录日志："AES 密钥已重新生成"          │
│  4. 历史密文数据无法解密                    │
│  5. 用户发送消息时:                          │
│     a. WechatClawbotSender 尝试解密         │
│     b. 解密失败，返回错误                    │
│     c. 消息状态标记为发送失败                │
│     d. 提示用户重新绑定渠道                  │
│  6. 用户重新绑定:                            │
│     a. 使用新密钥加密新凭证                  │
│     b. 更新渠道配置                         │
│     c. 恢复正常推送                          │
│                                             │
└─────────────────────────────────────────────┘
```

**错误提示**：

当解密失败时，后端返回：

```json
{
  "code": 110905,
  "message": "凭证解密失败，可能是服务器重启导致密钥变更，请重新绑定微信ClawBot"
}
```

### 7.2 网络中断

**场景**：绑定过程中网络中断。

| 阶段 | 中断影响 | 用户操作 |
|------|---------|---------|
| 获取二维码时 | 二维码获取失败 | 点击"重新开始" |
| 轮询状态时 | 状态查询失败 | 前端自动重试（保留现有逻辑） |
| 提交创建时 | 渠道创建失败 | 重新提交，密文不变 |

### 7.3 二维码过期

**场景**：用户 5 分钟内未完成扫码。

| 处理逻辑 | 说明 |
|---------|------|
| 前端提示 | "二维码已过期，请重新开始" |
| 用户操作 | 点击"重新开始"按钮 |
| 后端清理 | 清理过期的 BindSession |

### 7.4 并发绑定

**场景**：用户同时发起多个绑定请求。

| 处理逻辑 | 说明 |
|---------|------|
| BindStore | 基于 bindID 隔离，不会冲突 |
| 前端限制 | 同一时间只允许一个活跃绑定会话 |

### 7.5 数据库异常

**场景**：渠道创建时数据库写入失败。

| 处理逻辑 | 说明 |
|---------|------|
| 事务回滚 | 渠道数据不写入 |
| 前端提示 | "创建失败，请重试" |
| 凭证安全 | 密文未持久化，无泄漏风险 |

### 7.6 加密器未初始化

**场景**：服务器启动时 AES 加密器初始化失败。

| 处理逻辑 | 说明 |
|---------|------|
| 服务启动失败 | 服务器无法正常启动 |
| 日志记录 | 记录详细错误信息 |
| 告警通知 | 触发运维告警 |

---

## 8. 前后端开发任务清单

### 8.1 后端任务

| ID | 任务 | 文件 | 预估工作量 | 优先级 |
|----|------|------|-----------|--------|
| B-1 | 创建 `pkg/aescipher` 包 | `backend/pkg/aescipher/aescipher.go`<br>`backend/pkg/aescipher/aescipher_test.go` | 2h | P0 |
| B-2 | 修改 `BindCredentials` 结构 | `backend/internal/service/bind_store.go` | 0.5h | P0 |
| B-3 | 修改 `GetBindStatus` 响应逻辑 | `backend/internal/handler/channel_handler.go`<br>`backend/internal/service/channel_service.go` | 1h | P0 |
| B-4 | 修改 `CreateChannel` 处理逻辑 | `backend/internal/service/channel_service.go` | 1h | P0 |
| B-5 | 修改 `WechatClawbotSender` 支持解密 | `backend/internal/sender/wechat_clawbot.go` | 1h | P0 |
| B-6 | 服务器启动初始化 AES 密钥 | `backend/internal/server/server.go` | 0.5h | P0 |
| B-7 | 新增加密相关错误码 | `backend/pkg/xerr/codes.go` | 0.5h | P1 |
| B-8 | 更新 Swagger 文档 | `backend/docs/swagger/` | 1h | P1 |
| B-9 | 单元测试覆盖 | 各模块对应测试文件 | 2h | P1 |
| B-10 | 编写迁移脚本（可选） | 迁移脚本 | 1h | P2 |

**总计预估工作量**: 约 9.5 小时

### 8.2 前端任务

| ID | 任务 | 文件 | 预估工作量 | 优先级 |
|----|------|------|-----------|--------|
| F-1 | 更新 `BindCredentials` 类型定义 | `frontend/src/api/channels.ts` | 0.5h | P0 |
| F-2 | 修改绑定成功后的凭证填充逻辑 | `frontend/src/views/channels/ChannelCreateView.vue` | 0.5h | P0 |
| F-3 | 修改 BotToken 字段显示为掩码 | `frontend/src/views/channels/ChannelCreateView.vue` | 0.5h | P0 |
| F-4 | 更新表单验证逻辑 | `frontend/src/views/channels/ChannelCreateView.vue` | 0.5h | P0 |
| F-5 | 更新 API 类型定义 | `frontend/src/types/api.ts` | 0.5h | P1 |

**总计预估工作量**: 约 2.5 小时

### 8.3 测试任务

| ID | 任务 | 预估工作量 | 优先级 |
|----|------|-----------|--------|
| T-1 | 端到端绑定流程测试 | 2h | P0 |
| T-2 | 加密/解密功能测试 | 1h | P0 |
| T-3 | 异常场景测试（重启、网络中断等） | 1.5h | P1 |
| T-4 | 安全测试（XSS、网络抓包等） | 1h | P1 |

**总计预估工作量**: 约 5.5 小时

### 8.4 开发优先级说明

**P0（必须完成）：**
- 核心加密/解密功能
- 接口响应格式变更
- 前端绑定流程适配

**P1（重要但可稍后）：**
- 错误码完善
- 文档更新
- 测试覆盖

**P2（可选）：**
- 历史数据迁移脚本

---

## 9. 数据库变更

### 9.1 当前数据库结构

**渠道表 (channels)**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| user_id | BIGINT | 用户 ID |
| type | VARCHAR(32) | 渠道类型 |
| name | VARCHAR(128) | 渠道名称 |
| config | JSON | 渠道配置 |
| status | TINYINT | 状态 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| last_used_at | DATETIME | 最后使用时间 |

### 9.2 变更后 config 字段内容

**wechat_clawbot 类型的 config 字段**：

```json
{
  "bot_token_ciphertext": "base64 编码的密文",
  "bot_token_nonce": "base64 编码的 nonce",
  "ilink_bot_id": "字符串",
  "ilink_user_id": "字符串"
}
```

**注意**：

- 无需修改数据库表结构（config 字段为 JSON 类型，灵活存储）
- 旧数据的 `bot_token` 明文仍保留，`WechatClawbotSender` 需要兼容处理
- 新数据仅存储密文，不存储明文

### 9.3 数据迁移策略（可选）

如果希望将历史明文数据也转换为密文，可执行以下迁移：

```sql
-- 伪代码，实际需要使用 Go 脚本执行
-- 1. 查询所有 wechat_clawbot 类型的渠道
SELECT id, config FROM channels WHERE type = 'wechat_clawbot';

-- 2. 对每条记录：
--    a. 解析 config 中的 bot_token 明文
--    b. 使用当前 AES 密钥加密
--    c. 更新 config 为密文格式
UPDATE channels 
SET config = '{"bot_token_ciphertext":"...","bot_token_nonce":"...","ilink_bot_id":"...","ilink_user_id":"..."}'
WHERE id = ?;
```

**迁移注意事项**：

| 注意事项 | 说明 |
|---------|------|
| 备份数据 | 迁移前备份数据库 |
| 停机窗口 | 建议在低峰期执行 |
| 回滚方案 | 保留原始明文数据作为备份 |
| 验证 | 迁移后验证所有渠道可正常推送 |

---

## 附录

### A. Go AES-256-GCM 实现参考

```go
package aescipher

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
    "sync"
)

var (
    globalEncryptor *AESEncryptor
    initOnce        sync.Once
)

// AESEncryptor AES-256-GCM 加密器
type AESEncryptor struct {
    key []byte
}

// NewAESEncryptor 创建加密器实例
func NewAESEncryptor() (*AESEncryptor, error) {
    key := make([]byte, 32) // 256 bit
    if _, err := io.ReadFull(rand.Reader, key); err != nil {
        return nil, fmt.Errorf("生成 AES 密钥失败: %w", err)
    }
    return &AESEncryptor{key: key}, nil
}

// SetGlobalEncryptor 设置全局加密器实例
func SetGlobalEncryptor(e *AESEncryptor) {
    initOnce.Do(func() {
        globalEncryptor = e
    })
}

// GetGlobalEncryptor 获取全局加密器实例
func GetGlobalEncryptor() *AESEncryptor {
    return globalEncryptor
}

// Encrypt 加密数据
func (e *AESEncryptor) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return nil, nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, nil, fmt.Errorf("创建 GCM 失败: %w", err)
    }

    nonce = make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, nil, fmt.Errorf("生成 Nonce 失败: %w", err)
    }

    // Seal 方法将 nonce 附加到密文前
    sealed := gcm.Seal(nil, nonce, plaintext, nil)
    return sealed, nonce, nil
}

// Decrypt 解密数据
func (e *AESEncryptor) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("创建 GCM 失败: %w", err)
    }

    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, fmt.Errorf("解密失败: %w", err)
    }

    return plaintext, nil
}

// EncryptBase64 加密并返回 Base64 编码的密文和 Nonce
func (e *AESEncryptor) EncryptBase64(plaintext []byte) (ciphertextB64, nonceB64 string, err error) {
    ciphertext, nonce, err := e.Encrypt(plaintext)
    if err != nil {
        return "", "", err
    }
    return base64.StdEncoding.EncodeToString(ciphertext), base64.StdEncoding.EncodeToString(nonce), nil
}

// DecryptBase64 解密 Base64 编码的密文和 Nonce
func (e *AESEncryptor) DecryptBase64(ciphertextB64, nonceB64 string) ([]byte, error) {
    ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
    if err != nil {
        return nil, fmt.Errorf("解码密文失败: %w", err)
    }
    nonce, err := base64.StdEncoding.DecodeString(nonceB64)
    if err != nil {
        return nil, fmt.Errorf("解码 Nonce 失败: %w", err)
    }
    return e.Decrypt(ciphertext, nonce)
}
```

### B. 错误处理规范

所有加密/解密操作的错误处理应遵循以下规范：

```go
// 加密失败
if err != nil {
    logger.Error("凭证加密失败", zap.Error(err))
    return nil, xerr.ErrBindEncryptFailed.WithInternal(err)
}

// 解密失败
if err != nil {
    logger.Error("凭证解密失败", zap.Error(err))
    return nil, xerr.ErrBindDecryptFailed.WithInternal(err)
}
```

### C. 日志安全规范

**禁止在日志中记录的内容：**

- 明文 BotToken
- 密文内容（过长，无意义）
- AES 密钥

**允许在日志中记录的内容：**

- 操作类型（加密/解密）
- 操作结果（成功/失败）
- 错误信息（脱敏后）

---

> **文档结束**