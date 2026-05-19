// Package xerr 提供应用级业务错误处理的单元测试
package xerr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// TestNew 测试 New 函数创建 AppError 的基本功能
func TestNew(t *testing.T) {
	err := New(100001, "test error")

	// 验证返回的错误对象不为空
	if err == nil {
		t.Fatal("期望错误对象不为 nil")
	}

	// 验证业务错误码正确
	if err.Code != 100001 {
		t.Errorf("期望错误码为 100001，得到 %d", err.Code)
	}

	// 验证用户提示信息正确
	if err.Msg != "test error" {
		t.Errorf("期望消息为 'test error'，得到 '%s'", err.Msg)
	}

	// 验证内部错误信息为空
	if err.Internal != "" {
		t.Errorf("期望 Internal 为空，得到 '%s'", err.Internal)
	}
}

// TestError_WithoutInternal 测试 Error() 方法在没有内部错误信息时的行为
func TestError_WithoutInternal(t *testing.T) {
	appErr := New(100001, "user message")

	// 没有内部错误时，Error() 应返回用户提示信息
	errMsg := appErr.Error()

	if errMsg != "user message" {
		t.Errorf("期望 Error() 返回 'user message'，得到 '%s'", errMsg)
	}
}

// TestError_WithInternal 测试 Error() 方法在附加内部错误信息后的行为
func TestError_WithInternal(t *testing.T) {
	internalErr := errors.New("database connection failed")
	appErr := New(100001, "user message").WithInternal(internalErr)

	// 有内部错误时，Error() 应返回 "用户提示: 内部错误"
	errMsg := appErr.Error()

	expectedMsg := "user message: database connection failed"
	if errMsg != expectedMsg {
		t.Errorf("期望 Error() 返回 '%s'，得到 '%s'", expectedMsg, errMsg)
	}

	// 验证 Internal 字段已设置
	if appErr.Internal != "database connection failed" {
		t.Errorf("期望 Internal 字段为 'database connection failed'，得到 '%s'", appErr.Internal)
	}

	// 验证原始 Msg 字段未被修改
	if appErr.Msg != "user message" {
		t.Errorf("期望 Msg 字段为 'user message'，得到 '%s'", appErr.Msg)
	}
}

// TestWithInternal_NilError 测试 WithInternal 方法传入 nil 时返回原始错误
func TestWithInternal_NilError(t *testing.T) {
	appErr := New(100001, "user message")

	result := appErr.WithInternal(nil)

	if result != appErr {
		t.Error("期望 WithInternal(nil) 返回原始错误")
	}
}

// TestWithInternal_DoesNotModifyOriginal 测试 WithInternal 方法不修改原始错误对象
func TestWithInternal_DoesNotModifyOriginal(t *testing.T) {
	original := New(100001, "user message")
	internalErr := errors.New("internal error")

	modified := original.WithInternal(internalErr)

	// 验证原始错误的 Internal 字段仍为空
	if original.Internal != "" {
		t.Error("期望原始错误的 Internal 字段保持为空")
	}

	// 验证新错误的 Internal 字段已设置
	if modified.Internal != "internal error" {
		t.Errorf("期望修改后的 Internal 为 'internal error'，得到 '%s'", modified.Internal)
	}

	// 验证返回的是新实例，不是原对象
	if original == modified {
		t.Error("期望 WithInternal 返回新实例")
	}
}

// TestWithInternal_Chaining 测试 WithInternal 方法的链式调用
func TestWithInternal_Chaining(t *testing.T) {
	appErr := New(100001, "user message")
	internal1 := errors.New("first error")
	internal2 := errors.New("second error")

	result1 := appErr.WithInternal(internal1)
	result2 := result1.WithInternal(internal2)

	// 验证第一次链式调用的结果
	if result1.Internal != "first error" {
		t.Errorf("期望 result1 的 Internal 为 'first error'，得到 '%s'", result1.Internal)
	}

	// 验证第二次链式调用的结果
	if result2.Internal != "second error" {
		t.Errorf("期望 result2 的 Internal 为 'second error'，得到 '%s'", result2.Internal)
	}
}

