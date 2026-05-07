// Package xerr 提供应用级业务错误处理的单元测试
package xerr

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestNew 测试 New 函数创建 AppError 的基本功能
func TestNew(t *testing.T) {
	err := New(100001, "test error")

	// 验证返回的错误对象不为空
	if err == nil {
		t.Fatal("expected error to be non-nil")
	}

	// 验证业务错误码正确
	if err.Code != 100001 {
		t.Errorf("expected code 100001, got %d", err.Code)
	}

	// 验证用户提示信息正确
	if err.Msg != "test error" {
		t.Errorf("expected msg 'test error', got '%s'", err.Msg)
	}

	// 验证内部错误信息为空
	if err.Internal != "" {
		t.Errorf("expected Internal to be empty, got '%s'", err.Internal)
	}
}

// TestError_WithoutInternal 测试 Error() 方法在没有内部错误信息时的行为
func TestError_WithoutInternal(t *testing.T) {
	appErr := New(100001, "user message")

	// 没有内部错误时，Error() 应返回用户提示信息
	errMsg := appErr.Error()

	if errMsg != "user message" {
		t.Errorf("expected Error() to return 'user message', got '%s'", errMsg)
	}
}

// TestError_WithInternal 测试 Error() 方法在附加内部错误信息后的行为
func TestError_WithInternal(t *testing.T) {
	internalErr := errors.New("database connection failed")
	appErr := New(100001, "user message").WithInternal(internalErr)

	// 有内部错误时，Error() 应返回内部错误信息
	errMsg := appErr.Error()

	if errMsg != "database connection failed" {
		t.Errorf("expected Error() to return internal error message, got '%s'", errMsg)
	}

	// 验证 Internal 字段已设置
	if appErr.Internal != "database connection failed" {
		t.Errorf("expected Internal field to be 'database connection failed', got '%s'", appErr.Internal)
	}

	// 验证原始 Msg 字段未被修改
	if appErr.Msg != "user message" {
		t.Errorf("expected Msg field to remain 'user message', got '%s'", appErr.Msg)
	}
}

// TestWithInternal_NilError 测试 WithInternal 方法传入 nil 时会触发 panic
func TestWithInternal_NilError(t *testing.T) {
	appErr := New(100001, "user message")

	// 使用 defer/recover 捕获 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected WithInternal(nil) to panic")
		}
	}()

	appErr.WithInternal(nil)
}

