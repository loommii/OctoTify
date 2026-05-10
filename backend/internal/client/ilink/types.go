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

// Credentials 绑定成功后的凭证
type Credentials struct {
	BotToken    string `json:"bot_token"`     // Bot Token（明文）
	IlinkBotID  string `json:"ilink_bot_id"`  // iLink 平台分配的 Bot ID
	IlinkUserID string `json:"ilink_user_id"` // iLink 平台的用户 ID
}

// QRCodeResponse 调用 iLink API 获取二维码的响应
type QRCodeResponse struct {
	QRCode           string `json:"qrcode"`
	QRCodeImgContent string `json:"qrcode_img_content"`
}

// QRStatusResponse 调用 iLink API 查询二维码状态的响应
type QRStatusResponse struct {
	Status      string `json:"status"`
	BotToken    string `json:"bot_token"`
	ILinkBotID  string `json:"ilink_bot_id"`
	ILinkUserID string `json:"ilink_user_id"`
}
