# TODO: 微信 ClawBot 激活状态检查功能设计文档

> **状态**: 待实施
> **优先级**: 中（主流程完成后实施）
> **最后更新**: 2026-05-08

---

## 一、需求背景

### 1.1 问题描述

当前扫码绑定微信 ClawBot 后，前端直接显示"绑定成功"，并提示"请在24小时内向Bot发送任意消息以激活推送"。但用户绑定后往往忘记发送消息激活，导致推送功能无法使用。

### 1.2 解决方案

增加"待激活"和"已激活"两个前端状态提示，让用户清楚知道是否已完成激活。

### 1.3 核心需求

- **前端状态展示**：显示"待激活"或"已激活"状态
- **不影响提交**：用户可以在未激活状态下直接点击"创建"按钮
- **后端不保存状态**：激活状态仅前端展示，后端不持久化
- **实时检测**：通过长轮询查询 iLink API 判断用户是否已发送消息

---

## 二、目前解决方案概要

### 2.1 后端方案

**新增接口**: `GET /api/channels/wechat-clawbot/bind/:bindId/check-activation`

**核心逻辑**:
1. 通过 bindId 从 BindStore 获取凭证（包括 completedSessions 缓存）
2. 校验凭证有效性（复用 `ProcessChannelConfig`）
3. 解密 BotToken
4. 调用 iLink getUpdates API（30秒超时）
5. 返回激活状态和消息内容

**关键发现**:
- `sessions` 是**内存存储**（`map[string]*BindSession`）
- confirmed 状态后，会话移到 `completedSessions` 缓存（保留 2 分钟）
- 获取会话模式：先查 `Get()`，不存在再查 `GetCompletedSession()`

### 2.2 前端方案

**新增 composable**: `useActivationPoll`

**核心逻辑**:
1. confirmed 状态后自动启动激活检查
2. 循环调用 check-activation 接口
3. `activated=true` → 显示"已激活" + 消息内容 → 停止轮询
4. `activated=false` → 等待 500ms → 继续轮询
5. 页面卸载时停止轮询

### 2.3 超时策略

| 层级 | 超时时间 | 说明 |
|------|----------|------|
| iLink getUpdates | 30s | 长轮询等待消息 |
| 后端 HTTP 响应 | 30s | 与 iLink 一致 |
| 前端 Axios | 35s | 覆盖后端超时 |

---

## 三、TODO 清单

### TODO 1: 后端开发

- [ ] **TODO**: 在 `ChannelService` 中新增 `CheckActivation` 方法
  - 调用 iLink getUpdates API
  - 30秒超时
  - 解析响应，查找 `message_type=1 && message_state=2` 的用户消息
  - 返回激活状态和消息内容

- [ ] **TODO**: 在 `ChannelHandler` 中新增 `CheckActivation` Handler
  - 通过 bindId 从 BindStore 获取凭证
  - 校验凭证有效性（复用 `ProcessChannelConfig`）
  - 解密 BotToken
  - 调用 `ChannelService.CheckActivation`
  - 返回结果

- [ ] **TODO**: 新增 `getBindSession` 辅助方法
  - 复用现有模式：先查 `Get()`，再查 `GetCompletedSession()`
  - 校验用户权限

- [ ] **TODO**: 注册路由
  - 在 `server.go` 中添加 `GET /bind/:bindId/check-activation`

- [ ] **TODO**: 新增错误码（可选）
  - `ErrActivationCheckFailed = New(110902, "激活状态检查失败")`

### TODO 2: 单元测试

- [ ] **TODO**: `ChannelService.CheckActivation` 单元测试
  - Mock iLink API 响应
  - 测试有用户消息的情况
  - 测试无用户消息的情况
  - 测试 API 调用失败的情况

- [ ] **TODO**: `ChannelHandler.CheckActivation` 单元测试
  - 测试正常流程
  - 测试凭证不存在
  - 测试凭证校验失败
  - 测试用户权限校验

### TODO 3: 前端开发

- [ ] **TODO**: 创建 `useActivationPoll` composable
  - 实现长轮询逻辑
  - 实现重试机制（500ms/1s）
  - 实现取消逻辑

- [ ] **TODO**: 新增 `checkActivationByBindId` API 调用
  - 在 `channels.ts` 中添加

- [ ] **TODO**: 修改 `ChannelCreateView.vue`
  - confirmed 状态下显示"待激活"提示
  - 激活成功后显示"已激活" + 消息内容
  - 添加加载动画和样式

- [ ] **TODO**: 修改 `useWechatBind`
  - confirmed 状态后自动启动激活轮询
  - 暴露 `isActivated` 和 `messageContent` 状态

- [ ] **TODO**: 页面卸载时清理
  - 在 `onBeforeUnmount` 中调用 `stopActivationPoll()`

---

## 四、完整方案详情（供实施参考）

### 4.1 iLink getUpdates API 详情

参考项目：`tmp/weclaw-main/ilink/`

**请求格式**：
```http
POST https://ilinkai.weixin.qq.com/ilink/bot/getupdates
Content-Type: application/json
AuthorizationType: ilink_bot_token
Authorization: Bearer {bot_token}
X-WECHAT-UIN: {base64_encoded_uin}

{
  "get_updates_buf": "",
  "base_info": {
    "channel_version": "1.0.0"
  }
}
```

