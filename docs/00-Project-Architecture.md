# 项目架构说明

> OctoTify 是一个消息总线，支持多种消息来源和推送渠道。

***

## 一、项目概述

OctoTify 是一个消息总线平台，核心功能是：

1. **消息来源（Source）**：外部系统通过 Token 向平台推送消息
2. **推送渠道（Channel）**：平台将消息推送到多种渠道（微信 ClawBot、飞书、Telegram、钉钉、邮件等）
3. **灵活绑定**：一个 Source 可以关联多个 Channel，实现一对多推送

### 技术栈

| 层级   | 技术           |
| ---- | ------------ |
| 后端   | Go + Gin     |
| ORM  | GORM Gen     |
| 配置管理 | Viper        |
| 日志库  | zap          |
| 前端   | Vue 3 + Vite |
| 数据库  | SQLite       |
| 部署   | Docker 单容器   |

### 规范文档索引

| 文档                                                         | 说明                          |
| ---------------------------------------------------------- | --------------------------- |
| [01-RequestID.md](./01-RequestID.md)                       | 请求 ID 规范                    |
| [02-Logging.md](./02-Logging.md)                           | 日志规范                        |
| [03-API-Specification.md](./03-API-Specification.md)       | 后端 API 规范（响应格式、HTTP 状态码、分页） |
| [04-i18n.md](./04-i18n.md)                                 | 多语言规范                       |
| [05-Error-Codes.md](./05-Error-Codes.md)                   | 错误码规范                       |

***

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
          ├── 类型: 微信 ClawBot / 飞书 / Telegram / 钉钉 / 邮件
          └── 配置: 各自的凭证/配置
