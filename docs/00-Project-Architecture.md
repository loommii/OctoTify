# 项目架构说明

> OctoTify 是一个消息总线，支持多种消息来源和推送渠道。

---

## 一、项目概述

OctoTify 是一个消息总线平台，核心功能是：

1. **消息来源（Source）**：外部系统通过 Token 向平台推送消息
2. **推送渠道（Channel）**：平台将消息推送到多种渠道（微信、Telegram、钉钉、邮件等）
3. **灵活绑定**：一个 Source 可以关联多个 Channel，实现一对多推送

### 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go + Gin |
| ORM | GORM Gen |
| 配置管理 | Viper |
| 日志库 | zap |
| 前端 | Vue 3 + Vite |
| 数据库 | SQLite |
| 部署 | Docker 单容器 |

---

## 二、核心领域模型

### 实体关系

```
User (用户)
  │
  ├── 创建 Source (消息来源) → 获得 Token
  │       │
  │       └── 绑定 Channel 1, 2, 3...
  │
  └── 创建 Channel (推送渠道)
          ├── 类型: 微信 / Telegram / 钉钉 / 邮件 / Webhook
          └── 配置: 各自的凭证/配置
```

### 核心实体

| 实体 | 说明 |
|------|------|
| **User** | 用户账户，支持双令牌认证 |
| **Source** | 消息来源，系统自动生成前缀 + UUIDv4 的随机 Token |
| **Channel** | 推送渠道，绑定某种 Bot/推送方式 |
| **Message** | 消息记录，记录推送状态 |

### 关系

```
User 1 ──── N Source      (一个用户可以创建多个消息来源)
User 1 ──── N Channel     (一个用户可以绑定多个推送渠道)
Source 1 ──── N SourceChannel (关联表) ──── N Channel
Source 1 ──── N Message   (一个来源可以产生多条消息)
Channel 1 ──── N Message  (一个渠道有多条消息)
Message 1 ──── 1 Channel  (一条消息只属于一个渠道)
```

---

## 三、数据库设计

### users 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 主键 |
| username | string | 用户名 |
| password_hash | string | 密码哈希 |
| created_at | int64 | 创建时间（Unix 毫秒时间戳） |
| updated_at | int64 | 更新时间（Unix 毫秒时间戳） |

### sources 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 主键 |
| user_id | int64 | 用户ID |
| name | string | 来源名称 |
| token | string | 推送 Token（前缀 `src` + UUIDv4 无连字符，共 35 位） |
| description | string | 描述 |
| status | int | 1=正常, 2=禁用, -1=删除 |
| created_at | int64 | 创建时间（Unix 毫秒时间戳） |
| updated_at | int64 | 更新时间（Unix 毫秒时间戳） |
| last_used_at | int64 | 最后使用时间（Unix 毫秒时间戳，0 表示未使用） |

### channels 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 主键 |
| user_id | int64 | 用户ID |
| type | string | 渠道类型: wechat, telegram, dingtalk, email... |
| name | string | 渠道名称 |
| config | JSON | 各渠道的配置信息 |
| status | int | 1=正常, 2=禁用, -1=删除 |
| created_at | int64 | 创建时间（Unix 毫秒时间戳） |
| updated_at | int64 | 更新时间（Unix 毫秒时间戳） |
| last_used_at | int64 | 最后使用时间（Unix 毫秒时间戳，0 表示未使用） |

### source_channels 表（关联表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 主键 |
| source_id | int64 | 来源ID |
| channel_id | int64 | 渠道ID |
| created_at | int64 | 创建时间（Unix 毫秒时间戳） |

### messages 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 主键 |
| source_id | int64 | 来源ID |
| channel_id | int64 | 渠道ID |
| title | string | 消息标题 |
| content | text | 消息内容 |
| status | int | 100=待推送, 200=成功, 300=失败... |
| created_at | int64 | 创建时间（Unix 毫秒时间戳） |
| updated_at | int64 | 更新时间（Unix 毫秒时间戳） |

### refresh_tokens 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 主键 |
| jti | string | JWT ID，令牌唯一标识 |
| user_id | int64 | 用户ID |
| revoked | bool | 是否已撤销 |
| expires_at | int64 | 过期时间（Unix 毫秒时间戳） |
| created_at | int64 | 创建时间（Unix 毫秒时间戳） |

---

## 四、推送流程

### 外部系统调用

```http
POST /api/message/push
Authorization: Bearer src0196a3b2c4d50000a1b2c3d4e5f67890
Content-Type: application/json

{
  "title": "CI Build",
  "message": "Build #123 passed"
}
```

### 后端处理流程

```
1. 验证 Source Token
   ↓
2. 查询 Source 关联的 Channels
   ↓
3. 遍历 Channels，通过 SenderFactory 获取对应 Sender
   ↓
4. 并发推送消息到各 Channel
   ↓
5. 记录每条推送结果到 messages 表
   ↓
6. 返回推送结果
```

### 并发推送管理

消息推送到多个渠道时采用并发执行，具体策略如下：

