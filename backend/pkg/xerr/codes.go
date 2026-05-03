package xerr

// 通用错误 1000XX
var (
	ErrBadRequest       = New(100000, "请求参数错误")
	ErrUnauthorized     = New(100001, "未登录或Token已过期")
	ErrForbidden        = New(100002, "权限不足")
	ErrNotFound         = New(100003, "资源不存在")
	ErrInternalServer   = New(100004, "服务器内部错误")
	ErrTooManyRequest   = New(100005, "请求过于频繁")
	ErrMethodNotAllowed = New(100006, "请求方法不允许")
)

// 注册模块错误 1101XX
var (
	ErrRegisterUsernameEmpty   = New(110100, "用户名不能为空")
	ErrRegisterPasswordEmpty   = New(110101, "密码不能为空")
	ErrRegisterUsernameInvalid = New(110102, "用户名格式不合法")
	ErrRegisterPasswordInvalid = New(110103, "密码格式不合法")
	ErrRegisterUsernameExists  = New(110104, "用户名已存在")
	ErrRegisterFailed          = New(110105, "注册失败")
)

// 登录模块错误 1102XX
var (
	ErrLoginInvalidCredentials = New(110200, "用户名或密码错误")
	ErrLoginFailed             = New(110201, "登录失败")
)

// 刷新令牌模块错误 1103XX
var (
	ErrRefreshTokenInvalid = New(110300, "刷新令牌无效")
	ErrRefreshTokenRevoked = New(110301, "刷新令牌已撤销")
	ErrRefreshTokenExpired = New(110302, "刷新令牌已过期")
	ErrRefreshTokenFailed  = New(110303, "刷新令牌失败")
)

// 密码管理模块错误 1104XX
var (
	ErrChangePasswordOldEmpty     = New(110400, "旧密码不能为空")
	ErrChangePasswordNewEmpty     = New(110401, "新密码不能为空")
	ErrChangePasswordOldIncorrect = New(110402, "旧密码错误")
	ErrChangePasswordFailed       = New(110403, "密码修改失败")
)

// Source 模块错误 1105XX
var (
	ErrSourceParamNameEmpty  = New(110500, "来源名称不能为空")
	ErrSourceInsertFailed    = New(110501, "创建来源失败")
	ErrSourceTokenFailed     = New(110502, "生成来源Token失败")
	ErrSourceNotFound        = New(110503, "来源不存在")
	ErrSourceNoPermission    = New(110504, "无权操作该来源")
	ErrSourceQueryFailed     = New(110505, "查询来源失败")
	ErrSourceDeleteFailed    = New(110506, "删除来源失败")
	ErrSourceAlreadyDisabled = New(110507, "来源已停用")
	ErrSourceAlreadyEnabled  = New(110508, "来源已启用")
	ErrSourceAlreadyDeleted  = New(110509, "来源已删除")
)

// Channel 模块错误 1106XX
var (
	ErrChannelParamNameEmpty  = New(110600, "渠道名称不能为空")
	ErrChannelInvalidType     = New(110601, "无效的渠道类型")
	ErrChannelInsertFailed    = New(110602, "创建渠道失败")
	ErrChannelNotFound        = New(110603, "渠道不存在")
	ErrChannelNoPermission    = New(110604, "无权操作该渠道")
	ErrChannelQueryFailed     = New(110605, "查询渠道失败")
	ErrChannelDeleteFailed    = New(110606, "删除渠道失败")
	ErrChannelAlreadyDisabled = New(110607, "渠道已停用")
	ErrChannelAlreadyEnabled  = New(110608, "渠道已启用")
	ErrChannelAlreadyDeleted  = New(110609, "渠道已删除")
)

// Message 模块错误 1107XX
var (
	ErrMessageParamTitleEmpty   = New(110700, "消息标题不能为空")
	ErrMessageParamContentEmpty = New(110701, "消息内容不能为空")
	ErrMessagePushFailed        = New(110702, "消息推送失败")
	ErrMessageNoChannels        = New(110703, "来源未绑定任何渠道")
	ErrMessageRecordFailed      = New(110704, "记录消息状态失败")
	ErrMessageAlreadyDeleted    = New(110705, "消息已删除")
)

// 用户管理模块错误 1108XX
var (
	ErrUserProfileNotFound    = New(110800, "用户不存在")
	ErrUserProfileQueryFailed = New(110801, "查询用户信息失败")
)

// 第三方服务错误 2000XX
var (
	ErrThirdPartyCallFailed = New(200000, "第三方接口调用失败")
)