// TestPredefinedErrors 测试通用错误码定义（1000xx 段）
func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedCode int
	}{
		{"ErrBadRequest", ErrBadRequest, 100000},             // 请求参数错误
		{"ErrUnauthorized", ErrUnauthorized, 100001},         // 未登录或Token已过期
		{"ErrForbidden", ErrForbidden, 100002},               // 权限不足
		{"ErrNotFound", ErrNotFound, 100003},                 // 资源不存在
		{"ErrInternalServer", ErrInternalServer, 100004},     // 服务器内部错误
		{"ErrTooManyRequest", ErrTooManyRequest, 100005},     // 请求过于频繁
		{"ErrMethodNotAllowed", ErrMethodNotAllowed, 100006}, // 请求方法不允许
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("期望 %s 不为 nil", tt.name)
			}

			// 验证错误码正确
			if tt.err.Code != tt.expectedCode {
				t.Errorf("期望 %s 错误码为 %d，得到 %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestRegisterErrors 测试注册模块错误码定义（1101xx 段）
func TestRegisterErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedCode int
	}{
		{"ErrRegisterUsernameEmpty", ErrRegisterUsernameEmpty, 110100},     // 用户名不能为空
		{"ErrRegisterPasswordEmpty", ErrRegisterPasswordEmpty, 110101},     // 密码不能为空
		{"ErrRegisterUsernameInvalid", ErrRegisterUsernameInvalid, 110102}, // 用户名格式不合法
		{"ErrRegisterPasswordInvalid", ErrRegisterPasswordInvalid, 110103}, // 密码格式不合法
		{"ErrRegisterUsernameExists", ErrRegisterUsernameExists, 110104},   // 用户名已存在
		{"ErrRegisterFailed", ErrRegisterFailed, 110105},                   // 注册失败
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("期望 %s 不为 nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("期望 %s 错误码为 %d，得到 %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestLoginErrors 测试登录模块错误码定义（1102xx 段）
func TestLoginErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedCode int
	}{
		{"ErrLoginInvalidCredentials", ErrLoginInvalidCredentials, 110200}, // 用户名或密码错误
		{"ErrLoginFailed", ErrLoginFailed, 110201},                         // 登录失败
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("期望 %s 不为 nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("期望 %s 错误码为 %d，得到 %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestRefreshTokenErrors 测试刷新令牌模块错误码定义（1103xx 段）
func TestRefreshTokenErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedCode int
	}{
		{"ErrRefreshTokenInvalid", ErrRefreshTokenInvalid, 110300}, // 刷新令牌无效
		{"ErrRefreshTokenRevoked", ErrRefreshTokenRevoked, 110301}, // 刷新令牌已撤销
		{"ErrRefreshTokenExpired", ErrRefreshTokenExpired, 110302}, // 刷新令牌已过期
		{"ErrRefreshTokenFailed", ErrRefreshTokenFailed, 110303},   // 刷新令牌失败
		{"ErrLogoutFailed", ErrLogoutFailed, 110304},               // 退出登录失败
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("期望 %s 不为 nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("期望 %s 错误码为 %d，得到 %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestChangePasswordErrors 测试密码管理模块错误码定义（1104xx 段）
func TestChangePasswordErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedCode int
	}{
		{"ErrChangePasswordOldEmpty", ErrChangePasswordOldEmpty, 110400},         // 旧密码不能为空
		{"ErrChangePasswordNewEmpty", ErrChangePasswordNewEmpty, 110401},         // 新密码不能为空
		{"ErrChangePasswordOldIncorrect", ErrChangePasswordOldIncorrect, 110402}, // 旧密码错误
		{"ErrChangePasswordFailed", ErrChangePasswordFailed, 110403},             // 密码修改失败
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("期望 %s 不为 nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("期望 %s 错误码为 %d，得到 %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestSourceErrors 测试 Source 模块错误码定义（1105xx 段）
func TestSourceErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedCode int
	}{
		{"ErrSourceParamNameEmpty", ErrSourceParamNameEmpty, 110500},   // 来源名称不能为空
		{"ErrSourceInsertFailed", ErrSourceInsertFailed, 110501},       // 创建来源失败
		{"ErrSourceTokenFailed", ErrSourceTokenFailed, 110502},         // 生成来源Token失败
		{"ErrSourceNotFound", ErrSourceNotFound, 110503},               // 来源不存在
		{"ErrSourceNoPermission", ErrSourceNoPermission, 110504},       // 无权操作该来源
		{"ErrSourceQueryFailed", ErrSourceQueryFailed, 110505},         // 查询来源失败
		{"ErrSourceDeleteFailed", ErrSourceDeleteFailed, 110506},       // 删除来源失败
		{"ErrSourceAlreadyDisabled", ErrSourceAlreadyDisabled, 110507}, // 来源已停用
		{"ErrSourceAlreadyEnabled", ErrSourceAlreadyEnabled, 110508},   // 来源已启用
		{"ErrSourceAlreadyDeleted", ErrSourceAlreadyDeleted, 110509},   // 来源已删除
		{"ErrSourceUpdateFailed", ErrSourceUpdateFailed, 110510},       // 更新来源失败
		{"ErrSourceDisabled", ErrSourceDisabled, 110511},               // 来源已停用，无法推送
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("期望 %s 不为 nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("期望 %s 错误码为 %d，得到 %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestChannelErrors 测试 Channel 模块错误码定义（1106xx 段）
func TestChannelErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedCode int
	}{
		{"ErrChannelParamNameEmpty", ErrChannelParamNameEmpty, 110600},   // 渠道名称不能为空
		{"ErrChannelInvalidType", ErrChannelInvalidType, 110601},         // 无效的渠道类型
		{"ErrChannelInsertFailed", ErrChannelInsertFailed, 110602},       // 创建渠道失败
		{"ErrChannelNotFound", ErrChannelNotFound, 110603},               // 渠道不存在
		{"ErrChannelNoPermission", ErrChannelNoPermission, 110604},       // 无权操作该渠道
		{"ErrChannelQueryFailed", ErrChannelQueryFailed, 110605},         // 查询渠道失败
		{"ErrChannelDeleteFailed", ErrChannelDeleteFailed, 110606},       // 删除渠道失败
		{"ErrChannelAlreadyDisabled", ErrChannelAlreadyDisabled, 110607}, // 渠道已停用
		{"ErrChannelAlreadyEnabled", ErrChannelAlreadyEnabled, 110608},   // 渠道已启用
		{"ErrChannelAlreadyDeleted", ErrChannelAlreadyDeleted, 110609},   // 渠道已删除
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("期望 %s 不为 nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("期望 %s 错误码为 %d，得到 %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestMessageErrors 测试 Message 模块错误码定义（1107xx 段）
func TestMessageErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedCode int
	}{
		{"ErrMessageParamTitleEmpty", ErrMessageParamTitleEmpty, 110700},     // 消息标题不能为空
		{"ErrMessageParamContentEmpty", ErrMessageParamContentEmpty, 110701}, // 消息内容不能为空
		{"ErrMessagePushFailed", ErrMessagePushFailed, 110702},               // 消息推送失败
		{"ErrMessageNoChannels", ErrMessageNoChannels, 110703},               // 来源未绑定任何渠道
		{"ErrMessageRecordFailed", ErrMessageRecordFailed, 110704},           // 记录消息状态失败
		{"ErrMessageAlreadyDeleted", ErrMessageAlreadyDeleted, 110705},       // 消息已删除
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("期望 %s 不为 nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("期望 %s 错误码为 %d，得到 %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestUserProfileErrors 测试用户管理模块错误码定义（1108xx 段）
func TestUserProfileErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedCode int
	}{
		{"ErrUserProfileNotFound", ErrUserProfileNotFound, 110800},       // 用户不存在
		{"ErrUserProfileQueryFailed", ErrUserProfileQueryFailed, 110801}, // 查询用户信息失败
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("期望 %s 不为 nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("期望 %s 错误码为 %d，得到 %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestThirdPartyErrors 测试第三方服务错误码定义（2000xx 段）
func TestThirdPartyErrors(t *testing.T) {
	if ErrThirdPartyCallFailed == nil {
		t.Fatal("期望 ErrThirdPartyCallFailed 不为 nil")
	}

	if ErrThirdPartyCallFailed.Code != 200000 {
		t.Errorf("期望错误码为 200000，得到 %d", ErrThirdPartyCallFailed.Code)
	}
}

// TestAppError_ImplementsErrorInterface 测试 AppError 实现了 error 接口
func TestAppError_ImplementsErrorInterface(t *testing.T) {
	appErr := New(100001, "test")

	// 验证 AppError 可以赋值给 error 接口
	var err error = appErr

	if err == nil {
		t.Error("期望 AppError 实现 error 接口")
	}

	if err.Error() != "test" {
		t.Errorf("期望错误消息为 'test'，得到 '%s'", err.Error())
	}
}

// TestAppError_ErrorWithFormattedInternal 测试 Error() 方法返回格式化后的内部错误信息
func TestAppError_ErrorWithFormattedInternal(t *testing.T) {
	originalErr := fmt.Errorf("database error: %w", errors.New("connection timeout"))
	appErr := New(100001, "user message").WithInternal(originalErr)

	errMsg := appErr.Error()

	// 验证错误信息包含关键内容
	if !strings.Contains(errMsg, "database error") {
		t.Errorf("期望错误消息包含 'database error'，得到 '%s'", errMsg)
	}
}

// TestAppError_JSONTags 测试 AppError 结构体的 JSON 标签定义
func TestAppError_JSONTags(t *testing.T) {
	appErr := New(100001, "user message")

	// 验证 Code 字段
	if appErr.Code != 100001 {
		t.Errorf("期望错误码为 100001，得到 %d", appErr.Code)
	}

	// 验证 Msg 字段（json tag: "msg"）
	if appErr.Msg != "user message" {
		t.Errorf("期望消息为 'user message'，得到 '%s'", appErr.Msg)
	}

	// 验证 Internal 字段（json tag: "-"，不应序列化）
	if appErr.Internal != "" {
		t.Errorf("期望 Internal 为空，得到 '%s'", appErr.Internal)
	}
}

// TestWithInternal_PreservesOriginalFields 测试 WithInternal 方法保留原始字段值
func TestWithInternal_PreservesOriginalFields(t *testing.T) {
	original := New(100001, "user message")
	internalErr := errors.New("internal error")

	modified := original.WithInternal(internalErr)

	// 验证 Code 字段保持不变
	if modified.Code != original.Code {
		t.Errorf("期望错误码保持不变: %d vs %d", original.Code, modified.Code)
	}

	// 验证 Msg 字段保持不变
	if modified.Msg != original.Msg {
		t.Errorf("期望消息保持不变: '%s' vs '%s'", original.Msg, modified.Msg)
	}

	// 验证 Internal 字段已正确设置
	if modified.Internal != internalErr.Error() {
		t.Errorf("期望 Internal 已设置: '%s' vs '%s'", internalErr.Error(), modified.Internal)
	}
}

// TestMultipleWithInternalCalls 测试多次调用 WithInternal 的独立性
func TestMultipleWithInternalCalls(t *testing.T) {
	base := New(100000, "base error")

	// 从同一个基础错误生成多个变体
	err1 := base.WithInternal(errors.New("error 1"))
	err2 := base.WithInternal(errors.New("error 2"))
	err3 := base.WithInternal(errors.New("error 3"))

	// 验证每个变体的 Internal 字段相互独立
	if err1.Internal != "error 1" {
		t.Errorf("期望 err1 的 Internal 为 'error 1'，得到 '%s'", err1.Internal)
	}

	if err2.Internal != "error 2" {
		t.Errorf("期望 err2 的 Internal 为 'error 2'，得到 '%s'", err2.Internal)
	}

	if err3.Internal != "error 3" {
		t.Errorf("期望 err3 的 Internal 为 'error 3'，得到 '%s'", err3.Internal)
	}

	// 验证基础错误保持不变
	if base.Internal != "" {
		t.Error("期望基础错误的 Internal 保持为空")
	}
}

// TestUnwrap 测试 Unwrap 方法支持错误链
func TestUnwrap(t *testing.T) {
	originalErr := errors.New("database connection failed")
	appErr := New(100001, "user message").WithInternal(originalErr)

	// 验证 Unwrap 返回原始错误
	unwrapped := appErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("期望 Unwrap 返回原始错误，得到 %v", unwrapped)
	}
}

// TestUnwrap_NilCause 测试无 Cause 时 Unwrap 返回 nil
func TestUnwrap_NilCause(t *testing.T) {
	appErr := New(100001, "user message")

	if appErr.Unwrap() != nil {
		t.Error("期望无 Cause 时 Unwrap 返回 nil")
	}
}

// TestErrorsIs 测试 errors.Is 支持错误链遍历
func TestErrorsIs(t *testing.T) {
	originalErr := errors.New("database connection failed")
	appErr := New(100001, "user message").WithInternal(originalErr)

	// 验证 errors.Is 能找到链中的原始错误
	if !errors.Is(appErr, originalErr) {
		t.Error("期望 errors.Is 能在错误链中找到原始错误")
	}

	// 验证 errors.Is 不会匹配不相关的错误
	unrelatedErr := errors.New("unrelated error")
	if errors.Is(appErr, unrelatedErr) {
		t.Error("期望 errors.Is 不匹配不相关的错误")
	}
}

// TestErrorsAs 测试 errors.As 支持错误链类型匹配
func TestErrorsAs(t *testing.T) {
	dbErr := &DBError{Message: "connection timeout"}
	appErr := New(100001, "user message").WithInternal(dbErr)

	// 验证 errors.As 能提取链中的自定义错误类型
	var extracted *DBError
	if !errors.As(appErr, &extracted) {
		t.Error("期望 errors.As 能从错误链中提取 DBError")
	}

	if extracted.Message != "connection timeout" {
		t.Errorf("期望提取的消息为 'connection timeout'，得到 '%s'", extracted.Message)
	}
}

// DBError 自定义数据库错误类型，用于测试 errors.As
type DBError struct {
	Message string
}

func (e *DBError) Error() string {
	return e.Message
}

// TestWithCause 测试 WithCause 方法仅附加原始错误
func TestWithCause(t *testing.T) {
	originalErr := errors.New("database error")
	appErr := New(100001, "user message").WithCause(originalErr)

	// 验证 Cause 字段已设置
	if appErr.Cause != originalErr {
		t.Error("期望 Cause 字段已设置为原始错误")
	}

	// 验证 Internal 字段保持为空
	if appErr.Internal != "" {
		t.Errorf("期望 Internal 字段保持为空，得到 '%s'", appErr.Internal)
	}

	// 验证 Error() 返回 Msg 而不是 Internal
	if appErr.Error() != "user message" {
		t.Errorf("期望 Error() 返回 'user message'，得到 '%s'", appErr.Error())
	}
}

// TestWithCause_NilError 测试 WithCause 方法传入 nil 时返回原始错误
func TestWithCause_NilError(t *testing.T) {
	appErr := New(100001, "user message")

	result := appErr.WithCause(nil)

	if result != appErr {
		t.Error("期望 WithCause(nil) 返回原始错误")
	}
}

// TestWrap 测试 Wrap 方法包装错误添加上下文
func TestWrap(t *testing.T) {
	appErr := New(100001, "user message").Wrap("查询用户失败: %s", "database timeout")

	// 验证 Internal 字段已设置格式化消息
	if appErr.Internal != "查询用户失败: database timeout" {
		t.Errorf("期望 Internal 为 '查询用户失败: database timeout'，得到 '%s'", appErr.Internal)
	}

	// 验证原始字段保持不变
	if appErr.Code != 100001 {
		t.Errorf("期望错误码为 100001，得到 %d", appErr.Code)
	}

	if appErr.Msg != "user message" {
		t.Errorf("期望消息为 'user message'，得到 '%s'", appErr.Msg)
	}
}

// TestHTTPStatus 测试 HTTPStatus 方法返回正确状态码
func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name         string
		err          *AppError
		expectedHTTP int
	}{
		{"BadRequest", ErrBadRequest, 200},
		{"Unauthorized", ErrUnauthorized, 401},
		{"Forbidden", ErrForbidden, 200},
		{"NotFound", ErrNotFound, 200},
		{"InternalServer", ErrInternalServer, 200},
		{"TooManyRequest", ErrTooManyRequest, 200},
		{"MethodNotAllowed", ErrMethodNotAllowed, 200},
		{"SourceNotFound", ErrSourceNotFound, 200},
		{"ThirdPartyCallFailed", ErrThirdPartyCallFailed, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := tt.err.HTTPStatus()
			if status != tt.expectedHTTP {
				t.Errorf("期望 HTTP 状态码为 %d，得到 %d", tt.expectedHTTP, status)
			}
		})
	}
}

// TestHTTPStatusFromCode 测试 HTTPStatusFromCode 函数映射
func TestHTTPStatusFromCode(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		expectedHTTP int
	}{
		{"Success", 0, 200},
		{"BadRequest", 100000, 200},
		{"Unauthorized", 100001, 401},
		{"Forbidden", 100002, 200},
		{"NotFound", 100003, 200},
		{"InternalServer", 100004, 200},
		{"TooManyRequests", 100005, 200},
		{"MethodNotAllowed", 100006, 200},
		{"RegisterError", 110100, 200},
		{"LoginError", 110200, 200},
		{"RefreshTokenError", 110300, 200},
		{"SourceError", 110500, 200},
		{"ChannelError", 110600, 200},
		{"MessageError", 110700, 200},
		{"ThirdPartyError", 200000, 200},
		{"UnknownError", 999999, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := HTTPStatusFromCode(tt.code)
			if status != tt.expectedHTTP {
				t.Errorf("期望错误码 %d 对应 HTTP 状态码 %d，得到 %d", tt.code, tt.expectedHTTP, status)
			}
		})
	}
}

// TestErrorChain_MultipleLevels 测试多层错误链
func TestErrorChain_MultipleLevels(t *testing.T) {
	// 模拟多层错误包装
	dbErr := errors.New("connection timeout")
	repoErr := New(110503, "来源不存在").WithInternal(fmt.Errorf("查询数据库失败: %w", dbErr))

	// 验证 errors.Is 能找到最底层的原始错误
	if !errors.Is(repoErr, dbErr) {
		t.Error("期望 errors.Is 能在错误链中找到最底层的 dbErr")
	}

	// 验证 AppError 本身也能被识别
	var appErr *AppError
	if !errors.As(repoErr, &appErr) {
		t.Error("期望 errors.As 能在错误链中识别 AppError")
	}

	if appErr.Code != 110503 {
		t.Errorf("期望错误码为 110503，得到 %d", appErr.Code)
	}
}

// TestWithInternal_PreservesCause 测试 WithInternal 保留 Cause 字段
func TestWithInternal_PreservesCause(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := New(100001, "user message").WithInternal(originalErr)

	if appErr.Cause != originalErr {
		t.Error("期望 Cause 字段被设置为原始错误")
	}

	if appErr.Internal != "original error" {
		t.Errorf("期望 Internal 为 'original error'，得到 '%s'", appErr.Internal)
	}
}

// TestNew_WithoutMsg 测试 New 函数不传消息时 Msg 为空
func TestNew_WithoutMsg(t *testing.T) {
	err := New(CodeBadRequest)

	if err.Code != CodeBadRequest {
		t.Errorf("期望错误码为 %d，得到 %d", CodeBadRequest, err.Code)
	}

	if err.Msg != "" {
		t.Errorf("期望 Msg 为空，得到 '%s'", err.Msg)
	}
}

// TestTranslateMsg_Zh 测试中文翻译
func TestTranslateMsg_Zh(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{CodeBadRequest, "请求参数错误"},
		{CodeUnauthorized, "未登录或Token已过期"},
		{CodeLoginInvalidCredentials, "用户名或密码错误"},
		{CodeSourceNotFound, "来源不存在"},
	}

	for _, tt := range tests {
		msg := TranslateMsg(tt.code, "zh")
		if msg != tt.expected {
			t.Errorf("错误码 %d 期望翻译 '%s'，得到 '%s'", tt.code, tt.expected, msg)
		}
	}
}

// TestTranslateMsg_En 测试英文翻译
func TestTranslateMsg_En(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{CodeBadRequest, "Invalid request parameters"},
		{CodeUnauthorized, "Not logged in or token expired"},
		{CodeLoginInvalidCredentials, "Invalid username or password"},
		{CodeSourceNotFound, "Source not found"},
	}

	for _, tt := range tests {
		msg := TranslateMsg(tt.code, "en")
		if msg != tt.expected {
			t.Errorf("错误码 %d 期望翻译 '%s'，得到 '%s'", tt.code, tt.expected, msg)
		}
	}
}

// TestTranslateMsg_DefaultLang 测试默认语言回退
func TestTranslateMsg_DefaultLang(t *testing.T) {
	// 空语言应回退到中文
	msg := TranslateMsg(CodeBadRequest, "")
	if msg != "请求参数错误" {
		t.Errorf("期望默认语言为中文 '请求参数错误'，得到 '%s'", msg)
	}

	// 不支持的语言应回退到中文
	msg = TranslateMsg(CodeBadRequest, "ja")
	if msg != "请求参数错误" {
		t.Errorf("期望回退到中文 '请求参数错误'，得到 '%s'", msg)
	}
}

// TestTranslateMsg_UnknownCode 测试未知错误码返回空字符串
func TestTranslateMsg_UnknownCode(t *testing.T) {
	msg := TranslateMsg(999999, "zh")
	if msg != "" {
		t.Errorf("期望未知错误码返回空字符串，得到 '%s'", msg)
	}
}

// TestTranslateError 测试 TranslateError 函数
func TestTranslateError(t *testing.T) {
	err := New(CodeBadRequest)

	// 测试英文翻译
	enErr := TranslateError(err, "en")
	if enErr.Msg != "Invalid request parameters" {
		t.Errorf("期望英文翻译 'Invalid request parameters'，得到 '%s'", enErr.Msg)
	}

	// 验证原始错误未被修改
	if err.Msg != "" {
		t.Errorf("期望原始错误 Msg 为空，得到 '%s'", err.Msg)
	}

	// 测试中文翻译
	zhErr := TranslateError(err, "zh")
	if zhErr.Msg != "请求参数错误" {
		t.Errorf("期望中文翻译 '请求参数错误'，得到 '%s'", zhErr.Msg)
	}
}

// TestTranslateError_Nil 测试 TranslateError 传入 nil
func TestTranslateError_Nil(t *testing.T) {
	result := TranslateError(nil, "en")
	if result != nil {
		t.Error("期望 TranslateError(nil) 返回 nil")
	}
}

// TestTranslateError_UnknownCode 测试 TranslateError 未知错误码返回原错误
func TestTranslateError_UnknownCode(t *testing.T) {
	err := New(999999)
	result := TranslateError(err, "en")

	// 未知错误码应返回原错误
	if result != err {
		t.Error("期望未知错误码返回原错误")
	}
}

// TestNew_EmptyMsg 测试 New 函数传空字符串 msg
func TestNew_EmptyMsg(t *testing.T) {
	err := New(CodeBadRequest, "")

	if err.Msg != "" {
		t.Errorf("期望 Msg 为空，得到 '%s'", err.Msg)
	}
}

// TestError_OnlyInternal 测试 Error() 方法仅有 Internal 无 Cause
func TestError_OnlyInternal(t *testing.T) {
	appErr := &AppError{
		Code:     CodeBadRequest,
		Msg:      "user message",
		Internal: "internal details",
	}

	errMsg := appErr.Error()
	if errMsg != "internal details" {
		t.Errorf("期望 Error() 返回 Internal '%s'，得到 '%s'", "internal details", errMsg)
	}
}

// TestWrap_NoArgs 测试 Wrap 不传 args
func TestWrap_NoArgs(t *testing.T) {
	appErr := New(CodeBadRequest, "user message").Wrap("operation failed")

	if appErr.Internal != "operation failed" {
		t.Errorf("期望 Internal 为 'operation failed'，得到 '%s'", appErr.Internal)
	}
}

// TestWrap_MultipleArgs 测试 Wrap 传多个 args
func TestWrap_MultipleArgs(t *testing.T) {
	appErr := New(CodeBadRequest, "user message").Wrap("failed: %s, code: %d, retry: %t", "timeout", 500, true)

	expected := "failed: timeout, code: 500, retry: true"
	if appErr.Internal != expected {
		t.Errorf("期望 Internal 为 '%s'，得到 '%s'", expected, appErr.Internal)
	}
}

// TestIs_SameCode 测试 Is 方法相同错误码等价
func TestIs_SameCode(t *testing.T) {
	err1 := New(CodeBadRequest, "msg1")
	err2 := New(CodeBadRequest, "msg2")

	if !err1.Is(err2) {
		t.Error("期望相同错误码等价")
	}
}

// TestIs_DifferentCode 测试 Is 方法不同错误码不等价
func TestIs_DifferentCode(t *testing.T) {
	err1 := New(CodeBadRequest, "msg1")
	err2 := New(CodeNotFound, "msg2")

	if err1.Is(err2) {
		t.Error("期望不同错误码不等价")
	}
}

// TestIs_NonAppError 测试 Is 方法与非 AppError 比较
func TestIs_NonAppError(t *testing.T) {
	err := New(CodeBadRequest, "msg")
	target := errors.New("standard error")

	if err.Is(target) {
		t.Error("期望与非 AppError 不等价")
	}
}

// TestIsBadRequest_Nil 测试 IsBadRequest 传入 nil
func TestIsBadRequest_Nil(t *testing.T) {
	if IsBadRequest(nil) {
		t.Error("期望 nil 不是 BadRequest")
	}
}

// TestIsServerError_Nil 测试 IsServerError 传入 nil
func TestIsServerError_Nil(t *testing.T) {
	if IsServerError(nil) {
		t.Error("期望 nil 不是 ServerError")
	}
}

// TestIsBadRequest_NonAppError 测试 IsBadRequest 传入非 AppError
func TestIsBadRequest_NonAppError(t *testing.T) {
	err := errors.New("standard error")
	if IsBadRequest(err) {
		t.Error("期望非 AppError 不是 BadRequest")
	}
}

// TestIsServerError_NonAppError 测试 IsServerError 传入非 AppError
func TestIsServerError_NonAppError(t *testing.T) {
	err := errors.New("standard error")
	if IsServerError(err) {
		t.Error("期望非 AppError 不是 ServerError")
	}
}

// TestIsBadRequest_TableDriven 测试 IsBadRequest 表驱动
func TestIsBadRequest_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"BadRequest", ErrBadRequest, false},
		{"Unauthorized", ErrUnauthorized, true},
		{"InternalServer", ErrInternalServer, false},
		{"Nil", nil, false},
		{"StandardError", errors.New("test"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBadRequest(tt.err)
			if result != tt.expected {
				t.Errorf("期望 IsBadRequest 返回 %v，得到 %v", tt.expected, result)
			}
		})
	}
}

