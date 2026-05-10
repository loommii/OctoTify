// i18n.go 定义错误消息的多语言翻译映射
//
// 提供以下功能：
//   - errorMessages: 存储所有错误码的中英文翻译
//   - TranslateMsg: 根据错误码和语言返回对应翻译消息
//   - DefaultLang: 默认语言（中文）
package xerr

// DefaultLang 默认语言
const DefaultLang = "zh"

// errorMessages 错误消息的多语言翻译映射
// key 为错误码，value 为对应语言的消息
var errorMessages = map[int]map[string]string{
	// 通用错误 1000XX
	CodeBadRequest: {
		"zh": "请求参数错误",
		"en": "Invalid request parameters",
	},
	CodeUnauthorized: {
		"zh": "未登录或Token已过期",
		"en": "Not logged in or token expired",
	},
	CodeForbidden: {
		"zh": "权限不足",
		"en": "Insufficient permissions",
	},
	CodeNotFound: {
		"zh": "资源不存在",
		"en": "Resource not found",
	},
	CodeInternalServer: {
		"zh": "服务器内部错误",
		"en": "Internal server error",
	},
	CodeTooManyRequest: {
		"zh": "请求过于频繁",
		"en": "Too many requests",
	},
	CodeMethodNotAllowed: {
		"zh": "请求方法不允许",
		"en": "Method not allowed",
	},

	// 注册模块错误 1101XX
	CodeRegisterUsernameEmpty: {
		"zh": "用户名不能为空",
		"en": "Username cannot be empty",
	},
	CodeRegisterPasswordEmpty: {
		"zh": "密码不能为空",
		"en": "Password cannot be empty",
	},
	CodeRegisterUsernameInvalid: {
		"zh": "用户名格式不合法",
		"en": "Invalid username format",
	},
	CodeRegisterPasswordInvalid: {
		"zh": "密码格式不合法",
		"en": "Invalid password format",
	},
	CodeRegisterUsernameExists: {
		"zh": "用户名已存在",
		"en": "Username already exists",
	},
	CodeRegisterFailed: {
		"zh": "注册失败",
		"en": "Registration failed",
	},

	// 登录模块错误 1102XX
	CodeLoginInvalidCredentials: {
		"zh": "用户名或密码错误",
		"en": "Invalid username or password",
	},
	CodeLoginFailed: {
		"zh": "登录失败",
		"en": "Login failed",
	},

	// 刷新令牌模块错误 1103XX
	CodeRefreshTokenInvalid: {
		"zh": "刷新令牌无效",
		"en": "Invalid refresh token",
	},
	CodeRefreshTokenRevoked: {
		"zh": "刷新令牌已撤销",
		"en": "Refresh token revoked",
	},
	CodeRefreshTokenExpired: {
		"zh": "刷新令牌已过期",
		"en": "Refresh token expired",
	},
	CodeRefreshTokenFailed: {
		"zh": "刷新令牌失败",
		"en": "Refresh token failed",
	},
	CodeLogoutFailed: {
		"zh": "退出登录失败",
		"en": "Logout failed",
	},

	// 密码管理模块错误 1104XX
	CodeChangePasswordOldEmpty: {
		"zh": "旧密码不能为空",
		"en": "Old password cannot be empty",
	},
	CodeChangePasswordNewEmpty: {
		"zh": "新密码不能为空",
		"en": "New password cannot be empty",
	},
	CodeChangePasswordOldIncorrect: {
		"zh": "旧密码错误",
		"en": "Incorrect old password",
	},
	CodeChangePasswordFailed: {
		"zh": "密码修改失败",
		"en": "Password change failed",
	},

	// Source 模块错误 1105XX
	CodeSourceParamNameEmpty: {
		"zh": "来源名称不能为空",
		"en": "Source name cannot be empty",
	},
	CodeSourceInsertFailed: {
		"zh": "创建来源失败",
		"en": "Failed to create source",
	},
	CodeSourceTokenFailed: {
		"zh": "生成来源Token失败",
		"en": "Failed to generate source token",
	},
	CodeSourceNotFound: {
		"zh": "来源不存在",
		"en": "Source not found",
	},
	CodeSourceNoPermission: {
		"zh": "无权操作该来源",
		"en": "No permission to operate this source",
	},
	CodeSourceQueryFailed: {
		"zh": "查询来源失败",
		"en": "Failed to query source",
	},
	CodeSourceDeleteFailed: {
		"zh": "删除来源失败",
		"en": "Failed to delete source",
	},
	CodeSourceAlreadyDisabled: {
		"zh": "来源已停用",
		"en": "Source already disabled",
	},
	CodeSourceAlreadyEnabled: {
		"zh": "来源已启用",
		"en": "Source already enabled",
	},
	CodeSourceAlreadyDeleted: {
		"zh": "来源已删除",
		"en": "Source already deleted",
	},
	CodeSourceUpdateFailed: {
		"zh": "更新来源失败",
		"en": "Failed to update source",
	},
	CodeSourceDisabled: {
		"zh": "来源已停用，无法推送",
		"en": "Source disabled, cannot push",
	},

	// Channel 模块错误 1106XX
	CodeChannelParamNameEmpty: {
		"zh": "渠道名称不能为空",
		"en": "Channel name cannot be empty",
	},
	CodeChannelInvalidType: {
		"zh": "无效的渠道类型",
		"en": "Invalid channel type",
	},
	CodeChannelInsertFailed: {
		"zh": "创建渠道失败",
		"en": "Failed to create channel",
	},
	CodeChannelNotFound: {
		"zh": "渠道不存在",
		"en": "Channel not found",
	},
	CodeChannelNoPermission: {
		"zh": "无权操作该渠道",
		"en": "No permission to operate this channel",
	},
	CodeChannelQueryFailed: {
		"zh": "查询渠道失败",
		"en": "Failed to query channel",
	},
	CodeChannelDeleteFailed: {
		"zh": "删除渠道失败",
		"en": "Failed to delete channel",
	},
	CodeChannelAlreadyDisabled: {
		"zh": "渠道已停用",
		"en": "Channel already disabled",
	},
	CodeChannelAlreadyEnabled: {
		"zh": "渠道已启用",
		"en": "Channel already enabled",
	},
	CodeChannelAlreadyDeleted: {
		"zh": "渠道已删除",
		"en": "Channel already deleted",
	},

	// Message 模块错误 1107XX
	CodeMessageParamTitleEmpty: {
		"zh": "消息标题不能为空",
		"en": "Message title cannot be empty",
	},
	CodeMessageParamContentEmpty: {
		"zh": "消息内容不能为空",
		"en": "Message content cannot be empty",
	},
	CodeMessagePushFailed: {
		"zh": "消息推送失败",
		"en": "Message push failed",
	},
	CodeMessageNoChannels: {
		"zh": "来源未绑定任何渠道",
		"en": "Source not bound to any channel",
	},
	CodeMessageRecordFailed: {
		"zh": "记录消息状态失败",
		"en": "Failed to record message status",
	},
	CodeMessageAlreadyDeleted: {
		"zh": "消息已删除",
		"en": "Message already deleted",
	},

	// 用户管理模块错误 1108XX
	CodeUserProfileNotFound: {
		"zh": "用户不存在",
		"en": "User not found",
	},
	CodeUserProfileQueryFailed: {
		"zh": "查询用户信息失败",
		"en": "Failed to query user profile",
	},

	// 微信绑定模块错误 1109XX
	CodeBindExpired: {
		"zh": "绑定二维码已过期",
		"en": "Binding QR code expired",
	},
	CodeCredentialEncryptFailed: {
		"zh": "凭证加密失败",
		"en": "Credential encryption failed",
	},
	CodeCredentialDecryptFailed: {
		"zh": "凭证解密失败",
		"en": "Credential decryption failed",
	},

	// 第三方服务错误 2000XX
	CodeThirdPartyCallFailed: {
		"zh": "第三方接口调用失败",
		"en": "Third-party API call failed",
	},
	CodeQRCodeFetchFailed: {
		"zh": "获取绑定二维码失败",
		"en": "Failed to get binding QR code",
	},
	CodeBindStatusFailed: {
		"zh": "查询绑定状态失败",
		"en": "Failed to query binding status",
	},
}

// TranslateMsg 根据错误码和语言返回对应翻译消息
// 如果未找到翻译，返回空字符串
func TranslateMsg(code int, lang string) string {
	if lang == "" {
		lang = DefaultLang
	}
	if msgs, ok := errorMessages[code]; ok {
		if msg, ok := msgs[lang]; ok {
			return msg
		}
		// 回退到默认语言
		if msg, ok := msgs[DefaultLang]; ok {
			return msg
		}
	}
	return ""
}

// TranslateError 翻译 AppError 的消息，返回新对象
func TranslateError(err *AppError, lang string) *AppError {
	if err == nil {
		return nil
	}
	msg := TranslateMsg(err.Code, lang)
	if msg == "" {
		return err
	}
	newErr := *err
	newErr.Msg = msg
	return &newErr
}
