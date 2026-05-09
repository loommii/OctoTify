# 微信 ClawBot 绑定流程无状态重构方案

> 文档版本: v5.0
> 创建日期: 2026-05-09
> 更新日期: 2026-05-09
> 状态: 评审修订版（已整合前后端审查意见）
> 作者: 系统架构师

***

## 目录

1. [当前架构分析](#1-当前架构分析)
2. [新架构设计](#2-新架构设计)
3. [接口变更对比](#3-接口变更对比)
4. [代码变更清单](#4-代码变更清单)
5. [技术风险](#5-技术风险)
6. [审查意见与修订记录](#6-审查意见与修订记录)

***

## 1. 当前架构分析

### 1.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        当前绑定流程架构                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌──────────┐    ① StartBind              ┌──────────┐            │
│   │  Frontend│─────────────────────────────►│  Handler │            │
│   └──────────┘                              └────┬─────┘            │
│                                                   │                 │
│                                        ② 调用 iLink 获取二维码      │
│                                                   │                 │
│                                        ┌──────────▼──────────┐     │
│                                        │  ChannelService      │     │
│                                        │  fetchQRCode()       │     │
│                                        └──────────┬──────────┘     │
│                                                   │                 │
│                                        ③ 返回 BindSession          │
│                                                   │                 │
│                                                   ▼                 │
│                                        ┌──────────────────┐         │
│                                   ④ Set│  BindStore         │         │
│                                  ┌────►│  (内存存储)        │         │
│                                  │     │                   │         │
│                                  │     │  sessions map     │         │
│                                  │     │  completedSessions│         │
│                                  │     └──────┬───────────┘         │
│                                  │            │                     │
│   ┌──────────┐    ⑤ GetStatus   │     ┌───────▼────────┐          │
│   │  Frontend│──────────────────┘     │  PollStatus    │          │
│   └──────────┘                        │  (长轮询 API)   │          │
│        │                              └───────┬────────┘          │
│        │ ⑥ 返回状态 + credentials             │                   │
│        │◄─────────────────────────────────────┘                   │
│                                                                    │
│   ┌──────────────────────────────────────────────────────────┐    │
│   │              后台 Goroutine (2 个)                        │    │
│   │  ┌────────────────────────┐  ┌────────────────────────┐ │    │
│   │  │ cleanupExpiredSessions │  │ cleanupCompletedSessions│ │    │
│   │  │   每 1 分钟执行         │  │   每 1 分钟执行         │ │    │
│   │  │   清理过期活跃会话      │  │   清理已完成会话缓存    │ │    │
│   │  └────────────────────────┘  └────────────────────────┘ │    │
│   └──────────────────────────────────────────────────────────┘    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

### 1.2 存在的问题

#### 问题 1: 大量内存缓存

**涉及组件**: `BindStore` ([bind\_store.go](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/service/bind_store.go#L80-L89))

```go
type BindStore struct {
    mu                sync.RWMutex                 // 活跃会话的读写锁
    sessions          map[string]*BindSession      // 活跃绑定会话集合
    completedMu       sync.RWMutex                 // 已完成会话的读写锁
    completedSessions map[string]*completedSession // 已完成会话缓存
    poller            BindStatusPoller             // 状态轮询器
    logger            *zap.Logger
    done              chan struct{}                // 关闭信号通道
    closeOnce         sync.Once                    // 确保只关闭一次
}
```

**问题分析**:

- 维护两个独立的 `map` 存储（`sessions` + `completedSessions`）
- 每个活跃会话占用约 200+ 字节内存（不含 map 开销）
- 高并发场景下可能积累大量会话数据
- 双锁机制增加代码复杂度和死锁风险

#### 问题 2: 两个后台 Goroutine

**涉及代码**: [bind\_store.go#L106-L108](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/service/bind_store.go#L106-L108)

```go
go store.cleanupExpiredSessions()   // 启动过期会话清理器
go store.cleanupCompletedSessions() // 启动已完成会话清理器
```

**问题分析**:

- 每 1 分钟遍历整个 map 查找过期条目
- 服务关闭需要等待 goroutine 优雅退出
- goroutine 泄漏风险（如果 done channel 未正确关闭）
- 增加测试复杂度（需要模拟时间相关行为）

#### 问题 3: 服务重启导致状态丢失

**影响场景**:

- 服务发布/重启 → 所有活跃绑定会话丢失
- 用户正在扫码绑定时服务重启 → 前端无法查询到状态
- 多实例部署时，不同实例的内存状态不一致

**现有代码中的表现**:

- [channel\_handler.go#L522-L538](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/handler/channel_handler.go#L522-L538): 当 `bindStore.Get()` 和 `bindStore.GetCompletedSession()` 都找不到数据时返回错误
- 前端需要用户重新发起绑定

#### 问题 4: 复杂的会话状态机

**当前状态流转**:

```
pending → scanned → confirmed/expired
  │                      │
  └──→ cleanupExpired ───┘
         │
         ▼
  保存到 completedSessions（保留 2 分钟）
         │
         ▼
  cleanupCompletedSessions 清理
```

**问题分析**:

- 活跃会话和已完成会话需要分别管理
- 状态更新时需要跨两个存储操作（加两把锁）
- [bind\_store.go#L203-L242](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/service/bind_store.go#L203-L242): `UpdateSessionStatus` 方法需要同时操作 `sessions` 和 `completedSessions`
- 前端需要理解复杂的缓存逻辑才能正确处理状态

#### 问题 5: 长轮询依赖内存状态

**涉及代码**: [channel\_handler.go#L515-L573](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/handler/channel_handler.go#L515-L573)

```go
func (h *ChannelHandler) GetBindStatusLongPolling(c *gin.Context) {
    // 1. 从内存中获取 session
    session, exists := h.bindStore.Get(bindID)
    // 2. 如果不存在，检查 completedSessions
    completedSession, completedExists := h.bindStore.GetCompletedSession(bindID)
    // 3. 调用 PollStatus（内部调用 iLink API）
    h.bindStore.PollStatus(bindID)
    // 4. 再次从内存中获取最新状态
    latestSession, stillExists := h.bindStore.Get(bindID)
    // ...
}
```

**问题分析**:

- 整个长轮询流程依赖内存中的 session 对象
- `PollStatus` 内部调用 iLink API（40s 超时），阻塞 goroutine
- 多个并发请求可能导致重复调用 iLink API
- 内存 session 是 iLink qrcode 的唯一载体，丢失后无法查询

#### 问题 6: 无法水平扩展

**部署限制**:

```
Instance A (sessions: {bind_1, bind_2})
Instance B (sessions: {bind_3, bind_4})

前端请求 bind_1 到达 Instance B → 404 找不到
```

**根本原因**:

- 会话状态存储在各实例本地内存
- 无共享存储机制（Redis/DB）
- 负载均衡器需要 sticky session 才能正常工作

### 1.3 现有架构数据流时序图

```
用户            前端             Handler         BindStore      ChannelService     iLink API
 |              |                |                 |               |                  |
 |--开始绑定--->|                |                 |               |                  |
 |              |--POST /bind--->|                 |               |                  |
 |              |                |--StartBind----->|               |                  |
 |              |                |                 |               |--fetchQRCode---->|
 |              |                |                 |               |<--QRCode--------|
 |              |                |                 |--Set(session) |                  |
 |              |<--bind_id +    |                 |               |                  |
 |              |   qrcode_url---|                 |               |                  |
 |              |                |                 |               |                  |
 |              |--GET /bind/:id |                 |               |                  |
 |              |  /wait-------->|                 |               |                  |
 |              |                |--Get(bind_id)--->|               |                  |
 |              |                |<--session-------|               |                  |
 |              |                |--PollStatus---->|               |                  |
 |              |                |                 |--PollBindStatus                  |
 |              |                |                 |               |-->get_qrcode->|
 |              |                |                 |               |<--status------|
 |              |                |                 |--UpdateStatus |                  |
 |              |                |                 |               |                  |
 |              |                |--Get(bind_id)--->|               |                  |
 |              |                |<--latest_session|               |                  |
 |              |<--status +     |                 |               |                  |
 |              |  credentials---|                 |               |                  |
 |              |                |                 |               |                  |
```

### 1.4 问题总结表

| 问题编号 | 问题描述               | 影响范围        | 严重程度 |
| ---- | ------------------ | ----------- | ---- |
| P1   | 大量内存缓存（双 map + 双锁） | 内存占用、代码复杂度  | 中    |
| P2   | 两个后台清理 Goroutine   | 服务生命周期管理、测试 | 中    |
| P3   | 服务重启状态丢失           | 用户体验、可靠性    | 高    |
| P4   | 复杂的状态机逻辑           | 维护成本、bug 风险 | 中    |
| P5   | 长轮询依赖内存状态          | 并发安全、性能     | 中    |
| P6   | 无法水平扩展             | 部署架构        | 高    |

***

## 2. 新架构设计

### 2.1 设计原则

| 原则                | 说明                                 |
| ----------------- | ---------------------------------- |
| **完全无状态**         | 不维护任何运行时内存状态，不引入 bind\_id 概念       |
| **qrcode 作为唯一标识** | 前端直接保存并使用 iLink 返回的 qrcode 值进行状态查询 |
| **按需查询**          | 每次请求直接调用 iLink API 获取最新状态，不缓存      |
| **极简接口**          | 去掉 bind\_id 生成、校验、映射等所有中间环节        |
| **简化生命周期**        | 移除所有后台 goroutine，消除生命周期管理复杂度       |
| **水平可扩展**         | 任何实例都可以处理任意 qrcode 的请求             |

### 2.2 核心设计思路

**关键洞察**: iLink API 本身就是状态源，qrcode 是天然的可复用标识

当前架构引入了 `bind_id` 作为会话标识，将 iLink 的状态复制到内存中，再通过 bind\_id 关联查询。但实际上：

- iLink 的 `get_qrcode_status` API 已经提供了实时状态查询能力
- qrcode 本身就是 iLink 生成的唯一标识，可以直接用于查询
- 引入 bind\_id 增加了无必要的复杂度（生成、存储、校验）
- **BindSession 中间结构体同样是不必要的抽象**，只需传递 qrcode 字符串即可

```
旧方案: iLink API → 生成 bind_id → 封装 BindSession → 存入内存 → 前端通过 bind_id 查询
新方案: 前端保存 qrcode → 直接传 qrcode 字符串 → 后端调用 iLink API → 返回结果
```

### 2.3 新架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        无状态绑定流程架构                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   ┌──────────┐    ① StartBind              ┌──────────┐            │
│   │  Frontend│─────────────────────────────►│  Handler │            │
│   └──────────┘                              └────┬─────┘            │
│                                                   │                 │
│                                        ② 调用 iLink 获取二维码      │
│                                                   │                 │
│                                        ┌──────────▼──────────┐     │
│                                        │  ChannelService      │     │
│                                        │  fetchQRCode()       │     │
│                                        └──────────┬──────────┘     │
│                                                   │                 │
│                                        ③ 返回 qrcode + qrcodeURL    │
│                                       （两个字符串，无中间结构体）   │
│                                                   │                 │
│                                                   ▼                 │
│                                        ┌──────────────────┐         │
│                                   ④ 返回│  (无存储)         │         │
│                                  ┌─────│  qrcode 作为      │         │
│                                  │     │  唯一临时标识     │         │
│                                  │     └──────────────────┘         │
│                                  │                                  │
│   ┌──────────┐    ⑤ 长轮询状态  │                                  │
│   │  Frontend│──────────────────┘                                  │
│   │  (POST)  │                                                     │
│   └──────────┘                                                     │
│        │                                                           │
│        │ ⑥ 直接传 qrcode 字符串调用 iLink API                      │
│        │◄─────────────────────┐                                   │
│                              │                                   │
│                     ┌────────▼──────────┐                        │
│                     │  ChannelService    │                        │
│                     │  PollBindStatus()  │                        │
│                     │  (直接调用 iLink)   │                        │
│                     └────────┬──────────┘                        │
│                              │                                   │
│                              │ ⑦ get_qrcode_status               │
│                              └──────────────► iLink API          │
│                                             (状态源)             │
│                                                                    │
│   ┌──────────────────────────────────────────────────────────┐    │
│   │              无 BindSession 结构体                         │    │
│   │              无 bind_id 概念                               │    │
│   │              无后台 Goroutine                             │    │
│   │              无内存缓存                                   │    │
│   │              无状态管理                                   │    │
│   └──────────────────────────────────────────────────────────┘    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

### 2.4 qrcode 作为唯一标识

**新方案**: 完全移除 bind\_id 和 BindSession 结构体，前端直接使用 iLink 返回的 `qrcode` 值作为后续状态查询的唯一标识。

```
POST /bind 响应:
{
  "qrcode_url": "data:image/png;base64,...",  // 二维码图片（用于展示）
  "qrcode": "ilink_qr_abc123xyz"              // 二维码原始值（用于查询）
}

POST /bind/status 请求:
{
  "qrcode": "ilink_qr_abc123xyz"              // 直接传 qrcode 字符串
}
```

**设计理由**:

- qrcode 是 iLink 生成的唯一标识，天然具备唯一性
- 前端保存 qrcode 即可在任意时刻查询状态
- 后端无需维护任何映射关系，收到 qrcode 直接查 iLink
- 任何实例都可以处理任意 qrcode 的查询请求，支持水平扩展
- 去掉了 bind\_id 生成、存储、校验的全部复杂度
- **去掉了 BindSession 中间结构体，只传递必要的字符串参数**

### 2.5 新架构数据流时序图

```
用户            前端             Handler         ChannelService       iLink API
 |              |                |                  |                   |
 |--开始绑定--->|                |                  |                   |
 |              |--POST /bind--->|                  |                   |
 |              |                |--StartBind------>|                   |
 |              |                |                  |--fetchQRCode----->|
 |              |                |                  |<--QRCode+URL-----|
 |              |                |<--qrcode_url,    |                   |
 |              |                |   qrcode         |                   |
 |              |                |  (两个字符串)     |                   |
 |              |                |                  |                   |
 |  展示二维码  |                  |                  |                   |
 |              |                |                  |                   |
 |  (用户扫码)  |                  |                  |                   |
 |              |                |                  |                   |
 |              |--POST /bind/  |                  |                   |
 |              |   status------>|                  |                   |
 |              |   {qrcode:xxx} |                  |                   |
 |              |                |--PollBindStatus->|                   |
 |              |                |  (qrcode string) |                   |
 |              |                |                  |-->get_qrcode--->|
 |              |                |                  |<--status+creds--|
 |              |                |<--status + creds-|                   |
 |              |                |                  |                   |
 |              |<--status +     |                  |                   |
 |              |  credentials---|                  |                   |
 |              |                |                  |                   |
```

### 2.6 核心变更点

#### 2.6.1 StartBind 接口变更

**变更前** ([channel\_handler.go#L472-L492](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/handler/channel_handler.go#L472-L492)):

```go
func (h *ChannelHandler) StartBind(c *gin.Context) {
    session, err := h.channelService.StartBind(c.Request.Context(), userID)
    if err != nil {
        c.Error(err)
        return
    }
    bindID := uuid.New().String()           // UUID 随机生成
    h.bindStore.Set(bindID, session)        // 存入内存
    response.Success(c, gin.H{
        "bind_id":    bindID,
        "qrcode_url": session.QRCodeURL,
        "expires_at": session.ExpiresAt.Unix(),
    })
}
```

**变更后**:

```go
func (h *ChannelHandler) StartBind(c *gin.Context) {
    userID, ok := h.getUserID(c)
    if !ok {
        return
    }

    // 审计日志：记录用户发起绑定
    h.logger.Info("发起微信ClawBot绑定",
        zap.Int64("user_id", userID),
    )

    qrcode, qrcodeURL, err := h.channelService.StartBind(c.Request.Context(), userID)
    if err != nil {
        c.Error(err)
        return
    }

    // 直接返回两个字符串，无需 BindSession 中间结构体
    response.Success(c, gin.H{
        "qrcode_url": qrcodeURL,
        "qrcode":     qrcode,
    })
}
```

**关键变化**:

- `StartBind` 不再返回 `*BindSession`，而是直接返回 `(qrcode, qrcodeURL string, err error)`
- 前端需要保存 `qrcode` 值，在后续调用 `POST /bind/status` 时传回后端
- Handler 层不再需要构造任何中间对象
- **新增 userID 审计日志**（审查意见采纳）

#### 2.6.2 查询状态接口变更：GET → POST

**变更前** ([channel\_handler.go#L515-L573](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/handler/channel_handler.go#L515-L573)):

```go
// GET /bind/:bindId/wait
func (h *ChannelHandler) GetBindStatusLongPolling(c *gin.Context) {
    bindID := c.Param("bindId")
    session, exists := h.bindStore.Get(bindID)           // 从内存获取
    // ... 复杂的 BindStore 查找逻辑 ...
    h.bindStore.PollStatus(bindID)                       // 更新内存
    latestSession, stillExists := h.bindStore.Get(bindID) // 从内存读取
    // ...
}
```

**变更后**:

```go
// POST /bind/status
func (h *ChannelHandler) GetBindStatus(c *gin.Context) {
    userID, ok := h.getUserID(c)
    if !ok {
        return
    }

    var req dto.GetBindStatusReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, xerr.ErrBadRequest.Code, "请求参数格式错误")
        return
    }

    // 审计日志：记录用户查询绑定状态
    h.logger.Info("查询绑定状态",
        zap.Int64("user_id", userID),
    )

    // 直接用 qrcode 字符串调用 service，无需构造 BindSession
    // 注意：iLink 的 get_qrcode_status API 本身是长轮询设计（40s 超时），
    // 后端只是中转，前端依然是长轮询模式，行为与旧方案一致
    status, credentials, err := h.channelService.PollBindStatus(c.Request.Context(), req.QRCode)
    if err != nil {
        // iLink API 调用失败，返回 pending 让前端重试
        response.Success(c, gin.H{"status": service.BindStatusPending})
        return
    }

    h.sendBindResponse(c, status, credentials)
}
```

**关键变化**:

- 路由从 `GET /bind/:bindId/wait` 变为 `POST /bind/status`
- 请求体直接传 `{ qrcode: "xxx" }`，不再需要 bind\_id
- **不再有任何 BindStore 或 BindSession 操作，直接传递 qrcode 字符串调用 service**
- `PollBindStatus` 方法签名简化为 `(ctx context.Context, qrcode string)`
- **新增 userID 审计日志**（审查意见采纳）
- **前端依然是长轮询模式**：iLink 的 `get_qrcode_status` API 本身就是长轮询设计（40s 超时），后端只是中转，行为与旧方案完全一致

#### 2.6.3 BindStore 和 BindSession 彻底删除

**删除的组件**:

- 整个 `BindStore` 类（bind\_store.go 文件，\~365 行）
- **`BindSession`** **结构体**（所有定义和引用）
- 所有内存缓存逻辑
- 两个后台清理 goroutine
- `Close()` 和 `CloseWithTimeout()` 生命周期管理
- bind\_id 相关的生成、校验、映射逻辑

### 2.7 新架构下的数据结构

#### 仅保留 BindCredentials 结构体

```go
// BindCredentials 绑定成功后的凭证
// 仅在绑定成功时返回给前端，用于保存加密后的 bot token 等信息
type BindCredentials struct {
    BotTokenCiphertext string `json:"bot_token_ciphertext"`
    BotTokenNonce      string `json:"bot_token_nonce"`
    IlinkBotID         string `json:"ilink_bot_id"`
    IlinkUserID        string `json:"ilink_user_id"`
}
```

#### 绑定状态常量

```go
// 绑定状态常量
const (
    BindStatusPending   = "pending"
    BindStatusScanned   = "scanned"
    BindStatusConfirmed = "confirmed"
    BindStatusExpired   = "expired"
)
```

#### 新增请求 DTO

```go
// GetBindStatusReq 查询绑定状态请求参数
type GetBindStatusReq struct {
    QRCode string `json:"qrcode" binding:"required"` // iLink 返回的二维码原始值
}
```

**设计说明**:

- **BindSession 结构体已完全删除**，不再作为数据传输对象
- Service 层方法直接返回基本类型（字符串、结构体指针）
- Handler 层直接使用 service 返回值，无需中间转换
- 只保留必要的 `BindCredentials` 用于返回凭证数据

### 2.8 新架构优势对比

| 维度               | 旧架构（有状态）                    | 新架构（无状态）     | 改进幅度   |
| ---------------- | --------------------------- | ------------ | ------ |
| **内存占用**         | O(N) N=活跃会话数                | O(1)         | 显著降低   |
| **后台 Goroutine** | 2 个清理线程                     | 0 个          | 消除     |
| **服务重启影响**       | 丢失所有会话                      | 无影响          | 完全解决   |
| **水平扩展**         | 需要 sticky session           | 任意路由         | 完全支持   |
| **代码行数**         | \~500 行（含测试）                | \~100 行      | 减少 80% |
| **锁竞争**          | 双读写锁                        | 无锁           | 消除     |
| **测试复杂度**        | 需要 mock 时间/并发               | 简单 mock API  | 显著降低   |
| **状态一致性**        | 依赖内存同步                      | 直接查询权威源      | 完全一致   |
| **概念复杂度**        | bind\_id + BindSession + 映射 | 仅 qrcode 字符串 | 极简     |

***

## 3. 接口变更对比

### 3.1 POST /api/channels/wechat-clawbot/bind

#### 请求参数（无变更）

```json
// 无需请求体，通过 JWT token 识别用户
Authorization: Bearer <access_token>
```

#### 响应格式变更

**旧响应**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "bind_id": "550e8400-e29b-41d4-a716-446655440000",
    "qrcode_url": "data:image/png;base64,...",
    "expires_at": 1715234567
  }
}
```

**新响应**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "qrcode_url": "data:image/png;base64,...",
    "qrcode": "ilink_qr_abc123xyz"
  }
}
```

**变更说明**:

| 字段           | 旧值       | 新值     | 说明                                  |
| ------------ | -------- | ------ | ----------------------------------- |
| `bind_id`    | UUID v4  | **移除** | 不再需要 bind\_id                       |
| `qrcode`     | 无        | **新增** | iLink 返回的二维码原始值，前端保存用于查询            |
| `qrcode_url` | 不变       | 不变     | 二维码图片数据 URL                         |
| `expires_at` | Unix 时间戳 | **移除** | 二维码过期由 iLink API 返回 `expired` 状态时判定 |

#### 前端适配要求

前端需要在 `StartBind` 成功后**保存** **`qrcode`** **值**，用于后续状态查询。

```javascript
// 旧代码
const { bind_id, qrcode_url } = await startBind()

// 新代码
const { qrcode_url, qrcode } = await startBind()
// qrcode_url 用于展示二维码图片
// qrcode 用于后续状态查询（保存在内存或 localStorage）
// 二维码过期由 iLink API 返回 expired 状态时判定
```

### 3.2 POST /api/channels/wechat-clawbot/bind/status

> **注意**: 原 `GET /bind/:bindId/wait` 接口完全移除，替换为新的 `POST /bind/status`

#### 请求参数

**旧接口** (已移除):

```
GET /api/channels/wechat-clawbot/bind/{bindId}/wait
Authorization: Bearer <access_token>
```

**新接口**:

```
POST /api/channels/wechat-clawbot/bind/status
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "qrcode": "ilink_qr_abc123xyz"
}
```

**请求体参数**:

| 参数       | 位置           | 类型     | 必填 | 说明              |
| -------- | ------------ | ------ | -- | --------------- |
| `qrcode` | Request Body | string | 是  | iLink 返回的二维码原始值 |

#### 响应格式

**旧响应**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "status": "confirmed",
    "credentials": {
      "bot_token_ciphertext": "...",
      "bot_token_nonce": "...",
      "ilink_bot_id": "...",
      "ilink_user_id": "..."
    }
  }
}
```

**新响应**（格式不变，仅来源不同）:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "status": "confirmed",
    "credentials": {
      "bot_token_ciphertext": "...",
      "bot_token_nonce": "...",
      "ilink_bot_id": "...",
      "ilink_user_id": "..."
    }
  }
}
```

**关键差异**:

- 旧方案: 通过 bind\_id 从 BindSession 内存缓存读取
- 新方案: 通过 qrcode 字符串实时从 iLink API 获取并加密

#### 错误码变更

| 场景             | 旧错误码             | 新错误码            | 说明        |
| -------------- | ---------------- | --------------- | --------- |
| bind\_id 不存在   | 110900 (绑定会话不存在) | **不再出现**        | 无状态设计无此概念 |
| qrcode 参数缺失    | 无                | 100002 (参数错误)   | 新增校验      |
| iLink API 调用失败 | 无错误（返回 pending）  | 无错误（返回 pending） | 保持不变      |

**ErrBindSessionNotFound 错误码清理**（审查意见采纳）:

- `ErrBindSessionNotFound`（错误码 110900）定义在 [codes.go#L95](file:///Users/lin/Documents/code/projects/OctoTify/backend/pkg/xerr/codes.go#L95)，当前仅被 `channel_handler.go` 引用
- 重构后该错误码不再使用，应从 `pkg/xerr/codes.go` 中删除
- 同一模块的其他错误码（`ErrBindQRCodeFailed` 110901、`ErrBindStatusFailed` 110902、`ErrBindExpired` 110903、`ErrBindEncryptFailed` 110904、`ErrBindDecryptFailed` 110905）仍需保留，它们在其他场景中继续使用

### 3.3 前端适配变更清单

#### 3.3.1 发起绑定流程

```javascript
// 旧实现
async function startBind() {
  const res = await axios.post('/api/channels/wechat-clawbot/bind')
  bindID.value = res.data.bind_id
  qrcodeURL.value = res.data.qrcode_url
  startPolling()
}

// 新实现
async function startBind() {
  const res = await axios.post('/api/channels/wechat-clawbot/bind')
  qrcodeURL.value = res.data.qrcode_url
  qrcode.value = res.data.qrcode        // 保存 qrcode 用于后续查询
  startPolling()
}
```

#### 3.3.2 查询状态流程

```javascript
// 旧实现
async function checkBindStatus() {
  const res = await axios.get(`/api/channels/wechat-clawbot/bind/${bindID.value}/wait`)
  return res.data
}

// 新实现
async function checkBindStatus() {
  const res = await axios.post('/api/channels/wechat-clawbot/bind/status', {
    qrcode: qrcode.value  // 直接传 qrcode 字符串
  })
  return res.data
}
```

**重要说明**：前端依然是长轮询模式，不需要改为主动轮询。iLink 的 `get_qrcode_status` API 本身就是长轮询设计（40s 超时），后端只是中转。前端发起请求后等待后端响应即可，行为与旧方案完全一致。

#### 3.3.3 状态处理（含 scanned 状态）

```javascript
// 旧实现：bind_id 对应的 session 过期后返回错误
// 新实现：二维码过期由 iLink API 返回 expired 状态时判定

async function checkBindStatus() {
  const res = await axios.post('/api/channels/wechat-clawbot/bind/status', {
    qrcode: qrcode.value
  })

  bindStatus.value = res.data.status

  switch (res.data.status) {
    case 'pending':
      // 仍在等待扫码，继续长轮询
      break
    case 'scanned':
      // 用户已扫码，待确认
      // 前端可展示"请在手机上确认"提示，继续长轮询等待 confirmed
      break
    case 'confirmed':
      // 绑定成功，处理凭证
      handleBindSuccess(res.data.credentials)
      break
    case 'expired':
      // 二维码已过期，提示用户重新获取
      break
  }
}
```

**scanned 状态处理说明**（审查意见采纳）：

- `scanned` 状态表示用户已扫码但尚未在手机上确认
- 前端收到 `scanned` 后应展示"请在手机上确认绑定"的提示
- 无需改变轮询行为，继续发起长轮询请求等待 `confirmed` 或 `expired`
- 该状态在旧方案中同样存在，前端处理逻辑保持一致

### 3.4 接口兼容性分析

| 接口                 | 向后兼容    | 需要前端改造 | 说明                    |
| ------------------ | ------- | ------ | --------------------- |
| POST /bind         | **不兼容** | 是      | 移除 bind\_id，新增 qrcode |
| GET /bind/:id/wait | **移除**  | 是      | 替换为 POST /bind/status |
| POST /bind/status  | **新增**  | 是      | 极简无状态接口               |

**迁移建议**: 由于接口变更需要前端配合，建议前后端同步发布。

***

## 4. 代码变更清单

### 4.1 需要删除的文件

| 文件路径                                               | 行数      | 说明                            |
| -------------------------------------------------- | ------- | ----------------------------- |
| `/backend/internal/service/bind_store.go`          | \~365 行 | 整个 BindStore 和 BindSession 实现 |
| `/backend/internal/service/bind_store_test.go`（如有） | -       | BindStore/BindSession 相关测试    |

### 4.2 需要修改的文件

#### 4.2.1 ChannelService ([channel\_service.go](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/service/channel_service.go))

**变更摘要**:

- `StartBind` 方法签名改为返回 `(qrcode, qrcodeURL string, err error)`
- `PollBindStatus` 方法签名改为 `(ctx context.Context, qrcode string) (status string, credentials *BindCredentials, err error)`
- 移除对 `BindSession` 的所有依赖
- 移除 `BindStatusPoller` 接口依赖（不再需要）

**详细变更**:

```diff
-func (s *ChannelService) StartBind(ctx context.Context, userID int64) (*BindSession, error) {
+func (s *ChannelService) StartBind(ctx context.Context, userID int64) (qrcode string, qrcodeURL string, err error) {
     qrResp, err := s.fetchQRCode(ctx)
     if err != nil {
-        return nil, fmt.Errorf("获取二维码失败: %w", err)
+        return "", "", fmt.Errorf("获取二维码失败: %w", err)
     }

-    session := &BindSession{
-        UserID:    userID,
-        QRCode:    qrResp.QRCode,
-        QRCodeURL: qrResp.QRCodeImgContent,
-        Status:    BindStatusPending,
-    }
-
-    return session, nil
+    return qrResp.QRCode, qrResp.QRCodeImgContent, nil
 }
```

```diff
-func (s *ChannelService) PollBindStatus(ctx context.Context, session *BindSession) (status string, credentials *BindCredentials, err error) {
+func (s *ChannelService) PollBindStatus(ctx context.Context, qrcode string) (status string, credentials *BindCredentials, err error) {
     // 内部实现直接使用 qrcode 字符串调用 iLink API
-    resp, err := s.ilinkClient.GetQRCodeStatus(ctx, session.QRCode)
+    resp, err := s.ilinkClient.GetQRCodeStatus(ctx, qrcode)
     if err != nil {
         return "", nil, err
     }
     // ... 解析响应、加密凭证等逻辑保持不变 ...
 }
```

#### 4.2.2 ChannelHandler ([channel\_handler.go](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/handler/channel_handler.go))

**变更摘要**:

- 移除 `bindStore` 依赖
- 移除 `uuid` 导入
- 简化 `StartBind` 方法（直接接收字符串返回值）
- 将 `GetBindStatusLongPolling` 替换为新的 `GetBindStatus` 方法
- `sendBindResponse` 签名改为接收 `(status string, credentials *service.BindCredentials)`
- 删除所有 bind\_id 和 BindSession 相关逻辑

**详细变更**:

```diff
 import (
     "encoding/json"
     "strconv"

     "github.com/gin-gonic/gin"
-    "github.com/google/uuid"
     "go.uber.org/zap"

     "octotify/internal/handler/dto"
     "octotify/internal/middleware"
     "octotify/internal/service"
     "octotify/pkg/response"
     "octotify/pkg/xerr"
 )

 type ChannelHandler struct {
     channelService *service.ChannelService
-    bindStore      *service.BindStore
     logger         *zap.Logger
 }

-func NewChannelHandler(channelService *service.ChannelService, bindStore *service.BindStore, logger *zap.Logger) *ChannelHandler {
-    return &ChannelHandler{channelService: channelService, bindStore: bindStore, logger: logger}
+func NewChannelHandler(channelService *service.ChannelService, logger *zap.Logger) *ChannelHandler {
+    return &ChannelHandler{channelService: channelService, logger: logger}
 }
```

```diff
 func (h *ChannelHandler) StartBind(c *gin.Context) {
     userID, ok := h.getUserID(c)
     if !ok {
         return
     }

-    session, err := h.channelService.StartBind(c.Request.Context(), userID)
+    qrcode, qrcodeURL, err := h.channelService.StartBind(c.Request.Context(), userID)
     if err != nil {
         c.Error(err)
         return
     }

-    bindID := uuid.New().String()
-    h.bindStore.Set(bindID, session)
-
     response.Success(c, gin.H{
-        "bind_id":    bindID,
-        "qrcode_url": session.QRCodeURL,
-        "expires_at": session.ExpiresAt.Unix(),
+        "qrcode_url": qrcodeURL,
+        "qrcode":     qrcode,
     })
 }
```

```diff
-func (h *ChannelHandler) GetBindStatusLongPolling(c *gin.Context) {
-    userID, ok := h.getUserID(c)
-    if !ok {
-        return
-    }
-
-    bindID := c.Param("bindId")
-    session, exists := h.bindStore.Get(bindID)
-    if !exists {
-        completedSession, completedExists := h.bindStore.GetCompletedSession(bindID)
-        if completedExists {
-            if completedSession.UserID != userID {
-                h.logger.Warn("绑定会话用户不匹配",
-                    zap.Int64("session_user_id", completedSession.UserID),
-                    zap.Int64("request_user_id", userID),
-                )
-                response.Fail(c, xerr.ErrUnauthorized.Code, "未提供认证令牌")
-                return
-            }
-            h.sendBindResponse(c, completedSession.Status, completedSession)
-            return
-        }
-        response.Fail(c, xerr.ErrBindSessionNotFound.Code, "绑定会话不存在或已过期")
-        return
-    }
-
-    if session.UserID != userID {
-        h.logger.Warn("绑定会话用户不匹配",
-            zap.Int64("session_user_id", session.UserID),
-            zap.Int64("request_user_id", userID),
-        )
-        response.Fail(c, xerr.ErrUnauthorized.Code, "未提供认证令牌")
-        return
-    }
-
-    if session.Status == service.BindStatusConfirmed || session.Status == service.BindStatusExpired {
-        h.sendBindResponse(c, session.Status, session)
-        return
-    }
-
-    h.bindStore.PollStatus(bindID)
-
-    latestSession, stillExists := h.bindStore.Get(bindID)
-    if !stillExists {
-        completedSession, completedExists := h.bindStore.GetCompletedSession(bindID)
-        if completedExists {
-            h.sendBindResponse(c, completedSession.Status, completedSession)
-            return
-        }
-
-        h.logger.Warn("长轮询：会话已删除且 completedSessions 中无数据，使用原始 session 返回",
-            zap.String("bind_id", bindID),
-        )
-        h.sendBindResponse(c, session.Status, session)
-        return
-    }
-
-    h.sendBindResponse(c, latestSession.Status, latestSession)
-}
+
+// GetBindStatus godoc
+// @Summary      查询微信ClawBot绑定状态
+// @Description  通过 qrcode 调用 iLink API 查询绑定状态（长轮询，40s 超时）。
+// @Description  ## 工作原理
+// @Description  1. 前端传 qrcode 字符串给后端
+// @Description  2. 后端直接调用 iLink API 查询状态（iLink 本身是长轮询设计，40s 超时）
+// @Description  3. iLink 返回状态时立即返回给前端
+// @Description  4. 如果 iLink 超时，返回 pending，前端可重新发起请求
+// @Description  ## 返回状态
+// @Description  - pending: 仍在等待中（超时返回）
+// @Description  - scanned: 用户已扫码，待确认
+// @Description  - confirmed: 绑定成功，返回凭证
+// @Description  - expired: 二维码已过期
+// @Description  ## 错误码说明
+// @Description  - 100001: 未提供认证令牌
+// @Description  - 100002: 请求参数格式错误
+// @Tags         推送渠道管理
+// @Accept       json
+// @Produce      json
+// @Param        body  body      dto.GetBindStatusReq  true  "查询绑定状态请求参数"
+// @Success      200   {object}  response.Response     "查询成功"
+// @Router       /channels/wechat-clawbot/bind/status [post]
+// @Security     BearerAuth
+func (h *ChannelHandler) GetBindStatus(c *gin.Context) {
+    userID, ok := h.getUserID(c)
+    if !ok {
+        return
+    }
+
+    var req dto.GetBindStatusReq
+    if err := c.ShouldBindJSON(&req); err != nil {
+        response.Fail(c, xerr.ErrBadRequest.Code, "请求参数格式错误")
+        return
+    }
+
+    // 审计日志：记录用户查询绑定状态
+    h.logger.Info("查询绑定状态",
+        zap.Int64("user_id", userID),
+    )
+
+    // 直接用 qrcode 字符串调用 service，无需构造任何中间结构体
+    // 注意：iLink 的 get_qrcode_status API 本身是长轮询设计（40s 超时），
+    // 后端只是中转，前端依然是长轮询模式，行为与旧方案一致
+    status, credentials, err := h.channelService.PollBindStatus(c.Request.Context(), req.QRCode)
+    if err != nil {
+        // iLink API 调用失败，返回 pending 让前端重试
+        response.Success(c, gin.H{"status": service.BindStatusPending})
+        return
+    }
+
+    h.sendBindResponse(c, status, credentials)
+}
```

**修改后的 sendBindResponse 方法**:

```go
// sendBindResponse 统一发送绑定状态响应
func (h *ChannelHandler) sendBindResponse(c *gin.Context, status string, credentials *service.BindCredentials) {
    result := gin.H{
        "status": status,
    }
    if status == service.BindStatusConfirmed && credentials != nil {
        result["credentials"] = gin.H{
            "bot_token_ciphertext": credentials.BotTokenCiphertext,
            "bot_token_nonce":      credentials.BotTokenNonce,
            "ilink_bot_id":         credentials.IlinkBotID,
            "ilink_user_id":        credentials.IlinkUserID,
        }
    }

    response.Success(c, result)
}
```

#### 4.2.3 Server ([server.go](file:///Users/lin/Documents/code/projects/OctoTify/backend/internal/server/server.go))

**变更摘要**:

- 移除 `bindStore` 字段
- 移除 `NewBindStore` 初始化
- 移除 `bindStore.Close()` 调用
- 修改 `NewChannelHandler` 调用参数
- 更新路由注册：`GET /bind/:bindId/wait` 替换为 `POST /bind/status`

**详细变更**:

```diff
 type Server struct {
     engine     *gin.Engine
     addr       string
     serverName string
     logger     *zap.Logger
     cfg        *config.Config

     authHandler     *handler.AuthHandler
     userHandler     *handler.UserHandler
     sourceHandler   *handler.SourceHandler
     channelHandler  *handler.ChannelHandler
     messageHandler  *handler.MessageHandler
     pushHandler     *handler.PushHandler
     accessJWTHelper *jwtx.JWTHelper
-    bindStore       *service.BindStore
 }
```

```diff
 func (s *Server) initDependencies(cfg *config.Config, db *gorm.DB, logger *zap.Logger) {
     // ... 其他初始化逻辑 ...

     senderFactory := sender.NewSenderFactory(s.logger)
     channelService := service.NewChannelService(db, logger, senderFactory)
-    bindStore := service.NewBindStore(channelService, logger)
-    s.bindStore = bindStore
-    s.channelHandler = handler.NewChannelHandler(channelService, bindStore, logger)
+    s.channelHandler = handler.NewChannelHandler(channelService, logger)

     // ... 其他初始化逻辑 ...
 }
```

```diff
 func (s *Server) Close() {
-    if s.bindStore != nil {
-        s.bindStore.Close()
-    }
     s.logger.Info("服务器资源已关闭")
 }

 func (s *Server) Shutdown(ctx context.Context) error {
-    if s.bindStore != nil {
-        s.bindStore.CloseWithTimeout(ctx)
-    }
     s.logger.Info("服务器资源已关闭")
     return nil
 }
```

**路由注册变更**:

```diff
 // 绑定相关路由
 wechatGroup.POST("/bind", h.channelHandler.StartBind)
-wechatGroup.GET("/bind/:bindId/wait", h.channelHandler.GetBindStatusLongPolling)
+wechatGroup.POST("/bind/status", h.channelHandler.GetBindStatus)
```

#### 4.2.4 DTO 定义 (新建)

**文件**: `/backend/internal/handler/dto/bind_dto.go`（新建）

```go
package dto

// GetBindStatusReq 查询绑定状态请求参数
type GetBindStatusReq struct {
    QRCode string `json:"qrcode" binding:"required"` // iLink 返回的二维码原始值
}
```

**校验说明**:

- 仅校验 `qrcode` 字段必填（`binding:"required"`），不做长度校验
- qrcode 值由第三方 iLink 生成，长度不确定且可能变化，不应硬编码长度限制
- 完全依赖第三方 API 的合法性校验

#### 4.2.5 BindCredentials、状态常量和 PollAPITimeout (保留)

**文件**: `/backend/internal/service/bind_types.go`（新建，保留必要类型）

```go
package service

import "time"

// 绑定状态常量
const (
    BindStatusPending   = "pending"
    BindStatusScanned   = "scanned"
    BindStatusConfirmed = "confirmed"
    BindStatusExpired   = "expired"
)

// PollAPITimeout 轮询 iLink API 的 HTTP 请求超时
// iLink 的 get_qrcode_status 接口本身就是长轮询设计，需要 35-40 秒才返回
// 从 bind_store.go 迁移至此，因为 BindStore 删除后原文件不再存在
const PollAPITimeout = 40 * time.Second

// BindCredentials 绑定成功后的凭证
type BindCredentials struct {
    BotTokenCiphertext string `json:"bot_token_ciphertext"`
    BotTokenNonce      string `json:"bot_token_nonce"`
    IlinkBotID         string `json:"ilink_bot_id"`
    IlinkUserID        string `json:"ilink_user_id"`
}
```

**设计说明**:

- **BindSession 结构体已完全删除**
- 仅保留 `BindCredentials` 用于返回凭证数据
- 状态常量保留在 service 包中供 Handler 层引用
- **`PollAPITimeout` 常量从 `bind_store.go` 迁移至此**（审查意见采纳）：BindStore 删除后原文件不存在，该常量仍被 `channel_service.go` 中的 `PollBindStatus` 方法使用
- `bind_store.go` 中其他仅服务于 BindStore 的常量（`LongPollTimeout`、`CleanupInterval`、`SessionTTL`、`CompletedSessionTTL`）随文件一起删除
- 所有方法签名均使用基本类型而非结构体

### 4.3 代码变更统计

| 操作     | 文件数     | 行数变化             | 说明                              |
| ------ | ------- | ---------------- | ------------------------------- |
| 删除     | 1-2     | -365 \~ -400     | bind\_store.go 及其测试文件           |
| 修改     | 4-5     | +30 \~ -120      | channel\_handler.go, channel\_service.go, server.go, codes.go |
| 新增     | 1-2     | +40 \~ +60       | bind\_types.go, bind\_dto.go    |
| **合计** | **6-9** | **-295 \~ -460** |                                 |

**新增修改文件**（审查意见采纳后）:

- `pkg/xerr/codes.go`：删除 `ErrBindSessionNotFound` 错误码定义

### 4.4 依赖关系变更

#### 删除的依赖

```go
// bind_store.go 中的依赖（整个文件删除）
import (
    "context"      // 不再需要（goroutine 中）
    "sync"         // 不再需要（无锁）
    "time"         // 部分保留
    "go.uber.org/zap" // 部分保留
)
```

#### channel\_handler.go 中的依赖变更

```diff
 import (
     "encoding/json"
     "strconv"

     "github.com/gin-gonic/gin"
-    "github.com/google/uuid"  // 不再需要

     "go.uber.org/zap"

     "octotify/internal/handler/dto"
     "octotify/internal/middleware"
     "octotify/internal/service"
     "octotify/pkg/response"
     "octotify/pkg/xerr"
 )
```

***

## 5. 技术风险

### 5.1 风险清单

| 风险编号 | 风险描述                | 影响程度 | 发生概率 | 缓解措施                   |
| ---- | ------------------- | ---- | ---- | ---------------------- |
| R1   | iLink API 稳定性       | 高    | 低    | 保持超时重试机制，失败返回 pending  |
| R2   | 前端未传递 qrcode 参数     | 中    | 中    | 严格的参数校验和错误提示           |
| R3   | QRCode 值泄露风险        | 低    | 低    | QRCode 本身无敏感信息，仅用于状态查询 |
| R4   | 并发查询导致 iLink API 限流 | 中    | 低    | 前端控制请求频率，保持现有重试间隔      |

### 5.2 风险缓解方案

#### R1: iLink API 稳定性

**风险场景**: iLink API 不可用或响应缓慢。

**缓解措施**:

```go
// 保持现有的 40 秒超时
ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
defer cancel()

// 调用失败时返回 pending，让前端重试
status, credentials, err := s.PollBindStatus(ctx, qrcode)
if err != nil {
    return BindStatusPending, nil  // 前端将重新发起请求
}
```

**监控指标**:

- iLink API 调用成功率
- 平均响应时间
- 超时率

#### R2: 前端未传递 qrcode 参数

**风险场景**: 前端代码未升级或遗漏 qrcode 参数。

**缓解措施**:

1. 服务端严格校验参数
2. 返回明确的错误信息
3. 前端增加防御性检查

```go
var req dto.GetBindStatusReq
if err := c.ShouldBindJSON(&req); err != nil || req.QRCode == "" {
    response.Fail(c, xerr.ErrBadRequest.Code, "缺少 qrcode 参数，请刷新页面重试")
    return
}
```

#### R3: QRCode 值泄露风险

**风险场景**: qrcode 值被第三方获取。

**影响分析**: qrcode 仅用于状态查询，不包含敏感用户信息或凭证数据，泄露风险极低。

#### R4: 并发查询导致 iLink API 限流

**风险场景**: 前端频繁轮询导致 iLink API 触发限流。

**缓解措施**:

- 前端保持现有轮询间隔（如每 3-5 秒一次）
- 绑定完成后立即停止轮询
- 可在服务端增加请求频率限制

***

## 附录

### A. 无状态设计的扩展性分析

#### A.1 水平扩展能力

```
┌─────────────────────────────────────────────────┐
│              多实例部署架构                       │
├─────────────────────────────────────────────────┤
│                                                 │
│                    ┌─────────┐                  │
│                    │   LB    │                  │
│                    │ (Nginx) │                  │
│                    └────┬────┘                  │
│                  ┌──────┼──────┐               │
│                  ▼      ▼      ▼               │
│            ┌─────────┐ ┌─────────┐ ┌─────────┐ │
│            │Instance A│ │Instance B│ │Instance C│ │
│            │         │ │         │ │         │ │
│            │ 无状态  │ │ 无状态  │ │ 无状态  │ │
│            │ 可处理  │ │ 可处理  │ │ 可处理  │ │
│            │ 任意请求│ │ 任意请求│ │ 任意请求│ │
│            └─────────┘ └─────────┘ └─────────┘ │
│                                                 │
│   任意请求路由策略:                              │
│   - 轮询 (Round Robin)                          │
│   - 最少连接 (Least Connections)                │
│   - 随机 (Random)                               │
│   均可正常工作，无需 sticky session              │
│                                                 │
└─────────────────────────────────────────────────┘
```

#### A.2 性能对比

| 场景     | 旧架构                           | 新架构              |
| ------ | ----------------------------- | ---------------- |
| 单次查询延迟 | 内存读取（\~1μs）+ iLink API（\~40s） | iLink API（\~40s） |
| 并发查询能力 | 受限于内存锁竞争                      | 无锁，完全并行          |
| 内存占用   | O(N) N=会话数                    | O(1)             |
| 实例扩展   | 需要 sticky session             | 任意扩展             |

**结论**: 由于 iLink API 调用耗时（\~40s）远大于内存操作（\~1μs），移除内存缓存对整体延迟的影响可忽略不计。

### B. 测试策略变更

#### B.1 旧架构测试复杂度

```go
// 需要 mock:
// - 并发（用于测试锁竞争）
// - goroutine 生命周期
// - 双 map 状态同步

func TestBindStore_CleanupExpiredSessions(t *testing.T) {
    // 需要模拟时间流逝
    // 需要等待 goroutine 执行
    // 需要验证 map 状态
}
```

#### B.2 新架构测试策略

**当前阶段：先跑通主流程，不写单元测试**（审查意见采纳）

重构的首要目标是验证无状态方案的可行性，优先确保主流程端到端跑通。单元测试在主流程稳定后再补充。

后续补充测试时，只需 mock iLink API 响应：

```go
func TestChannelService_PollBindStatus(t *testing.T) {
    // 简单的 HTTP mock
    // 验证状态解析
    // 验证凭证加密
}

func TestGetBindStatus_Handler(t *testing.T) {
    // 验证参数校验
    // 验证 iLink API 调用
    // 验证响应格式
}
```

### C. 监控指标建议

#### C.1 关键指标

| 指标名称                              | 指标类型      | 告警阈值      | 说明               |
| --------------------------------- | --------- | --------- | ---------------- |
| `bind_start_total`                | Counter   | -         | 发起绑定总数           |
| `bind_status_query_total`         | Counter   | -         | 状态查询总数           |
| `bind_ilink_api_duration_seconds` | Histogram | p99 > 45s | iLink API 调用耗时   |
| `bind_ilink_api_errors_total`     | Counter   | 错误率 > 5%  | iLink API 错误数    |
| `bind_qrcode_missing_total`       | Counter   | -         | 缺少 qrcode 参数的请求数 |

#### C.2 日志建议

```go
// StartBind
logger.Info("发起微信ClawBot绑定",
    zap.Int64("user_id", userID),
)

// GetBindStatus
logger.Info("查询绑定状态",
    zap.Int64("user_id", userID),
    zap.String("status", status),
)

// 错误日志
logger.Error("iLink API 调用失败",
    zap.Int64("user_id", userID),
    zap.Error(err),
)
```

***

## 6. 审查意见与修订记录

### 6.1 前端审查意见纠正

#### "后端从长轮询改为直接返回" -- 此意见错误

**原审查意见**: 认为新方案将后端从长轮询改为直接返回结果。

**纠正**: 前端依然是长轮询模式，行为没有变化。

| 维度     | 旧方案                                       | 新方案                                       | 是否变化 |
| ------ | ------------------------------------------ | ------------------------------------------ | ---- |
| 前端请求方式 | 前端请求 → 等待后端响应                              | 前端请求 → 等待后端响应                              | 无变化  |
| 后端处理流程 | 从内存取 session → 调用 iLink API（长轮询，40s 超时）→ 返回 | 直接调用 iLink API（长轮询，40s 超时）→ 返回             | 仅省去内存读取 |
| iLink API | get\_qrcode\_status，长轮询设计，40s 超时           | get\_qrcode\_status，长轮询设计，40s 超时           | 无变化  |
| 前端等待时间 | 最长 40s                                     | 最长 40s                                     | 无变化  |

**结论**: iLink 的 `get_qrcode_status` API 本身就是长轮询设计，后端只是中转。新方案去掉了内存 session 的中间环节，但前端的长轮询行为完全不变。前端不需要改为主动轮询。

### 6.2 已采纳的审查意见

| 编号  | 审查意见                          | 来源 | 修订位置                   | 修订内容                                              |
| --- | ----------------------------- | -- | ---------------------- | ------------------------------------------------- |
| A1  | PollAPITimeout 常量迁移到 bind\_types.go | 后端 | 4.2.5 节                | 将 `PollAPITimeout` 从 `bind_store.go` 迁移至 `bind_types.go`，其余 BindStore 专用常量随文件删除 |
| A2  | ErrBindSessionNotFound 错误码清理  | 后端 | 3.2 节错误码变更             | 从 `pkg/xerr/codes.go` 中删除 `ErrBindSessionNotFound`（110900），同模块其他错误码保留 |
| A3  | Handler 层增加 userID 审计日志       | 后端 | 2.6.1 节、2.6.2 节        | StartBind 和 GetBindStatus 方法中增加 `zap.Int64("user_id", userID)` 审计日志 |
| A4  | 不需要 qrcode 长度校验               | 前端 | 4.2.4 节 DTO 定义         | 仅 `binding:"required"`，不做长度校验，完全依赖第三方 API 合法性    |
| A5  | 不需要写单元测试，先跑通主流程              | 后端 | 附录 B.2 节               | 测试策略调整为"先跑通主流程，不写单元测试"，后续稳定后再补充                  |
| A6  | 补充 scanned 状态的前端处理说明          | 前端 | 3.3.3 节、2.6.2 节 godoc  | 增加 scanned 状态的 switch-case 处理说明和 godoc 描述         |
| A7  | 前端依然是长轮询，不需要改为主动轮询           | 前端 | 3.3.2 节、2.6.2 节        | 明确标注"前端依然是长轮询模式"，纠正"后端从长轮询改为直接返回"的错误认知          |

### 6.3 后端审查意见分析

以下三条后端审查意见经分析后给出判断：

#### 6.3.1 是否需要 singleflight 请求去重？

**审查意见**: 对同一 qrcode 的并发请求使用 `golang.org/x/sync/singleflight` 进行去重，避免重复调用 iLink API。

**分析结论: 当前阶段不需要，遵循 YAGNI 原则。**

理由：

1. **并发场景不存在**: 前端是串行长轮询模式（等上一个请求返回才发下一个），同一用户同一 qrcode 的并发请求在正常使用中不会发生
2. **多标签页场景有限**: 即使同一用户打开多个标签页，每个标签页各自维护独立的轮询循环，iLink API 本身支持并发查询
3. **引入复杂度大于收益**: singleflight 需要管理 key 生命周期、处理 context 取消、处理共享结果的生命周期，增加代码复杂度
4. **与无状态设计冲突**: singleflight 本质是内存级的有状态缓存，与"完全无状态"的设计原则相悖
5. **iLink API 无限流风险**: 当前用户规模下，iLink API 不会因为少量并发请求触发限流

**预留方案**: 如果未来用户规模增长导致 iLink API 限流，可在 Handler 层引入基于 Redis 的分布式请求去重，而非进程内 singleflight。

#### 6.3.2 是否需要 BindQRCodeResult 结构体替代多返回值？

**审查意见**: 将 `PollBindStatus` 的 `(status string, credentials *BindCredentials, err error)` 三返回值封装为 `BindQRCodeResult` 结构体。

**分析结论: 不需要，当前多返回值足够清晰。**

理由：

1. **Go 惯用风格**: Go 语言中 2-3 个返回值是惯用写法，标准库和主流框架大量使用多返回值（如 `map` 的 `value, ok`、`channel` 的 `value, ok`）
2. **返回值语义清晰**: `status`、`credentials`、`err` 三个变量名已经自解释，不存在歧义
3. **结构体增加代码量**: 引入 `BindQRCodeResult` 需要定义结构体、构造实例、访问字段，代码量增加但可读性不提升
4. **与已有模式一致**: 项目中 `StartBind` 返回 `(qrcode, qrcodeURL string, err error)` 也是多返回值风格，保持一致性
5. **仅在返回值超过 4 个时才考虑结构体**: 当前场景不满足此阈值

#### 6.3.3 是否需要旧接口保留过渡期？

**审查意见**: 保留旧接口 `GET /bind/:bindId/wait` 一段时间，给前端逐步迁移的缓冲期。

**分析结论: 不需要保留旧接口，前后端同步发布即可。**

理由：

1. **旧接口依赖 BindStore**: 旧接口 `GET /bind/:bindId/wait` 的核心依赖是 `BindStore` 内存存储。如果保留旧接口，就必须保留 `BindStore`，这完全违背了重构的初衷——消除有状态设计
2. **新旧接口无法共存**: 新接口使用 qrcode 查询 iLink API，旧接口使用 bind\_id 查询内存，两者没有共享的状态源，无法平滑过渡
3. **前后端同步发布成本低**: 绑定功能是低频操作（用户仅在首次绑定时使用），不存在高并发切换风险，前后端同步发布的成本和风险都很低
4. **用户影响极小**: 绑定流程在页面刷新后就会重新开始，不存在需要保留旧接口兼容的长期会话
5. **蓝绿部署已提供回滚能力**: 部署策略已规划蓝绿部署，如出现问题可直接回滚到旧版本

### 6.4 修订版本对照

| 文档版本 | 修订日期       | 修订内容                                       |
| ---- | ---------- | ------------------------------------------ |
| v4.0 | 2026-05-09 | 初始重构方案                                     |
| v5.0 | 2026-05-09 | 整合前后端审查意见：纠正长轮询误解、采纳 7 条意见、分析 3 条后端意见并给出结论 |

***

## 总结

### 核心改进

| 改进维度  | 旧架构                                              | 新架构                  | 收益         |
| ----- | ------------------------------------------------ | -------------------- | ---------- |
| 架构复杂度 | 有状态 + 双缓存 + 双 Goroutine + bind\_id + BindSession | 完全无状态 + 仅 qrcode 字符串 | 降低 80% 代码量 |
| 可扩展性  | 需要 sticky session                                | 任意负载均衡               | 支持水平扩展     |
| 可靠性   | 重启丢失状态                                           | 状态持久于 iLink          | 消除单点故障     |
| 测试难度  | 需要 mock 时间/并发/锁                                  | 简单 HTTP mock         | 测试覆盖率提升    |
| 运维成本  | 需要监控内存/ goroutine                                | 只需监控 API             | 降低运维负担     |
| 概念理解  | bind\_id + BindSession + 映射 + 校验                 | 仅 qrcode 字符串         | 极简认知负担     |

### 实施建议

1. **优先级**: P0（核心业务优化）
2. **预估工期**: 3-5 天（含前后端，不含单元测试）
3. **部署策略**: 蓝绿部署 + 灰度发布，前后端同步发布（旧接口不保留过渡期）
4. **回滚方案**: 保留旧版本代码 7 天，蓝绿部署可直接回滚
5. **测试策略**: 先跑通主流程端到端验证，单元测试后续补充

***

> **文档结束**
>
> 如需进一步讨论或调整方案，请联系架构团队。

