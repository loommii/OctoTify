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

// WithKeyType 设置密钥类型（HMAC 或 RSA）
func WithKeyType(keyType KeyType) Option {
	return func(helper *JWTHelper) {
		helper.keyType = keyType
	}
}

// WithSecret 设置 HMAC 密钥
func WithSecret(secret []byte) Option {
	return func(helper *JWTHelper) {
		helper.secretKey = secret
	}
}

// WithTTL 设置令牌有效期（覆盖默认 1 小时）
func WithTTL(ttl time.Duration) Option {
	return func(helper *JWTHelper) {
		helper.expiredTime = ttl
	}
}

// WithIssuer 设置签发者
func WithIssuer(issuer string) Option {
	return func(helper *JWTHelper) {
		helper.issuer = issuer
	}
}

// WithTokenType 设置默认令牌类型（access 或 refresh）
func WithTokenType(tokenType string) Option {
	return func(helper *JWTHelper) {
		helper.defaultTokenType = tokenType
	}
}
