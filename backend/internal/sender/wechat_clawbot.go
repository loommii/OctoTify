package sender

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// WechatClawbotSender 微信ClawBot消息发送器（基于iLink协议）
// 官方文档：https://ilinkai.weixin.qq.com/
type WechatClawbotSender struct {
	logger *zap.Logger // 日志记录器
}

// NewWechatClawbotSender 创建微信ClawBot发送器实例
func NewWechatClawbotSender(logger *zap.Logger) *WechatClawbotSender {
	return &WechatClawbotSender{logger: logger}
}

// wechatClawbotConfig 微信ClawBot渠道配置（数据库存储明文格式）
type wechatClawbotConfig struct {
	BotToken    string `json:"bot_token"`     // Bot Token（明文）
	ILinkBotID  string `json:"ilink_bot_id"`  // iLink 机器人 ID，格式如 "bot123@im.bot"
	ILinkUserID string `json:"ilink_user_id"` // iLink 用户 ID，格式如 "user123@im.wechat"
}

// ilinkSendMessageRequest iLink 发送消息请求体（参考 weclaw-main/ilink/types.go）
type ilinkSendMessageRequest struct {
	Msg      ilinkSendMsg  `json:"msg"`       // 消息内容
	BaseInfo ilinkBaseInfo `json:"base_info"` // 基础信息
}

// ilinkSendMsg iLink 消息体
type ilinkSendMsg struct {
	FromUserID   string             `json:"from_user_id"`  // 发送者 ID（Bot ID）
	ToUserID     string             `json:"to_user_id"`    // 接收者 ID（用户 ID）
	ClientID     string             `json:"client_id"`     // 客户端 ID（通常与 Bot ID 相同）
	MessageType  int                `json:"message_type"`  // 消息类型：2 = Bot 消息
	MessageState int                `json:"message_state"` // 消息状态：2 = 完成
	ItemList     []ilinkMessageItem `json:"item_list"`     // 消息项列表
}

// ilinkMessageItem iLink 消息项
type ilinkMessageItem struct {
	Type     int            `json:"type"`                // 消息项类型：1 = 文本
	TextItem *ilinkTextItem `json:"text_item,omitempty"` // 文本消息项
}

// ilinkTextItem iLink 文本消息项
type ilinkTextItem struct {
	Text string `json:"text"` // 文本内容
}

// ilinkBaseInfo iLink 基础信息
type ilinkBaseInfo struct {
	ChannelVersion string `json:"channel_version,omitempty"` // 渠道版本号
}

// ilinkSendMessageResponse iLink API 响应体
type ilinkSendMessageResponse struct {
	Ret    int    `json:"ret"`              // 业务状态码：0 表示成功
	ErrMsg string `json:"errmsg,omitempty"` // 错误信息
}

// wechatClawbotHTTPClient iLink HTTP 客户端（复用连接池，超时时间 30 秒）
var wechatClawbotHTTPClient = &http.Client{Timeout: 30 * time.Second}

// wechatClawbotBaseURL iLink API 基础地址（可通过测试覆盖）
var wechatClawbotBaseURL = "https://ilinkai.weixin.qq.com"

// wechatClawbotUIN UIN 全局实例（每个进程生成一次，参考 weclaw-main/ilink/client.go）
var wechatClawbotUIN string
var wechatClawbotUINOnce sync.Once

