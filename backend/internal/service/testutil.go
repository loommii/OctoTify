// Package testutil 提供测试辅助函数
package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"octotify/internal/model"
	"octotify/internal/sender"
	pkgjwtx "octotify/pkg/jwtx"
)

// SetupTestDB 创建 SQLite 内存数据库并自动迁移所有表
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "创建测试数据库失败")

	// 限制为单连接，避免并发测试时 SQLite :memory: 数据库出现多连接不一致的问题
	sqlDB, err := db.DB()
	require.NoError(t, err, "获取底层 sql.DB 失败")
	sqlDB.SetMaxOpenConns(1)

	// 迁移所有模型表
	err = db.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
		&model.Source{},
		&model.SourceChannel{},
		&model.Channel{},
		&model.Message{},
	)
	require.NoError(t, err, "数据库迁移失败")

	return db
}

// SetupTestLogger 创建 Noop 日志记录器（测试时不输出日志）
func SetupTestLogger(t *testing.T) *zap.Logger {
	t.Helper()

	// 创建无操作的 logger，避免测试时输出大量日志
	logger, _ := zap.NewDevelopment(zap.WithFatalHook(nil))
	return logger
}

// SetupTestJWTHelper 创建真实的 JWT 辅助工具（使用测试密钥）
func SetupTestJWTHelper(t *testing.T, tokenType string) *pkgjwtx.JWTHelper {
	t.Helper()

	// 使用测试密钥（HMAC-SHA256）
	var ttl time.Duration
	if tokenType == pkgjwtx.Access {
		ttl = 15 * time.Minute
	} else {
		ttl = 7 * 24 * time.Hour
	}

	helper := pkgjwtx.NewJWTHelper(
		pkgjwtx.WithKeyType(pkgjwtx.HMAC),
		pkgjwtx.WithSecret([]byte("test-secret-key-for-unit-testing")),
		pkgjwtx.WithTTL(ttl),
		pkgjwtx.WithIssuer("octotify-test"),
		pkgjwtx.WithTokenType(tokenType),
	)

	return helper
}

// CreateTestUser 创建测试用户并返回用户对象和密码明文
func CreateTestUser(t *testing.T, db *gorm.DB, username string, password string) *model.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "生成密码哈希失败")

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
	}

	err = db.Create(user).Error
	require.NoError(t, err, "创建测试用户失败")

	return user
}

// CreateTestSource 创建测试来源
func CreateTestSource(t *testing.T, db *gorm.DB, userID int64, name string, token string) *model.Source {
	t.Helper()

	if token == "" {
		token = "src" + uuid.New().String()
	}

	source := &model.Source{
		UserID:      userID,
		Name:        name,
		Description: "test source description",
		Token:       token,
		Status:      model.SourceStatusActive,
	}

	err := db.Create(source).Error
	require.NoError(t, err, "创建测试来源失败")

	return source
}

// CreateTestChannel 创建测试渠道
func CreateTestChannel(t *testing.T, db *gorm.DB, userID int64, channelType string, name string) *model.Channel {
	t.Helper()

	channel := &model.Channel{
		UserID: userID,
		Type:   channelType,
		Name:   name,
		Config: datatypes.JSON(`{"webhook_url":"https://example.com/webhook"}`),
		Status: model.ChannelStatusActive,
	}

	err := db.Create(channel).Error
	require.NoError(t, err, "创建测试渠道失败")

	return channel
}

// CreateTestRefreshToken 创建测试刷新令牌
func CreateTestRefreshToken(t *testing.T, db *gorm.DB, userID int64, jti string) *model.RefreshToken {
	t.Helper()

	token := &model.RefreshToken{
		JTI:     jti,
		UserID:  userID,
		Revoked: false,
	}

	err := db.Create(token).Error
	require.NoError(t, err, "创建测试刷新令牌失败")

	return token
}

// BindSourceToChannel 绑定来源到渠道
func BindSourceToChannel(t *testing.T, db *gorm.DB, sourceID int64, channelID int64) {
	t.Helper()

	sourceChannel := &model.SourceChannel{
		SourceID:  sourceID,
		ChannelID: channelID,
		Status:    model.SourceChannelStatusActive,
	}

	err := db.Create(sourceChannel).Error
	require.NoError(t, err, "绑定来源到渠道失败")
}

// GenerateTestJWT 生成测试用的 JWT Token
func GenerateTestJWT(t *testing.T, helper *pkgjwtx.JWTHelper, uid string, tokenType string, jti string) string {
	t.Helper()

	claims := pkgjwtx.JWTClaims{
		UID:       uid,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: jti,
		},
	}

	token, err := helper.GenerateToken(claims)
	require.NoError(t, err, "生成 JWT Token 失败")

	return token
}

// AssertJWTTokenValid 验证 JWT Token 是否有效并返回 Claims
func AssertJWTTokenValid(t *testing.T, helper *pkgjwtx.JWTHelper, tokenStr string) *pkgjwtx.JWTClaims {
	t.Helper()

	_, claims, err := helper.ValidateToken(tokenStr)
	require.NoError(t, err, "JWT Token 验证失败")
	require.NotNil(t, claims, "JWT Claims 不应为 nil")

	return claims
}

// NewChannelServiceForTest 创建测试用的 ChannelService
func NewChannelServiceForTest(db *gorm.DB, logger *zap.Logger, senderFactory *sender.SenderFactory) *ChannelService {
	return &ChannelService{
		db:            db,
		logger:        logger,
		senderFactory: senderFactory,
	}
}
