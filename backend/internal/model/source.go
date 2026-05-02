package model

import (
	"time"
)

const (
	SourceStatusActive   = 1
	SourceStatusDisabled = 2
	SourceStatusDeleted  = -1
)

type Source struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64     `gorm:"index;not null" json:"user_id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Token       string    `gorm:"uniqueIndex;size:64;not null" json:"token"`
	Description string    `gorm:"size:512" json:"description"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
}

func (Source) TableName() string {
	return "sources"
}