// generateWechatClawbotUIN 生成随机 UIN（User Identification Number）
// 使用加密安全的随机数生成 32 位无符号整数，并进行 Base64 编码
func generateWechatClawbotUIN() string {
	var n uint32
	_ = binary.Read(rand.Reader, binary.LittleEndian, &n)
	s := fmt.Sprintf("%d", n)
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// getWechatClawbotUIN 获取全局 UIN（线程安全，只生成一次）
func getWechatClawbotUIN() string {
	wechatClawbotUINOnce.Do(func() {
		wechatClawbotUIN = generateWechatClawbotUIN()
	})
	return wechatClawbotUIN
}

// 消息内容限制（按字符数计算，非字节数）
const (
	wechatClawbotMaxContentLen  = 2000           // 最大消息内容长度（字符数）
	wechatClawbotTruncateSuffix = "\n...[消息已截断]" // 消息截断后缀
)

// Send 发送微信ClawBot消息
// 参数:
//   - ctx: 上下文，用于控制请求生命周期
//   - config: 渠道配置 JSON（包含加密的 BotToken 和 iLink 标识符）
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

	// 5. 构造消息内容（格式：【标题】\n内容）
	msgContent := fmt.Sprintf("【%s】\n%s", title, content)
	if utf8.RuneCountInString(msgContent) > wechatClawbotMaxContentLen {
		// 超长消息截断处理（按字符数截断）
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

	// 6. 构造 iLink 请求体（参考 weclaw-main/ilink/types.go）
	msg := ilinkSendMessageRequest{
		Msg: ilinkSendMsg{
			FromUserID:   cfg.ILinkBotID,  // 发送者为 Bot
			ToUserID:     cfg.ILinkUserID, // 接收者为用户
			ClientID:     cfg.ILinkBotID,  // 客户端 ID 与 Bot ID 相同
			MessageType:  2,               // 2 = Bot 消息
			MessageState: 2,               // 2 = 完成
			ItemList: []ilinkMessageItem{
				{
					Type: 1, // 1 = 文本消息
					TextItem: &ilinkTextItem{
						Text: msgContent,
					},
				},
			},
		},
		BaseInfo: ilinkBaseInfo{
			ChannelVersion: "1.0.0", // 渠道版本号
		},
	}

	// 序列化请求体
	bodyBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("构造微信ClawBot消息体失败: %w", err)
	}

	s.logger.Debug("微信ClawBot请求体",
		zap.String("body", string(bodyBytes)),
	)

	// 7. 创建 HTTP 请求
	apiURL := wechatClawbotBaseURL + "/ilink/bot/sendmessage"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建微信ClawBot HTTP请求失败: %w", err)
	}

	// 8. 设置 iLink 必需请求头（参考 weclaw-main/ilink/client.go:setHeaders）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AuthorizationType", "ilink_bot_token") // 认证类型：Bot Token
	req.Header.Set("Authorization", "Bearer "+cfg.BotToken)  // Bot Token 认证
	req.Header.Set("X-WECHAT-UIN", getWechatClawbotUIN())  // 微信 UIN 标识

	// 9. 发送请求
	resp, err := wechatClawbotHTTPClient.Do(req)
	if err != nil {
		s.logger.Error("微信ClawBot网络请求失败",
			zap.String("url", apiURL),
			zap.Error(err),
		)
		return fmt.Errorf("发送微信ClawBot请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 10. 读取响应体（限制最大 1MB 防止恶意大响应）
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("读取微信ClawBot响应失败: %w", err)
	}

	s.logger.Debug("微信ClawBot原始响应",
		zap.Int("http_status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	// 11. 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("微信ClawBot返回HTTP错误: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 12. 检查响应体是否为空（iLink 成功时返回 {}，属于正常响应，不视为空响应）
	if len(respBody) == 0 {
		s.logger.Error("微信ClawBot返回空响应",
			zap.Int("http_status", resp.StatusCode),
		)
		return fmt.Errorf("微信ClawBot返回空响应（HTTP %d），请检查Bot Token和路由配置是否正确", resp.StatusCode)
	}

	// 13. 解析响应 JSON（iLink sendmessage 成功时返回 {}，此时 ret 默认为 0，视为成功）
	var clawbotResp ilinkSendMessageResponse
	if err := json.Unmarshal(respBody, &clawbotResp); err != nil {
		s.logger.Error("解析微信ClawBot响应失败",
			zap.String("raw_body", string(respBody)),
			zap.Error(err),
		)
		return fmt.Errorf("解析微信ClawBot响应失败: %w", err)
	}

	s.logger.Debug("微信ClawBot响应解析",
		zap.Int("ret", clawbotResp.Ret),
		zap.String("errmsg", clawbotResp.ErrMsg),
	)

	// 14. 检查业务状态码
	if clawbotResp.Ret != 0 {
		s.logger.Warn("微信ClawBot推送业务失败",
			zap.Int("ret", clawbotResp.Ret),
			zap.String("errmsg", clawbotResp.ErrMsg),
		)
		return fmt.Errorf("微信ClawBot推送失败: %s (ret: %d)", clawbotResp.ErrMsg, clawbotResp.Ret)
	}

	s.logger.Info("微信ClawBot推送成功")
	return nil
}
