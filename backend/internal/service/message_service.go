package service

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/internal/query"
	"octotify/internal/sender"
	"octotify/pkg/xerr"
)

type MessageService struct {
	db            *gorm.DB
	logger        *zap.Logger
	senderFactory *sender.SenderFactory
}

func NewMessageService(db *gorm.DB, logger *zap.Logger, senderFactory *sender.SenderFactory) *MessageService {
	return &MessageService{
		db:            db,
		logger:        logger,
		senderFactory: senderFactory,
	}
}

func (s *MessageService) ListMessages(ctx context.Context, userID int64, pageReq *dto.PageReq) ([]*dto.MessageDTO, int64, error) {
	q := query.Use(s.db)

	pageReq.Normalize()
	offset := (pageReq.Page - 1) * pageReq.PageSize

	total, err := q.Message.WithContext(ctx).
		Join(q.Source, q.Source.ID.EqCol(q.Message.SourceID)).
		Where(
			q.Source.UserID.Eq(userID),
			q.Message.Status.Neq(model.MessageStatusDeleted),
		).
		Count()
	if err != nil {
		s.logger.Error("查询消息总数失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, 0, xerr.ErrMessageRecordFailed.WithInternal(err)
	}

	messages, err := q.Message.WithContext(ctx).
		Join(q.Source, q.Source.ID.EqCol(q.Message.SourceID)).
		Where(
			q.Source.UserID.Eq(userID),
			q.Message.Status.Neq(model.MessageStatusDeleted),
		).
		Order(q.Message.CreatedAt.Desc()).
		Offset(offset).
		Limit(pageReq.PageSize).
		Find()
	if err != nil {
		s.logger.Error("查询消息列表失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, 0, xerr.ErrMessageRecordFailed.WithInternal(err)
	}

	list := make([]*dto.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		list = append(list, &dto.MessageDTO{
			ID:        msg.ID,
			SourceID:  msg.SourceID,
			ChannelID: msg.ChannelID,
			Title:     msg.Title,
			Content:   msg.Content,
			Status:    msg.Status,
			CreatedAt: msg.CreatedAt.UnixMilli(),
			UpdatedAt: msg.UpdatedAt.UnixMilli(),
		})
	}

	s.logger.Info("查询消息列表成功",
		zap.Int64("user_id", userID),
		zap.Int("page", pageReq.Page),
		zap.Int("page_size", pageReq.PageSize),
		zap.Int64("total", total),
	)

	return list, total, nil
}

func (s *MessageService) FilterMessages(ctx context.Context, userID int64, filter *dto.MessageFilterReq) ([]*dto.MessageDTO, int64, error) {
	q := query.Use(s.db)

	filter.Normalize()
	offset := (filter.Page - 1) * filter.PageSize

	baseQuery := q.Message.WithContext(ctx).
		Join(q.Source, q.Source.ID.EqCol(q.Message.SourceID)).
		Where(
			q.Source.UserID.Eq(userID),
			q.Message.Status.Neq(model.MessageStatusDeleted),
		)

	if filter.SourceID != nil {
		baseQuery = baseQuery.Where(q.Message.SourceID.Eq(*filter.SourceID))
	}
	if filter.ChannelID != nil {
		baseQuery = baseQuery.Where(q.Message.ChannelID.Eq(*filter.ChannelID))
	}
	if filter.Status != nil {
		baseQuery = baseQuery.Where(q.Message.Status.Eq(*filter.Status))
	}
	if filter.StartDate != nil {
		baseQuery = baseQuery.Where(q.Message.CreatedAt.Gte(time.UnixMilli(*filter.StartDate)))
	}
	if filter.EndDate != nil {
		baseQuery = baseQuery.Where(q.Message.CreatedAt.Lte(time.UnixMilli(*filter.EndDate)))
	}
	if filter.Keyword != "" {
		baseQuery = baseQuery.Where(q.Message.Title.Like("%" + filter.Keyword + "%"))
		baseQuery = baseQuery.Or(q.Message.WithContext(ctx).
			Join(q.Source, q.Source.ID.EqCol(q.Message.SourceID)).
			Where(
				q.Source.UserID.Eq(userID),
				q.Message.Status.Neq(model.MessageStatusDeleted),
			).
			Where(q.Message.Content.Like("%" + filter.Keyword + "%")))
	}

	total, err := baseQuery.Count()
	if err != nil {
		s.logger.Error("查询消息总数失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, 0, xerr.ErrMessageRecordFailed.WithInternal(err)
	}

	messages, err := baseQuery.
		Order(q.Message.CreatedAt.Desc()).
		Offset(offset).
		Limit(filter.PageSize).
		Find()
	if err != nil {
		s.logger.Error("查询消息列表失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, 0, xerr.ErrMessageRecordFailed.WithInternal(err)
	}

	list := make([]*dto.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		list = append(list, &dto.MessageDTO{
			ID:        msg.ID,
			SourceID:  msg.SourceID,
			ChannelID: msg.ChannelID,
			Title:     msg.Title,
			Content:   msg.Content,
			Status:    msg.Status,
			CreatedAt: msg.CreatedAt.UnixMilli(),
			UpdatedAt: msg.UpdatedAt.UnixMilli(),
		})
	}

	s.logger.Info("筛选消息列表成功",
		zap.Int64("user_id", userID),
		zap.Int("page", filter.Page),
		zap.Int("page_size", filter.PageSize),
		zap.Int64("total", total),
	)

	return list, total, nil
}

func (s *MessageService) GetMessageByID(ctx context.Context, userID int64, messageID int64) (*dto.MessageDetailDTO, error) {
	var result struct {
		ID          int64
		SourceID    int64
		SourceName  string
		ChannelID   int64
		ChannelName string
		ChannelType string
		Title       string
		Content     string
		Status      int
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	err := s.db.WithContext(ctx).
		Raw(`SELECT m.id, m.source_id, s.name as source_name, m.channel_id, c.name as channel_name, c.type as channel_type, 
				m.title, m.content, m.status, m.created_at, m.updated_at 
				FROM messages m 
				LEFT JOIN sources s ON m.source_id = s.id 
				LEFT JOIN channels c ON m.channel_id = c.id 
				WHERE m.id = ? AND s.user_id = ? AND m.status != -1`, messageID, userID).
		Scan(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("消息不存在",
				zap.Int64("message_id", messageID),
				zap.Int64("user_id", userID),
			)
			return nil, xerr.ErrNotFound
		}
		s.logger.Error("查询消息详情失败",
			zap.Error(err),
			zap.Int64("message_id", messageID),
		)
		return nil, xerr.ErrMessageRecordFailed.WithInternal(err)
	}

	if result.ID == 0 {
		s.logger.Error("消息不存在",
			zap.Int64("message_id", messageID),
			zap.Int64("user_id", userID),
		)
		return nil, xerr.ErrNotFound
	}

	return &dto.MessageDetailDTO{
		ID:          result.ID,
		SourceID:    result.SourceID,
		SourceName:  result.SourceName,
		ChannelID:   result.ChannelID,
		ChannelName: result.ChannelName,
		ChannelType: result.ChannelType,
		Title:       result.Title,
		Content:     result.Content,
		Status:      result.Status,
		CreatedAt:   result.CreatedAt.UnixMilli(),
		UpdatedAt:   result.UpdatedAt.UnixMilli(),
	}, nil
}

func (s *MessageService) DeleteMessage(ctx context.Context, userID int64, messageID int64) error {
	q := query.Use(s.db)

	message, err := q.Message.WithContext(ctx).
		Join(q.Source, q.Source.ID.EqCol(q.Message.SourceID)).
		Where(
			q.Source.UserID.Eq(userID),
			q.Message.ID.Eq(messageID),
		).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("消息不存在",
				zap.Int64("message_id", messageID),
				zap.Int64("user_id", userID),
			)
			return xerr.ErrNotFound
		}
		s.logger.Error("查询消息失败",
			zap.Error(err),
			zap.Int64("message_id", messageID),
		)
		return xerr.ErrMessageRecordFailed.WithInternal(err)
	}

	if message.Status == model.MessageStatusDeleted {
		s.logger.Error("消息已删除",
			zap.Int64("message_id", messageID),
		)
		return xerr.ErrMessageAlreadyDeleted
	}

	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Message.WithContext(ctx).
			Where(tx.Message.ID.Eq(messageID)).
			Updates(map[string]interface{}{
				"status":     model.MessageStatusDeleted,
				"updated_at": time.Now(),
			})
		return err
	})
	if err != nil {
		s.logger.Error("删除消息失败",
			zap.Error(err),
			zap.Int64("message_id", messageID),
		)
		return xerr.ErrMessageRecordFailed.WithInternal(err)
	}

	s.logger.Info("消息已删除",
		zap.Int64("message_id", messageID),
	)

	return nil
}

