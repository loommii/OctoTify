// Package jwtx 提供基于 RSA 非对称加密的 JWT 令牌生成与验证功能
package jwtx

import "github.com/golang-jwt/jwt/v5"

/*
这里存放的该项目自己用到的一些内容
*/

const (
	Access  = "access"  // Access Token 类型标识
	Refresh = "refresh" // Refresh Token 类型标识
)

// JWTClaims JWT 自定义声明，包含用户标识和令牌类型
type JWTClaims struct {
	UID       string `json:"uid"`        // 用户唯一标识（如用户ID）
	TokenType string `json:"token_type"` // 令牌类型：access 或 refresh
	jwt.RegisteredClaims
}

// IsAccessToken 判断当前令牌是否为 Access Token
func (j JWTClaims) IsAccessToken() bool {
	return j.TokenType == Access
}

// IsRefreshToken 判断当前令牌是否为 Refresh Token
func (j JWTClaims) IsRefreshToken() bool {
	return j.TokenType == Refresh
}
