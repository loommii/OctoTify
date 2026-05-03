// Package jwtx 提供基于 RSA 非对称加密的 JWT 令牌生成与验证功能
package jwtx

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	Access  = "access"  // Access Token 类型标识
	Refresh = "refresh" // Refresh Token 类型标识
)

// Claims JWT 自定义声明，包含用户标识和令牌类型
type Claims struct {
	UID       string `json:"uid"`        // 用户唯一标识（如用户ID）
	TokenType string `json:"token_type"` // 令牌类型：access 或 refresh
	jwt.RegisteredClaims
}

// IsAccessToken 判断当前令牌是否为 Access Token
func (c Claims) IsAccessToken() bool {
	return c.TokenType == Access
}

// IsRefreshToken 判断当前令牌是否为 Refresh Token
func (c Claims) IsRefreshToken() bool {
	return c.TokenType == Refresh
}

// JWTHelper JWT 辅助工具，封装 RSA 非对称加密的令牌操作
type JWTHelper struct {
	publicKey     *rsa.PublicKey    // RSA 公钥，用于验证令牌签名
	privateKey    *rsa.PrivateKey   // RSA 私钥，用于生成令牌签名
	expiredTime   time.Duration     // 令牌默认有效期
	signingMethod jwt.SigningMethod // 签名算法（默认 RS256）
}

// Option 函数选项类型，用于灵活配置 JWTHelper
type Option func(*JWTHelper)

// WithPrivateKey 设置 RSA 私钥
func WithPrivateKey(privateKey *rsa.PrivateKey) Option {
	return func(h *JWTHelper) {
		h.privateKey = privateKey
	}
}

// WithPublicKey 设置 RSA 公钥
func WithPublicKey(publicKey *rsa.PublicKey) Option {
	return func(h *JWTHelper) {
		h.publicKey = publicKey
	}
}

// WithExpiredTime 设置令牌默认有效期
func WithExpiredTime(expiredTime time.Duration) Option {
	return func(h *JWTHelper) {
		h.expiredTime = expiredTime
	}
}

// WithSigningMethod 设置签名算法
func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(h *JWTHelper) {
		h.signingMethod = method
	}
}

// NewJWTHelper 使用函数选项模式创建 JWTHelper 实例，默认使用 RS256 算法
func NewJWTHelper(opts ...Option) *JWTHelper {
	helper := &JWTHelper{
		signingMethod: jwt.SigningMethodRS256,
	}
	for _, opt := range opts {
		opt(helper)
	}
	return helper
}

// TokenPair 令牌对，包含 Access Token 和 Refresh Token 以及 Refresh Token 的 Claims
type TokenPair struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	RefreshClaims Claims `json:"-"` // Refresh Token 的声明，用于存储到数据库
}

// GenerateTokenPair 一次性生成 Access Token 和 Refresh Token
func (h *JWTHelper) GenerateTokenPair(uid string) (*TokenPair, error) {
	now := time.Now()
	refreshJTI := uuid.New().String()
	refreshExpiresAt := jwt.NewNumericDate(now.Add(h.expiredTime))

	accessToken, err := h.GenerateToken(Claims{
		UID:       uid,
		TokenType: Access,
	})
	if err != nil {
		return nil, err
	}

	refreshClaims := Claims{
		UID:       uid,
		TokenType: Refresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			ExpiresAt: refreshExpiresAt,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	refreshToken, err := h.GenerateToken(refreshClaims)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		RefreshClaims: refreshClaims,
	}, nil
}

// GenerateToken 根据自定义声明生成单个 JWT 令牌
// 注意：此方法会覆盖 claims 中的时间字段（IssuedAt、NotBefore、ExpiresAt）
func (h *JWTHelper) GenerateToken(claims Claims) (string, error) {
	now := time.Now()

	claims.ID = uuid.New().String()
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(h.expiredTime))
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.NotBefore = jwt.NewNumericDate(now)

	token := jwt.NewWithClaims(h.signingMethod, claims)
	signedToken, err := token.SignedString(h.privateKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign token")
	}

	return signedToken, nil
}

// ValidateToken 验证 JWT 令牌字符串，返回解析后的令牌和自定义声明
// 使用公钥验证签名，支持 5 秒的时间容差
func (h *JWTHelper) ValidateToken(tokenString string) (*jwt.Token, *Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.publicKey, nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return token, claims, nil
	}

	return nil, nil, errors.New("invalid token")
}

// ParseRSAPrivateKeyFromPEM 从 PEM 格式的字节数据解析 RSA 私钥
func ParseRSAPrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse RSA private key")
	}
	return privateKey, nil
}

// ParseRSAPublicKeyFromPEM 从 PEM 格式的字节数据解析 RSA 公钥
func ParseRSAPublicKeyFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse RSA public key")
	}
	return publicKey, nil
}

// EnsureRSAKeyPair 检查密钥文件是否存在，不存在则自动生成 RSA 密钥对
func EnsureRSAKeyPair(privateKeyPath, publicKeyPath string) error {
	privateKeyExists := fileExists(privateKeyPath)
	publicKeyExists := fileExists(publicKeyPath)

	if privateKeyExists && publicKeyExists {
		return nil
	}

	// 确保密钥目录存在
	privateKeyDir := filepath.Dir(privateKeyPath)
	if err := os.MkdirAll(privateKeyDir, 0700); err != nil {
		return errors.Wrap(err, "failed to create key directory")
	}

	// 生成 RSA 私钥（2048 位）
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "failed to generate RSA private key")
	}

	// 序列化私钥为 PEM 格式
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// 序列化公钥为 PEM 格式
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})

	// 写入私钥文件
	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		return errors.Wrap(err, "failed to write private key")
	}

	// 写入公钥文件
	if err := os.WriteFile(publicKeyPath, publicKeyPEM, 0644); err != nil {
		return errors.Wrap(err, "failed to write public key")
	}

	return nil
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
