# OctoTify

<p align="center">
  <img src="./logo.png" alt="OctoTify Logo" width="200" />
</p>

<p align="center">
  <strong>A Message Bus Platform for Multi-Source, Multi-Channel Notifications</strong>
</p>

<p align="center">
  <a href="./README.zh-CN.md">中文</a> · English
</p>

---

## 🚀 Overview

OctoTify is a **message bus platform** that bridges external systems (CI/CD, monitoring, custom apps) with multiple notification channels (Feishu, WeCom, Telegram, DingTalk, Email, Webhook).

**How it works:**

1. Create a **Source** in OctoTify — you get a unique push token
2. Bind the Source to one or more **Channels** (Feishu group, Telegram bot, etc.)
3. External systems push messages via `POST /api/push/{token}`
4. OctoTify concurrently delivers messages to all bound channels

---

## ✨ Key Features

- **Decoupled Push** — External systems push to a single token, no need to be aware of downstream channels
- **Flexible Routing** — Bind/unbind channels anytime, one source → many destinations
- **Independent Fault Tolerance** — One channel failing doesn't affect others, each channel has an independent 30-second timeout
- **Extensible by Design** — Strategy pattern architecture, add new channels without modifying existing code
- **Full Audit Trail** — Every push attempt is recorded, including status and error details

---

## 🚦 Quick Start

### Docker Compose (Recommended)

Create `docker-compose.yml`:

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

Start:

```bash
docker compose up -d
```

> **Persistence recommendation:** Use named volume `octotify-data` to persist the database and logs, preventing data loss when the container is removed.

After starting, visit `http://localhost:5233`.

### Docker

```bash
docker run -d \
  --name octotify \
  -p 5233:5233 \
  -v octotify-data:/app/data \
  loommii/octotify:latest
```

### From Source

**Prerequisites:**

- Go 1.26+
- Node.js 20+

**Backend:**

```bash
cd backend
go run cmd/server/main.go
```

**Frontend:**

```bash
cd frontend
npm install
npm run dev
```

Or use the startup scripts in the `run/` directory.

---

## 📡 Push API

External systems push messages via the following endpoint:

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
      { "channel_id": 2, "channel_name": "WeCom", "message_id": 11, "success": true }
    ]
  }
}
```

> Except for JWT authentication failures which return HTTP 401, all business errors return HTTP 200, with the `code` field distinguishing error types.

---

## 🏗️ Architecture Overview

```
External System ──POST /api/push/{token}──→ OctoTify
                                                │
                                          Validate Source Token
                                                │
                                    Query associated Channels
                                                │
                              Concurrent push to each channel
                              (independent goroutine per channel, 30s timeout)
                                                │
                                    Record results → Return push summary
```

Backend uses **Go + Gin**, frontend uses **Vue 3 + Vite**, data storage uses **SQLite**, and the data access layer uses **GORM Gen**.

---

## 📚 Documentation

For detailed design docs, API specifications, and UML diagrams, see the [docs](./docs/) directory:

| Document | Description |
|----------|-------------|
| [Architecture](./docs/00-Project-Architecture.md) | Project architecture, domain models, database design |
| [API Specification](./docs/03-API-Specification.md) | API design rules, pagination, response format |
| [Error Codes](./docs/05-Error-Codes.md) | Error code definitions by module |
| [UML Diagrams](./docs/uml/) | Use case, sequence, class, state, and activity diagrams |

---

<p align="center">
  <a href="https://github.com/loommii/OctoTify">loommii/OctoTify</a>
</p>
