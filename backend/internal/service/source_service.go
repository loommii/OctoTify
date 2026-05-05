package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

// CreateSource 创建消息来源，生成唯一 Token 并绑定渠道
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

	// 在事务中创建来源记录并绑定渠道
	err = q.Transaction(func(tx *query.Query) error {
		if err := tx.Source.WithContext(ctx).Create(source); err != nil {
			return err
		}

		// 绑定渠道：批量插入 source_channels 关联记录
		if len(req.ChannelIDs) > 0 {
			sourceChannels := make([]*model.SourceChannel, 0, len(req.ChannelIDs))
			now := time.Now()
			for _, channelID := range req.ChannelIDs {
				sourceChannels = append(sourceChannels, &model.SourceChannel{
					SourceID:  source.ID,
					ChannelID: channelID,
					Status:    model.SourceChannelStatusActive,
					CreatedAt: now,
				})
			}
			if err := tx.SourceChannel.WithContext(ctx).CreateInBatches(sourceChannels, len(sourceChannels)); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		s.logger.Error("创建来源失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.String("name", req.Name),
		)
		return nil, xerr.ErrSourceInsertFailed.WithInternal(err)
	}

	s.logger.Info("来源创建成功",
		zap.Int64("source_id", source.ID),
		zap.String("name", source.Name),
		zap.Int("channel_count", len(req.ChannelIDs)),
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

// ListSources 分页查询消息来源列表
func (s *SourceService) ListSources(ctx context.Context, userID int64, pageReq *dto.PageReq) ([]*dto.SourceDTO, int64, error) {
	q := query.Use(s.db)

	pageReq.Normalize()
	offset := (pageReq.Page - 1) * pageReq.PageSize

	// 查询总数
	total, err := q.Source.WithContext(ctx).
		Where(
			q.Source.UserID.Eq(userID),
			q.Source.Status.Neq(model.SourceStatusDeleted),
		).
		Count()
	if err != nil {
		s.logger.Error("查询来源总数失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, 0, xerr.ErrSourceQueryFailed.WithInternal(err)
	}

	// 查询列表，按创建时间降序（最新优先）
	sources, err := q.Source.WithContext(ctx).
		Where(
			q.Source.UserID.Eq(userID),
			q.Source.Status.Neq(model.SourceStatusDeleted),
		).
		Order(q.Source.CreatedAt.Desc()).
		Offset(offset).
		Limit(pageReq.PageSize).
		Find()
	if err != nil {
		s.logger.Error("查询来源列表失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, 0, xerr.ErrSourceQueryFailed.WithInternal(err)
	}

	// 转换为 DTO
	list := make([]*dto.SourceDTO, 0, len(sources))
	for _, src := range sources {
		list = append(list, &dto.SourceDTO{
			ID:          src.ID,
			UserID:      src.UserID,
			Name:        src.Name,
			Description: src.Description,
			Status:      src.Status,
			CreatedAt:   src.CreatedAt.UnixMilli(),
		})
	}

	s.logger.Info("查询来源列表成功",
		zap.Int64("user_id", userID),
		zap.Int("page", pageReq.Page),
		zap.Int("page_size", pageReq.PageSize),
		zap.Int64("total", total),
	)

	return list, total, nil
}

// generateUniqueToken 生成唯一的 Source Token（前缀 src + UUID 无连字符，共 35 位）
// 优先使用 UUID v7（时间有序），降级到 UUID v4
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
			s.logger.Error("校验 Token 唯一性失败", zap.Error(err))
			return "", xerr.ErrSourceTokenFailed.WithInternal(err)
		}
		if count == 0 {
			return token, nil
		}
	}

	return "", xerr.ErrSourceTokenFailed.WithInternal(fmt.Errorf("生成唯一 Token 失败，已达最大重试次数"))
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

// GetSourceDetail 查询消息来源详情（包含绑定的有效渠道列表）
func (s *SourceService) GetSourceDetail(ctx context.Context, sourceID int64, userID int64) (*dto.SourceDetailResponse, error) {
	q := query.Use(s.db)

	// 查询来源，通过 id 和 user_id 联合查询确保权限隔离，过滤已删除记录
	source, err := q.Source.WithContext(ctx).
		Where(
			q.Source.ID.Eq(sourceID),
			q.Source.UserID.Eq(userID),
			q.Source.Status.Neq(model.SourceStatusDeleted),
		).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("来源不存在",
				zap.Int64("source_id", sourceID),
				zap.Int64("user_id", userID),
			)
			return nil, xerr.ErrSourceNotFound
		}
		s.logger.Error("查询来源详情失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return nil, xerr.ErrSourceQueryFailed.WithInternal(err)
	}

	// 查询绑定的有效渠道（通过 source_channels 关联表）
	channels, err := q.Channel.WithContext(ctx).
		Select(q.Channel.ID, q.Channel.UserID, q.Channel.Type, q.Channel.Name, q.Channel.Status, q.Channel.CreatedAt).
		Join(q.SourceChannel, q.SourceChannel.ChannelID.EqCol(q.Channel.ID)).
		Where(
			q.SourceChannel.SourceID.Eq(sourceID),
			q.Channel.Status.Neq(model.ChannelStatusDeleted),
		).
		Find()
	if err != nil {
		s.logger.Error("查询绑定渠道失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return nil, xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	// 转换为 SourceDetailDTO（详情接口暴露 Token）
	// LastUsedAt 为指针类型，nil 表示未使用，转换为 0 返回给前端
	lastUsedAt := int64(0)
	if source.LastUsedAt != nil {
		lastUsedAt = source.LastUsedAt.UnixMilli()
	}
	sourceDTO := &dto.SourceDetailDTO{
		ID:          source.ID,
		UserID:      source.UserID,
		Name:        source.Name,
		Token:       source.Token,
		Description: source.Description,
		Status:      source.Status,
		CreatedAt:   source.CreatedAt.UnixMilli(),
		UpdatedAt:   source.UpdatedAt.UnixMilli(),
		LastUsedAt:  lastUsedAt,
	}

	// 转换为 ChannelDTO
	channelDTOs := make([]*dto.ChannelDTO, 0, len(channels))
	for _, ch := range channels {
		channelDTOs = append(channelDTOs, &dto.ChannelDTO{
			ID:        ch.ID,
			UserID:    ch.UserID,
			Type:      ch.Type,
			Name:      ch.Name,
			Status:    ch.Status,
			CreatedAt: ch.CreatedAt.UnixMilli(),
		})
	}

	s.logger.Info("查询来源详情成功",
		zap.Int64("source_id", sourceID),
		zap.Int("channel_count", len(channelDTOs)),
	)

	return &dto.SourceDetailResponse{
		Source:   sourceDTO,
		Channels: channelDTOs,
	}, nil
}

// GetSourceToken 查询来源令牌
func (s *SourceService) GetSourceToken(ctx context.Context, sourceID int64, userID int64) (string, error) {
	q := query.Use(s.db)

	source, err := q.Source.WithContext(ctx).
		Where(
			q.Source.ID.Eq(sourceID),
			q.Source.UserID.Eq(userID),
			q.Source.Status.Neq(model.SourceStatusDeleted),
		).
		Select(q.Source.Token).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("来源不存在",
				zap.Int64("source_id", sourceID),
				zap.Int64("user_id", userID),
			)
			return "", xerr.ErrSourceNotFound
		}
		s.logger.Error("查询来源令牌失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return "", xerr.ErrSourceQueryFailed.WithInternal(err)
	}

	s.logger.Info("查询来源令牌成功",
		zap.Int64("source_id", sourceID),
	)

	return source.Token, nil
}

// VerifyPassword 验证用户密码
func (s *SourceService) VerifyPassword(ctx context.Context, userID int64, password string) error {
	q := query.Use(s.db)

	user, err := q.User.WithContext(ctx).
		Where(q.User.ID.Eq(userID)).
		Select(q.User.PasswordHash).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrUnauthorized
		}
		s.logger.Error("查询用户失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return xerr.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("密码验证失败",
			zap.Int64("user_id", userID),
		)
		return xerr.ErrUnauthorized
	}

	return nil
}

// ResetSourceToken 重置来源令牌（旧 Token 立即失效）
func (s *SourceService) ResetSourceToken(ctx context.Context, sourceID int64, userID int64) (string, error) {
	q := query.Use(s.db)

	// 查询来源，校验用户权限
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
			return "", xerr.ErrSourceNotFound
		}
		s.logger.Error("查询来源失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return "", xerr.ErrSourceQueryFailed.WithInternal(err)
	}

	// 检查来源是否已删除
	if source.Status == model.SourceStatusDeleted {
		s.logger.Error("来源已删除",
			zap.Int64("source_id", sourceID),
		)
		return "", xerr.ErrSourceAlreadyDeleted
	}

	// 生成新的唯一 Token
	newToken, err := s.generateUniqueToken(q, ctx)
	if err != nil {
		return "", err
	}

	// 在事务中更新 Token
	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Source.WithContext(ctx).
			Where(tx.Source.ID.Eq(sourceID), tx.Source.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"token":      newToken,
				"updated_at": time.Now(),
			})
		return err
	})
	if err != nil {
		s.logger.Error("重置来源令牌失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return "", xerr.ErrSourceTokenFailed.WithInternal(err)
	}

	s.logger.Info("来源令牌重置成功",
		zap.Int64("source_id", sourceID),
	)

	return newToken, nil
}

