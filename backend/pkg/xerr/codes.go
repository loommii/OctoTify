// Package xerr 提供应用级业务错误处理
//
// 本包包含以下功能：
//   - AppError 结构体：统一的业务错误类型，支持错误码、用户提示、内部错误信息
//   - 错误链支持：实现 Go 1.13+ 的 Unwrap/Is 接口，兼容 errors.Is/As
//   - HTTP 状态码映射：自动将业务错误码映射为 HTTP 状态码
//   - 多语言支持：根据错误码和语言返回对应翻译消息
//   - 预定义错误：各模块的常见错误常量
//
// 文件说明：
//   - errors.go:      AppError 核心定义和方法（New、WithInternal、Wrap 等）
//   - codes.go:       预定义错误常量（ErrBadRequest、ErrUnauthorized 等）
//   - http_status.go: HTTP 状态码映射逻辑（HTTPStatusFromCode）
//   - i18n.go:        多语言翻译映射（中英文错误消息）
//
// 使用示例：
//
//	// 创建错误
//	err := xerr.New(100001)
//
//	// 获取翻译消息
//	msg := xerr.TranslateMsg(100001, "en")  // "Not logged in or token expired"
//
//	// 附加内部错误信息
//	err := xerr.ErrNotFound.WithInternal(dbErr)
//
//	// 包装错误
//	err := xerr.ErrSourceNotFound.Wrap("来源ID: %d", sourceID)
//
//	// 错误判断
//	if errors.Is(err, xerr.ErrNotFound) { ... }
package xerr

// 错误码常量定义
const (
	// 通用错误 1000XX
	CodeBadRequest       = 100000
	CodeUnauthorized     = 100001
	CodeForbidden        = 100002
	CodeNotFound         = 100003
	CodeInternalServer   = 100004
	CodeTooManyRequest   = 100005
	CodeMethodNotAllowed = 100006

	// 注册模块错误 1101XX
	CodeRegisterUsernameEmpty   = 110100
	CodeRegisterPasswordEmpty   = 110101
	CodeRegisterUsernameInvalid = 110102
	CodeRegisterPasswordInvalid = 110103
	CodeRegisterUsernameExists  = 110104
	CodeRegisterFailed          = 110105

	// 登录模块错误 1102XX
	CodeLoginInvalidCredentials = 110200
	CodeLoginFailed             = 110201

	// 刷新令牌模块错误 1103XX
	CodeRefreshTokenInvalid = 110300
	CodeRefreshTokenRevoked = 110301
	CodeRefreshTokenExpired = 110302
	CodeRefreshTokenFailed  = 110303
	CodeLogoutFailed        = 110304

	// 密码管理模块错误 1104XX
	CodeChangePasswordOldEmpty     = 110400
	CodeChangePasswordNewEmpty     = 110401
	CodeChangePasswordOldIncorrect = 110402
	CodeChangePasswordFailed       = 110403

	// Source 模块错误 1105XX
	CodeSourceParamNameEmpty  = 110500
	CodeSourceInsertFailed    = 110501
	CodeSourceTokenFailed     = 110502
	CodeSourceNotFound        = 110503
	CodeSourceNoPermission    = 110504
	CodeSourceQueryFailed     = 110505
	CodeSourceDeleteFailed    = 110506
	CodeSourceAlreadyDisabled = 110507
	CodeSourceAlreadyEnabled  = 110508
	CodeSourceAlreadyDeleted  = 110509
	CodeSourceUpdateFailed    = 110510
	CodeSourceDisabled        = 110511

	// Channel 模块错误 1106XX
	CodeChannelParamNameEmpty  = 110600
	CodeChannelInvalidType     = 110601
	CodeChannelInsertFailed    = 110602
	CodeChannelNotFound        = 110603
	CodeChannelNoPermission    = 110604
	CodeChannelQueryFailed     = 110605
	CodeChannelDeleteFailed    = 110606
	CodeChannelAlreadyDisabled = 110607
	CodeChannelAlreadyEnabled  = 110608
	CodeChannelAlreadyDeleted  = 110609

	// Message 模块错误 1107XX
	CodeMessageParamTitleEmpty   = 110700
	CodeMessageParamContentEmpty = 110701
	CodeMessagePushFailed        = 110702
	CodeMessageNoChannels        = 110703
	CodeMessageRecordFailed      = 110704
	CodeMessageAlreadyDeleted    = 110705

	// 用户管理模块错误 1108XX
	CodeUserProfileNotFound    = 110800
	CodeUserProfileQueryFailed = 110801

	// 微信绑定模块错误 1109XX
	CodeBindExpired             = 110901
	CodeCredentialEncryptFailed = 110902
	CodeCredentialDecryptFailed = 110903

	// 第三方服务错误 2000XX
	CodeThirdPartyCallFailed = 200000
	CodeQRCodeFetchFailed    = 200001
	CodeBindStatusFailed     = 200002
)

// 通用错误 1000XX
var (
	ErrBadRequest       = New(CodeBadRequest)
	ErrUnauthorized     = New(CodeUnauthorized)
	ErrForbidden        = New(CodeForbidden)
	ErrNotFound         = New(CodeNotFound)
	ErrInternalServer   = New(CodeInternalServer)
	ErrTooManyRequest   = New(CodeTooManyRequest)
	ErrMethodNotAllowed = New(CodeMethodNotAllowed)
)