// TestWithInternal_DoesNotModifyOriginal 测试 WithInternal 方法不修改原始错误对象
func TestWithInternal_DoesNotModifyOriginal(t *testing.T) {
	original := New(100001, "user message")
	internalErr := errors.New("internal error")

	modified := original.WithInternal(internalErr)

	// 验证原始错误的 Internal 字段仍为空
	if original.Internal != "" {
		t.Error("expected original error Internal to remain empty")
	}

	// 验证新错误的 Internal 字段已设置
	if modified.Internal != "internal error" {
		t.Errorf("expected modified error Internal to be 'internal error', got '%s'", modified.Internal)
	}

	// 验证返回的是新实例，不是原对象
	if original == modified {
		t.Error("expected WithInternal to return new instance")
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
		t.Errorf("expected result1 Internal to be 'first error', got '%s'", result1.Internal)
	}

	// 验证第二次链式调用的结果
	if result2.Internal != "second error" {
		t.Errorf("expected result2 Internal to be 'second error', got '%s'", result2.Internal)
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
				t.Fatalf("expected %s to be non-nil", tt.name)
			}

			// 验证错误码正确
			if tt.err.Code != tt.expectedCode {
				t.Errorf("expected %s code to be %d, got %d", tt.name, tt.expectedCode, tt.err.Code)
			}

			// 验证提示信息不为空
			if tt.err.Msg == "" {
				t.Errorf("expected %s to have non-empty Msg", tt.name)
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
				t.Fatalf("expected %s to be non-nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("expected %s code to be %d, got %d", tt.name, tt.expectedCode, tt.err.Code)
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
				t.Fatalf("expected %s to be non-nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("expected %s code to be %d, got %d", tt.name, tt.expectedCode, tt.err.Code)
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
				t.Fatalf("expected %s to be non-nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("expected %s code to be %d, got %d", tt.name, tt.expectedCode, tt.err.Code)
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
				t.Fatalf("expected %s to be non-nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("expected %s code to be %d, got %d", tt.name, tt.expectedCode, tt.err.Code)
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
				t.Fatalf("expected %s to be non-nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("expected %s code to be %d, got %d", tt.name, tt.expectedCode, tt.err.Code)
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
				t.Fatalf("expected %s to be non-nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("expected %s code to be %d, got %d", tt.name, tt.expectedCode, tt.err.Code)
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
				t.Fatalf("expected %s to be non-nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("expected %s code to be %d, got %d", tt.name, tt.expectedCode, tt.err.Code)
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
				t.Fatalf("expected %s to be non-nil", tt.name)
			}

			if tt.err.Code != tt.expectedCode {
				t.Errorf("expected %s code to be %d, got %d", tt.name, tt.expectedCode, tt.err.Code)
			}
		})
	}
}

// TestThirdPartyErrors 测试第三方服务错误码定义（2000xx 段）
func TestThirdPartyErrors(t *testing.T) {
	if ErrThirdPartyCallFailed == nil {
		t.Fatal("expected ErrThirdPartyCallFailed to be non-nil")
	}

	if ErrThirdPartyCallFailed.Code != 200000 {
		t.Errorf("expected code 200000, got %d", ErrThirdPartyCallFailed.Code)
	}
}

// TestAppError_ImplementsErrorInterface 测试 AppError 实现了 error 接口
func TestAppError_ImplementsErrorInterface(t *testing.T) {
	appErr := New(100001, "test")

	// 验证 AppError 可以赋值给 error 接口
	var err error = appErr

	if err == nil {
		t.Error("expected AppError to implement error interface")
	}

	if err.Error() != "test" {
		t.Errorf("expected error message 'test', got '%s'", err.Error())
	}
}

// TestAppError_ErrorWithFormattedInternal 测试 Error() 方法返回格式化后的内部错误信息
func TestAppError_ErrorWithFormattedInternal(t *testing.T) {
	originalErr := fmt.Errorf("database error: %w", errors.New("connection timeout"))
	appErr := New(100001, "user message").WithInternal(originalErr)

	errMsg := appErr.Error()

	// 验证错误信息包含关键内容
	if !strings.Contains(errMsg, "database error") {
		t.Errorf("expected error message to contain 'database error', got '%s'", errMsg)
	}
}

// TestAppError_JSONTags 测试 AppError 结构体的 JSON 标签定义
func TestAppError_JSONTags(t *testing.T) {
	appErr := New(100001, "user message")

	// 验证 Code 字段
	if appErr.Code != 100001 {
		t.Errorf("expected code 100001, got %d", appErr.Code)
	}

	// 验证 Msg 字段（json tag: "msg"）
	if appErr.Msg != "user message" {
		t.Errorf("expected msg 'user message', got '%s'", appErr.Msg)
	}

	// 验证 Internal 字段（json tag: "-"，不应序列化）
	if appErr.Internal != "" {
		t.Errorf("expected internal to be empty, got '%s'", appErr.Internal)
	}
}

// TestWithInternal_PreservesOriginalFields 测试 WithInternal 方法保留原始字段值
func TestWithInternal_PreservesOriginalFields(t *testing.T) {
	original := New(100001, "user message")
	internalErr := errors.New("internal error")

	modified := original.WithInternal(internalErr)

	// 验证 Code 字段保持不变
	if modified.Code != original.Code {
		t.Errorf("expected code to be preserved: %d vs %d", original.Code, modified.Code)
	}

	// 验证 Msg 字段保持不变
	if modified.Msg != original.Msg {
		t.Errorf("expected msg to be preserved: '%s' vs '%s'", original.Msg, modified.Msg)
	}

	// 验证 Internal 字段已正确设置
	if modified.Internal != internalErr.Error() {
		t.Errorf("expected internal to be set: '%s' vs '%s'", internalErr.Error(), modified.Internal)
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
		t.Errorf("expected err1 internal to be 'error 1', got '%s'", err1.Internal)
	}

	if err2.Internal != "error 2" {
		t.Errorf("expected err2 internal to be 'error 2', got '%s'", err2.Internal)
	}

	if err3.Internal != "error 3" {
		t.Errorf("expected err3 internal to be 'error 3', got '%s'", err3.Internal)
	}

	// 验证基础错误保持不变
	if base.Internal != "" {
		t.Error("expected base error internal to remain empty")
	}
}
