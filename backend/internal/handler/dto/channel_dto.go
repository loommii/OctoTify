package dto

import "gorm.io/datatypes"

const (
	ChannelTypeWechat   = "wechat"
	ChannelTypeTelegram = "telegram"
	ChannelTypeDingtalk = "dingtalk"
	ChannelTypeEmail    = "email"
	ChannelTypeWebhook  = "webhook"
)

var ValidChannelTypes = map[string]bool{
	ChannelTypeWechat:   true,
	ChannelTypeTelegram: true,
	ChannelTypeDingtalk: true,
	ChannelTypeEmail:    true,
	ChannelTypeWebhook:  true,
}

// CreateChannelReq 创建推送渠道请求
// @Description 创建新的推送渠道，需要指定渠道类型、名称和配置信息
// @Description ## 渠道配置示例
// @Description - wechat: {"webhook": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"}
// @Description - telegram: {"bot_token": "xxx", "chat_id": "xxx"}
// @Description - dingtalk: {"webhook": "https://oapi.dingtalk.com/robot/send?access_token=xxx"}
// @Description - email: {"smtp_host": "smtp.example.com", "smtp_port": 587, "username": "xxx", "password": "xxx", "to": "xxx@example.com"}
// @Description - webhook: {"url": "https://example.com/webhook", "method": "POST", "headers": {"Content-Type": "application/json"}}
type CreateChannelReq struct {
	Type   string         `json:"type" binding:"required,channel_type" example:"dingtalk"` // 渠道类型，可选值：wechat, telegram, dingtalk, email, webhook
	Name   string         `json:"name" binding:"required,min=1,max=128" example:"钉钉-运维群"`  // 渠道名称，1-128 个字符
	Config datatypes.JSON `json:"config" binding:"required"`                               // 渠道配置，JSON 格式，不同渠道类型配置不同
}

// UpdateChannelReq 编辑推送渠道请求
// @Description 编辑已有推送渠道的名称和配置信息
// @Description ## 渠道配置示例
// @Description - wechat: {"webhook": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"}
// @Description - telegram: {"bot_token": "xxx", "chat_id": "xxx"}
// @Description - dingtalk: {"webhook": "https://oapi.dingtalk.com/robot/send?access_token=xxx"}
// @Description - email: {"smtp_host": "smtp.example.com", "smtp_port": 587, "username": "xxx", "password": "xxx", "to": "xxx@example.com"}
// @Description - webhook: {"url": "https://example.com/webhook", "method": "POST", "headers": {"Content-Type": "application/json"}}
type UpdateChannelReq struct {
	Name   string         `json:"name" binding:"required,min=1,max=128" example:"钉钉-运维群"` // 渠道名称，1-128 个字符
	Config datatypes.JSON `json:"config" binding:"required"`                              // 渠道配置，JSON 格式，不同渠道类型配置不同
}

// ChannelDTO 推送渠道响应
// @Description 推送渠道的详细信息
type ChannelDTO struct {
	ID         int64          `json:"id" example:"1"`                       // 渠道 ID
	UserID     int64          `json:"user_id" example:"1"`                  // 所属用户 ID
	Type       string         `json:"type" example:"dingtalk"`              // 渠道类型：wechat, telegram, dingtalk, email, webhook
	Name       string         `json:"name" example:"钉钉-运维群"`                // 渠道名称
	Config     datatypes.JSON `json:"config"`                               // 渠道配置，JSON 格式
	Status     int            `json:"status" example:"1"`                   // 状态：1-启用 2-停用 -1-已删除
	CreatedAt  int64          `json:"created_at" example:"1714636800000"`   // 创建时间（Unix 毫秒时间戳）
	UpdatedAt  int64          `json:"updated_at" example:"1714636800000"`   // 更新时间（Unix 毫秒时间戳）
	LastUsedAt int64          `json:"last_used_at" example:"1714636800000"` // 最后使用时间（Unix 毫秒时间戳）
}
