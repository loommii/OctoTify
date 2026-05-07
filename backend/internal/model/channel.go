package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	ChannelStatusActive   = 1  // 渠道状态：启用
	ChannelStatusDisabled = 2  // 渠道状态：停用
	ChannelStatusDeleted  = -1 // 渠道状态：已删除
)

// Channel 渠道表（如钉钉、飞书、邮件等推送渠道）
type Channel struct {
	ID         int64          `gorm:"primaryKey;autoIncrement" json:"id"` // 主键ID
	UserID     int64          `gorm:"index;not null" json:"user_id"`      // 所属用户ID
	Type       string         `gorm:"size:32;not null" json:"type"`       // 渠道类型（dingtalk/email/webhook等）
	Name       string         `gorm:"size:128;not null" json:"name"`      // 渠道名称
	Config     datatypes.JSON `gorm:"type:json" json:"config"`            // 渠道配置（JSON格式，各类型不同）
	Status     int            `gorm:"default:1" json:"status"`            // 状态：1-启用 2-停用 -1-已删除
	CreatedAt  time.Time      `json:"created_at"`                         // 创建时间
	UpdatedAt  time.Time      `json:"updated_at"`                         // 更新时间
	LastUsedAt *time.Time     `json:"last_used_at"`                       // 最后使用时间
}

func (Channel) TableName() string {
	return "channels"
}
