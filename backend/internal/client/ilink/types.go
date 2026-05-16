package ilink

import "time"

// iLink 平台返回的绑定状态值
const (
	StatusWait      = "wait"      // 等待用户扫码
	StatusScanned   = "scanned"   // 用户已扫码，待确认
	StatusConfirmed = "confirmed" // 绑定成功
	StatusExpired   = "expired"   // 二维码已过期
)

// PollAPITimeout 轮询 iLink API 的 HTTP 请求超时
// iLink 的 get_qrcode_status 接口本身就是长轮询设计，需要 35-40 秒才返回
const PollAPITimeout = 40 * time.Second

// SendAPITimeout 发送消息 API 的 HTTP 请求超时
// 与 weclaw 项目保持一致，设置为 15 秒
const SendAPITimeout = 15 * time.Second

// GetUpdatesTimeout 获取消息 API 的 HTTP 请求超时
// iLink 的 getupdates 接口是长轮询设计，服务端最长挂起 35 秒
// 与 weclaw 保持一致（35+5=40 秒），额外留 5 秒缓冲
const GetUpdatesTimeout = 45 * time.Second

// Credentials 绑定成功后的凭证
// 包含调用 iLink API 所需的全部认证信息
type Credentials struct {
	BotToken    string `json:"bot_token"`     // Bot Token（明文），用于 API 认证
	IlinkBotID  string `json:"ilink_bot_id"`  // iLink 平台分配的 Bot ID，格式如 "bot123@im.bot"
	IlinkUserID string `json:"ilink_user_id"` // iLink 平台的用户 ID，格式如 "user123@im.wechat"
}

// SendMessageRequest 发送消息请求体
// 对应 iLink API POST /ilink/bot/sendmessage 的请求结构
type SendMessageRequest struct {
	Msg      SendMsg  `json:"msg"`       // 消息内容
	BaseInfo BaseInfo `json:"base_info"` // 基础信息（如渠道版本号）
}

// SendMsg 消息体
// 包含消息的发送者、接收者、类型和具体内容
type SendMsg struct {
	FromUserID   string        `json:"from_user_id"`   // 发送者 ID（Bot ID）
	ToUserID     string        `json:"to_user_id"`     // 接收者 ID（用户 ID）
	ClientID     string        `json:"client_id"`      // 客户端关联 ID，每次发送应生成唯一值，防止 iLink 消息去重
	MessageType  int           `json:"message_type"`   // 消息类型：1=用户消息，2=Bot 消息
	MessageState int           `json:"message_state"`  // 消息状态：0=新建，1=生成中，2=完成
	ItemList     []MessageItem `json:"item_list"`      // 消息项列表（文本、图片、语音等）
}

// MessageItem 消息项
// 一条消息可以包含多个不同类型的消息项
type MessageItem struct {
	Type     int       `json:"type"`                 // 消息项类型：0=无，1=文本，2=图片，3=语音，4=文件，5=视频
	TextItem *TextItem `json:"text_item,omitempty"` // 文本消息项
}

// TextItem 文本消息项
type TextItem struct {
	Text string `json:"text"` // 文本内容
}

// BaseInfo 基础信息
// 随请求体发送的元数据
type BaseInfo struct {
	ChannelVersion string `json:"channel_version,omitempty"` // 渠道版本号，如 "1.0.0"
}

// SendMessageResponse 发送消息响应体
// iLink API 返回的发送结果
type SendMessageResponse struct {
	Ret     int    `json:"ret"`               // 保留字段，JSON 缺失时 Go 零值为 0，不可用于判断成功/失败
	ErrCode int    `json:"errcode,omitempty"` // 业务错误码：0 或不存在表示成功，非 0 表示失败
	ErrMsg  string `json:"errmsg,omitempty"`  // 错误信息，仅在失败时返回
}

// QRCodeResponse 调用 iLink API 获取二维码的响应
type QRCodeResponse struct {
	RetCode          int    `json:"ret"`                // 返回码，0 表示成功
	QRCode           string `json:"qrcode"`             // 二维码原始值
	QRCodeImgContent string `json:"qrcode_img_content"` // 二维码图片 URL
}

// QRStatusResponse 调用 iLink API 查询二维码状态的响应
type QRStatusResponse struct {
	Status      string `json:"status"`
	BotToken    string `json:"bot_token"`
	ILinkBotID  string `json:"ilink_bot_id"`
	ILinkUserID string `json:"ilink_user_id"`
}

// GetUpdatesRequest 获取用户消息请求体
// 对应 iLink API POST /bot/getupdates 的请求结构
type GetUpdatesRequest struct {
	GetUpdatesBuf string   `json:"get_updates_buf"`
	BaseInfo      BaseInfo `json:"base_info"`
}

// GetUpdatesResponse 获取用户消息响应体
type GetUpdatesResponse struct {
	Ret                  int             `json:"ret"`
	ErrCode              int             `json:"errcode,omitempty"`
	ErrMsg               string          `json:"errmsg,omitempty"`
	Msgs                 []WeixinMessage `json:"msgs"`
	GetUpdatesBuf        string          `json:"get_updates_buf"`
	LongPollingTimeoutMs int             `json:"longpolling_timeout_ms,omitempty"`
}

// WeixinMessage 微信消息结构
type WeixinMessage struct {
	FromUserID   string        `json:"from_user_id"`
	ToUserID     string        `json:"to_user_id"`
	MessageType  int           `json:"message_type"`
	MessageState int           `json:"message_state"`
	ItemList     []MessageItem `json:"item_list"`
	ContextToken string        `json:"context_token"`
}