func (s *MessageService) PushMessage(ctx context.Context, sourceToken string, req *dto.PushMessageReq) (*dto.PushResponse, error) {
	q := query.Use(s.db)

	source, err := q.Source.WithContext(ctx).
		Where(
			q.Source.Token.Eq(sourceToken),
			q.Source.Status.Eq(model.SourceStatusActive),
		).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("来源不存在或已禁用",
				zap.String("token", sourceToken),
			)
			return nil, xerr.ErrSourceNotFound
		}
		s.logger.Error("查询来源失败",
			zap.Error(err),
			zap.String("token", sourceToken),
		)
		return nil, xerr.ErrSourceQueryFailed.WithInternal(err)
	}

	channels, err := q.Channel.WithContext(ctx).
		Join(q.SourceChannel, q.SourceChannel.ChannelID.EqCol(q.Channel.ID)).
		Where(
			q.SourceChannel.SourceID.Eq(source.ID),
			q.Channel.Status.Eq(model.ChannelStatusActive),
		).
		Find()
	if err != nil {
		s.logger.Error("查询渠道失败",
			zap.Error(err),
			zap.Int64("source_id", source.ID),
		)
		return nil, xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	if len(channels) == 0 {
		s.logger.Error("来源未绑定任何渠道",
			zap.Int64("source_id", source.ID),
		)
		return nil, xerr.ErrMessageNoChannels
	}

	return s.pushToChannels(ctx, source, channels, req)
}