| 策略 | 说明 |
|------|------|
| **并发模型** | 每个渠道启动独立 goroutine 推送 |
| **并发控制** | 使用 goroutine pool 限制最大并发数（默认 10） |
| **超时控制** | 单个渠道推送超时时间 30 秒，通过 context.WithTimeout 实现 |
| **错误隔离** | 单个渠道推送失败不影响其他渠道，独立记录状态 |
| **错误聚合** | 使用 errgroup 收集所有 goroutine 的推送结果 |
| **结果汇总** | 等待所有 goroutine 完成后，汇总成功/失败数量返回 |

---

## 五、项目结构

```
OctoTify/
├── docs/                     # 项目文档
├── backend/                  # Go 后端
│   ├── cmd/server/main.go    # 服务入口
│   ├── tools/generate/       # GORM Gen 代码生成工具
│   ├── internal/             # 私有代码
│   │   ├── config/           # 配置加载（Viper）
│   │   ├── log/              # 日志初始化（zap）
│   │   ├── database/         # 数据库初始化（GORM + SQLite）
│   │   ├── model/            # 数据模型
│   │   ├── query/            # GORM Gen 生成的查询代码
│   │   ├── server/           # HTTP Server（Gin）
│   │   ├── handler/          # HTTP 处理器
│   │   ├── service/          # 业务逻辑 + 数据访问（直接调用 query）
│   │   ├── sender/           # 推送策略层
│   │   │   ├── sender.go     # Sender 接口
│   │   │   ├── factory.go    # Sender 工厂
│   │   │   ├── wechat/       # 微信实现
│   │   │   ├── telegram/     # Telegram实现
│   │   │   ├── email/        # 邮件实现
│   │   │   └── webhook/      # Webhook实现
│   │   └── middleware/       # 中间件
│   ├── pkg/                  # 可复用的公共库
│   │   ├── response/         # 统一响应
│   │   ├── errors/           # 错误码
│   │   ├── i18n/             # 多语言
│   │   ├── jwt/              # JWT 工具
│   │   └── validator/        # 自定义参数验证器
│   ├── config/               # 配置文件
│   ├── data/                 # SQLite 数据库文件
│   ├── go.mod
│   └── go.sum
├── go.work                   # Go 工作区配置
├── go.work.sum
├── frontend/                 # Vue 3 前端
│   ├── src/
│   │   ├── api/              # API 请求封装
│   │   ├── composables/      # 组合式函数
│   │   ├── components/       # 可复用组件
│   │   ├── views/            # 页面组件
│   │   ├── stores/           # Pinia 状态
│   │   ├── types/            # TypeScript 类型
│   │   ├── router/           # 路由
│   │   └── lib/              # 基础设施
│   ├── public/
│   ├── vite.config.ts
│   └── package.json
└── Dockerfile
```

---

## 六、设计模式

| 模式 | 应用位置 | 说明 |
|------|----------|------|
| **策略模式** | Sender 接口 | 每种推送渠道实现统一的 Sender 接口 |
| **工厂模式** | SenderFactory | 根据 Channel 类型创建对应的 Sender |
| **观察者模式** | 消息推送 | 消息到达后通知所有已绑定的 Channel |
| **适配器模式** | 各渠道实现 | 不同 Bot API 适配成统一消息格式 |

---

## 七、状态字段设计

### Source 状态

| 状态值 | 含义 | 说明 |
|--------|------|------|
| `1` | 正常 | Token 有效，可推送 |
| `2` | 禁用 | Token 失效，不可推送，可恢复 |
| `-1` | 已删除 | 软删除，Token 失效，不可恢复 |

### Channel 状态

| 状态值 | 含义 | 说明 |
|--------|------|------|
| `1` | 正常 | 渠道可用 |
| `2` | 禁用 | 渠道临时停用 |
| `-1` | 已删除 | 软删除 |

### Message 状态

| 状态码 | 含义 | 说明 |
|--------|------|------|
| `100` | 待推送（pending） | 消息已创建，等待推送 |
| `200` | 推送成功（sent） | 推送成功 |
| `300` | 推送失败（failed） | 推送失败 |
| `-1` | 已删除（软删除） | 软删除 |

---

## 八、鉴权机制

### 双令牌认证（用户管理后台）

| Token 类型 | 有效期 | 用途 |
|------------|--------|------|
| Access Token | 1 小时 | 管理后台 API 鉴权 |
| Refresh Token | 7 天 | 刷新 Access Token |

### Source Token（推送 API）

- 格式：前缀 `src` + UUIDv4 无连字符（共 35 位，仅小写字母和数字）
- 示例：`src0196a3b2c4d50000a1b2c3d4e5f67890`
- 系统自动生成，创建 Source 时由后端生成，用户不可自定义
- 永久有效（直到 Source 被删除或禁用）
- 通过 `Authorization: Bearer` 传递

---

## 九、规范文档索引

| 文档 | 说明 |
|------|------|
| [01-RequestID.md](./01-RequestID.md) | 请求 ID 规范 |
| [02-Logging.md](./02-Logging.md) | 日志规范 |
| [03-API-Specification.md](./03-API-Specification.md) | 后端 API 规范 |
| [04-i18n.md](./04-i18n.md) | 多语言规范 |
| [05-Error-Codes.md](./05-Error-Codes.md) | 错误码规范 |

---

## 版本历史

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-02 | 1.0 | 初始版本，定义项目架构和核心模型 |
