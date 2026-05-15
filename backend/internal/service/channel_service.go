package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"octotify/internal/client/ilink"
	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/internal/query"
	"octotify/internal/sender"
	"octotify/pkg/aescipher"
	"octotify/pkg/ctxutil"
	"octotify/pkg/xerr"
)

// ChannelService 推送渠道服务层
// 负责处理推送渠道相关的业务逻辑，包括渠道的增删改查、状态管理、连接测试等
type ChannelService struct {
	db            *gorm.DB              // 数据库连接
	logger        *zap.Logger           // 日志记录器
	senderFactory *sender.SenderFactory // 渠道发送器工厂，用于创建不同渠道的消息发送器
	ilinkClient   *ilink.Client         // iLink 平台 API 客户端
}

// NewChannelService 创建推送渠道服务实例
// 参数:
//   - db: 数据库连接
//   - logger: 日志记录器
//   - senderFactory: 渠道发送器工厂
//   - ilinkClient: iLink 平台 API 客户端
//
// 返回: 初始化后的 ChannelService 实例
func NewChannelService(db *gorm.DB, logger *zap.Logger, senderFactory *sender.SenderFactory, ilinkClient *ilink.Client) *ChannelService {
	return &ChannelService{
		db:            db,
		logger:        logger,
		senderFactory: senderFactory,
		ilinkClient:   ilinkClient,
	}
}

