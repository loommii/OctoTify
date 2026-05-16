# 错误码规范

> 定义 OctoTify 项目中错误码的分配、命名和使用规范。

---

## 一、错误码定义

错误码按模块划分：

| 错误码范围 | 模块 | 说明 |
|------------|------|------|
| `0` | 通用 | 请求成功 |
| `100000` ~ `100099` | 通用错误 | 参数错误等 |
| `110100` ~ `110199` | 注册模块 | 注册相关错误 |
| `110200` ~ `110299` | 登录模块 | 登录相关错误 |
| `110300` ~ `110399` | 刷新令牌模块 | Token 刷新相关错误 |
| `110400` ~ `110499` | 密码管理模块 | 密码相关错误 |
| `110500` ~ `110599` | Source 模块 | 消息来源管理相关错误 |
| `110600` ~ `110699` | Channel 模块 | 推送渠道管理相关错误 |
| `110700` ~ `110799` | Message 模块 | 消息推送相关错误 |
| `110800` ~ `110899` | 用户管理模块 | 用户管理相关错误 |
| `110900` ~ `110999` | 微信绑定模块 | 微信绑定本地业务逻辑错误（二维码过期、凭证加解密） |
| `111000` ~ `111099` | JWT 鉴权模块 | JWT 令牌解析、验证相关错误 |
| `200000` ~ `200099` | 第三方服务 | 第三方接口调用错误（含微信 iLink API） |

---

## 二、使用规范

1. **按模块分配范围**：每个模块分配 100 个错误码
2. **语义化命名**：错误码变量名应清晰表达错误含义
3. **不要复用错误码**：不同场景使用不同错误码，即使消息相同

---

## 三、错误码清单

### 3.1 通用错误 1000XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 100000 | ErrBadRequest | 请求参数错误 | Invalid request parameters |
| 100001 | ErrUnauthorized | 未登录或Token已过期 | Not logged in or token expired |
| 100002 | ErrForbidden | 权限不足 | Insufficient permissions |
| 100003 | ErrNotFound | 资源不存在 | Resource not found |
| 100004 | ErrInternalServer | 服务器内部错误 | Internal server error |
| 100005 | ErrTooManyRequest | 请求过于频繁 | Too many requests |
| 100006 | ErrMethodNotAllowed | 请求方法不允许 | Method not allowed |

### 3.2 注册模块 1101XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 110100 | ErrRegisterUsernameEmpty | 用户名不能为空 | Username cannot be empty |
| 110101 | ErrRegisterPasswordEmpty | 密码不能为空 | Password cannot be empty |
| 110102 | ErrRegisterUsernameInvalid | 用户名格式不合法 | Invalid username format |
| 110103 | ErrRegisterPasswordInvalid | 密码格式不合法 | Invalid password format |
| 110104 | ErrRegisterUsernameExists | 用户名已存在 | Username already exists |
| 110105 | ErrRegisterFailed | 注册失败 | Registration failed |

### 3.3 登录模块 1102XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 110200 | ErrLoginInvalidCredentials | 用户名或密码错误 | Invalid username or password |
| 110201 | ErrLoginFailed | 登录失败 | Login failed |

### 3.4 刷新令牌模块 1103XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 110300 | ErrRefreshTokenInvalid | 刷新令牌无效 | Invalid refresh token |
| 110301 | ErrRefreshTokenRevoked | 刷新令牌已撤销 | Refresh token revoked |
| 110302 | ErrRefreshTokenExpired | 刷新令牌已过期 | Refresh token expired |
| 110303 | ErrRefreshTokenFailed | 刷新令牌失败 | Refresh token failed |
| 110304 | ErrLogoutFailed | 退出登录失败 | Logout failed |

### 3.5 密码管理模块 1104XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 110400 | ErrChangePasswordOldEmpty | 旧密码不能为空 | Old password cannot be empty |
| 110401 | ErrChangePasswordNewEmpty | 新密码不能为空 | New password cannot be empty |
| 110402 | ErrChangePasswordOldIncorrect | 旧密码错误 | Incorrect old password |
| 110403 | ErrChangePasswordFailed | 密码修改失败 | Password change failed |

### 3.6 Source 模块 1105XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 110500 | ErrSourceParamNameEmpty | 来源名称不能为空 | Source name cannot be empty |
| 110501 | ErrSourceInsertFailed | 创建来源失败 | Failed to create source |
| 110502 | ErrSourceTokenFailed | 生成来源Token失败 | Failed to generate source token |
| 110503 | ErrSourceNotFound | 来源不存在 | Source not found |
| 110504 | ErrSourceNoPermission | 无权操作该来源 | No permission to operate this source |
| 110505 | ErrSourceQueryFailed | 查询来源失败 | Failed to query source |
| 110506 | ErrSourceDeleteFailed | 删除来源失败 | Failed to delete source |
| 110507 | ErrSourceAlreadyDisabled | 来源已停用 | Source already disabled |
| 110508 | ErrSourceAlreadyEnabled | 来源已启用 | Source already enabled |
| 110509 | ErrSourceAlreadyDeleted | 来源已删除 | Source already deleted |
| 110510 | ErrSourceUpdateFailed | 更新来源失败 | Failed to update source |
| 110511 | ErrSourceDisabled | 来源已停用，无法推送 | Source disabled, cannot push |

