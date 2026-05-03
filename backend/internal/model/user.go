package model

import (
	"time"
)

// User 用户表
type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`           // 主键ID
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"` // 用户名，唯一索引
	PasswordHash string    `gorm:"size:255;not null" json:"-"`                   // 密码哈希，JSON序列化时忽略
	CreatedAt    time.Time `json:"created_at"`                                   // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`                                   // 更新时间
}

func (User) TableName() string {
	return "users"
}