func (s *MessageService) pushToChannels(ctx context.Context, source *model.Source, channels []*model.Channel, req *dto.PushMessageReq) (*dto.PushResponse, error) {
	q := query.Use(s.db)

	g, ctx := errgroup.WithContext(ctx)
	results := make([]*dto.PushResult, len(channels))

	for i, ch := range channels {
		i, ch := i, ch
		g.Go(func() error {
			message := &model.Message{
				SourceID:  source.ID,
				ChannelID: ch.ID,
				Title:     req.Title,
				Content:   req.Message,
				Status:    model.MessageStatusPending,
			}

			err := q.Transaction(func(tx *query.Query) error {
				if err := tx.Message.WithContext(ctx).Create(message); err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				s.logger.Error("创建消息记录失败",
					zap.Error(err),
					zap.Int64("channel_id", ch.ID),
				)
				results[i] = &dto.PushResult{
					ChannelID:   ch.ID,
					ChannelName: ch.Name,
					Success:     false,
					Error:       "创建消息记录失败",
				}
				return nil
			}

			snd, err := s.senderFactory.Create(ch.Type)
			if err != nil {
				s.logger.Error("创建发送器失败",
					zap.Error(err),
					zap.String("channel_type", ch.Type),
				)
				s.updateMessageStatus(ctx, message.ID, model.MessageStatusFailed)
				results[i] = &dto.PushResult{
					ChannelID:   ch.ID,
					ChannelName: ch.Name,
					MessageID:   message.ID,
					Success:     false,
					Error:       "不支持的渠道类型",
				}
				return nil
			}

			sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			err = snd.Send(sendCtx, ch.Config, req.Title, req.Message)
			if err != nil {
				s.logger.Error("推送消息失败",
					zap.Error(err),
					zap.Int64("channel_id", ch.ID),
					zap.Int64("message_id", message.ID),
				)
				s.updateMessageStatus(ctx, message.ID, model.MessageStatusFailed)
				results[i] = &dto.PushResult{
					ChannelID:   ch.ID,
					ChannelName: ch.Name,
					MessageID:   message.ID,
					Success:     false,
					Error:       err.Error(),
				}
				return nil
			}

			s.updateMessageStatus(ctx, message.ID, model.MessageStatusSuccess)
			results[i] = &dto.PushResult{
				ChannelID:   ch.ID,
				ChannelName: ch.Name,
				MessageID:   message.ID,
				Success:     true,
			}

			s.logger.Info("推送消息成功",
				zap.Int64("channel_id", ch.ID),
				zap.Int64("message_id", message.ID),
			)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		s.logger.Error("并发推送消息失败",
			zap.Error(err),
		)
		return nil, xerr.ErrMessagePushFailed.WithInternal(err)
	}

	var successCount, failedCount int
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	s.logger.Info("消息推送完成",
		zap.Int64("source_id", source.ID),
		zap.Int("total", len(results)),
		zap.Int("success", successCount),
		zap.Int("failed", failedCount),
	)

	return &dto.PushResponse{
		Total:   len(results),
		Success: successCount,
		Failed:  failedCount,
		Results: results,
	}, nil
}

func (s *MessageService) updateMessageStatus(ctx context.Context, messageID int64, status int) {
	q := query.Use(s.db)

	_, err := q.Message.WithContext(ctx).
		Where(q.Message.ID.Eq(messageID)).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})
	if err != nil {
		s.logger.Error("更新消息状态失败",
			zap.Error(err),
			zap.Int64("message_id", messageID),
			zap.Int("status", status),
		)
	}
}