```

### 核心实体

| 实体          | 说明                               |
| ----------- | -------------------------------- |
| **User**    | 用户账户，支持双令牌认证                     |
| **Source**  | 消息来源，系统自动生成前缀 + UUIDv4 的随机 Token |
| **Channel** | 推送渠道，绑定某种 Bot/推送方式               |
| **Message** | 消息记录，记录推送状态                      |

### 关系

```
User 1 ──── N Source      (一个用户可以创建多个消息来源)
User 1 ──── N Channel     (一个用户可以绑定多个推送渠道)
Source 1 ──── N SourceChannel (关联表) ──── N Channel
Source 1 ──── N Message   (一个来源可以产生多条消息)
Channel 1 ──── N Message  (一个渠道有多条消息)
Message 1 ──── 1 Channel  (一条消息只属于一个渠道)
```

***

## 三、项目结构

```
OctoTify/
├── docs/                     # 项目文档
│   ├── uml/                  # PlantUML 图
│   └── *.md                  # 规范文档
├── backend/                  # Go 后端
│   ├── cmd/server/main.go    # 服务入口
│   ├── internal/             # 私有代码
│   │   ├── client/           # 第三方 API 客户端
│   │   │   └── ilink/        # iLink 平台客户端（微信 ClawBot 绑定）
│   │   ├── config/           # 配置加载（Viper）
│   │   ├── database/         # 数据库初始化（GORM + SQLite）
│   │   ├── handler/          # HTTP 处理器
│   │   │   └── dto/          # 请求/响应 DTO
│   │   ├── log/              # 日志初始化（zap）
│   │   ├── middleware/       # 中间件
│   │   ├── model/            # 数据模型
│   │   ├── query/            # GORM Gen 生成的查询代码
│   │   ├── sender/           # 推送策略层
│   │   │   ├── sender.go         # Sender 接口
│   │   │   ├── factory.go        # Sender 工厂
│   │   │   ├── wechat_clawbot.go # 微信 ClawBot 实现
│   │   │   ├── telegram.go       # Telegram 实现
│   │   │   ├── email.go          # 邮件实现
│   │   │   ├── dingtalk.go       # 钉钉实现
│   │   │   └── feishu.go         # 飞书实现
│   │   ├── server/           # HTTP Server（Gin + OpenAPI）
│   │   └── service/          # 业务逻辑 + 数据访问
│   ├── pkg/                  # 可复用的公共库
│   │   ├── aescipher/        # AES 加密
│   │   ├── ctxutil/          # Context 工具
│   │   ├── jwtx/             # JWT 工具
│   │   ├── response/         # 统一响应
│   │   └── xerr/             # 错误码
│   ├── config/               # 配置文件
│   ├── go.mod
│   └── go.sum
├── docker/                   # Docker 部署配置
│   ├── Dockerfile
│   ├── entrypoint.sh
│   ├── nginx.conf
│   └── supervisord.conf
├── run/                      # 开发运行脚本
│   ├── backend.sh
│   ├── dev.sh
│   └── frontend.sh
├── frontend/                 # Vue 3 前端（vue-vben-admin monorepo）
│   ├── apps/
│   │   └── web-ele/          # Element Plus 应用
│   │       ├── public/       # 静态资源
│   │       ├── src/
│   │       │   ├── adapter/      # 框架适配器（表单、组件等）
│   │       │   ├── api/          # API 客户端层
│   │       │   │   └── modules/  # 按模块组织的 API 函数
│   │       │   ├── components/   # 业务组件
│   │       │   ├── composables/  # 组合式函数
│   │       │   ├── layouts/      # 布局组件
│   │       │   ├── locales/      # 多语言
│   │       │   ├── router/       # 路由配置
│   │       │   ├── store/        # Pinia 状态管理
│   │       │   ├── utils/        # 工具函数
│   │       │   ├── views/        # 页面组件
│   │       │   ├── app.vue       # 根组件
│   │       │   ├── bootstrap.ts  # 应用启动引导
│   │       │   ├── main.ts       # 入口文件
│   │       │   └── preferences.ts# 偏好配置
│   │       ├── .env*             # 环境变量
│   │       ├── index.html
│   │       ├── vite.config.ts
│   │       └── tsconfig.json
│   ├── internal/             # 内部构建工具
│   │   ├── lint-configs/     # ESLint/Oxlint/Stylelint 配置
│   │   ├── node-utils/       # Node 工具
│   │   ├── tailwind-config/  # Tailwind 配置
│   │   ├── tsconfig/         # TypeScript 基础配置
│   │   └── vite-config/      # Vite 插件配置
│   ├── packages/             # 共享包
│   │   ├── @core/            # 核心 UI 组件库
│   │   ├── constants/        # 常量
│   │   ├── effects/          # 功能模块（access、layouts、request 等）
│   │   ├── icons/            # 图标
│   │   ├── locales/          # 多语言
│   │   ├── preferences/      # 偏好设置
│   │   ├── stores/           # Pinia 基础封装
│   │   ├── styles/           # 样式
│   │   ├── types/            # 类型定义
│   │   └── utils/            # 工具函数
│   ├── package.json
│   ├── pnpm-workspace.yaml
│   ├── turbo.json
│   └── openapi-ts.config.ts  # OpenAPI 代码生成配置
├── go.work                   # Go 工作区配置
└── logo.png
```

***

## 四、设计模式

| 模式        | 应用位置          | 说明                        |
| --------- | ------------- | ------------------------- |
| **策略模式**  | Sender 接口     | 每种推送渠道实现统一的 Sender 接口     |
| **工厂模式**  | SenderFactory | 根据 Channel 类型创建对应的 Sender |
| **观察者模式** | 消息推送          | 消息到达后通知所有已绑定的 Channel     |
| **适配器模式** | 各渠道实现         | 不同 Bot API 适配成统一消息格式      |

***

## 五、鉴权机制

### 双令牌认证（用户管理后台）

| Token 类型      | 有效期  | 用途              |
| ------------- | ---- | --------------- |
| Access Token  | 1 小时 | 管理后台 API 鉴权     |
| Refresh Token | 7 天  | 刷新 Access Token |

### Source Token（推送 API）

- 格式：前缀 `src` + UUIDv4 无连字符（共 35 位，仅小写字母和数字）
- 示例：`src0196a3b2c4d50000a1b2c3d4e5f67890`
- 系统自动生成，创建 Source 时由后端生成，用户不可自定义
- 永久有效（直到 Source 被删除或禁用）
- 通过 URL 路径参数传递：`POST /api/push/:token`

***

## 六、规范文档索引

| 文档                                                   | 说明        |
| ---------------------------------------------------- | --------- |
| [01-RequestID.md](./01-RequestID.md)                 | 请求 ID 规范  |
| [02-Logging.md](./02-Logging.md)                     | 日志规范      |
| [03-API-Specification.md](./03-API-Specification.md) | 后端 API 规范 |
| [04-i18n.md](./04-i18n.md)                           | 多语言规范     |
| [05-Error-Codes.md](./05-Error-Codes.md)             | 错误码规范     |

***

## 版本历史

| 日期         | 版本    | 变更               |
| ---------- | ----- | ---------------- |
| 2026-05-02 | 1.0.0 | 初始版本，定义项目架构和核心模型 |
| 2026-05-16 | 1.2.0 | 精简架构文档，移除数据库设计、推送流程、状态字段等实现细节 |
| 2026-05-16 | 1.2.0 | 修正推送 API 路径 |
| 2026-05-16 | 1.2.0 | 修正 OpenAPI SourceTokenAuth 安全方案描述 |
| 2026-05-16 | 1.2.0 | 修正 sender 为平铺文件结构，删除空 api/ 目录 |
