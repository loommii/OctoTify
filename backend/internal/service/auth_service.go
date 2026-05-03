package service

import (
	"context"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	db               *gorm.DB           // 数据库连接
	accessJWTHelper  *pkgjwtx.JWTHelper // Access Token 令牌辅助工具
	refreshJWTHelper *pkgjwtx.JWTHelper // Refresh Token 令牌辅助工具
	logger           *zap.Logger        // 日志记录器
}

// NewAuthService 创建认证服务实例
func NewAuthService(
	db *gorm.DB,
	accessJWTHelper *pkgjwtx.JWTHelper,
	refreshJWTHelper *pkgjwtx.JWTHelper,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		db:               db,
		accessJWTHelper:  accessJWTHelper,
		refreshJWTHelper: refreshJWTHelper,
		logger:           logger,
	}
}

// Login 用户登录
// 流程：查询用户 -> 验证密码 -> 生成令牌对 -> 保存刷新令牌
func (s *AuthService) Login(ctx context.Context, req *dto.LoginReq) (*dto.AuthResp, error) {
	q := query.Use(s.db)

	// 查询用户
	user, err := q.WithContext(ctx).User.Where(q.User.Username.Eq(req.Username)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrLoginInvalidCredentials
		}
		s.logger.Error("query user by username failed", zap.Error(err))
		return nil, xerr.ErrLoginFailed.WithInternal(err)
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, xerr.ErrLoginInvalidCredentials
	}

	uid := strconv.FormatInt(user.ID, 10)

	// 生成 Access Token
	accessToken, err := s.accessJWTHelper.GenerateToken(pkgjwtx.JWTClaims{
		UID:       uid,
		TokenType: pkgjwtx.Access,
	})
	if err != nil {
		s.logger.Error("generate access token failed", zap.Error(err))
		return nil, xerr.ErrLoginFailed.WithInternal(err)
	}

	// 生成 Refresh Token（预分配 JTI 用于数据库存储）
	refreshJTI := uuid.New().String()
	refreshToken, err := s.refreshJWTHelper.GenerateToken(pkgjwtx.JWTClaims{
		UID:       uid,
		TokenType: pkgjwtx.Refresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: refreshJTI,
		},
	})
	if err != nil {
		s.logger.Error("generate refresh token failed", zap.Error(err))
		return nil, xerr.ErrLoginFailed.WithInternal(err)
	}

	// 保存刷新令牌
	refreshTokenRecord := &model.RefreshToken{
		JTI:    refreshJTI,
		UserID: user.ID,
	}
	if err := q.RefreshToken.WithContext(ctx).Create(refreshTokenRecord); err != nil {
		s.logger.Error("create refresh token failed", zap.Error(err))
		return nil, xerr.ErrLoginFailed.WithInternal(err)
	}

	s.logger.Info("user logged in",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	// 返回认证响应
	return &dto.AuthResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.UnixMilli(),
		},
	}, nil
}

// RefreshAccessToken 刷新访问令牌
// 流程：验证JWT令牌 -> 查询数据库记录 -> 检查撤销状态 -> 生成新 Access Token -> 复用原 Refresh Token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (*dto.AuthResp, error) {
	q := query.Use(s.db)

	// 验证 JWT 令牌
	_, claims, err := s.refreshJWTHelper.ValidateToken(refreshTokenStr)
	if err != nil {
		s.logger.Error("validate refresh token failed", zap.Error(err))
		return nil, xerr.ErrRefreshTokenInvalid
	}

	// 检查令牌类型是否为 refresh
	if !claims.IsRefreshToken() {
		s.logger.Error("token type is not refresh", zap.String("token_type", claims.TokenType))
		return nil, xerr.ErrRefreshTokenInvalid
	}

	// 查询数据库中的刷新令牌记录
	rt, err := q.RefreshToken.WithContext(ctx).Where(q.RefreshToken.JTI.Eq(claims.ID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrRefreshTokenInvalid
		}
		s.logger.Error("query refresh token failed", zap.Error(err))
		return nil, xerr.ErrRefreshTokenFailed.WithInternal(err)
	}

	// 检查令牌是否已撤销
	if rt.Revoked {
		return nil, xerr.ErrRefreshTokenRevoked
	}

	// 查询用户信息
	user, err := q.User.WithContext(ctx).Where(q.User.ID.Eq(rt.UserID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrRefreshTokenInvalid
		}
		s.logger.Error("query user by id failed", zap.Error(err))
		return nil, xerr.ErrRefreshTokenFailed.WithInternal(err)
	}

	uid := strconv.FormatInt(user.ID, 10)

	// 生成新的 Access Token
	newAccessToken, err := s.accessJWTHelper.GenerateToken(pkgjwtx.JWTClaims{
		UID:       uid,
		TokenType: pkgjwtx.Access,
	})
	if err != nil {
		s.logger.Error("generate new access token failed", zap.Error(err))
		return nil, xerr.ErrRefreshTokenFailed.WithInternal(err)
	}

	s.logger.Info("access token refreshed",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	// 返回新的认证响应（复用原 Refresh Token）
	return &dto.AuthResp{
		AccessToken:  newAccessToken,
		RefreshToken: refreshTokenStr,
		User: dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.UnixMilli(),
		},
	}, nil
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

	var accessToken string
	var refreshToken string
	var refreshJTI string

	// 使用事务确保数据一致性
	err = q.Transaction(func(tx *query.Query) error {
		// 创建用户记录（执行后 user.ID 会被自动填充）
		if err := tx.User.Create(user); err != nil {
			return err
		}

		uid := strconv.FormatInt(user.ID, 10)

		// 生成 Access Token
		accessToken, err = s.accessJWTHelper.GenerateToken(pkgjwtx.JWTClaims{
			UID:       uid,
			TokenType: pkgjwtx.Access,
		})
		if err != nil {
			s.logger.Error("generate access token failed", zap.Error(err))
			return err
		}

		// 生成 Refresh Token（预分配 JTI）
		refreshJTI = uuid.New().String()
		refreshToken, err = s.refreshJWTHelper.GenerateToken(pkgjwtx.JWTClaims{
			UID:       uid,
			TokenType: pkgjwtx.Refresh,
			RegisteredClaims: jwt.RegisteredClaims{
				ID: refreshJTI,
			},
		})
		if err != nil {
			s.logger.Error("generate refresh token failed", zap.Error(err))
			return err
		}

		// 保存刷新令牌到数据库
		refreshTokenRecord := &model.RefreshToken{
			JTI:    refreshJTI,
			UserID: user.ID,
		}
		if err := tx.RefreshToken.Create(refreshTokenRecord); err != nil {
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
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.UnixMilli(),
		},
	}, nil
}
