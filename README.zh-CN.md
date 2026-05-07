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

OctoTify 是一个**消息总线平台**，连接外部系统（CI/CD、监控、自定义应用）与多种通知渠道。

**版本：** 1.1.0

**已支持的渠道：**

| 渠道 | 状态 | 说明 |
|------|------|------|
| 飞书 | ✅ | 飞书自定义机器人 |
| Telegram | ✅ | Telegram Bot 推送 |
| 邮件 | ✅ | SMTP 邮件推送 |
| 钉钉 | ✅ | 钉钉群机器人，支持 HMAC-SHA256 签名校验 |
| Gotify | ✅ | 自托管 Gotify 消息推送 |
| 企业微信 | 🚧 | 企业微信群机器人（即将上线） |
| Webhook | 🚧 | 自定义 Webhook（即将上线） |

**工作流程：**

1. 在 OctoTify 中创建一个**来源（Source）** — 获得唯一推送令牌
2. 将来源绑定到一个或多个**渠道（Channel）**（飞书群、Telegram 机器人等）
3. 外部系统通过 `POST /api/push/{token}` 推送消息
4. OctoTify 并发将消息投递到所有绑定的渠道

---

## ✨ 核心特性

- **推送解耦** — 外部系统只需推送到一个令牌，无需感知下游渠道
- **灵活路由** — 随时绑定/解绑渠道，一个来源 → 多个目的地
- **独立容错** — 一个渠道失败不影响其他渠道，每个渠道独立 30 秒超时
- **易于扩展** — 策略模式架构，新增渠道无需改动现有代码
- **完整审计** — 每次推送尝试均有记录，包含状态和错误详情

---

## 🚦 快速开始

### Docker Compose（推荐）

创建 `docker-compose.yml`：

```yaml
services:
  octotify:
    image: loommii/octotify:latest
    container_name: octotify
    restart: unless-stopped
    ports:
      - "5233:5233"
    volumes:
      - octotify-data:/app/data

volumes:
  octotify-data:
```

启动：

```bash
docker compose up -d
```

> **持久化建议：** 使用命名数据卷 `octotify-data` 持久化数据库和日志，防止容器删除后数据丢失。

启动后访问 `http://localhost:5233` 即可使用。

### Docker

```bash
docker run -d \
  --name octotify \
  -p 5233:5233 \
  -v octotify-data:/app/data \
  loommii/octotify:latest
```

### 源码部署

**前置要求：**

- Go 1.26+
- Node.js 20+

**后端：**

```bash
cd backend
go run cmd/server/main.go
```

**前端：**

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

## 📝 更新日志

### v1.1.0

**新增渠道：**
- 新增钉钉群机器人渠道，支持 HMAC-SHA256 加签校验
- 新增 Telegram Bot 推送渠道，支持可选 HTTP 代理
- 新增 Email SMTP 邮件推送，支持端口 465（隐式 TLS）、587（STARTTLS）、25（明文）
- 新增 Gotify 自托管消息推送渠道，支持 Markdown 消息格式和优先级配置

**改进：**
- 前端渠道表单提交时自动规范化数字类型字段
- 新增全渠道 Playwright E2E 端到端测试

---

<p align="center">
  <a href="https://github.com/loommii/OctoTify">loommii/OctoTify</a>
</p>