### 3.7 Channel 模块 1106XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 110600 | ErrChannelParamNameEmpty | 渠道名称不能为空 | Channel name cannot be empty |
| 110601 | ErrChannelInvalidType | 无效的渠道类型 | Invalid channel type |
| 110602 | ErrChannelInsertFailed | 创建渠道失败 | Failed to create channel |
| 110603 | ErrChannelNotFound | 渠道不存在 | Channel not found |
| 110604 | ErrChannelNoPermission | 无权操作该渠道 | No permission to operate this channel |
| 110605 | ErrChannelQueryFailed | 查询渠道失败 | Failed to query channel |
| 110606 | ErrChannelDeleteFailed | 删除渠道失败 | Failed to delete channel |
| 110607 | ErrChannelAlreadyDisabled | 渠道已停用 | Channel already disabled |
| 110608 | ErrChannelAlreadyEnabled | 渠道已启用 | Channel already enabled |
| 110609 | ErrChannelAlreadyDeleted | 渠道已删除 | Channel already deleted |

### 3.8 Message 模块 1107XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 110700 | ErrMessageParamTitleEmpty | 消息标题不能为空 | Message title cannot be empty |
| 110701 | ErrMessageParamContentEmpty | 消息内容不能为空 | Message content cannot be empty |
| 110702 | ErrMessagePushFailed | 消息推送失败 | Message push failed |
| 110703 | ErrMessageNoChannels | 来源未绑定任何渠道 | Source not bound to any channel |
| 110704 | ErrMessageRecordFailed | 记录消息状态失败 | Failed to record message status |
| 110705 | ErrMessageAlreadyDeleted | 消息已删除 | Message already deleted |

### 3.9 用户管理模块 1108XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 110800 | ErrUserProfileNotFound | 用户不存在 | User not found |
| 110801 | ErrUserProfileQueryFailed | 查询用户信息失败 | Failed to query user profile |

### 3.10 微信绑定模块 1109XX（本地业务逻辑）

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 110901 | ErrBindExpired | 绑定二维码已过期 | Binding QR code expired |
| 110902 | ErrCredentialEncryptFailed | 凭证加密失败 | Credential encryption failed |
| 110903 | ErrCredentialDecryptFailed | 凭证解密失败 | Credential decryption failed |

### 3.11 JWT 鉴权模块 1110XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 111000 | ErrJWTMissingToken | 未提供认证令牌 | Authentication token not provided |
| 111001 | ErrJWTInvalidFormat | 认证令牌格式错误 | Invalid authentication token format |
| 111002 | ErrJWTInvalidToken | 认证令牌无效或已过期 | Invalid or expired authentication token |
| 111003 | ErrJWTWrongTokenType | 无效的令牌类型 | Invalid token type |

### 3.12 第三方服务 2000XX

| 错误码 | 错误常量 | 错误消息 (中文) | Error Message (English) |
|--------|---------|----------------|------------------------|
| 200000 | ErrThirdPartyCallFailed | 第三方接口调用失败 | Third-party API call failed |
| 200001 | ErrQRCodeFetchFailed | 获取绑定二维码失败 | Failed to get binding QR code |
| 200002 | ErrBindStatusFailed | 查询绑定状态失败 | Failed to query binding status |

---

## 四、错误处理规范

必须使用预定义错误码，不得直接返回 error 字符串：

```go
// 正确示例：使用预定义错误码
if req.Name == "" {
    return nil, ErrSourceParamNameEmpty
}

// 错误示例：不要直接返回 error 字符串
if req.Name == "" {
    return nil, errors.New("来源名称不能为空")  // 不要这样做
}
```

---

## 五、多语言

错误消息的多语言处理遵循 [API 多语言规范](./04-i18n.md)。

核心要点：
- 错误码 `code` 始终不变，作为错误的唯一标识
- 错误消息 `msg` 根据 `Accept-Language` 返回对应语言
- 日志中记录错误码和默认语言消息

---

## 版本历史

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-16 | 1.2.0 | 新增 JWT 鉴权模块 (1110XX)、修正 4 处英文翻译 |
| 2026-05-10 | 1.1.0 | 新增模块、拆分第三方错误码、双语错误消息、表格化清单 |
| 2026-05-02 | 1.0.0 | 初始版本 |
