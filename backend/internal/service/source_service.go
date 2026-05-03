package service

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	// 生成唯一的 Source Token
	token, err := s.generateUniqueToken(q, ctx)
	if err != nil {
		return nil, err
	}

	// 构建 Source 实体
	source := &model.Source{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Token:       token,
		Status:      model.SourceStatusActive,
	}

	// 在事务中创建来源记录
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

	// 返回 DTO 响应数据
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
		// 优先使用 UUID v7（时间有序），降级到 UUID v4
		u, err := uuid.NewV7()
		if err != nil {
			u = uuid.New()
		}
		token := "src" + strings.ReplaceAll(u.String(), "-", "")

		// 查询数据库校验 Token 唯一性
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

// UpdateSource 编辑消息来源
func (s *SourceService) UpdateSource(ctx context.Context, sourceID int64, userID int64, req *dto.UpdateSourceReq) error {
	q := query.Use(s.db)

	// 查询来源，通过 id 和 user_id 联合查询确保权限隔离
	source, err := q.Source.WithContext(ctx).Where(
		q.Source.ID.Eq(sourceID),
		q.Source.UserID.Eq(userID),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("来源不存在",
				zap.Int64("source_id", sourceID),
				zap.Int64("user_id", userID),
			)
			return xerr.ErrSourceNotFound
		}
		s.logger.Error("查询来源失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return xerr.ErrSourceQueryFailed.WithInternal(err)
	}

	// 检查来源是否已删除
	if source.Status == model.SourceStatusDeleted {
		s.logger.Error("来源已删除",
			zap.Int64("source_id", sourceID),
		)
		return xerr.ErrSourceAlreadyDeleted
	}

	// 在事务中更新来源
	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Source.WithContext(ctx).
			Where(tx.Source.ID.Eq(sourceID), tx.Source.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"name":        req.Name,
				"description": req.Description,
				"updated_at":  time.Now(),
			})
		return err
	})
	if err != nil {
		s.logger.Error("更新来源失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return xerr.ErrSourceUpdateFailed.WithInternal(err)
	}

	s.logger.Info("来源更新成功",
		zap.Int64("source_id", sourceID),
		zap.String("name", req.Name),
	)

	return nil
}
