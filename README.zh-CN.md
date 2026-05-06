# OctoTify

<p align="center">
  <img src="./logo.png" alt="OctoTify Logo" width="200" />
</p>

<p align="center">
  <strong>多来源、多渠道消息总线平台</strong>
</p>

<p align="center">
  <a href="./README.md">English</a> · 中文
</p>

---

## 🚀 项目简介

OctoTify 是一个**消息总线平台**，连接外部系统（CI/CD、监控、自定义应用）与多种通知渠道（飞书、企业微信、Telegram、钉钉、邮件、Webhook）。

**工作流程：**

1. 在 OctoTify 中创建一个**来源（Source）** — 获得唯一推送令牌
2. 将来源绑定到一个或多个**渠道（Channel）**（飞书群、Telegram 机器人等）
3. 外部系统通过 `POST /api/push/{token}` 推送消息
4. OctoTify 并发将消息投递到所有绑定的渠道

---

## ✨ 核心特性

- **一个令牌，多渠道推送** — 一次推送，多端送达
- **并发投递** — 每个渠道独立 goroutine，单渠道 30 秒超时
- **完整审计** — 每条消息的投递状态均有记录
- **双令牌认证** — JWT Access Token（1 小时）+ Refresh Token（7 天）
- **国际化** — API 错误消息支持多语言
- **请求追踪** — `X-Request-ID` 便于调试和日志关联

---

## 🚦 快速开始

### 前置要求

- Go 1.26+
- Node.js 20+

### 后端

```bash
cd backend
go run cmd/server/main.go
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

或使用 `run/` 目录下的启动脚本。

---

## 📡 推送 API

外部系统通过以下接口推送消息：

```bash
POST /api/push/{sourceToken}
Authorization: Bearer src019df95961e6743a85bb86bc1e42e181
Content-Type: application/json

{
  "title": "CI Build",
  "message": "Build #123 passed"
}
```

**响应：**

```json
{
  "code": 0,
  "msg": "请求成功",
  "data": {
    "total": 2,
    "success": 2,
    "failed": 0,
    "results": [
      { "channel_id": 1, "channel_name": "飞书", "message_id": 10, "success": true },
      { "channel_id": 2, "channel_name": "企业微信", "message_id": 11, "success": true }
    ]
  }
}
```

> 除 JWT 鉴权失败返回 HTTP 401 外，所有业务错误均返回 HTTP 200，通过 `code` 字段区分错误类型。

---

## 🏗️ 架构概览

```
外部系统 ──POST /api/push/{token}──→ OctoTify
                                          │
                                    验证 Source Token
                                          │
                              查询关联的 Channels
                                          │
                        并发推送到各渠道
                        （每个渠道独立 goroutine，30 秒超时）
                                          │
                              记录结果 → 返回推送摘要
```

后端使用 **Go + Gin**，前端使用 **Vue 3 + Vite**，数据存储使用 **SQLite**，数据访问层使用 **GORM Gen**。

---

## 📚 文档索引

详细的设计文档、API 规范和 UML 图请查阅 [docs](./docs/) 目录：

| 文档 | 说明 |
|------|------|
| [项目架构](./docs/00-Project-Architecture.md) | 项目架构、领域模型、数据库设计 |
| [API 规范](./docs/03-API-Specification.md) | API 设计规则、分页、响应格式 |
| [错误码](./docs/05-Error-Codes.md) | 按模块划分的错误码定义 |
| [UML 图](./docs/uml/) | 用例图、时序图、类图、状态图、活动图 |

---

<p align="center">Made with ❤️ using Go & Vue</p>