// 注册模块错误 1101XX
var (
	ErrRegisterUsernameEmpty   = New(CodeRegisterUsernameEmpty)
	ErrRegisterPasswordEmpty   = New(CodeRegisterPasswordEmpty)
	ErrRegisterUsernameInvalid = New(CodeRegisterUsernameInvalid)
	ErrRegisterPasswordInvalid = New(CodeRegisterPasswordInvalid)
	ErrRegisterUsernameExists  = New(CodeRegisterUsernameExists)
	ErrRegisterFailed          = New(CodeRegisterFailed)
)

// 登录模块错误 1102XX
var (
	ErrLoginInvalidCredentials = New(CodeLoginInvalidCredentials)
	ErrLoginFailed             = New(CodeLoginFailed)
)

// 刷新令牌模块错误 1103XX
var (
	ErrRefreshTokenInvalid = New(CodeRefreshTokenInvalid)
	ErrRefreshTokenRevoked = New(CodeRefreshTokenRevoked)
	ErrRefreshTokenExpired = New(CodeRefreshTokenExpired)
	ErrRefreshTokenFailed  = New(CodeRefreshTokenFailed)
	ErrLogoutFailed        = New(CodeLogoutFailed)
)

// 密码管理模块错误 1104XX
var (
	ErrChangePasswordOldEmpty     = New(CodeChangePasswordOldEmpty)
	ErrChangePasswordNewEmpty     = New(CodeChangePasswordNewEmpty)
	ErrChangePasswordOldIncorrect = New(CodeChangePasswordOldIncorrect)
	ErrChangePasswordFailed       = New(CodeChangePasswordFailed)
)

// Source 模块错误 1105XX
var (
	ErrSourceParamNameEmpty  = New(CodeSourceParamNameEmpty)
	ErrSourceInsertFailed    = New(CodeSourceInsertFailed)
	ErrSourceTokenFailed     = New(CodeSourceTokenFailed)
	ErrSourceNotFound        = New(CodeSourceNotFound)
	ErrSourceNoPermission    = New(CodeSourceNoPermission)
	ErrSourceQueryFailed     = New(CodeSourceQueryFailed)
	ErrSourceDeleteFailed    = New(CodeSourceDeleteFailed)
	ErrSourceAlreadyDisabled = New(CodeSourceAlreadyDisabled)
	ErrSourceAlreadyEnabled  = New(CodeSourceAlreadyEnabled)
	ErrSourceAlreadyDeleted  = New(CodeSourceAlreadyDeleted)
	ErrSourceUpdateFailed    = New(CodeSourceUpdateFailed)
	ErrSourceDisabled        = New(CodeSourceDisabled)
)

// Channel 模块错误 1106XX
var (
	ErrChannelParamNameEmpty  = New(CodeChannelParamNameEmpty)
	ErrChannelInvalidType     = New(CodeChannelInvalidType)
	ErrChannelInsertFailed    = New(CodeChannelInsertFailed)
	ErrChannelNotFound        = New(CodeChannelNotFound)
	ErrChannelNoPermission    = New(CodeChannelNoPermission)
	ErrChannelQueryFailed     = New(CodeChannelQueryFailed)
	ErrChannelDeleteFailed    = New(CodeChannelDeleteFailed)
	ErrChannelAlreadyDisabled = New(CodeChannelAlreadyDisabled)
	ErrChannelAlreadyEnabled  = New(CodeChannelAlreadyEnabled)
	ErrChannelAlreadyDeleted  = New(CodeChannelAlreadyDeleted)
)

// Message 模块错误 1107XX
var (
	ErrMessageParamTitleEmpty   = New(CodeMessageParamTitleEmpty)
	ErrMessageParamContentEmpty = New(CodeMessageParamContentEmpty)
	ErrMessagePushFailed        = New(CodeMessagePushFailed)
	ErrMessageNoChannels        = New(CodeMessageNoChannels)
	ErrMessageRecordFailed      = New(CodeMessageRecordFailed)
	ErrMessageAlreadyDeleted    = New(CodeMessageAlreadyDeleted)
)

// 用户管理模块错误 1108XX
var (
	ErrUserProfileNotFound    = New(CodeUserProfileNotFound)
	ErrUserProfileQueryFailed = New(CodeUserProfileQueryFailed)
)

// 微信绑定模块错误 1109XX（本地业务逻辑）
var (
	ErrBindExpired             = New(CodeBindExpired)
	ErrCredentialEncryptFailed = New(CodeCredentialEncryptFailed)
	ErrCredentialDecryptFailed = New(CodeCredentialDecryptFailed)
)

// 第三方服务错误 2000XX
var (
	ErrThirdPartyCallFailed = New(CodeThirdPartyCallFailed)

	// 微信 iLink API 调用错误
	ErrQRCodeFetchFailed = New(CodeQRCodeFetchFailed)
	ErrBindStatusFailed  = New(CodeBindStatusFailed)
)
