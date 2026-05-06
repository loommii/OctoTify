# OctoTify

<p align="center">
  <img src="./logo.png" alt="OctoTify Logo" width="200" />
</p>

<p align="center">
  <strong>A Message Bus Platform for Multi-Source, Multi-Channel Notifications</strong>
</p>

<p align="center">
  English · <a href="./README.zh-CN.md">中文</a>
</p>

---

## 🚀 Overview

OctoTify is a **message bus platform** that bridges external systems (CI/CD, monitoring, custom apps) with multiple notification channels (Feishu, WeChat, Telegram, DingTalk, Email, Webhook).

**How it works:**

1. Create a **Source** in OctoTify — you get a unique token
2. Bind the Source to one or more **Channels** (Feishu group, Telegram bot, etc.)
3. External systems push messages to `POST /api/push/{token}`
4. OctoTify delivers the message to all bound channels concurrently

---

## ✨ Key Features

- **One Token, Many Channels** — Push once, deliver everywhere
- **Concurrent Delivery** — Each channel gets an independent goroutine with 30s timeout
- **Full Audit Trail** — Every message delivery is recorded with status
- **Dual-Token Auth** — JWT Access Token (1h) + Refresh Token (7d) for admin panel
- **i18n Ready** — API error messages support multiple languages
- **Request Tracing** — `X-Request-ID` for debugging and log correlation

---

## � Quick Start

### Prerequisites

- Go 1.26+
- Node.js 20+

### Backend

```bash
cd backend
go run cmd/server/main.go
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Or use the startup scripts in `run/`.

---

## � Push API

External systems use this endpoint to push messages:

```bash
POST /api/push/{sourceToken}
Authorization: Bearer src019df95961e6743a85bb86bc1e42e181
Content-Type: application/json

{
  "title": "CI Build",
  "message": "Build #123 passed"
}
```

**Response:**

```json
{
  "code": 0,
  "msg": "Success",
  "data": {
    "total": 2,
    "success": 2,
    "failed": 0,
    "results": [
      { "channel_id": 1, "channel_name": "Feishu", "message_id": 10, "success": true },
      { "channel_id": 2, "channel_name": "WeChat", "message_id": 11, "success": true }
    ]
  }
}
```

> All business errors return HTTP 200 with a non-zero `code`. Only JWT auth errors return HTTP 401.

---

## 🏗️ Architecture

```
External System ──POST /api/push/{token}──→ OctoTify
                                                │
                                          Validate Token
                                                │
                                    Query bound Channels
                                                │
                              Concurrent push to each Channel
                              (independent goroutine, 30s timeout per channel)
                                                │
                                    Record results → Return summary
```

Built with **Go + Gin** (backend) and **Vue 3 + Vite** (frontend), using **SQLite** for storage and **GORM Gen** for data access.

---

## � Documentation

For detailed design docs, API specifications, and UML diagrams, see the [docs](./docs/) directory:

| Document | Description |
|----------|-------------|
| [Architecture](./docs/00-Project-Architecture.md) | Project architecture, domain models, database design |
| [API Spec](./docs/03-API-Specification.md) | API design rules, pagination, response format |
| [Error Codes](./docs/05-Error-Codes.md) | Error code definitions by module |
| [UML Diagrams](./docs/uml/) | Use case, sequence, class, state, and activity diagrams |

---

<p align="center">Made with ❤️ using Go & Vue</p>
