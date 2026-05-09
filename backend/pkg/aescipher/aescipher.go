package aescipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
)

var (
	globalEncryptor *AESEncryptor
	globalMu        sync.RWMutex
)

func init() {
	encryptor, err := NewAESEncryptor()
	if err != nil {
		panic("初始化 AES 加密器失败: " + err.Error())
	}
	globalEncryptor = encryptor
}

// AESEncryptor AES-256-GCM 加密器
// 密钥仅驻留内存，不写入磁盘，服务重启后自动失效
type AESEncryptor struct {
	key []byte // 32 字节 AES-256 密钥
}

// NewAESEncryptor 创建加密器实例，使用 crypto/rand 生成 32 字节随机密钥
func NewAESEncryptor() (*AESEncryptor, error) {
	key := make([]byte, 32) // AES-256 需要 32 字节密钥
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("生成 AES 密钥失败: %w", err)
	}
	return &AESEncryptor{key: key}, nil
}

// SetGlobalEncryptor 设置全局加密器实例（线程安全）
func SetGlobalEncryptor(e *AESEncryptor) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalEncryptor = e
}

// GetGlobalEncryptor 获取全局加密器实例（线程安全）
func GetGlobalEncryptor() *AESEncryptor {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalEncryptor
}

// Encrypt 使用 AES-256-GCM 加密明文数据
// 返回 (ciphertext, nonce, error)
// ciphertext 已包含 GCM 的认证标签（tag）
func (e *AESEncryptor) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("创建 GCM 模式失败: %w", err)
	}

	// 生成 12 字节（96 bit）随机 Nonce（GCM 推荐长度）
	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("生成 Nonce 失败: %w", err)
	}

	// Seal 将明文加密，并在密文末尾附加认证标签
	// 返回值格式: ciphertext（含 tag），不含 nonce
	sealed := gcm.Seal(nil, nonce, plaintext, nil)
	return sealed, nonce, nil
}

// Decrypt 使用 AES-256-GCM 解密密文数据
// 自动验证完整性，若密文被篡改或密钥不匹配则返回错误
func (e *AESEncryptor) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建 GCM 模式失败: %w", err)
	}

	// Open 方法同时解密和验证完整性
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("解密失败（密文可能被篡改或密钥不匹配）: %w", err)
	}

	return plaintext, nil
}

// EncryptBase64 加密并返回 Base64 编码的密文和 Nonce
// 适用于需要通过 JSON 传输的场景
func (e *AESEncryptor) EncryptBase64(plaintext []byte) (ciphertextB64, nonceB64 string, err error) {
	ciphertext, nonce, err := e.Encrypt(plaintext)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), base64.StdEncoding.EncodeToString(nonce), nil
}

// DecryptBase64 解密 Base64 编码的密文和 Nonce
func (e *AESEncryptor) DecryptBase64(ciphertextB64, nonceB64 string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("解码密文失败: %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return nil, fmt.Errorf("解码 Nonce 失败: %w", err)
	}
	return e.Decrypt(ciphertext, nonce)
}

// GlobalEncryptBase64 使用全局加密器加密并返回 Base64 编码
func GlobalEncryptBase64(plaintext []byte) (ciphertextB64, nonceB64 string, err error) {
	e := GetGlobalEncryptor()
	if e == nil {
		return "", "", fmt.Errorf("全局 AES 加密器未初始化")
	}
	return e.EncryptBase64(plaintext)
}

// GlobalDecryptBase64 使用全局加密器解密 Base64 编码的密文
func GlobalDecryptBase64(ciphertextB64, nonceB64 string) ([]byte, error) {
	e := GetGlobalEncryptor()
	if e == nil {
		return nil, fmt.Errorf("全局 AES 加密器未初始化")
	}
	return e.DecryptBase64(ciphertextB64, nonceB64)
}
