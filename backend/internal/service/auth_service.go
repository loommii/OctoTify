package service

import (
	"context"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/internal/query"
	"octotify/pkg/ctxutil"
	pkgjwtx "octotify/pkg/jwtx"
	"octotify/pkg/xerr"
)

// AuthService 认证服务，处理用户登录、令牌刷新等认证相关业务逻辑
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

func (s *AuthService) log(ctx context.Context) *zap.Logger {
	if rid := ctxutil.GetRequestID(ctx); rid != "" {
		return s.logger.With(zap.String("request_id", rid))
	}
	return s.logger
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
		s.log(ctx).Error("根据用户名查询用户失败", zap.Error(err))
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
		s.log(ctx).Error("生成 Access Token 失败", zap.Error(err))
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
		s.log(ctx).Error("生成 Refresh Token 失败", zap.Error(err))
		return nil, xerr.ErrLoginFailed.WithInternal(err)
	}

	// 保存刷新令牌
	refreshTokenRecord := &model.RefreshToken{
		JTI:    refreshJTI,
		UserID: user.ID,
	}
	if err := q.RefreshToken.WithContext(ctx).Create(refreshTokenRecord); err != nil {
		s.log(ctx).Error("保存 Refresh Token 失败", zap.Error(err))
		return nil, xerr.ErrLoginFailed.WithInternal(err)
	}

	s.log(ctx).Info("用户登录成功",
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

// Logout 退出登录
// 流程：撤销该用户所有 Refresh Token
func (s *AuthService) Logout(ctx context.Context, userID int64) error {
	q := query.Use(s.db)

	_, err := q.RefreshToken.WithContext(ctx).Where(q.RefreshToken.UserID.Eq(userID)).Update(q.RefreshToken.Revoked, model.RefreshTokenRevoked)
	if err != nil {
		s.log(ctx).Error("撤销所有 Refresh Token 失败", zap.Error(err))
		return xerr.ErrLogoutFailed.WithInternal(err)
	}

	s.log(ctx).Info("用户已退出登录",
		zap.Int64("user_id", userID),
	)

	return nil
}

// RefreshAccessToken 刷新访问令牌
// 流程：验证JWT令牌 -> 查询数据库记录 -> 检查撤销状态 -> 生成新 Access Token -> 重新签发 Refresh Token（相同 JTI）
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshTokenStr string) (*dto.AuthResp, error) {
	q := query.Use(s.db)

	// 验证 JWT 令牌
	_, claims, err := s.refreshJWTHelper.ValidateToken(refreshTokenStr)
	if err != nil {
		s.log(ctx).Error("验证 Refresh Token 失败", zap.Error(err))
		return nil, xerr.ErrRefreshTokenInvalid
	}

	// 检查令牌类型是否为 refresh
	if !claims.IsRefreshToken() {
		s.log(ctx).Error("令牌类型不是 Refresh Token", zap.String("token_type", claims.TokenType))
		return nil, xerr.ErrRefreshTokenInvalid
	}

	// 查询数据库中的刷新令牌记录
	rt, err := q.RefreshToken.WithContext(ctx).Where(q.RefreshToken.JTI.Eq(claims.ID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrRefreshTokenInvalid
		}
		s.log(ctx).Error("查询 Refresh Token 失败", zap.Error(err))
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
		s.log(ctx).Error("根据用户 ID 查询用户失败", zap.Error(err))
		return nil, xerr.ErrRefreshTokenFailed.WithInternal(err)
	}

	uid := strconv.FormatInt(user.ID, 10)

	// 生成新的 Access Token
	newAccessToken, err := s.accessJWTHelper.GenerateToken(pkgjwtx.JWTClaims{
		UID:       uid,
		TokenType: pkgjwtx.Access,
	})
	if err != nil {
		s.log(ctx).Error("生成新 Access Token 失败", zap.Error(err))
		return nil, xerr.ErrRefreshTokenFailed.WithInternal(err)
	}

	// 重新签发 Refresh Token（使用相同的 JTI，刷新过期时间，无需更新数据库）
	newRefreshToken, err := s.refreshJWTHelper.GenerateToken(pkgjwtx.JWTClaims{
		UID:       uid,
		TokenType: pkgjwtx.Refresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: claims.ID,
		},
	})
	if err != nil {
		s.log(ctx).Error("生成新 Refresh Token 失败", zap.Error(err))
		return nil, xerr.ErrRefreshTokenFailed.WithInternal(err)
	}

	s.log(ctx).Info("Access Token 刷新成功",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	// 返回新的认证响应（重新签发 Refresh Token）
	return &dto.AuthResp{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		User: dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.UnixMilli(),
		},
	}, nil
}
