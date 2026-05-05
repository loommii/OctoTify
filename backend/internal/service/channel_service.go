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

// ChannelService 推送渠道服务层
// 负责处理推送渠道相关的业务逻辑，包括渠道的增删改查、状态管理、连接测试等
type ChannelService struct {
	db            *gorm.DB              // 数据库连接
	logger        *zap.Logger           // 日志记录器
	senderFactory *sender.SenderFactory // 渠道发送器工厂，用于创建不同渠道的消息发送器
}

// NewChannelService 创建推送渠道服务实例
// 参数:
//   - db: 数据库连接
//   - logger: 日志记录器
//   - senderFactory: 渠道发送器工厂
//
// 返回: 初始化后的 ChannelService 实例
func NewChannelService(db *gorm.DB, logger *zap.Logger, senderFactory *sender.SenderFactory) *ChannelService {
	return &ChannelService{
		db:            db,
		logger:        logger,
		senderFactory: senderFactory,
	}
}

// GetChannelTypes 获取系统支持的所有渠道类型元数据
// 返回渠道类型列表，包含每种类型的配置字段定义，用于前端动态渲染创建渠道表单
// 返回: 渠道类型元数据列表
func (s *ChannelService) GetChannelTypes() []dto.ChannelTypeMeta {
	return dto.ChannelTypeMetas
}

// CreateChannel 创建新的推送渠道
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID
//   - req: 创建渠道请求参数，包含渠道类型、名称和配置
//
// 返回:
//   - 创建成功的渠道信息
//   - 错误信息（如果创建失败）
func (s *ChannelService) CreateChannel(ctx context.Context, userID int64, req *dto.CreateChannelReq) (*dto.ChannelDTO, error) {
	q := query.Use(s.db)

	// 构建渠道模型实例
	channel := &model.Channel{
		UserID: userID,
		Type:   req.Type,                   // 渠道类型（如 feishu, wechat 等）
		Name:   req.Name,                   // 渠道名称
		Config: datatypes.JSON(req.Config), // 渠道配置，将 string 转换为 datatypes.JSON
		Status: model.ChannelStatusActive,  // 默认启用状态
	}

	// 使用事务创建渠道，确保数据一致性
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

	// LastUsedAt 为指针类型，nil 表示未使用，转换为 0 返回给前端
	lastUsedAt := int64(0)
	if channel.LastUsedAt != nil {
		lastUsedAt = channel.LastUsedAt.UnixMilli()
	}

	// 将模型转换为 DTO 返回给前端
	return &dto.ChannelDTO{
		ID:         channel.ID,
		UserID:     channel.UserID,
		Type:       channel.Type,
		Name:       channel.Name,
		Config:     string(channel.Config), // 将 datatypes.JSON 转换为 string
		Status:     channel.Status,
		CreatedAt:  channel.CreatedAt.UnixMilli(),
		UpdatedAt:  channel.UpdatedAt.UnixMilli(),
		LastUsedAt: lastUsedAt,
	}, nil
}

