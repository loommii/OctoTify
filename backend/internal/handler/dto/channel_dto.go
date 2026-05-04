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
type CreateChannelReq struct {
	Type   string         `json:"type" binding:"required,channel_type" example:"dingtalk"`                                                     // 渠道类型
	Name   string         `json:"name" binding:"required,min=1,max=128" example:"钉钉-运维群"`                                                      // 渠道名称
	Config datatypes.JSON `json:"config" binding:"required" example:"{\"webhook\":\"https://oapi.dingtalk.com/robot/send?access_token=xxx\"}"` // 渠道配置
}

// UpdateChannelReq 编辑推送渠道请求
type UpdateChannelReq struct {
	Name   string         `json:"name" binding:"required,min=1,max=128" example:"钉钉-运维群"`                                                      // 渠道名称
	Config datatypes.JSON `json:"config" binding:"required" example:"{\"webhook\":\"https://oapi.dingtalk.com/robot/send?access_token=xxx\"}"` // 渠道配置
}

// ChannelDTO 推送渠道响应
type ChannelDTO struct {
	ID         int64          `json:"id" example:"1"`                       // 渠道 ID
	UserID     int64          `json:"user_id" example:"1"`                  // 所属用户 ID
	Type       string         `json:"type" example:"dingtalk"`              // 渠道类型
	Name       string         `json:"name" example:"钉钉-运维群"`                // 渠道名称
	Config     datatypes.JSON `json:"config"`                               // 渠道配置
	Status     int            `json:"status" example:"1"`                   // 状态：1-启用 2-停用 -1-已删除
	CreatedAt  int64          `json:"created_at" example:"1714636800000"`   // 创建时间（Unix 毫秒时间戳）
	UpdatedAt  int64          `json:"updated_at" example:"1714636800000"`   // 更新时间（Unix 毫秒时间戳）
	LastUsedAt int64          `json:"last_used_at" example:"1714636800000"` // 最后使用时间（Unix 毫秒时间戳）
}
