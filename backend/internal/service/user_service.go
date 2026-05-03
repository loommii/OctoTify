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
	pkgjwtx "octotify/pkg/jwtx"
	"octotify/pkg/xerr"
)

type UserService struct {
	db               *gorm.DB
	accessJWTHelper  *pkgjwtx.JWTHelper
	refreshJWTHelper *pkgjwtx.JWTHelper
	logger           *zap.Logger
}

func NewUserService(
	db *gorm.DB,
	accessJWTHelper *pkgjwtx.JWTHelper,
	refreshJWTHelper *pkgjwtx.JWTHelper,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		db:               db,
		accessJWTHelper:  accessJWTHelper,
		refreshJWTHelper: refreshJWTHelper,
		logger:           logger,
	}
}

func (s *UserService) GetUserProfile(ctx context.Context, userID int64) (*dto.UserDTO, error) {
	q := query.Use(s.db)

	user, err := q.User.WithContext(ctx).Where(q.User.ID.Eq(userID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrUserProfileNotFound
		}
		s.logger.Error("query user by id failed", zap.Error(err))
		return nil, xerr.ErrUserProfileQueryFailed.WithInternal(err)
	}

	return &dto.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.UnixMilli(),
	}, nil
}

func (s *UserService) GetUserProfileByID(ctx context.Context, userIDStr string) (*dto.UserDTO, error) {
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		s.logger.Error("parse user id failed", zap.String("user_id", userIDStr), zap.Error(err))
		return nil, xerr.ErrUserProfileNotFound
	}

	return s.GetUserProfile(ctx, userID)
}

func (s *UserService) Register(ctx context.Context, req *dto.RegisterReq) (*dto.AuthResp, error) {
	q := query.Use(s.db)

	existing, err := q.WithContext(ctx).User.Where(q.User.Username.Eq(req.Username)).First()
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Error("query user by username failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}
	if existing != nil {
		return nil, xerr.ErrRegisterUsernameExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("bcrypt hash failed", zap.Error(err))
		return nil, xerr.ErrRegisterFailed.WithInternal(err)
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
	}

	var accessToken string
	var refreshToken string
	var refreshJTI string

	err = q.Transaction(func(tx *query.Query) error {
		if err := tx.User.Create(user); err != nil {
			return err
		}

		uid := strconv.FormatInt(user.ID, 10)

		accessToken, err = s.accessJWTHelper.GenerateToken(pkgjwtx.JWTClaims{
			UID:       uid,
			TokenType: pkgjwtx.Access,
		})
		if err != nil {
			s.logger.Error("generate access token failed", zap.Error(err))
			return err
		}

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

func (s *UserService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	q := query.Use(s.db)

	user, err := q.User.WithContext(ctx).Where(q.User.ID.Eq(userID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrChangePasswordFailed
		}
		s.logger.Error("query user by id failed", zap.Error(err))
		return xerr.ErrChangePasswordFailed.WithInternal(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return xerr.ErrChangePasswordOldIncorrect
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("bcrypt hash new password failed", zap.Error(err))
		return xerr.ErrChangePasswordFailed.WithInternal(err)
	}

	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.User.WithContext(ctx).Where(tx.User.ID.Eq(userID)).Update(tx.User.PasswordHash, string(newHash))
		if err != nil {
			s.logger.Error("update password failed", zap.Error(err))
			return err
		}

		_, err = tx.RefreshToken.WithContext(ctx).Where(tx.RefreshToken.UserID.Eq(userID)).Update(tx.RefreshToken.Revoked, true)
		if err != nil {
			s.logger.Error("revoke all refresh tokens failed", zap.Error(err))
			return err
		}

		return nil
	})
	if err != nil {
		s.logger.Error("change password transaction failed", zap.Error(err))
		return xerr.ErrChangePasswordFailed.WithInternal(err)
	}

	s.logger.Info("user password changed",
		zap.Int64("user_id", userID),
		zap.String("username", user.Username),
	)

	return nil
}