// TestIsServerError_TableDriven 测试 IsServerError 表驱动
func TestIsServerError_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"InternalServer", ErrInternalServer, false},
		{"BadRequest", ErrBadRequest, false},
		{"Nil", nil, false},
		{"StandardError", errors.New("test"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServerError(tt.err)
			if result != tt.expected {
				t.Errorf("期望 IsServerError 返回 %v，得到 %v", tt.expected, result)
			}
		})
	}
}

// TestHTTPStatus_ZeroFallback 测试 HTTPStatus 在 httpStatus 为 0 时回退
func TestHTTPStatus_ZeroFallback(t *testing.T) {
	appErr := &AppError{
		Code:       CodeBadRequest,
		Msg:        "test",
		httpStatus: 0,
	}

	if appErr.HTTPStatus() != http.StatusOK {
		t.Errorf("期望回退到 200，得到 %d", appErr.HTTPStatus())
	}
}

// TestErrorChain_MultipleWrap 测试多次 Wrap 的独立性
func TestErrorChain_MultipleWrap(t *testing.T) {
	base := New(CodeBadRequest, "base")

	wrapped1 := base.Wrap("context 1: %s", "error A")
	wrapped2 := base.Wrap("context 2: %s", "error B")

	if wrapped1.Internal == wrapped2.Internal {
		t.Error("期望多次 Wrap 相互独立")
	}

	// 验证基础错误未被修改
	if base.Internal != "" {
		t.Error("期望基础错误 Internal 保持为空")
	}
}

