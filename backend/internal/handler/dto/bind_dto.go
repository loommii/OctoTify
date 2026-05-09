package dto

// GetBindStatusReq 查询绑定状态请求参数
type GetBindStatusReq struct {
	QRCode string `json:"qrcode" binding:"required"` // iLink 返回的二维码原始值
}

// BindCredentialsDTO 绑定成功时返回的凭证信息
type BindCredentialsDTO struct {
	BotTokenCiphertext string `json:"bot_token_ciphertext"` // 加密后的 Bot Token（密文）
	BotTokenNonce      string `json:"bot_token_nonce"`      // Token 加密使用的 Nonce
	IlinkBotID         string `json:"ilink_bot_id"`         // iLink 平台分配的 Bot ID
	IlinkUserID        string `json:"ilink_user_id"`        // iLink 平台的用户 ID
}

// BindStatusResp 查询微信ClawBot绑定状态响应
type BindStatusResp struct {
	Status      string              `json:"status"`                // 绑定状态：pending/scanned/confirmed/expired
	Credentials *BindCredentialsDTO `json:"credentials,omitempty"` // 绑定成功时返回的凭证
}

// StartBindResp 发起微信ClawBot绑定响应
type StartBindResp struct {
	QRCodeURL string `json:"qrcode_url"` // 二维码 URL，用户扫码使用
	QRCode    string `json:"qrcode"`     // 二维码原始值，用于后续查询绑定状态
}