**响应格式**：
```json
{
  "ret": 0,
  "errcode": 0,
  "errmsg": "",
  "msgs": [
    {
      "seq": 123,
      "message_id": 456,
      "from_user_id": "user123@im.wechat",
      "to_user_id": "bot456@im.bot",
      "message_type": 1,
      "message_state": 2,
      "item_list": [
        {
          "type": 1,
          "text_item": {
            "text": "你好"
          }
        }
      ],
      "context_token": "xxx"
    }
  ],
  "get_updates_buf": "新的游标",
  "longpolling_timeout_ms": 35000
}
```

**消息类型常量**：
- `message_type = 1`：用户发送的消息
- `message_type = 2`：Bot 发送的消息
- `message_state = 0`：新建
- `message_state = 1`：生成中
- `message_state = 2`：完成

### 4.2 BindStore 会话管理

```go
// 活跃会话（绑定中）
sessions map[string]*BindSession

// 已完成会话缓存（confirmed/expired 后保留 2 分钟）
completedSessions map[string]*completedSession
```

**凭证结构**：
```go
type BindCredentials struct {
    BotTokenCiphertext string `json:"bot_token_ciphertext"` // 加密的 BotToken
    BotTokenNonce      string `json:"bot_token_nonce"`      // 加密 nonce
    IlinkBotID         string `json:"ilink_bot_id"`         // iLink 机器人 ID
    IlinkUserID        string `json:"ilink_user_id"`        // iLink 用户 ID
}
```

### 4.3 后端接口响应格式

已激活：
```json
{
  "code": 200,
  "data": {
    "activated": true,
    "message_content": "你好",
    "received_at": "2026-05-08T10:30:00Z"
  }
}
```

未激活：
```json
{
  "code": 200,
  "data": {
    "activated": false
  }
}
```

### 4.4 前端重试逻辑

```
confirmed 状态
    ↓
发起 check-activation (30s超时)
    ↓
activated=true → 显示"已激活" + 消息内容 → 停止轮询
    ↓
activated=false → 等待 500ms → 继续轮询
    ↓
请求失败/超时 → 等待 1s → 继续轮询
    ↓
页面卸载/bindStatus改变 → 停止轮询
```

---

## 五、关键设计决策

### 5.1 为什么不修改数据库？

1. **激活状态是临时状态**：一旦用户创建渠道，激活状态即完成使命
2. **降低系统复杂度**：无需新增表或字段
3. **可接受的重试成本**：服务重启后用户可重新绑定

### 5.2 为什么从 BindStore 获取凭证？

1. **confirmed 状态后，凭证已保存在 completedSessions 缓存**
2. **前端不需要传递凭证**，减少安全风险
3. **和现有 GetBindStatusLongPolling 模式一致**

### 5.3 为什么不使用 get_updates_buf 游标？

1. **激活检查是独立场景**：每次都是查询"是否有消息"，不需要增量同步
2. **简化实现**：空游标会返回最新消息，满足需求
3. **无需持久化**：服务重启后不影响功能

---

## 六、完整时序图

```
用户           前端                后端                iLink API
 |             |                   |                     |
 |--扫码绑定-->|                   |                     |
 |             |--POST /bind------>|                     |
 |             |                   |--get_qrcode_status->|
 |             |<--confirmed-------|                     |
 |             |  +bindID          |                     |
 |             |  +credentials     |                     |
 |             |                   |                     |
 |<--显示绑定成功-|                  |                     |
 |  显示"待激活" |                  |                     |
 |             |                   |                     |
 |             |--GET /check------>|                     |
 |             |  activation       |                     |
 |             |  /bind/:bindId    |                     |
 |             |                   |--从 BindStore 获取凭证|
 |             |                   |--校验凭证 + 解密     |
 |             |                   |                     |
 |             |                   |--getUpdates(30s)--->|
 |             |                   |                     |
 |             |                   |<--msgs:[]-----------|
 |             |<--activated:false-|                     |
 |             |                   |                     |
 |             |--等待 500ms------>|                     |
 |             |                   |                     |
 |             |--GET /check------>|                     |
 |             |  activation       |                     |
 |             |                   |--getUpdates-------->|
 |             |                   |<--msgs:[用户消息]---|
 |             |                   |                     |
 |             |<--activated:true--|                     |
 |             |  +message_content |                     |
 |             |                   |                     |
 |<--显示已激活-|                   |                     |
 |  "✅ Bot收到消息：你好"          |                     |
 |             |                   |                     |
 |--点击创建-->|                   |                     |
 |             |--POST /channels-->|                     |
 |             |  (credentials)    |                     |
 |             |<--创建成功--------|                     |
 |             |                   |                     |
```

---

## 七、风险与应对

| 风险 | 影响 | 应对措施 |
|------|------|----------|
| iLink getUpdates API 格式不确定 | 开发延期 | 先写 Mock，后续替换真实 API |
| 长轮询连接数过多 | 服务器负载 | 前端控制重试频率，500ms 间隔 |
| 凭证过期 | 功能失效 | completedSessions 保留 2 分钟，超时后提示重新绑定 |
| 内存泄漏 | 服务不稳定 | BindStore 已有自动清理机制 |

---

## 八、参考文档

- [iLink API 源码](../tmp/weclaw-main/ilink/) - client.go, types.go, monitor.go
- [现有绑定流程](../backend/internal/service/bind_store.go)
- [现有轮询逻辑](../backend/internal/handler/channel_handler.go)
