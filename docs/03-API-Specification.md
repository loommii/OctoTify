# 后端 API 规范

> 记录 OctoTify 后端 API 的设计规范，包括响应格式、错误处理、鉴权机制等。

---

## 四、JSON 字段命名规范

### 4.1 基本原则

**所有 JSON 字段统一使用 snake_case（下划线命名法）。**

### 4.2 命名规则

| 类型 | 规则 | 示例 |
|------|------|------|
| 单单词字段 | 全小写 | `id`、`name`、`timestamp` |
| 多单词字段 | 下划线分隔，全小写 | `server_name`、`created_at`、`user_id` |
| 布尔值字段 | `is_` / `has_` / `can_` 前缀 | `is_active`、`has_permission` |
| 枚举值字段 | `_type` / `_status` 后缀 | `source_type`、`order_status` |

### 4.3 JSON 示例

```json
{
  "id": 1,
  "name": "CI Pipeline",
  "is_active": true,
  "source_type": "ci",
  "created_at_ts": 1705298400123
}
```

### 4.4 禁止的命名风格

| 风格 | 示例 | 说明 |
|------|------|------|
| camelCase | `serverName` | 不使用，与前端 JavaScript 变量命名冲突 |
| PascalCase | `ServerName` | 不使用 |
| SCREAMING_SNAKE_CASE | `SERVER_NAME` | 不使用，仅用于常量 |

### 4.5 设计理由

1. **与数据库一致**：数据库字段通常使用 snake_case，减少映射转换
2. **RESTful 惯例**：大多数 RESTful API 使用 snake_case 作为 JSON 字段命名
3. **可读性**：多单词字段用下划线分隔，比 camelCase 更易读

---

## 五、时间字段规范

### 5.1 响应格式

| 字段后缀 | 类型 | 格式 | 示例 |
|----------|------|------|------|
| `_at` | string | `YYYY-MM-DD HH:MM:SS` | `2024-01-15 14:30:00` |
| `_ts` | int64 | Unix 毫秒时间戳 | `1705298400123` |

### 5.2 字段命名规则

| 含义 | `_at` 字段（字符串） | `_ts` 字段（时间戳） |
|------|---------------------|---------------------|
| 创建时间 | `created_at` | `created_at_ts` |
| 更新时间 | `updated_at` | `updated_at_ts` |
| 删除时间 | `deleted_at` | `deleted_at_ts` |

**说明：**
- `_at` 结尾：可读时间字符串，格式 `YYYY-MM-DD HH:MM:SS`
- `_ts` 结尾：Unix 毫秒时间戳，int64 类型
- 两者**不要求同时返回**，但字段后缀必须与类型严格匹配
- 其他时间字段（如 `last_used_at` / `last_used_at_ts`）遵循相同规则

### 5.3 JSON 示例

```json
{
  "id": 1,
  "name": "CI Pipeline",
  "created_at": "2024-01-15 14:30:00",
  "updated_at_ts": 1705298400123
}
```

---

## 六、统一响应格式

所有 API 返回统一格式：

```json
{
  "code": 0,
  "msg": "请求成功",
  "data": { ... }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 业务错误码，`0` 表示成功，非 `0` 表示失败 |
| `msg` | string | 响应消息，成功时为 "请求成功"，失败时为错误描述 |
| `data` | any | 响应数据，成功时返回业务数据，失败时可选 |

### 响应头要求

所有响应必须包含 `X-Request-ID` 头，值与请求头中的 `X-Request-ID` 一致。详见 [Request ID 规范](./01-RequestID.md)。

---

## 七、HTTP 状态码规范

### 2.1 基本原则

**除 JWT 鉴权失败外，所有业务错误均返回 HTTP 200，通过 `code` 字段区分错误类型。**

### 2.2 状态码映射

| HTTP 状态码 | 触发条件 | 说明 |
|-------------|----------|------|
| **200** | 业务请求成功 | 包含 `code=0` 的成功响应 |
| **200** | 业务逻辑错误 | 包含非 `0` 的 `code` |
| **401** | JWT 鉴权失败 | Access Token 过期、无效或签名错误 |

### 2.3 设计理由

1. **前端统一处理**：响应拦截器只需处理 401（触发 Token 刷新），其他错误通过 `code` 判断
2. **简化错误流**：避免 HTTP 状态码与业务错误码双重判断
3. **保持一致性**：所有业务接口行为统一，降低调试成本

---

## 八、分页规范

所有列表接口统一使用分页查询，参数和约束如下：

### 8.1 请求参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `page` | int | 否 | 1 | 页码，从 1 开始 |
| `page_size` | int | 否 | 20 | 每页条数 |

### 8.2 约束规则

| 规则 | 说明 |
|------|------|
| `page` 最小值 | 1，小于 1 自动修正为 1 |
| `page_size` 范围 | 1-100，小于 1 使用默认值 20，超过 100 修正为 100 |
| 默认排序 | 按 `created_at_ts` 降序（最新优先） |

### 8.3 响应格式

```json
{
  "code": 0,
  "msg": "请求成功",
  "data": {
    "list": [...],
    "total": 150,
    "page": 1,
    "page_size": 20
  }
}
```

### 8.4 请求示例

```
GET /api/sources?page=2&page_size=50
Authorization: Bearer {accessToken}
```