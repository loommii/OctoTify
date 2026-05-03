package model

import "time"

type SourceChannel struct {
	SourceID  int64     `gorm:"primaryKey" json:"source_id"`
	ChannelID int64     `gorm:"primaryKey" json:"channel_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (SourceChannel) TableName() string {
	return "source_channels"
}
