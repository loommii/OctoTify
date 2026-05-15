package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/google/uuid"
	"octotify/internal/client/ilink"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// WechatClawbotSender 微信ClawBot消息发送器（基于iLink协议）
// 官方文档：https://ilinkai.weixin.qq.com/
// 负责将业务消息通过 iLink 协议推送到微信用户
type WechatClawbotSender struct {
	logger      *zap.Logger     // 日志记录器
	ilinkClient *ilink.Client   // iLink API 客户端，负责底层 HTTP 通信
}

// NewWechatClawbotSender 创建微信ClawBot发送器实例
// 参数:
//   - logger: 日志记录器
//   - ilinkClient: iLink API 客户端（全局共享，多 bot 复用）
func NewWechatClawbotSender(logger *zap.Logger, ilinkClient *ilink.Client) *WechatClawbotSender {
	return &WechatClawbotSender{
		logger:      logger,
		ilinkClient: ilinkClient,
	}
}

// wechatClawbotConfig 微信ClawBot渠道配置（数据库存储明文格式）
// 绑定成功后从 iLink 平台获取，存储在 channels 表的 config 字段中
type wechatClawbotConfig struct {
	BotToken    string `json:"bot_token"`     // Bot Token（明文），用于 API 认证
	ILinkBotID  string `json:"ilink_bot_id"`  // iLink 机器人 ID，格式如 "bot123@im.bot"
	ILinkUserID string `json:"ilink_user_id"` // iLink 用户 ID，格式如 "user123@im.wechat"
}

// 消息内容限制（按字符数计算，非字节数）
// iLink 平台对单条消息长度有限制，超长消息会被截断
const (
	wechatClawbotMaxContentLen  = 2000              // 最大消息内容长度（字符数）
	wechatClawbotTruncateSuffix = "\n...[消息已截断]" // 消息截断后缀
)

// Send 发送微信ClawBot消息
// 参数:
//   - ctx: 上下文，用于控制请求生命周期
//   - config: 渠道配置 JSON（包含 BotToken 和 iLink 标识符）
//   - title: 消息标题
//   - content: 消息内容
//
// 返回:
//   - 错误信息（如果发送失败）
func (s *WechatClawbotSender) Send(ctx context.Context, config datatypes.JSON, title string, content string) error {
	// 1. 解析渠道配置
	var cfg wechatClawbotConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("解析微信ClawBot渠道配置失败: %w", err)
	}

	// 2. 校验必填字段
	if cfg.BotToken == "" {
		return fmt.Errorf("bot_token 不能为空")
	}
	if cfg.ILinkBotID == "" {
		return fmt.Errorf("微信ClawBot IlinkBotID 不能为空")
	}
	if cfg.ILinkUserID == "" {
		return fmt.Errorf("微信ClawBot IlinkUserID 不能为空")
	}

	// 3. 打印脱敏后的渠道配置（Debug 级别）
	s.logger.Debug("微信ClawBot渠道配置",
		zap.String("bot_token", maskToken(cfg.BotToken)),
		zap.String("ilink_bot_id", maskIdentifier(cfg.ILinkBotID)),
		zap.String("ilink_user_id", maskIdentifier(cfg.ILinkUserID)),
	)

	// 4. 构造消息内容（格式：【标题】\n内容）
	// 超长消息按字符数截断，确保不截断多字节字符（如中文）
	msgContent := fmt.Sprintf("【%s】\n%s", title, content)
	if utf8.RuneCountInString(msgContent) > wechatClawbotMaxContentLen {
		suffixLen := utf8.RuneCountInString(wechatClawbotTruncateSuffix)
		maxLen := wechatClawbotMaxContentLen - suffixLen
		if maxLen <= 0 {
			maxLen = wechatClawbotMaxContentLen
		}
		// 将字符数转换为字节索引，确保不截断多字节字符
		byteIndex := 0
		charIndex := 0
		for _, r := range msgContent {
			if charIndex >= maxLen {
				break
			}
			byteIndex += utf8.RuneLen(r)
			charIndex++
		}
		msgContent = msgContent[:byteIndex] + wechatClawbotTruncateSuffix
	}

	// 5. 构造 iLink 请求体
	// ClientID 每次生成唯一 UUID，防止 iLink 服务端消息去重
	msg := &ilink.SendMessageRequest{
		Msg: ilink.SendMsg{
			FromUserID:   cfg.ILinkBotID,      // 发送者为 Bot
			ToUserID:     cfg.ILinkUserID,     // 接收者为用户
			ClientID:     uuid.New().String(), // 每次发送生成唯一 ID，防止消息去重
			MessageType:  2,                   // 2 = Bot 消息
			MessageState: 2,                   // 2 = 完成
			ItemList: []ilink.MessageItem{
				{
					Type: 1, // 1 = 文本消息
					TextItem: &ilink.TextItem{
						Text: msgContent,
					},
				},
			},
		},
		BaseInfo: ilink.BaseInfo{
			ChannelVersion: "1.0.0", // 渠道版本号
		},
	}

	// 6. 调用 ilink client 发送消息
	// 使用动态凭证（botToken），支持多 bot 场景
	_, err := s.ilinkClient.SendMessage(ctx, msg, cfg.BotToken)
	if err != nil {
		return fmt.Errorf("发送微信ClawBot消息失败: %w", err)
	}

	s.logger.Info("微信ClawBot推送成功")
	return nil
}
