package model

import (
	"time"
)

const (
	SourceStatusActive   = 1  // 来源状态：启用
	SourceStatusDisabled = 2  // 来源状态：停用
	SourceStatusDeleted  = -1 // 来源状态：已删除
)

// Source 来源表（消息推送的来源方）
type Source struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`               // 主键ID
	UserID      int64     `gorm:"index;not null" json:"user_id"`                    // 所属用户ID
	Name        string    `gorm:"size:128;not null" json:"name"`                    // 来源名称
	Token       string    `gorm:"uniqueIndex;size:64;not null" json:"token"`        // 来源Token，唯一索引，用于鉴权
	Description string    `gorm:"size:512" json:"description"`                      // 来源描述
	Status      int       `gorm:"default:1" json:"status"`                          // 状态：1-启用 2-停用 -1-已删除
	CreatedAt   time.Time `json:"created_at"`                                        // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`                                        // 更新时间
	LastUsedAt  time.Time `json:"last_used_at"`                                      // 最后使用时间
}

func (Source) TableName() string {
	return "sources"
}
