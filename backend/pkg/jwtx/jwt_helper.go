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

// JWTHelper JWT 辅助工具结构体，封装 JWT 令牌的生成与验证操作
type JWTHelper struct {
	publicKey       *rsa.PublicKey    // RSA 公钥，用于验证令牌签名
	privateKey      *rsa.PrivateKey   // RSA 私钥，用于生成令牌签名
	expiredTime     time.Duration     // 令牌默认有效期
	signingMethod   jwt.SigningMethod // 签名算法（默认 RS256）
	keyType         KeyType           // 密钥类型（HMAC 或 RSA）
	secretKey       []byte            // HMAC 密钥
	issuer          string            // 签发者
	defaultTokenType string           // 默认令牌类型
}

// NewJWTHelper 使用函数选项模式创建 JWTHelper 实例
// 默认使用 RS256 签名算法，默认令牌有效期为 1 小时
// 必须通过 Option 配置私钥和公钥
func NewJWTHelper(opts ...Option) *JWTHelper {
	helper := &JWTHelper{
		signingMethod: jwt.SigningMethodRS256,
		expiredTime:   time.Hour, // 默认 1 小时
	}
	for _, opt := range opts {
		opt(helper)
	}
	return helper
}

// GetExpiredTime 获取令牌有效期时长
func (j *JWTHelper) GetExpiredTime() time.Duration {
	return j.expiredTime
}

// GenerateToken 生成 JWT 令牌
// 注意：
//  1. 此函数入参为值拷贝，不会修改原始 claims
//  2. 此函数会覆盖时间字段（IssuedAt、NotBefore、ExpiresAt）
//  3. 如果 claims.ID 已设置则保留，否则自动生成 UUID
func (j *JWTHelper) GenerateToken(claims JWTClaims) (string, error) {
	now := time.Now()

	// 如果未设置令牌唯一标识，则自动生成
	if claims.ID == "" {
		claims.ID = uuid.New().String()
	}
	// 设置令牌过期时间、签发时间和生效时间
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(j.expiredTime))
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.NotBefore = jwt.NewNumericDate(now)

	// 如果设置了签发者，则覆盖
	if j.issuer != "" {
		claims.Issuer = j.issuer
	}

	// 如果未设置令牌类型，使用默认类型
	if claims.TokenType == "" && j.defaultTokenType != "" {
		claims.TokenType = j.defaultTokenType
	}

	// 根据密钥类型选择签名方式
	if j.keyType == HMAC {
		// HMAC 对称加密
		if len(j.secretKey) == 0 {
			return "", errors.New("HMAC secret key is not configured")
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(j.secretKey)
		if err != nil {
			return "", errors.Wrap(err, "failed to sign token with HMAC")
		}
		return signedToken, nil
	}

	// RSA 非对称加密（默认）
	if j.privateKey == nil {
		return "", errors.New("private key is not configured")
	}
	token := jwt.NewWithClaims(j.signingMethod, claims)
	signedToken, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign token")
	}

	return signedToken, nil
}

// ValidateToken 验证 JWT 令牌
// 返回解析后的令牌对象、自定义声明和错误信息
// 验证内容包括：签名有效性、过期时间、令牌格式等
func (j *JWTHelper) ValidateToken(tokenString string) (*jwt.Token, *JWTClaims, error) {
	// 根据密钥类型选择验证方式
	if j.keyType == HMAC {
		// HMAC 对称加密验证
		if len(j.secretKey) == 0 {
			return nil, nil, errors.New("HMAC secret key is not configured")
		}
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
			// 验证签名算法是否为 HMAC 类型
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return j.secretKey, nil
		}, jwt.WithLeeway(5*time.Second))

		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to parse token")
		}

		// 验证令牌是否有效且声明类型正确
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			return token, claims, nil
		}

		return nil, nil, errors.New("invalid token")
	}

	// RSA 非对称加密验证（默认）
	if j.publicKey == nil {
		return nil, nil, errors.New("public key is not configured")
	}

	// 解析并验证令牌，使用公钥验证签名，允许 5 秒的时间误差
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// 验证签名算法是否为 RSA 类型
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse token")
	}

	// 验证令牌是否有效且声明类型正确
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return token, claims, nil
	}

	return nil, nil, errors.New("invalid token")
}

