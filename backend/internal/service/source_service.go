package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/internal/query"
	"octotify/pkg/xerr"
)

// SourceService 消息来源业务逻辑
type SourceService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSourceService 创建 SourceService 实例
func NewSourceService(db *gorm.DB, logger *zap.Logger) *SourceService {
	return &SourceService{
		db:     db,
		logger: logger,
	}
}

// CreateSource 创建消息来源，生成唯一 Token 并写入数据库
func (s *SourceService) CreateSource(ctx context.Context, userID int64, req *dto.CreateSourceReq) (*dto.SourceDTO, error) {
	q := query.Use(s.db)

	token, err := s.generateUniqueToken(q, ctx)
	if err != nil {
		return nil, err
	}

	source := &model.Source{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Token:       token,
		Status:      model.SourceStatusActive,
	}

	err = q.Transaction(func(tx *query.Query) error {
		if err := tx.Source.WithContext(ctx).Create(source); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.logger.Error("create source failed",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.String("name", req.Name),
		)
		return nil, xerr.ErrSourceInsertFailed.WithInternal(err)
	}

	s.logger.Info("source created",
		zap.Int64("source_id", source.ID),
		zap.String("name", source.Name),
	)

	return &dto.SourceDTO{
		ID:          source.ID,
		UserID:      source.UserID,
		Name:        source.Name,
		Token:       source.Token,
		Description: source.Description,
		Status:      source.Status,
		CreatedAt:   source.CreatedAt.UnixMilli(),
	}, nil
}

// generateUniqueToken 生成唯一的 Source Token（前缀 src + UUIDv4 无连字符，共 35 位）
// 最多重试 3 次，防止极端碰撞情况
func (s *SourceService) generateUniqueToken(q *query.Query, ctx context.Context) (string, error) {
	for i := 0; i < 3; i++ {
		u, err := uuid.NewV7()
		if err != nil {
			u = uuid.New()
		}
		token := "src" + strings.ReplaceAll(u.String(), "-", "")

		count, err := q.Source.WithContext(ctx).Where(q.Source.Token.Eq(token)).Count()
		if err != nil {
			s.logger.Error("check token uniqueness failed", zap.Error(err))
			return "", xerr.ErrSourceTokenFailed.WithInternal(err)
		}
		if count == 0 {
			return token, nil
		}
	}

	return "", xerr.ErrSourceTokenFailed.WithInternal(fmt.Errorf("failed to generate unique token after 3 attempts"))
}
