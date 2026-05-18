// i18n.go 定义错误消息的多语言翻译映射
//
// 提供以下功能：
//   - errorMessages: 存储所有错误码的中英文翻译
//   - TranslateMsg: 根据错误码和语言返回对应翻译消息
//   - DefaultLang: 默认语言（中文）
package xerr

// DefaultLang 默认语言（BCP 47: 简体中文-中国）
const DefaultLang = "zh-CN"

// SupportedLangs 支持的语言列表，按回退优先级排序
var SupportedLangs = []string{"zh-CN", "en-US"}

// errorMessages 错误消息的多语言翻译映射
// key 为错误码，value 为对应语言的消息
var errorMessages = map[int]map[string]string{
	// 通用错误 1000XX
	CodeBadRequest: {
		"zh-CN": "请求参数错误",
		"en-US": "Invalid request parameters",
	},
	CodeUnauthorized: {
		"zh-CN": "未登录或Token已过期",
		"en-US": "Not logged in or token expired",
	},
	CodeForbidden: {
		"zh-CN": "权限不足",
		"en-US": "Insufficient permissions",
	},
	CodeNotFound: {
		"zh-CN": "资源不存在",
		"en-US": "Resource not found",
	},
	CodeInternalServer: {
		"zh-CN": "服务器内部错误",
		"en-US": "Internal server error",
	},
	CodeTooManyRequest: {
		"zh-CN": "请求过于频繁",
		"en-US": "Too many requests",
	},
	CodeMethodNotAllowed: {
		"zh-CN": "请求方法不允许",
		"en-US": "Method not allowed",
	},

	// 注册模块错误 1101XX
	CodeRegisterUsernameEmpty: {
		"zh-CN": "用户名不能为空",
		"en-US": "Username cannot be empty",
	},
	CodeRegisterPasswordEmpty: {
		"zh-CN": "密码不能为空",
		"en-US": "Password cannot be empty",
	},
	CodeRegisterUsernameInvalid: {
		"zh-CN": "用户名格式不合法",
		"en-US": "Invalid username format",
	},
	CodeRegisterPasswordInvalid: {
		"zh-CN": "密码格式不合法",
		"en-US": "Invalid password format",
	},
	CodeRegisterUsernameExists: {
		"zh-CN": "用户名已存在",
		"en-US": "Username already exists",
	},
	CodeRegisterFailed: {
		"zh-CN": "注册失败",
		"en-US": "Registration failed",
	},

	// 登录模块错误 1102XX
	CodeLoginInvalidCredentials: {
		"zh-CN": "用户名或密码错误",
		"en-US": "Invalid username or password",
	},
	CodeLoginFailed: {
		"zh-CN": "登录失败",
		"en-US": "Login failed",
	},

	// 刷新令牌模块错误 1103XX
	CodeRefreshTokenInvalid: {
		"zh-CN": "刷新令牌无效",
		"en-US": "Invalid refresh token",
	},
	CodeRefreshTokenRevoked: {
		"zh-CN": "刷新令牌已撤销",
		"en-US": "Refresh token revoked",
	},
	CodeRefreshTokenExpired: {
		"zh-CN": "刷新令牌已过期",
		"en-US": "Refresh token expired",
	},
	CodeRefreshTokenFailed: {
		"zh-CN": "刷新令牌失败",
		"en-US": "Refresh token failed",
	},
	CodeLogoutFailed: {
		"zh-CN": "退出登录失败",
		"en-US": "Logout failed",
	},

	// 密码管理模块错误 1104XX
	CodeChangePasswordOldEmpty: {
		"zh-CN": "旧密码不能为空",
		"en-US": "Old password cannot be empty",
	},
	CodeChangePasswordNewEmpty: {
		"zh-CN": "新密码不能为空",
		"en-US": "New password cannot be empty",
	},
	CodeChangePasswordOldIncorrect: {
		"zh-CN": "旧密码错误",
		"en-US": "Incorrect old password",
	},
	CodeChangePasswordFailed: {
		"zh-CN": "密码修改失败",
		"en-US": "Password change failed",
	},

	// Source 模块错误 1105XX
	CodeSourceParamNameEmpty: {
		"zh-CN": "来源名称不能为空",
		"en-US": "Source name cannot be empty",
	},
	CodeSourceInsertFailed: {
		"zh-CN": "创建来源失败",
		"en-US": "Failed to create source",
	},
	CodeSourceTokenFailed: {
		"zh-CN": "生成来源Token失败",
		"en-US": "Failed to generate source token",
	},
	CodeSourceNotFound: {
		"zh-CN": "来源不存在",
		"en-US": "Source not found",
	},
	CodeSourceNoPermission: {
		"zh-CN": "无权操作该来源",
		"en-US": "No permission to operate this source",
	},
	CodeSourceQueryFailed: {
		"zh-CN": "查询来源失败",
		"en-US": "Failed to query source",
	},
	CodeSourceDeleteFailed: {
		"zh-CN": "删除来源失败",
		"en-US": "Failed to delete source",
	},
	CodeSourceAlreadyDisabled: {
		"zh-CN": "来源已停用",
		"en-US": "Source already disabled",
	},
	CodeSourceAlreadyEnabled: {
		"zh-CN": "来源已启用",
		"en-US": "Source already enabled",
	},
	CodeSourceAlreadyDeleted: {
		"zh-CN": "来源已删除",
		"en-US": "Source already deleted",
	},
	CodeSourceUpdateFailed: {
		"zh-CN": "更新来源失败",
		"en-US": "Failed to update source",
	},
	CodeSourceDisabled: {
		"zh-CN": "来源已停用，无法推送",
		"en-US": "Source disabled, cannot push",
	},

	// Channel 模块错误 1106XX
	CodeChannelParamNameEmpty: {
		"zh-CN": "渠道名称不能为空",
		"en-US": "Channel name cannot be empty",
	},
	CodeChannelInvalidType: {
		"zh-CN": "无效的渠道类型",
		"en-US": "Invalid channel type",
	},
	CodeChannelInsertFailed: {
		"zh-CN": "创建渠道失败",
		"en-US": "Failed to create channel",
	},
	CodeChannelNotFound: {
		"zh-CN": "渠道不存在",
		"en-US": "Channel not found",
	},
	CodeChannelNoPermission: {
		"zh-CN": "无权操作该渠道",
		"en-US": "No permission to operate this channel",
	},
	CodeChannelQueryFailed: {
		"zh-CN": "查询渠道失败",
		"en-US": "Failed to query channel",
	},
	CodeChannelDeleteFailed: {
		"zh-CN": "删除渠道失败",
		"en-US": "Failed to delete channel",
	},
	CodeChannelAlreadyDisabled: {
		"zh-CN": "渠道已停用",
		"en-US": "Channel already disabled",
	},
	CodeChannelAlreadyEnabled: {
		"zh-CN": "渠道已启用",
		"en-US": "Channel already enabled",
	},
	CodeChannelAlreadyDeleted: {
		"zh-CN": "渠道已删除",
		"en-US": "Channel already deleted",
	},

	// Message 模块错误 1107XX
	CodeMessageParamTitleEmpty: {
		"zh-CN": "消息标题不能为空",
		"en-US": "Message title cannot be empty",
	},
	CodeMessageParamContentEmpty: {
		"zh-CN": "消息内容不能为空",
		"en-US": "Message content cannot be empty",
	},
	CodeMessagePushFailed: {
		"zh-CN": "消息推送失败",
		"en-US": "Message push failed",
	},
	CodeMessageNoChannels: {
		"zh-CN": "来源未绑定任何渠道",
		"en-US": "Source not bound to any channel",
	},
	CodeMessageRecordFailed: {
		"zh-CN": "记录消息状态失败",
		"en-US": "Failed to record message status",
	},
	CodeMessageAlreadyDeleted: {
		"zh-CN": "消息已删除",
		"en-US": "Message already deleted",
	},
	CodeMessageChannelsDisabled: {
		"zh-CN": "来源绑定的渠道均已停用",
		"en-US": "All bound channels are disabled",
	},
	CodeMessageChannelsDeleted: {
		"zh-CN": "来源绑定的渠道均已删除",
		"en-US": "All bound channels are deleted",
	},

	// 用户管理模块错误 1108XX
	CodeUserProfileNotFound: {
		"zh-CN": "用户不存在",
		"en-US": "User not found",
	},
	CodeUserProfileQueryFailed: {
		"zh-CN": "查询用户信息失败",
		"en-US": "Failed to query user profile",
	},

	// 微信绑定模块错误 1109XX
	CodeBindExpired: {
		"zh-CN": "绑定二维码已过期",
		"en-US": "Binding QR code expired",
	},
	CodeCredentialEncryptFailed: {
		"zh-CN": "凭证加密失败",
		"en-US": "Credential encryption failed",
	},
	CodeCredentialDecryptFailed: {
		"zh-CN": "凭证解密失败",
		"en-US": "Credential decryption failed",
	},

	// JWT 鉴权模块错误 1110XX
	CodeJWTMissingToken: {
		"zh-CN": "未提供认证令牌",
		"en-US": "Authentication token not provided",
	},
	CodeJWTInvalidFormat: {
		"zh-CN": "认证令牌格式错误",
		"en-US": "Invalid authentication token format",
	},
	CodeJWTInvalidToken: {
		"zh-CN": "认证令牌无效或已过期",
		"en-US": "Invalid or expired authentication token",
	},
	CodeJWTWrongTokenType: {
		"zh-CN": "无效的令牌类型",
		"en-US": "Invalid token type",
	},

	// 第三方服务错误 2000XX
	CodeThirdPartyCallFailed: {
		"zh-CN": "第三方接口调用失败",
		"en-US": "Third-party API call failed",
	},
	CodeQRCodeFetchFailed: {
		"zh-CN": "获取绑定二维码失败",
		"en-US": "Failed to get binding QR code",
	},
	CodeBindStatusFailed: {
		"zh-CN": "查询绑定状态失败",
		"en-US": "Failed to query binding status",
	},
}

// resolveLang 根据 BCP 47 语言标签解析实际使用的语言
// 支持精确匹配（zh-CN）和前缀回退（zh -> zh-CN）
func resolveLang(lang string) string {
	if lang == "" {
		return DefaultLang
	}

	// 精确匹配
	for _, supported := range SupportedLangs {
		if lang == supported {
			return lang
		}
	}

	// 前缀回退匹配：zh-CN/zh-TW -> zh-CN, en-US/en-GB -> en-US
	langPrefix := lang
	if idx := len(lang); idx > 2 {
		// 查找第一个 '-' 前的前缀
		for i, c := range lang {
			if c == '-' {
				langPrefix = lang[:i]
				break
			}
		}
	}

	for _, supported := range SupportedLangs {
		if len(supported) >= len(langPrefix) && supported[:len(langPrefix)] == langPrefix {
			return supported
		}
	}

	// 无法匹配，回退到默认语言
	return DefaultLang
}

// TranslateMsg 根据错误码和语言返回对应翻译消息
// 支持 BCP 47 语言标签，自动进行前缀回退匹配
// 如果未找到翻译，返回空字符串
func TranslateMsg(code int, lang string) string {
	lang = resolveLang(lang)
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
