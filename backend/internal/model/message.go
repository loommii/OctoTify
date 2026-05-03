package model

import (
	"time"
)

const (
	MessageStatusPending = 100  // 消息状态：待推送
	MessageStatusSuccess = 200  // 消息状态：推送成功
	MessageStatusFailed  = 300  // 消息状态：推送失败
	MessageStatusDeleted = -1   // 消息状态：已删除
)

// Message 消息表（推送记录）
type Message struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`    // 主键ID
	SourceID  int64     `gorm:"index;not null" json:"source_id"`       // 来源ID
	ChannelID int64     `gorm:"index;not null" json:"channel_id"`      // 渠道ID
	Title     string    `gorm:"size:256;not null" json:"title"`        // 消息标题
	Content   string    `gorm:"type:text" json:"content"`              // 消息内容
	Status    int       `gorm:"default:100" json:"status"`             // 状态：100-待推送 200-成功 300-失败 -1-已删除
	CreatedAt time.Time `json:"created_at"`                            // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                            // 更新时间
}

func (Message) TableName() string {
	return "messages"
}
