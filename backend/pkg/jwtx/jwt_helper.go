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

// JWTHelper 结构体，封装 JWT 的操作
type JWTHelper struct {
	publicKey     *rsa.PublicKey    // RSA 公钥，用于验证令牌签名
	privateKey    *rsa.PrivateKey   // RSA 私钥，用于生成令牌签名
	expiredTime   time.Duration     // 令牌默认有效期
	signingMethod jwt.SigningMethod // 签名算法（默认 RS256）
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

// GetExpiredTime 获取令牌有效期
func (j *JWTHelper) GetExpiredTime() time.Duration {
	return j.expiredTime
}

// GenerateToken 生成 JWT token
// 注意：
//  1. 此函数入参为值拷贝，不会修改原始 claims
//  2. 此函数会覆盖时间字段（IssuedAt、NotBefore、ExpiresAt）
//  3. 如果 claims.ID 已设置则保留，否则自动生成
func (j *JWTHelper) GenerateToken(claims JWTClaims) (string, error) {
	now := time.Now()

	if claims.ID == "" {
		claims.ID = uuid.New().String()
	}
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(j.expiredTime))
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.NotBefore = jwt.NewNumericDate(now)

	token := jwt.NewWithClaims(j.signingMethod, claims)
	signedToken, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign token")
	}

	return signedToken, nil
}

// ValidateToken 验证 JWT token
func (j *JWTHelper) ValidateToken(tokenString string) (*jwt.Token, *JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse token")
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
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

	privateKeyDir := filepath.Dir(privateKeyPath)
	if err := os.MkdirAll(privateKeyDir, 0700); err != nil {
		return errors.Wrap(err, "failed to create key directory")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "failed to generate RSA private key")
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})

	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		return errors.Wrap(err, "failed to write private key")
	}

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