// DisableSource 停用消息来源
func (s *SourceService) DisableSource(ctx context.Context, sourceID int64, userID int64) error {
	q := query.Use(s.db)

	// 查询来源，校验用户权限
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

	// 检查来源是否已停用
	if source.Status == model.SourceStatusDisabled {
		s.logger.Error("来源已停用",
			zap.Int64("source_id", sourceID),
		)
		return xerr.ErrSourceAlreadyDisabled
	}

	// 在事务中更新状态为停用
	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Source.WithContext(ctx).
			Where(tx.Source.ID.Eq(sourceID), tx.Source.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"status":     model.SourceStatusDisabled,
				"updated_at": time.Now(),
			})
		return err
	})
	if err != nil {
		s.logger.Error("停用来源失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return xerr.ErrSourceUpdateFailed.WithInternal(err)
	}

	s.logger.Info("来源已停用",
		zap.Int64("source_id", sourceID),
	)

	return nil
}

// EnableSource 启用消息来源
func (s *SourceService) EnableSource(ctx context.Context, sourceID int64, userID int64) error {
	q := query.Use(s.db)

	// 查询来源，校验用户权限
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

	// 检查来源是否已启用
	if source.Status == model.SourceStatusActive {
		s.logger.Error("来源已启用",
			zap.Int64("source_id", sourceID),
		)
		return xerr.ErrSourceAlreadyEnabled
	}

	// 在事务中更新状态为启用
	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Source.WithContext(ctx).
			Where(tx.Source.ID.Eq(sourceID), tx.Source.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"status":     model.SourceStatusActive,
				"updated_at": time.Now(),
			})
		return err
	})
	if err != nil {
		s.logger.Error("启用来源失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return xerr.ErrSourceUpdateFailed.WithInternal(err)
	}

	s.logger.Info("来源已启用",
		zap.Int64("source_id", sourceID),
	)

	return nil
}

// DeleteSource 删除消息来源（软删除），同时级联软删除所有关联的渠道绑定关系
func (s *SourceService) DeleteSource(ctx context.Context, sourceID int64, userID int64) error {
	q := query.Use(s.db)

	// 查询来源，校验用户权限
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

	// 在事务中软删除来源及其关联的渠道绑定关系
	err = q.Transaction(func(tx *query.Query) error {
		// 软删除来源
		_, err := tx.Source.WithContext(ctx).
			Where(tx.Source.ID.Eq(sourceID), tx.Source.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"status":     model.SourceStatusDeleted,
				"updated_at": time.Now(),
			})
		if err != nil {
			return err
		}

		// 级联软删除所有关联的渠道绑定关系
		_, err = tx.SourceChannel.WithContext(ctx).
			Where(tx.SourceChannel.SourceID.Eq(sourceID)).
			Updates(map[string]interface{}{
				"status": model.SourceChannelStatusDeleted,
			})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.logger.Error("删除来源失败",
			zap.Error(err),
			zap.Int64("source_id", sourceID),
		)
		return xerr.ErrSourceDeleteFailed.WithInternal(err)
	}

	s.logger.Info("来源已删除",
		zap.Int64("source_id", sourceID),
	)

	return nil
}
