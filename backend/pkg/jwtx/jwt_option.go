package jwtx

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Option 函数类型，用于配置 JWTHelper
type Option func(*JWTHelper)

// WithPrivateKey 设置 RSA 私钥
func WithPrivateKey(privateKey *rsa.PrivateKey) Option {
	return func(helper *JWTHelper) {
		helper.privateKey = privateKey
	}
}

// WithPublicKey 设置 RSA 公钥
func WithPublicKey(publicKey *rsa.PublicKey) Option {
	return func(helper *JWTHelper) {
		helper.publicKey = publicKey
	}
}

// WithExpiredTime 设置令牌默认有效期
func WithExpiredTime(expiredTime time.Duration) Option {
	return func(helper *JWTHelper) {
		helper.expiredTime = expiredTime
	}
}

// WithSigningMethod 设置签名算法
func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(helper *JWTHelper) {
		helper.signingMethod = method
	}
}
