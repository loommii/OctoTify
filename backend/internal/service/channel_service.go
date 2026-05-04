package service

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/internal/query"
	"octotify/internal/sender"
	"octotify/pkg/xerr"
)

type ChannelService struct {
	db            *gorm.DB
	logger        *zap.Logger
	senderFactory *sender.SenderFactory
}

func NewChannelService(db *gorm.DB, logger *zap.Logger, senderFactory *sender.SenderFactory) *ChannelService {
	return &ChannelService{
		db:            db,
		logger:        logger,
		senderFactory: senderFactory,
	}
}

func (s *ChannelService) CreateChannel(ctx context.Context, userID int64, req *dto.CreateChannelReq) (*dto.ChannelDTO, error) {
	q := query.Use(s.db)

	channel := &model.Channel{
		UserID: userID,
		Type:   req.Type,
		Name:   req.Name,
		Config: req.Config,
		Status: model.ChannelStatusActive,
	}

	err := q.Transaction(func(tx *query.Query) error {
		if err := tx.Channel.WithContext(ctx).Create(channel); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.logger.Error("创建渠道失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.String("name", req.Name),
		)
		return nil, xerr.ErrChannelInsertFailed.WithInternal(err)
	}

	s.logger.Info("渠道创建成功",
		zap.Int64("channel_id", channel.ID),
		zap.String("name", channel.Name),
	)

	return &dto.ChannelDTO{
		ID:         channel.ID,
		UserID:     channel.UserID,
		Type:       channel.Type,
		Name:       channel.Name,
		Config:     channel.Config,
		Status:     channel.Status,
		CreatedAt:  channel.CreatedAt.UnixMilli(),
		UpdatedAt:  channel.UpdatedAt.UnixMilli(),
		LastUsedAt: channel.LastUsedAt.UnixMilli(),
	}, nil
}

func (s *ChannelService) UpdateChannel(ctx context.Context, userID int64, channelID int64, req *dto.UpdateChannelReq) error {
	q := query.Use(s.db)

	channel, err := q.Channel.WithContext(ctx).Where(
		q.Channel.ID.Eq(channelID),
		q.Channel.UserID.Eq(userID),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("渠道不存在",
				zap.Int64("channel_id", channelID),
				zap.Int64("user_id", userID),
			)
			return xerr.ErrChannelNotFound
		}
		s.logger.Error("查询渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"name":       req.Name,
				"config":     datatypes.JSON(req.Config),
				"updated_at": time.Now(),
			})
		return err
	})
	if err != nil {
		s.logger.Error("更新渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	s.logger.Info("渠道更新成功",
		zap.Int64("channel_id", channelID),
		zap.String("name", req.Name),
	)

	return nil
}

