package service

import "time"

// iLink 平台返回的绑定状态值（仅用于测试 mock 和状态判断）
const (
	ILinkStatusWait      = "wait"      // 等待用户扫码
	ILinkStatusScanned   = "scanned"   // 用户已扫码，待确认
	ILinkStatusConfirmed = "confirmed" // 绑定成功
	ILinkStatusExpired   = "expired"   // 二维码已过期
)

// PollAPITimeout 轮询 iLink API 的 HTTP 请求超时
// iLink 的 get_qrcode_status 接口本身就是长轮询设计，需要 35-40 秒才返回
const PollAPITimeout = 40 * time.Second

// BindCredentials 绑定成功后的凭证（后端内部统一使用明文）
type BindCredentials struct {
	BotToken    string `json:"bot_token"`     // Bot Token（明文）
	IlinkBotID  string `json:"ilink_bot_id"`  // iLink 平台分配的 Bot ID
	IlinkUserID string `json:"ilink_user_id"` // iLink 平台的用户 ID
}
