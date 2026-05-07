# 错误码规范

> 定义 OctoTify 项目中错误码的分配、命名和使用规范。

---

## 一、错误码定义

错误码按模块划分：

| 错误码范围 | 模块 | 说明 |
|------------|------|------|
| `0` | 通用 | 请求成功 |
| `100000` ~ `100099` | 通用错误 | 参数错误、JWT 错误等 |
| `110100` ~ `110199` | 注册模块 | 注册相关错误 |
| `110200` ~ `110299` | 登录模块 | 登录相关错误 |
| `110300` ~ `110399` | 刷新令牌模块 | Token 刷新相关错误 |
| `110400` ~ `110499` | 密码管理模块 | 密码相关错误 |
| `110500` ~ `110599` | Source 模块 | 消息来源管理相关错误 |
| `110600` ~ `110699` | Channel 模块 | 推送渠道管理相关错误 |
| `110700` ~ `110799` | Message 模块 | 消息推送相关错误 |
| `200000` ~ `200099` | 第三方服务 | 第三方接口调用错误 |

---

## 二、使用规范

1. **按模块分配范围**：每个模块分配 100 个错误码
2. **语义化命名**：错误码变量名应清晰表达错误含义
3. **不要复用错误码**：不同场景使用不同错误码，即使消息相同

---

## 三、错误码定义示例

```go
// 注册模块错误 1101XX
var (
    ErrRegisterUsernameEmpty     = NewCodeError(110100, "用户名不能为空")
    ErrRegisterPasswordEmpty    = NewCodeError(110101, "密码不能为空")
    ErrRegisterUsernameInvalid  = NewCodeError(110102, "用户名格式不合法")
    ErrRegisterPasswordInvalid = NewCodeError(110103, "密码格式不合法")
    ErrRegisterUsernameExists   = NewCodeError(110104, "用户名已存在")
    ErrRegisterFailed           = NewCodeError(110105, "注册失败")
)

// 登录模块错误 1102XX
var (
    ErrLoginInvalidCredentials = NewCodeError(110200, "用户名或密码错误")
    ErrLoginFailed             = NewCodeError(110201, "登录失败")
)

// 刷新令牌模块错误 1103XX
var (
    ErrRefreshTokenInvalid  = NewCodeError(110300, "刷新令牌无效")
    ErrRefreshTokenRevoked  = NewCodeError(110301, "刷新令牌已撤销")
    ErrRefreshTokenExpired  = NewCodeError(110302, "刷新令牌已过期")
    ErrRefreshTokenFailed   = NewCodeError(110303, "刷新令牌失败")
)

// 密码管理模块错误 1104XX
var (
    ErrChangePasswordOldEmpty     = NewCodeError(110400, "旧密码不能为空")
    ErrChangePasswordNewEmpty     = NewCodeError(110401, "新密码不能为空")
    ErrChangePasswordOldIncorrect = NewCodeError(110402, "旧密码错误")
    ErrChangePasswordFailed       = NewCodeError(110403, "密码修改失败")
)

// Source 管理错误 1105XX
var (
    ErrSourceParamNameEmpty    = NewCodeError(110500, "来源名称不能为空")
    ErrSourceInsertFailed      = NewCodeError(110501, "创建来源失败")
    ErrSourceTokenFailed       = NewCodeError(110502, "生成来源Token失败")
    ErrSourceNotFound          = NewCodeError(110503, "来源不存在")
    ErrSourceNoPermission      = NewCodeError(110504, "无权操作该来源")
    ErrSourceQueryFailed       = NewCodeError(110505, "查询来源失败")
    ErrSourceDeleteFailed      = NewCodeError(110506, "删除来源失败")
    ErrSourceAlreadyDisabled   = NewCodeError(110507, "来源已停用")
    ErrSourceAlreadyEnabled    = NewCodeError(110508, "来源已启用")
    ErrSourceAlreadyDeleted    = NewCodeError(110509, "来源已删除")
    ErrSourceUpdateFailed      = NewCodeError(110510, "更新来源失败")
    ErrSourceDisabled          = NewCodeError(110511, "来源已停用，无法推送")
)

// Channel 管理错误 1106XX
var (
    ErrChannelParamNameEmpty   = NewCodeError(110600, "渠道名称不能为空")
    ErrChannelInvalidType      = NewCodeError(110601, "无效的渠道类型")
    ErrChannelInsertFailed     = NewCodeError(110602, "创建渠道失败")
    ErrChannelNotFound         = NewCodeError(110603, "渠道不存在")
    ErrChannelNoPermission     = NewCodeError(110604, "无权操作该渠道")
    ErrChannelQueryFailed      = NewCodeError(110605, "查询渠道失败")
    ErrChannelDeleteFailed     = NewCodeError(110606, "删除渠道失败")
    ErrChannelAlreadyDisabled  = NewCodeError(110607, "渠道已停用")
    ErrChannelAlreadyEnabled   = NewCodeError(110608, "渠道已启用")
    ErrChannelAlreadyDeleted   = NewCodeError(110609, "渠道已删除")
)

// Message 推送错误 1107XX
var (
    ErrMessageParamTitleEmpty   = NewCodeError(110700, "消息标题不能为空")
    ErrMessageParamContentEmpty = NewCodeError(110701, "消息内容不能为空")
    ErrMessagePushFailed        = NewCodeError(110702, "消息推送失败")
    ErrMessageNoChannels        = NewCodeError(110703, "来源未绑定任何渠道")
    ErrMessageRecordFailed      = NewCodeError(110704, "记录消息状态失败")
    ErrMessageAlreadyDeleted    = NewCodeError(110705, "消息已删除")
)
```

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
| 2026-05-02 | 1.0 | 初始版本，定义错误码分配和使用规范 |