// ParseRSAPrivateKeyFromPEM 从 PEM 格式的字节数据中解析 RSA 私钥
func ParseRSAPrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse RSA private key")
	}
	return privateKey, nil
}

// ParseRSAPublicKeyFromPEM 从 PEM 格式的字节数据中解析 RSA 公钥
func ParseRSAPublicKeyFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse RSA public key")
	}
	return publicKey, nil
}

// EnsureRSAKeyPair 检查密钥文件是否存在，根据 autoGenerate 参数决定是否自动生成 RSA 密钥对
// 参数说明：
//   - privateKeyPath: RSA 私钥文件路径
//   - publicKeyPath: RSA 公钥文件路径
//   - autoGenerate: 是否自动生成密钥对
//   - true: 密钥不存在时自动生成（适合新手用户，开箱即用）
//   - false: 密钥不存在时返回错误（适合生产环境，强制手动管理密钥）
//
// 智能处理逻辑：
//   - 私钥和公钥都存在：直接返回
//   - 私钥存在但公钥丢失：从私钥重新生成公钥（不会覆盖私钥）
//   - 私钥丢失：根据 autoGenerate 决定是生成新密钥对还是报错
func EnsureRSAKeyPair(privateKeyPath, publicKeyPath string, autoGenerate bool) error {
	privateKeyExists := fileExists(privateKeyPath)
	publicKeyExists := fileExists(publicKeyPath)

	// 密钥对完整存在，直接返回
	if privateKeyExists && publicKeyExists {
		return nil
	}

	// 不自动生成时，检查缺失的密钥并返回明确错误
	if !autoGenerate {
		if !privateKeyExists && !publicKeyExists {
			return errors.New("RSA key pair not found, please provide valid key files or enable auto_generate_keys")
		}
		if !privateKeyExists {
			return errors.Errorf("RSA private key not found: %s", privateKeyPath)
		}
		return errors.Errorf("RSA public key not found: %s", publicKeyPath)
	}

	// 确保密钥文件所在目录存在
	privateKeyDir := filepath.Dir(privateKeyPath)
	if err := os.MkdirAll(privateKeyDir, 0700); err != nil {
		return errors.Wrap(err, "failed to create key directory")
	}

	publicKeyDir := filepath.Dir(publicKeyPath)
	if err := os.MkdirAll(publicKeyDir, 0700); err != nil {
		return errors.Wrap(err, "failed to create key directory")
	}

	var privateKey *rsa.PrivateKey
	var err error

	// 私钥已存在，读取并解析私钥（用于生成公钥）
	if privateKeyExists {
		privateKeyPEM, readErr := os.ReadFile(privateKeyPath)
		if readErr != nil {
			return errors.Wrap(readErr, "failed to read existing private key")
		}
		var parseErr error
		privateKey, parseErr = ParseRSAPrivateKeyFromPEM(privateKeyPEM)
		if parseErr != nil {
			return errors.Wrap(parseErr, "failed to parse existing private key")
		}
	} else {
		// 私钥不存在，生成新的 RSA 密钥对
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return errors.Wrap(err, "failed to generate RSA private key")
		}

		// 将私钥编码为 PEM 格式并写入文件
		privateKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})

		if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
			return errors.Wrap(err, "failed to write private key")
		}
	}

	// 公钥不存在，从私钥生成公钥并写入文件
	if !publicKeyExists {
		publicKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
		})

		if err := os.WriteFile(publicKeyPath, publicKeyPEM, 0644); err != nil {
			return errors.Wrap(err, "failed to write public key")
		}
	}

	return nil
}

// fileExists 检查指定路径的文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
