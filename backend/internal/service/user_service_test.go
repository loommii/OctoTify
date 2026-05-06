package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"octotify/internal/handler/dto"
	"octotify/internal/model"
	pkgjwtx "octotify/pkg/jwtx"
	"octotify/pkg/xerr"
)

// ============================================================================
// TestUserService_Register 注册功能测试
// ============================================================================

// TestUserService_Register 测试用户注册功能
// 覆盖场景：正常注册、用户名已存在
func TestUserService_Register(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	tests := []struct {
		name        string           // 测试用例名称
		req         *dto.RegisterReq // 注册请求参数
		setup       func()           // 测试前置操作（可选）
		wantErrCode int              // 期望的错误码（0 表示无错误）
		wantSuccess bool             // 是否期望成功
	}{
		{
			name: "成功：使用有效的用户名和密码注册",
			req: &dto.RegisterReq{
				AuthCredentials: dto.AuthCredentials{
					Username: "newuser",
					Password: "Password1",
				},
			},
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name: "失败：用户名已存在，应返回 ErrRegisterUsernameExists",
			req: &dto.RegisterReq{
				AuthCredentials: dto.AuthCredentials{
					Username: "existinguser",
					Password: "Password1",
				},
			},
			setup: func() {
				CreateTestUser(t, db, "existinguser", "Password1")
			},
			wantErrCode: xerr.ErrRegisterUsernameExists.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			svc := NewUserService(db, accessHelper, refreshHelper, logger)
			resp, err := svc.Register(ctx, tt.req)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken, "Access Token 不应为空")
				assert.NotEmpty(t, resp.RefreshToken, "Refresh Token 不应为空")
				assert.Equal(t, tt.req.Username, resp.User.Username)
				assert.NotZero(t, resp.User.ID, "User ID 不应为零")
				assert.NotZero(t, resp.User.CreatedAt, "CreatedAt 不应为零")

				// 验证数据库中确实创建了用户
				var user model.User
				findErr := db.Where("username = ?", tt.req.Username).First(&user).Error
				assert.NoError(t, findErr, "数据库中应存在该用户")
				assert.Equal(t, tt.req.Username, user.Username)

				// 验证 RefreshToken 记录已创建
				var refreshTokenCount int64
				db.Table("refresh_tokens").Where("user_id = ?", user.ID).Count(&refreshTokenCount)
				assert.Equal(t, int64(1), refreshTokenCount, "应恰好创建一条 RefreshToken 记录")
			} else {
				assert.Error(t, err)
				assert.Nil(t, resp)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// TestUserService_Register_TxError 测试注册时数据库事务错误场景
// 说明：Register 的事务内部错误需要外部并发操作才能可靠触发。
// 在单线程单元测试中，通过关闭数据库连接的方式让查询阶段就报错，
// 从而间接验证 ErrRegisterFailed 的错误处理路径。
func TestUserService_Register_TxError(t *testing.T) {
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	// 关闭数据库连接，使 Register 内的查询操作报错
	sqlDB, err := db.DB()
	require.NoError(t, err, "获取底层 sql.DB 失败")
	err = sqlDB.Close()
	require.NoError(t, err, "关闭数据库连接失败")

	svc := NewUserService(db, accessHelper, refreshHelper, logger)
	resp, err := svc.Register(ctx, &dto.RegisterReq{
		AuthCredentials: dto.AuthCredentials{
			Username: "db_down_user",
			Password: "Password1",
		},
	})

	// 由于数据库连接已关闭，查询阶段就会失败，应返回 ErrRegisterFailed
	assert.Error(t, err)
	assert.Nil(t, resp)
	appErr, ok := err.(*xerr.AppError)
	assert.True(t, ok, "错误类型应为 *xerr.AppError")
	assert.Equal(t, xerr.ErrRegisterFailed.Code, appErr.Code, "应返回注册失败错误码")
}

// ============================================================================
// TestUserService_GetUserProfile 获取用户资料测试
// ============================================================================

// TestUserService_GetUserProfile 测试通过用户 ID 获取用户资料功能
// 覆盖场景：成功获取、用户不存在
func TestUserService_GetUserProfile(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	// 预创建测试用户
	testUser := CreateTestUser(t, db, "profile_user", "Password1")

	tests := []struct {
		name        string // 测试用例名称
		userID      int64  // 用户 ID
		wantErrCode int    // 期望的错误码（0 表示无错误）
		wantSuccess bool   // 是否期望成功
	}{
		{
			name:        "成功：获取现有用户的资料",
			userID:      testUser.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：用户不存在，应返回 ErrUserProfileNotFound",
			userID:      99999,
			wantErrCode: xerr.ErrUserProfileNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewUserService(db, accessHelper, refreshHelper, logger)
			resp, err := svc.GetUserProfile(ctx, tt.userID)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, testUser.ID, resp.User.ID)
				assert.Equal(t, "profile_user", resp.User.Username)
				assert.NotZero(t, resp.User.CreatedAt, "CreatedAt 不应为零")
			} else {
				assert.Error(t, err)
				assert.Nil(t, resp)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// TestUserService_GetUserProfile_DBError 测试数据库查询错误场景
// 通过关闭数据库连接来模拟查询失败，验证 ErrUserProfileQueryFailed 路径
func TestUserService_GetUserProfile_DBError(t *testing.T) {
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	// 获取底层的 sql.DB 并关闭它，模拟数据库连接断开
	sqlDB, err := db.DB()
	require.NoError(t, err, "获取底层 sql.DB 失败")
	err = sqlDB.Close()
	require.NoError(t, err, "关闭数据库连接失败")

	svc := NewUserService(db, accessHelper, refreshHelper, logger)
	resp, err := svc.GetUserProfile(ctx, 1)

	// 应该返回 ErrUserProfileQueryFailed
	assert.Error(t, err)
	assert.Nil(t, resp)
	appErr, ok := err.(*xerr.AppError)
	assert.True(t, ok, "错误类型应为 *xerr.AppError")
	assert.Equal(t, xerr.ErrUserProfileQueryFailed.Code, appErr.Code, "应返回数据库查询失败错误码")
}

// ============================================================================
// TestUserService_GetUserProfileByID 通过字符串 ID 获取用户资料测试
// ============================================================================

// TestUserService_GetUserProfileByID 测试通过字符串形式的用户 ID 获取用户资料功能
// 覆盖场景：有效 ID 字符串、无效 ID 字符串（非数字）
func TestUserService_GetUserProfileByID(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	// 预创建测试用户
	testUser := CreateTestUser(t, db, "profile_by_id_user", "Password1")

	tests := []struct {
		name        string // 测试用例名称
		userIDStr   string // 字符串形式的用户 ID
		wantErrCode int    // 期望的错误码（0 表示无错误）
		wantSuccess bool   // 是否期望成功
	}{
		{
			name:        "成功：使用有效的用户 ID 字符串获取资料",
			userIDStr:   "1",
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：无效的用户 ID 字符串（非数字），应返回 ErrUserProfileNotFound",
			userIDStr:   "abc",
			wantErrCode: xerr.ErrUserProfileNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewUserService(db, accessHelper, refreshHelper, logger)
			resp, err := svc.GetUserProfileByID(ctx, tt.userIDStr)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, testUser.ID, resp.User.ID)
				assert.Equal(t, "profile_by_id_user", resp.User.Username)
			} else {
				assert.Error(t, err)
				assert.Nil(t, resp)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestUserService_ChangePassword 修改密码功能测试
// ============================================================================

// TestUserService_ChangePassword 测试用户修改密码功能
// 覆盖场景：成功修改密码、用户不存在、旧密码错误
func TestUserService_ChangePassword(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	// 预创建测试用户和刷新令牌
	testUser := CreateTestUser(t, db, "changepassword_user", "Password1")
	CreateTestRefreshToken(t, db, testUser.ID, "jti-token-to-revoke-1")
	CreateTestRefreshToken(t, db, testUser.ID, "jti-token-to-revoke-2")

	tests := []struct {
		name        string // 测试用例名称
		userID      int64  // 用户 ID
		oldPassword string // 旧密码
		newPassword string // 新密码
		setup       func() // 测试前置操作（可选）
		wantErrCode int    // 期望的错误码（0 表示无错误）
		wantSuccess bool   // 是否期望成功
	}{
		{
			name:        "成功：使用正确的旧密码修改密码",
			userID:      testUser.ID,
			oldPassword: "Password1",
			newPassword: "NewPassword1",
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：用户不存在，应返回 ErrChangePasswordFailed",
			userID:      99999,
			oldPassword: "Password1",
			newPassword: "NewPassword1",
			wantErrCode: xerr.ErrChangePasswordFailed.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：旧密码错误，应返回 ErrChangePasswordOldIncorrect",
			userID:      testUser.ID,
			oldPassword: "WrongPassword1",
			newPassword: "NewPassword1",
			wantErrCode: xerr.ErrChangePasswordOldIncorrect.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			svc := NewUserService(db, accessHelper, refreshHelper, logger)
			err := svc.ChangePassword(ctx, tt.userID, tt.oldPassword, tt.newPassword)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证密码已更新：使用新密码可以验证通过
				var updatedUser model.User
				findErr := db.Where("id = ?", tt.userID).First(&updatedUser).Error
				assert.NoError(t, findErr, "应能查询到用户")

				// 验证新密码可以验证通过
				bcryptErr := bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte(tt.newPassword))
				assert.NoError(t, bcryptErr, "新密码应能验证通过")

				// 验证旧密码无法验证通过
				bcryptErr = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte(tt.oldPassword))
				assert.Error(t, bcryptErr, "旧密码应无法验证通过")

				// 验证所有 RefreshToken 已被撤销
				var activeTokenCount int64
				db.Table("refresh_tokens").Where("user_id = ? AND revoked = ?", tt.userID, false).Count(&activeTokenCount)
				assert.Equal(t, int64(0), activeTokenCount, "修改密码后所有 Refresh Token 应被撤销")
			} else {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// TestUserService_ChangePassword_TxError 测试修改密码时数据库事务错误场景
// 通过关闭数据库连接来模拟查询阶段失败，验证 ErrChangePasswordFailed 路径
func TestUserService_ChangePassword_TxError(t *testing.T) {
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	// 创建测试用户
	testUser := CreateTestUser(t, db, "tx_error_user", "Password1")
	CreateTestRefreshToken(t, db, testUser.ID, "jti-tx-error")

	svc := NewUserService(db, accessHelper, refreshHelper, logger)

	// 关闭数据库连接，使 ChangePassword 的查询阶段报错
	sqlDB, err := db.DB()
	require.NoError(t, err, "获取底层 sql.DB 失败")
	err = sqlDB.Close()
	require.NoError(t, err, "关闭数据库连接失败")

	err = svc.ChangePassword(ctx, testUser.ID, "Password1", "NewPassword1")
	assert.Error(t, err)
	appErr, ok := err.(*xerr.AppError)
	assert.True(t, ok, "错误类型应为 *xerr.AppError")
	// 由于查询阶段就会失败，应该返回 ErrChangePasswordFailed
	assert.Equal(t, xerr.ErrChangePasswordFailed.Code, appErr.Code, "应返回密码修改失败错误码")
}
