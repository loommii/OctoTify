package service

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	pkgjwtx "octotify/pkg/jwtx"
	"octotify/pkg/xerr"

	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/internal/query"
)

// AuthService 认证服务，处理用户注册、登录、令牌刷新等认证相关业务逻辑
type AuthService struct {
	db        *gorm.DB           // 数据库连接
	jwtHelper *pkgjwtx.JWTHelper // JWT 令牌辅助工具
	logger    *zap.Logger        // 日志记录器
}

// NewAuthService 创建认证服务实例
func NewAuthService(
	db *gorm.DB,
	jwtHelper *pkgjwtx.JWTHelper,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		db:        db,
		jwtHelper: jwtHelper,
		logger:    logger,
	}
}

// Register 用户注册
// 流程：检查用户名是否已存在 -> 密码哈希 -> 创建用户 -> 生成令牌对 -> 保存刷新令牌
func (s *AuthService) Register(ctx context.Context, req *dto.RegisterReq) (*dto.AuthResp, error) {
	q := query.Use(s.db)
	// 检查用户名是否已存在
	existing, err := q.WithContext(ctx).User.Where(q.User.Username.Eq(req.Username)).First()
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Error("query user by username failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}
	if existing != nil {
		return nil, xerr.ErrRegisterUsernameExists
	}

	// 对密码进行 bcrypt 哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("bcrypt hash failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}

	// 构建用户模型
	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
	}

	var tokenPair *pkgjwtx.TokenPair
	// 使用事务确保数据一致性
	err = q.Transaction(func(tx *query.Query) error {
		// 创建用户记录（执行后 user.ID 会被自动填充）
		if err := tx.User.Create(user); err != nil {
			return err
		}

		// 生成访问令牌和刷新令牌（依赖 user.ID，必须在用户创建后执行）
		tokenPair, err = s.jwtHelper.GenerateTokenPair(strconv.FormatInt(user.ID, 10))
		if err != nil {
			s.logger.Error("generate token pair failed", zap.Error(err))
			return err
		}

		// 保存刷新令牌到数据库
		refreshToken := &model.RefreshToken{
			JTI:       tokenPair.RefreshClaims.ID,
			UserID:    user.ID,
			Revoked:   false,
			ExpiresAt: tokenPair.RefreshClaims.ExpiresAt.Time,
		}
		if err := tx.RefreshToken.Create(refreshToken); err != nil {
			return err
		}

		s.logger.Info("user registered",
			zap.Int64("user_id", user.ID),
			zap.String("username", user.Username),
		)

		return nil
	})
	if err != nil {
		s.logger.Error("create user or refresh token failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}

	// 返回认证响应
	return &dto.AuthResp{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User: dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.UnixMilli(),
		},
	}, nil
}
