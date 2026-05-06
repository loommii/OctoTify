package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"octotify/internal/handler/dto"
	pkgjwtx "octotify/pkg/jwtx"
	"octotify/pkg/xerr"
)

// ============================================================================
// TestAuthService_Login 登录功能测试
// ============================================================================

// TestAuthService_Login 测试用户登录功能
// 覆盖场景：正常登录、用户名错误、密码错误、数据库查询异常
func TestAuthService_Login(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	// 预创建测试用户
	testUser := CreateTestUser(t, db, "testuser", "Password123!")

	tests := []struct {
		name        string            // 测试用例名称
		req         *dto.LoginReq     // 登录请求参数
		setup       func()            // 测试前置操作（可选）
		wantErrCode int               // 期望的错误码（0 表示无错误）
		wantSuccess bool              // 是否期望成功
	}{
		{
			name: "成功：使用正确的用户名和密码登录",
			req: &dto.LoginReq{
				AuthCredentials: dto.AuthCredentials{
					Username: "testuser",
					Password: "Password123!",
				},
			},
			wantErrCode:   0,
			wantSuccess:   true,
		},
		{
			name: "失败：用户名不存在，应返回 ErrLoginInvalidCredentials",
			req: &dto.LoginReq{
				AuthCredentials: dto.AuthCredentials{
					Username: "nonexistent_user",
					Password: "Password123!",
				},
			},
			wantErrCode:   xerr.ErrLoginInvalidCredentials.Code,
			wantSuccess:   false,
		},
		{
			name: "失败：密码错误，应返回 ErrLoginInvalidCredentials",
			req: &dto.LoginReq{
				AuthCredentials: dto.AuthCredentials{
					Username: "testuser",
					Password: "WrongPassword!",
				},
			},
			wantErrCode:   xerr.ErrLoginInvalidCredentials.Code,
			wantSuccess:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			svc := NewAuthService(db, accessHelper, refreshHelper, logger)
			resp, err := svc.Login(ctx, tt.req)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken, "Access Token 不应为空")
				assert.NotEmpty(t, resp.RefreshToken, "Refresh Token 不应为空")
				assert.Equal(t, testUser.ID, resp.User.ID)
				assert.Equal(t, testUser.Username, resp.User.Username)
				assert.NotZero(t, resp.User.CreatedAt)
			} else {
				assert.Error(t, err)
				assert.Nil(t, resp)
				// 验证错误码
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestAuthService_Logout 退出登录功能测试
// ============================================================================

// TestAuthService_Logout 测试用户退出登录功能
// 覆盖场景：正常退出登录、数据库更新异常
func TestAuthService_Logout(t *testing.T) {
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	testUser := CreateTestUser(t, db, "logout_user", "Password123!")
	CreateTestRefreshToken(t, db, testUser.ID, "jti-token-1")
	CreateTestRefreshToken(t, db, testUser.ID, "jti-token-2")

	tests := []struct {
		name         string  // 测试用例名称
		userID       int64   // 用户 ID
		wantErrCode  int     // 期望的错误码（0 表示无错误）
		wantSuccess  bool    // 是否期望成功
	}{
		{
			name:         "成功：退出登录并撤销该用户所有 Refresh Token",
			userID:       testUser.ID,
			wantErrCode:  0,
			wantSuccess:  true,
		},
		{
			name:         "成功：退出登录时该用户没有任何 Refresh Token（零值场景）",
			userID:       99999,
			wantErrCode:  0,
			wantSuccess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAuthService(db, accessHelper, refreshHelper, logger)
			err := svc.Logout(ctx, tt.userID)

			if tt.wantSuccess {
				assert.NoError(t, err)
				// 验证该用户的所有 Refresh Token 均已被撤销
				var count int64
				db.Table("refresh_tokens").Where("user_id = ? AND revoked = ?", tt.userID, false).Count(&count)
				assert.Equal(t, int64(0), count, "该用户不应存在未撤销的 Refresh Token")
			} else {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestAuthService_RefreshAccessToken 刷新令牌功能测试
// ============================================================================

// TestAuthService_RefreshAccessToken 测试刷新 Access Token 功能
// 覆盖场景：正常刷新、无效 JWT Token、令牌类型错误、令牌不存在、令牌已撤销、用户不存在、生成新令牌失败
func TestAuthService_RefreshAccessToken(t *testing.T) {
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	accessHelper := SetupTestJWTHelper(t, pkgjwtx.Access)
	refreshHelper := SetupTestJWTHelper(t, pkgjwtx.Refresh)
	ctx := context.Background()

	testUser := CreateTestUser(t, db, "refresh_user", "Password123!")
	validJTI := "valid-jti-uuid"
	CreateTestRefreshToken(t, db, testUser.ID, validJTI)
	validRefreshToken := GenerateTestJWT(t, refreshHelper, "1", pkgjwtx.Refresh, validJTI)

	// 生成一个 access token 用于测试令牌类型错误
	wrongTypeToken := GenerateTestJWT(t, accessHelper, "1", pkgjwtx.Access, "wrong-type-jti")

	tests := []struct {
		name          string  // 测试用例名称
		refreshToken  string  // 刷新令牌字符串
		setup         func()  // 测试前置操作（可选）
		wantErrCode   int     // 期望的错误码（0 表示无错误）
		wantSuccess   bool    // 是否期望成功
	}{
		{
			name:         "成功：使用有效的 Refresh Token 刷新 Access Token",
			refreshToken: validRefreshToken,
			wantErrCode:  0,
			wantSuccess:  true,
		},
		{
			name:         "失败：无效的 JWT Token（格式错误或签名错误），应返回 ErrRefreshTokenInvalid",
			refreshToken: "this-is-not-a-valid-jwt-token",
			wantErrCode:  xerr.ErrRefreshTokenInvalid.Code,
			wantSuccess:  false,
		},
		{
			name:         "失败：Token 类型不是 refresh（传入的是 access token），应返回 ErrRefreshTokenInvalid",
			refreshToken: wrongTypeToken,
			wantErrCode:  xerr.ErrRefreshTokenInvalid.Code,
			wantSuccess:  false,
		},
		{
			name:         "失败：Token 在数据库中不存在（JTI 未注册），应返回 ErrRefreshTokenInvalid",
			refreshToken: GenerateTestJWT(t, refreshHelper, "1", pkgjwtx.Refresh, "nonexistent-jti"),
			wantErrCode:  xerr.ErrRefreshTokenInvalid.Code,
			wantSuccess:  false,
		},
		{
			name:         "失败：Token 已被撤销，应返回 ErrRefreshTokenRevoked",
			refreshToken: GenerateTestJWT(t, refreshHelper, "1", pkgjwtx.Refresh, "revoked-jti"),
			setup: func() {
				// 创建已撤销的令牌记录
				CreateTestRefreshToken(t, db, testUser.ID, "revoked-jti")
				db.Table("refresh_tokens").Where("jti = ?", "revoked-jti").Update("revoked", true)
			},
			wantErrCode:  xerr.ErrRefreshTokenRevoked.Code,
			wantSuccess:  false,
		},
		{
			name:         "失败：Token 对应的用户不存在，应返回 ErrRefreshTokenInvalid",
			refreshToken: GenerateTestJWT(t, refreshHelper, "99999", pkgjwtx.Refresh, "orphan-jti"),
			setup: func() {
				// 创建令牌记录但对应的用户不存在
				CreateTestRefreshToken(t, db, 99999, "orphan-jti")
			},
			wantErrCode:  xerr.ErrRefreshTokenInvalid.Code,
			wantSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			svc := NewAuthService(db, accessHelper, refreshHelper, logger)
			resp, err := svc.RefreshAccessToken(ctx, tt.refreshToken)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken, "新的 Access Token 不应为空")
				assert.NotEmpty(t, resp.RefreshToken, "新的 Refresh Token 不应为空")
				assert.Equal(t, testUser.ID, resp.User.ID)
				assert.Equal(t, testUser.Username, resp.User.Username)
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
