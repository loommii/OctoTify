package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	ChannelStatusActive   = 1
	ChannelStatusDisabled = 2
	ChannelStatusDeleted  = -1
)

type Channel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64          `gorm:"index;not null" json:"user_id"`
	Type      string         `gorm:"size:32;not null" json:"type"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	Config    datatypes.JSON `gorm:"type:json" json:"config"`
	Status    int            `gorm:"default:1" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (Channel) TableName() string {
	return "channels"
}