func (s *ChannelService) log(ctx context.Context) *zap.Logger {
	if rid := ctxutil.GetRequestID(ctx); rid != "" {
		return s.logger.With(zap.String("request_id", rid))
	}
	return s.logger
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

	// 对微信ClawBot渠道，解密密文并校验，确保配置有效
	configData, err := s.processChannelConfig(ctx, req.Type, req.Config.ToJSON())
	if err != nil {
		return nil, err
	}

	// 构建渠道模型实例
	channel := &model.Channel{
		UserID: userID,
		Type:   req.Type,                  // 渠道类型（如 feishu, wechat 等）
		Name:   req.Name,                  // 渠道名称
		Config: configData,                // 渠道配置（ClawBot 存储密文）
		Status: model.ChannelStatusActive, // 默认启用状态
	}

	// 使用事务创建渠道，确保数据一致性
	err = q.Transaction(func(tx *query.Query) error {
		if err := tx.Channel.WithContext(ctx).Create(channel); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.log(ctx).Error("创建渠道失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
			zap.String("name", req.Name),
		)
		return nil, xerr.ErrChannelInsertFailed.WithInternal(err)
	}

	s.log(ctx).Info("渠道创建成功",
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
		Config:     dto.FromJSON(channel.Config),
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
			s.log(ctx).Warn("渠道不存在",
				zap.Int64("channel_id", channelID),
				zap.Int64("user_id", userID),
			)
			return xerr.ErrChannelNotFound
		}
		s.log(ctx).Error("查询渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	// 检查渠道是否已被删除
	if channel.Status == model.ChannelStatusDeleted {
		s.log(ctx).Warn("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyDeleted
	}

	// 对微信ClawBot渠道，解密密文并校验，确保配置有效
	configData, err := s.processChannelConfig(ctx, channel.Type, req.Config.ToJSON())
	if err != nil {
		return err
	}

	// 使用事务更新渠道信息
	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			UpdateSimple(
				tx.Channel.Name.Value(req.Name),
				tx.Channel.Config.Value(configData),
				tx.Channel.UpdatedAt.Value(time.Now()),
			)
		return err
	})
	if err != nil {
		s.log(ctx).Error("更新渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	s.log(ctx).Info("渠道更新成功",
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
		s.log(ctx).Error("查询渠道总数失败",
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
		s.log(ctx).Error("查询渠道列表失败",
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
			Config:     dto.FromJSON(ch.Config),
			Status:     ch.Status,
			CreatedAt:  ch.CreatedAt.UnixMilli(),
			UpdatedAt:  ch.UpdatedAt.UnixMilli(),
			LastUsedAt: lastUsedAt,
		})
	}

	s.log(ctx).Info("查询渠道列表成功",
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
			s.log(ctx).Warn("渠道不存在",
				zap.Int64("channel_id", channelID),
				zap.Int64("user_id", userID),
			)
			return nil, xerr.ErrChannelNotFound
		}
		s.log(ctx).Error("查询渠道详情失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return nil, xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	// 二次检查渠道状态（防御性编程）
	if channel.Status == model.ChannelStatusDeleted {
		s.log(ctx).Warn("渠道已删除",
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
		Config:     dto.FromJSON(channel.Config),
		Status:     channel.Status,
		CreatedAt:  channel.CreatedAt.UnixMilli(),
		UpdatedAt:  channel.UpdatedAt.UnixMilli(),
		LastUsedAt: lastUsedAt,
	}, nil
}

// getChannelForUser 查询渠道并校验用户权限和状态
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID
//   - channelID: 渠道 ID
//   - checkDeleted: 是否检查已删除状态
//   - checkDisabled: 是否检查已停用状态
//
// 返回: 渠道模型实例，错误信息
func (s *ChannelService) getChannelForUser(ctx context.Context, userID int64, channelID int64, checkDeleted bool, checkDisabled bool) (*model.Channel, error) {
	q := query.Use(s.db)

	channel, err := q.Channel.WithContext(ctx).Where(
		q.Channel.ID.Eq(channelID),
		q.Channel.UserID.Eq(userID),
	).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.log(ctx).Warn("渠道不存在",
				zap.Int64("channel_id", channelID),
				zap.Int64("user_id", userID),
			)
			return nil, xerr.ErrChannelNotFound
		}
		s.log(ctx).Error("查询渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return nil, xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	// 检查渠道是否已删除
	if checkDeleted && channel.Status == model.ChannelStatusDeleted {
		s.log(ctx).Warn("渠道已删除",
			zap.Int64("channel_id", channelID),
		)
		return nil, xerr.ErrChannelAlreadyDeleted
	}

	// 检查渠道是否已停用
	if checkDisabled && channel.Status == model.ChannelStatusDisabled {
		s.log(ctx).Warn("渠道已停用",
			zap.Int64("channel_id", channelID),
		)
		return nil, xerr.ErrChannelAlreadyDisabled
	}

	return channel, nil
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
	// 查询渠道并校验权限和状态
	channel, err := s.getChannelForUser(ctx, userID, channelID, true, true)
	if err != nil {
		return err
	}

	// 根据渠道类型创建对应的发送器
	snd, err := s.senderFactory.Create(channel.Type)
	if err != nil {
		s.log(ctx).Error("创建发送器失败",
			zap.Error(err),
			zap.String("channel_type", channel.Type),
		)
		return xerr.ErrThirdPartyCallFailed.WithInternal(err)
	}

	// 发送测试消息
	// 测试消息包含请求ID和时间戳，确保每次内容不同，避免 iLink 服务端消息去重
	requestID := ctxutil.GetRequestID(ctx)
	testContent := fmt.Sprintf("这是一条测试消息，用于验证渠道配置是否正确。\n请求ID: %s\n发送时间: %s", requestID, time.Now().Format("2006-01-02 15:04:05"))
	err = snd.Send(ctx, channel.Config, "OctoTify 测试消息", testContent)
	if err != nil {
		s.log(ctx).Error("测试渠道连接失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrThirdPartyCallFailed.WithInternal(err)
	}

	s.log(ctx).Info("测试渠道连接成功",
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
	// 查询渠道并校验权限和状态（仅校验，不使用返回的 channel 对象）
	if _, err := s.getChannelForUser(ctx, userID, channelID, true, true); err != nil {
		return err
	}

	q := query.Use(s.db)

	// 使用事务更新渠道状态为停用
	var err error
	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			UpdateSimple(
				tx.Channel.Status.Value(model.ChannelStatusDisabled),
				tx.Channel.UpdatedAt.Value(time.Now()),
			)
		return err
	})
	if err != nil {
		s.log(ctx).Error("停用渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	s.log(ctx).Info("渠道已停用",
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
	// 查询渠道并校验权限和状态（仅校验删除状态，不校验停用状态）
	channel, err := s.getChannelForUser(ctx, userID, channelID, true, false)
	if err != nil {
		return err
	}

	// 检查渠道是否已启用
	if channel.Status == model.ChannelStatusActive {
		s.log(ctx).Warn("渠道已启用",
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelAlreadyEnabled
	}

	q := query.Use(s.db)

	// 使用事务更新渠道状态为启用
	err = q.Transaction(func(tx *query.Query) error {
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			UpdateSimple(
				tx.Channel.Status.Value(model.ChannelStatusActive),
				tx.Channel.UpdatedAt.Value(time.Now()),
			)
		return err
	})
	if err != nil {
		s.log(ctx).Error("启用渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelQueryFailed.WithInternal(err)
	}

	s.log(ctx).Info("渠道已启用",
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
	// 查询渠道并校验权限和状态（仅校验，不使用返回的 channel 对象）
	if _, err := s.getChannelForUser(ctx, userID, channelID, true, false); err != nil {
		return err
	}

	q := query.Use(s.db)

	// 使用事务删除渠道和解除来源关联
	var err error
	err = q.Transaction(func(tx *query.Query) error {
		// 软删除渠道：将状态标记为已删除
		_, err := tx.Channel.WithContext(ctx).
			Where(tx.Channel.ID.Eq(channelID), tx.Channel.UserID.Eq(userID)).
			UpdateSimple(
				tx.Channel.Status.Value(model.ChannelStatusDeleted),
				tx.Channel.UpdatedAt.Value(time.Now()),
			)
		if err != nil {
			return err
		}

		// 解除该渠道与所有来源的关联关系
		_, err = tx.SourceChannel.WithContext(ctx).
			Where(tx.SourceChannel.ChannelID.Eq(channelID)).
			UpdateSimple(
				tx.SourceChannel.Status.Value(model.SourceChannelStatusDeleted),
			)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.log(ctx).Error("删除渠道失败",
			zap.Error(err),
			zap.Int64("channel_id", channelID),
		)
		return xerr.ErrChannelDeleteFailed.WithInternal(err)
	}

	s.log(ctx).Info("渠道已删除",
		zap.Int64("channel_id", channelID),
	)

	return nil
}

// StartBind 发起微信ClawBot绑定流程
// 参数:
//   - ctx: 上下文
//   - userID: 用户 ID
//
// 返回:
//   - qrcode: 二维码原始值（用于轮询状态）
//   - qrcodeURL: 二维码图片内容（用于前端展示）
//   - err: 错误信息
func (s *ChannelService) StartBind(ctx context.Context, userID int64) (qrcode string, qrcodeURL string, err error) {
	s.log(ctx).Info("发起微信ClawBot绑定",
		zap.Int64("user_id", userID),
	)

	// 调用 iLink API 获取二维码
	qrResp, err := s.ilinkClient.GetQRCode(ctx)
	if err != nil {
		s.log(ctx).Error("获取二维码失败",
			zap.Error(err),
		)
		return "", "", xerr.ErrQRCodeFetchFailed.WithInternal(err)
	}

	return qrResp.QRCode, qrResp.QRCodeImgContent, nil
}

// PollBindStatus 轮询 iLink API 查询绑定状态
// 参数:
//   - ctx: 上下文
//   - qrcode: iLink 返回的二维码原始值
//
// 返回:
//   - status: 当前绑定状态（iLink 原始状态值）
//   - credentials: 绑定成功时的凭证（仅在 confirmed 时非空）
//   - err: 错误信息（如果调用失败）
func (s *ChannelService) PollBindStatus(ctx context.Context, qrcode string) (status string, credentials *ilink.Credentials, err error) {
	// 调用 iLink API 查询状态
	statusResp, err := s.ilinkClient.GetQRCodeStatus(ctx, qrcode)
	if err != nil {
		return "", nil, xerr.ErrBindStatusFailed.WithInternal(err)
	}

	// 直接返回 iLink 原始状态，不做映射
	status = statusResp.Status

	// 如果绑定成功，返回明文凭证（handler 层负责加密后返回前端）
	if status == ilink.StatusConfirmed {
		credentials = &ilink.Credentials{
			BotToken:    statusResp.BotToken,
			IlinkBotID:  statusResp.ILinkBotID,
			IlinkUserID: statusResp.ILinkUserID,
		}
	}

	return status, credentials, nil
}

// processChannelConfig 处理渠道配置（针对 wechat_clawbot 类型进行解密和校验）
// 前端提交的 config 中 bot_token 为密文格式（bot_token_ciphertext + bot_token_nonce），
// 此方法解密后替换为明文 bot_token，存入数据库时为明文格式
func (s *ChannelService) processChannelConfig(ctx context.Context, channelType string, config datatypes.JSON) (datatypes.JSON, error) {
	if channelType != dto.ChannelTypeWechatClawbot {
		return config, nil
	}

	// 解析配置
	var cfg map[string]interface{}
	if err := json.Unmarshal(config, &cfg); err != nil {
		s.log(ctx).Error("解析渠道配置失败", zap.Error(err))
		return nil, fmt.Errorf("解析渠道配置失败: %w", err)
	}

	// 检查是否为密文格式（前端提交时加密传输）
	cipherText, hasCipher := cfg["bot_token_ciphertext"].(string)
	nonce, hasNonce := cfg["bot_token_nonce"].(string)

	if hasCipher && hasNonce {
		// 解密前端传来的密文，提取明文 bot_token
		plaintext, err := aescipher.GlobalDecryptBase64(cipherText, nonce)
		if err != nil {
			s.log(ctx).Error("BotToken 解密失败", zap.Error(err))
			return nil, fmt.Errorf("凭证解密失败，请重新绑定微信ClawBot: %w", err)
		}
		if len(plaintext) == 0 {
			return nil, fmt.Errorf("解密后的 BotToken 为空")
		}

		ilinkBotID, _ := cfg["ilink_bot_id"].(string)
		ilinkUserID, _ := cfg["ilink_user_id"].(string)
		if ilinkBotID == "" {
			return nil, fmt.Errorf("ilink_bot_id 不能为空")
		}
		if ilinkUserID == "" {
			return nil, fmt.Errorf("ilink_user_id 不能为空")
		}

		// 替换为明文格式存入数据库
		delete(cfg, "bot_token_ciphertext")
		delete(cfg, "bot_token_nonce")
		cfg["bot_token"] = string(plaintext)

		result, err := json.Marshal(cfg)
		if err != nil {
			s.log(ctx).Error("序列化渠道配置失败", zap.Error(err))
			return nil, fmt.Errorf("序列化渠道配置失败: %w", err)
		}
		return result, nil
	}

	// 兼容已有明文格式（数据库中直接存储 bot_token 明文）
	if botToken, hasToken := cfg["bot_token"].(string); hasToken && botToken != "" {
		ilinkBotID, _ := cfg["ilink_bot_id"].(string)
		ilinkUserID, _ := cfg["ilink_user_id"].(string)
		if ilinkBotID == "" {
			return nil, fmt.Errorf("ilink_bot_id 不能为空")
		}
		if ilinkUserID == "" {
			return nil, fmt.Errorf("ilink_user_id 不能为空")
		}
		return config, nil
	}

	return nil, fmt.Errorf("微信ClawBot配置缺少 bot_token_ciphertext/bot_token_nonce 或 bot_token 字段")
}
