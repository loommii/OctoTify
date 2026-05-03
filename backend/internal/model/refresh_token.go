package model

import "time"

// RefreshToken 撤销状态
const (
	RefreshTokenRevoked   = true  // 已撤销
	RefreshTokenNotRevoked = false // 未撤销
)

// RefreshToken 刷新令牌表（用于JWT令牌续期）
type RefreshToken struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`      // 主键ID
	JTI       string    `gorm:"uniqueIndex;size:64;not null" json:"jti"` // 令牌唯一标识（JWT ID），唯一索引
	UserID    int64     `gorm:"index;not null" json:"user_id"`           // 所属用户ID
	Revoked   bool      `gorm:"default:false" json:"revoked"`            // 是否已撤销
	CreatedAt time.Time `json:"created_at"`                              // 创建时间
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
