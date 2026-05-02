package model

import (
	"time"
)

const (
	MessageStatusPending = 100
	MessageStatusSuccess = 200
	MessageStatusFailed  = 300
	MessageStatusDeleted = -1
)

type Message struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceID  int64     `gorm:"index;not null" json:"source_id"`
	ChannelID int64     `gorm:"index;not null" json:"channel_id"`
	Title     string    `gorm:"size:256;not null" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	Status    int       `gorm:"default:100" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Message) TableName() string {
	return "messages"
}
