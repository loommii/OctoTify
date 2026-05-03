package service

import (
	"context"
	"regexp"
	"strconv"
	"unicode/utf8"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	pkgjwtx "octotify/pkg/jwtx"
	"octotify/pkg/xerr"

	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/internal/query"
	"octotify/internal/repository"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type AuthService struct {
	db        *gorm.DB
	userRepo  *repository.UserRepository
	tokenRepo *repository.RefreshTokenRepository
	jwtHelper *pkgjwtx.JWTHelper
	logger    *zap.Logger
}

func NewAuthService(
	db *gorm.DB,
	userRepo *repository.UserRepository,
	tokenRepo *repository.RefreshTokenRepository,
	jwtHelper *pkgjwtx.JWTHelper,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		db:        db,
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtHelper: jwtHelper,
		logger:    logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterReq) (*dto.AuthResp, error) {
	// 校验注册请求参数
	if err := s.validateRegister(req); err != nil {
		return nil, err
	}

	// 检查用户名是否已存在
	existing, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Error("query user by username failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}
	if existing != nil {
		return nil, xerr.ErrRegisterUsernameExists
	}

	// 密码哈希加密
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("bcrypt hash failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
	}

	// 事务中创建用户
	q := query.Use(s.db)
	err = q.Transaction(func(tx *query.Query) error {
		if err := tx.User.Create(user); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.logger.Error("create user failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}

	// 生成 JWT 令牌对
	tokenPair, err := s.jwtHelper.GenerateTokenPair(strconv.FormatInt(user.ID, 10))
	if err != nil {
		s.logger.Error("generate token pair failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}

	// 解析 Refresh Token 以获取 JTI 和过期时间
	_, refreshClaims, err := s.jwtHelper.ValidateToken(tokenPair.RefreshToken)
	if err != nil {
		s.logger.Error("parse refresh token failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}

	// 保存 Refresh Token 记录到数据库
	refreshToken := &model.RefreshToken{
		JTI:       refreshClaims.ID,
		UserID:    user.ID,
		Revoked:   false,
		ExpiresAt: refreshClaims.ExpiresAt.Time,
	}

	err = q.Transaction(func(tx *query.Query) error {
		if err := tx.RefreshToken.Create(refreshToken); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.logger.Error("create refresh token failed", zap.Error(err))
		return nil, xerr.ErrRefreshTokenFailed.WithInternal(err)
	}

	s.logger.Info("user registered",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
	)

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

func (s *AuthService) validateRegister(req *dto.RegisterReq) error {
	if req.Username == "" {
		return xerr.ErrRegisterUsernameEmpty
	}
	if req.Password == "" {
		return xerr.ErrRegisterPasswordEmpty
	}
	if utf8.RuneCountInString(req.Username) < 3 || utf8.RuneCountInString(req.Username) > 64 {
		return xerr.ErrRegisterUsernameInvalid
	}
	if !usernameRegex.MatchString(req.Username) {
		return xerr.ErrRegisterUsernameInvalid
	}
	if utf8.RuneCountInString(req.Password) < 6 || utf8.RuneCountInString(req.Password) > 128 {
		return xerr.ErrRegisterPasswordInvalid
	}
	return nil
}