// TestWithCause_DoesNotSetInternal 测试 WithCause 不设置 Internal
func TestWithCause_DoesNotSetInternal(t *testing.T) {
	cause := errors.New("root cause")
	appErr := New(CodeBadRequest, "user message").WithCause(cause)

	if appErr.Internal != "" {
		t.Errorf("期望 Internal 保持为空，得到 '%s'", appErr.Internal)
	}

	if appErr.Cause != cause {
		t.Error("期望 Cause 已设置")
	}
}

// TestError_WithCauseAndInternal 测试同时有 Cause 和 Internal 时 Error() 行为
func TestError_WithCauseAndInternal(t *testing.T) {
	cause := errors.New("root cause")
	appErr := New(CodeBadRequest, "user message").WithInternal(cause)

	errMsg := appErr.Error()
	expected := "user message: root cause"
	if errMsg != expected {
		t.Errorf("期望 Error() 返回 '%s'，得到 '%s'", expected, errMsg)
	}
}

// TestUnwrap_ChainTraversal 测试 Unwrap 支持错误链遍历
func TestUnwrap_ChainTraversal(t *testing.T) {
	rootCause := errors.New("root")
	appErr := New(CodeBadRequest, "user").WithInternal(rootCause)

	unwrapped := appErr.Unwrap()
	if unwrapped != rootCause {
		t.Errorf("期望 Unwrap 返回 rootCause，得到 %v", unwrapped)
	}
}

// TestErrorsIs_ChainWithAppError 测试 errors.Is 遍历包含 AppError 的链
func TestErrorsIs_ChainWithAppError(t *testing.T) {
	baseErr := ErrNotFound
	wrapped := baseErr.Wrap("query failed: %s", "timeout")

	if !errors.Is(wrapped, baseErr) {
		t.Error("期望 errors.Is 能匹配等价 AppError")
	}
}

// TestTranslateError_PreservesOtherFields 测试 TranslateError 保留其他字段
func TestTranslateError_PreservesOtherFields(t *testing.T) {
	cause := errors.New("root cause")
	original := New(CodeBadRequest).WithInternal(cause)

	translated := TranslateError(original, "en")

	if translated.Code != original.Code {
		t.Error("期望 Code 字段保持不变")
	}

	if translated.Cause != original.Cause {
		t.Error("期望 Cause 字段保持不变")
	}
}
