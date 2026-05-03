package model

import "time"

// SourceChannel 来源与渠道的关联表（多对多）
type SourceChannel struct {
	SourceID  int64     `gorm:"primaryKey" json:"source_id"`   // 来源ID，联合主键
	ChannelID int64     `gorm:"primaryKey" json:"channel_id"`  // 渠道ID，联合主键
	CreatedAt time.Time `json:"created_at"`                    // 创建时间
}

func (SourceChannel) TableName() string {
	return "source_channels"
}