func (s *ChannelService) ListChannels(ctx context.Context, userID int64, pageReq *dto.PageReq) ([]*dto.ChannelDTO, int64, error) {
	q := query.Use(s.db)

	pageReq.Normalize()
	offset := (pageReq.Page - 1) * pageReq.PageSize

	total, err := q.Channel.WithContext(ctx).
		Where(
			q.Channel.UserID.Eq(userID),
			q.Channel.Status.Neq(model.ChannelStatusDeleted),
		).
		Count()
	if err != nil {
		s.logger.Error("查询渠道总数失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, 0, xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	channels, err := q.Channel.WithContext(ctx).
		Where(
			q.Channel.UserID.Eq(userID),
			q.Channel.Status.Neq(model.ChannelStatusDeleted),
		).
		Order(q.Channel.CreatedAt.Desc()).
		Offset(offset).
		Limit(pageReq.PageSize).
		Find()
	if err != nil {
		s.logger.Error("查询渠道列表失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, 0, xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	list := make([]*dto.ChannelDTO, 0, len(channels))
	for _, ch := range channels {
		list = append(list, &dto.ChannelDTO{
			ID:         ch.ID,
			UserID:     ch.UserID,
			Type:       ch.Type,
			Name:       ch.Name,
			Config:     ch.Config,
			Status:     ch.Status,
			CreatedAt:  ch.CreatedAt.UnixMilli(),
			UpdatedAt:  ch.UpdatedAt.UnixMilli(),
			LastUsedAt: ch.LastUsedAt.UnixMilli(),
		})
	}

	s.logger.Info("查询渠道列表成功",
		zap.Int64("user_id", userID),
		zap.Int("page", pageReq.Page),
		zap.Int("page_size", pageReq.PageSize),
		zap.Int64("total", total),
	)

	return list, total, nil
}

func (s *ChannelService) GetChannelByID(ctx context.Context, userID int64, channelID int64) (*dto.ChannelDTO, error) {
	q := query.Use(s.db)

	channel, err := q.Channel.WithContext(ctx).
		Where(
			q.Channel.ID.Eq(channelID),
			q.Channel.UserID.Eq(userID),
			q.Channel.Status.Neq(model.ChannelStatusDeleted),
		).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("渠道不存在",
				zap.Int64("channel_id", channelID),
				zap.Int64("user_id", userID),
			)
			return nil, xerr.ErrChannelNotFound
		}
		s.logger.Error("查询渠道详情失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return nil, xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return nil, xerr.ErrChannelAlreadyDeleted
	}

	return &dto.ChannelDTO{
		ID:         channel.ID,
		UserID:     channel.UserID,
		Type:       channel.Type,
		Name:       channel.Name,
		Config:     channel.Config,
		Status:     channel.Status,
		CreatedAt:  channel.CreatedAt.UnixMilli(),
		UpdatedAt:  channel.UpdatedAt.UnixMilli(),
		LastUsedAt: channel.LastUsedAt.UnixMilli(),
	}, nil
}

func (s *ChannelService) TestChannel(ctx context.Context, userID int64, channelID int64) error {
	q := query.Use(s.db)

	channel, err := q.Channel.WithContext(ctx).Where(
		q.Channel.ID.Eq(channelID),
		q.Channel.UserID.Eq(userID),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("渠道不存在",
				zap.Int64("channel_id", channelID),
				zap.Int64("user_id", userID),
			)
			return xerr.ErrChannelNotFound
		}
		s.logger.Error("查询渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	if channel.Status == model.ChannelStatusDisabled {
		s.logger.Error("渠道已停用",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDisabled
	}

	snd, err := s.senderFactory.Create(channel.Type)
	if err != nil {
		s.logger.Error("创建发送器失败",
			zap.Error(err),
			zap.String("channel_type", channel.Type),
		)
		return xerr.ErrThirdPartyCallFailed.WithInternal(err)
	}

	err = snd.Send(ctx, channel.Config, "OctoTify 测试消息", "这是一条测试消息，用于验证渠道配置是否正确。")
	if err != nil {
		s.logger.Error("测试渠道连接失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrThirdPartyCallFailed.WithInternal(err)
	}

	s.logger.Info("测试渠道连接成功",
		zap.Int64("channel_id", channelID),
	)

	return nil
}

func (s *ChannelService) DisableChannel(ctx context.Context, userID int64, channelID int64) error {
	q := query.Use(s.db)

	channel, err := q.Channel.WithContext(ctx).Where(
		q.Channel.ID.Eq(channelID),
		q.Channel.UserID.Eq(userID),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("渠道不存在",
				zap.Int64("channel_id", channelID),
				zap.Int64("user_id", userID),
			)
			return xerr.ErrChannelNotFound
		}
		s.logger.Error("查询渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	if channel.Status == model.ChannelStatusDisabled {
		s.logger.Error("渠道已停用",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDisabled
	}

	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"status":     model.ChannelStatusDisabled,
				"updated_at": time.Now(),
			})
		return err
	})
	if err != nil {
		s.logger.Error("停用渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	s.logger.Info("渠道已停用",
		zap.Int64("channel_id", channelID),
	)

	return nil
}

func (s *ChannelService) EnableChannel(ctx context.Context, userID int64, channelID int64) error {
	q := query.Use(s.db)

	channel, err := q.Channel.WithContext(ctx).Where(
		q.Channel.ID.Eq(channelID),
		q.Channel.UserID.Eq(userID),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("渠道不存在",
				zap.Int64("channel_id", channelID),
				zap.Int64("user_id", userID),
			)
			return xerr.ErrChannelNotFound
		}
		s.logger.Error("查询渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	if channel.Status == model.ChannelStatusActive {
		s.logger.Error("渠道已启用",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyEnabled
	}

	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"status":     model.ChannelStatusActive,
				"updated_at": time.Now(),
			})
		return err
	})
	if err != nil {
		s.logger.Error("启用渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	s.logger.Info("渠道已启用",
		zap.Int64("channel_id", channelID),
	)

	return nil
}

func (s *ChannelService) DeleteChannel(ctx context.Context, userID int64, channelID int64) error {
	q := query.Use(s.db)

	channel, err := q.Channel.WithContext(ctx).Where(
		q.Channel.ID.Eq(channelID),
		q.Channel.UserID.Eq(userID),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("渠道不存在",
				zap.Int64("channel_id", channelID),
				zap.Int64("user_id", userID),
			)
			return xerr.ErrChannelNotFound
		}
		s.logger.Error("查询渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"status":     model.ChannelStatusDeleted,
				"updated_at": time.Now(),
			})
		if err != nil {
			return err
		}

		_, err = tx.SourceChannel.WithContext(ctx).
			Where(tx.SourceChannel.ChannelID.Eq(channelID)).
			Updates(map[string]interface{}{
				"status": model.SourceChannelStatusDeleted,
			})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.logger.Error("删除渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelDeleteFailed.WithInternal(err)
	}

	s.logger.Info("渠道已删除",
		zap.Int64("channel_id", channelID),
	)

	return nil
}