// UpdateChannel 编辑推送渠道
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID，用于权限校验
//   - channelID: 渠道 ID
//   - req: 更新渠道请求参数，包含新的名称和配置
//
// 返回: 错误信息（如果更新失败）
func (s *ChannelService) UpdateChannel(ctx context.Context, userID int64, channelID int64, req *dto.UpdateChannelReq) error {
	q := query.Use(s.db)

	// 查询渠道是否存在，并校验用户权限
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

	// 检查渠道是否已被删除
	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	// 使用事务更新渠道信息
	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"name":       req.Name,
				"config":     datatypes.JSON(req.Config), // 将 string 转换为 datatypes.JSON
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

// ListChannels 分页查询当前用户的推送渠道列表
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID
//   - pageReq: 分页请求参数，包含页码和每页条数
//
// 返回:
//   - 渠道列表
//   - 总记录数
//   - 错误信息（如果查询失败）
func (s *ChannelService) ListChannels(ctx context.Context, userID int64, pageReq *dto.PageReq) ([]*dto.ChannelDTO, int64, error) {
	q := query.Use(s.db)

	// 规范化分页参数，并计算偏移量
	pageReq.Normalize()
	offset := (pageReq.Page - 1) * pageReq.PageSize

	// 查询符合条件的渠道总数（排除已删除的渠道）
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

	// 分页查询渠道列表，按创建时间倒序排列
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

	// 将模型列表转换为 DTO 列表
	list := make([]*dto.ChannelDTO, 0, len(channels))
	for _, ch := range channels {
		// LastUsedAt 为指针类型，nil 表示未使用，转换为 0 返回给前端
		lastUsedAt := int64(0)
		if ch.LastUsedAt != nil {
			lastUsedAt = ch.LastUsedAt.UnixMilli()
		}
		list = append(list, &dto.ChannelDTO{
			ID:         ch.ID,
			UserID:     ch.UserID,
			Type:       ch.Type,
			Name:       ch.Name,
			Config:     string(ch.Config), // 将 datatypes.JSON 转换为 string
			Status:     ch.Status,
			CreatedAt:  ch.CreatedAt.UnixMilli(),
			UpdatedAt:  ch.UpdatedAt.UnixMilli(),
			LastUsedAt: lastUsedAt,
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

// GetChannelByID 根据 ID 查询渠道详情
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID，用于权限校验
//   - channelID: 渠道 ID
//
// 返回:
//   - 渠道详情
//   - 错误信息（如果查询失败或渠道不存在）
func (s *ChannelService) GetChannelByID(ctx context.Context, userID int64, channelID int64) (*dto.ChannelDTO, error) {
	q := query.Use(s.db)

	// 查询渠道详情，排除已删除的渠道
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

	// 二次检查渠道状态（防御性编程）
	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return nil, xerr.ErrChannelAlreadyDeleted
	}

	// 转换 LastUsedAt 字段
	lastUsedAt := int64(0)
	if channel.LastUsedAt != nil {
		lastUsedAt = channel.LastUsedAt.UnixMilli()
	}

	// 将模型转换为 DTO 返回
	return &dto.ChannelDTO{
		ID:         channel.ID,
		UserID:     channel.UserID,
		Type:       channel.Type,
		Name:       channel.Name,
		Config:     string(channel.Config), // 将 datatypes.JSON 转换为 string
		Status:     channel.Status,
		CreatedAt:  channel.CreatedAt.UnixMilli(),
		UpdatedAt:  channel.UpdatedAt.UnixMilli(),
		LastUsedAt: lastUsedAt,
	}, nil
}

// TestChannel 测试渠道连接
// 发送测试消息到指定渠道，验证渠道配置是否正确
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID，用于权限校验
//   - channelID: 渠道 ID
//
// 返回: 错误信息（如果测试失败）
func (s *ChannelService) TestChannel(ctx context.Context, userID int64, channelID int64) error {
	q := query.Use(s.db)

	// 查询渠道信息
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

	// 检查渠道是否已删除
	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	// 检查渠道是否已停用
	if channel.Status == model.ChannelStatusDisabled {
		s.logger.Error("渠道已停用",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDisabled
	}

	// 根据渠道类型创建对应的发送器
	snd, err := s.senderFactory.Create(channel.Type)
	if err != nil {
		s.logger.Error("创建发送器失败",
			zap.Error(err),
			zap.String("channel_type", channel.Type),
		)
		return xerr.ErrThirdPartyCallFailed.WithInternal(err)
	}

	// 发送测试消息
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

// DisableChannel 停用推送渠道
// 停用后该渠道不再接收消息推送，但可以随时重新启用
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID，用于权限校验
//   - channelID: 渠道 ID
//
// 返回: 错误信息（如果停用失败）
func (s *ChannelService) DisableChannel(ctx context.Context, userID int64, channelID int64) error {
	q := query.Use(s.db)

	// 查询渠道信息
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

	// 检查渠道是否已删除
	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	// 检查渠道是否已停用
	if channel.Status == model.ChannelStatusDisabled {
		s.logger.Error("渠道已停用",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDisabled
	}

	// 使用事务更新渠道状态为停用
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

// EnableChannel 启用推送渠道
// 恢复渠道的消息推送功能
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID，用于权限校验
//   - channelID: 渠道 ID
//
// 返回: 错误信息（如果启用失败）
func (s *ChannelService) EnableChannel(ctx context.Context, userID int64, channelID int64) error {
	q := query.Use(s.db)

	// 查询渠道信息
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

	// 检查渠道是否已删除
	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	// 检查渠道是否已启用
	if channel.Status == model.ChannelStatusActive {
		s.logger.Error("渠道已启用",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyEnabled
	}

	// 使用事务更新渠道状态为启用
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

// DeleteChannel 删除推送渠道（软删除）
// 删除渠道的同时解除与所有来源的关联关系
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID，用于权限校验
//   - channelID: 渠道 ID
//
// 返回: 错误信息（如果删除失败）
func (s *ChannelService) DeleteChannel(ctx context.Context, userID int64, channelID int64) error {
	q := query.Use(s.db)

	// 查询渠道信息
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

	// 检查渠道是否已删除
	if channel.Status == model.ChannelStatusDeleted {
		s.logger.Error("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	// 使用事务删除渠道和解除来源关联
	err = q.Transaction(func(tx *query.Query) error {
		// 软删除渠道：将状态标记为已删除
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			Updates(map[string]interface{}{
				"status":     model.ChannelStatusDeleted,
				"updated_at": time.Now(),
			})
		if err != nil {
			return err
		}

		// 解除该渠道与所有来源的关联关系
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
