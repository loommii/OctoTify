package aescipher

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestNewAESEncryptor(t *testing.T) {
	// 测试能否成功创建加密器
	e, err := NewAESEncryptor()
	if err != nil {
		t.Fatalf("创建 AESEncryptor 失败: %v", err)
	}
	if e == nil {
		t.Fatal("AESEncryptor 不应为 nil")
	}
	if len(e.key) != 32 {
		t.Fatalf("AES 密钥长度应为 32 字节，实际为 %d", len(e.key))
	}
}

func TestNewAESEncryptor_UniqueKeys(t *testing.T) {
	// 测试每次创建都生成不同的密钥
	e1, _ := NewAESEncryptor()
	e2, _ := NewAESEncryptor()
	if bytes.Equal(e1.key, e2.key) {
		t.Fatal("两次创建的 AES 密钥不应相同")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	e, err := NewAESEncryptor()
	if err != nil {
		t.Fatalf("创建 AESEncryptor 失败: %v", err)
	}

	plaintext := []byte("test-bot-token-123456")
	ciphertext, nonce, err := e.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatal("密文不应为空")
	}
	if len(nonce) != 12 {
		t.Fatalf("Nonce 长度应为 12 字节，实际为 %d", len(nonce))
	}

	decrypted, err := e.Decrypt(ciphertext, nonce)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("解密后的明文与原始明文不匹配: got %s, want %s", decrypted, plaintext)
	}
}

func TestEncrypt_DifferentNonces(t *testing.T) {
	e, err := NewAESEncryptor()
	if err != nil {
		t.Fatalf("创建 AESEncryptor 失败: %v", err)
	}

	plaintext := []byte("same-plaintext")

	// 同一明文加密两次，应产生不同的密文和 nonce
	ct1, nonce1, err := e.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("第一次加密失败: %v", err)
	}
	ct2, nonce2, err := e.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("第二次加密失败: %v", err)
	}

	if bytes.Equal(nonce1, nonce2) {
		t.Fatal("两次加密的 Nonce 不应相同")
	}
	if bytes.Equal(ct1, ct2) {
		t.Fatal("两次加密的密文不应相同")
	}

	// 两者都应能正确解密
	d1, _ := e.Decrypt(ct1, nonce1)
	d2, _ := e.Decrypt(ct2, nonce2)
	if !bytes.Equal(d1, d2) {
		t.Fatal("解密结果应一致")
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	e, err := NewAESEncryptor()
	if err != nil {
		t.Fatalf("创建 AESEncryptor 失败: %v", err)
	}

	plaintext := []byte("secret-token")
	ciphertext, nonce, _ := e.Encrypt(plaintext)

	// 篡改密文
	ciphertext[0] ^= 0xFF

	_, err = e.Decrypt(ciphertext, nonce)
	if err == nil {
		t.Fatal("篡改后的密文解密应失败")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	e1, _ := NewAESEncryptor()
	e2, _ := NewAESEncryptor()

	plaintext := []byte("secret-token")
	ciphertext, nonce, _ := e1.Encrypt(plaintext)

	// 使用不同的密钥解密
	_, err := e2.Decrypt(ciphertext, nonce)
	if err == nil {
		t.Fatal("使用错误的密钥解密应失败")
	}
}

func TestEncryptBase64_RoundTrip(t *testing.T) {
	e, err := NewAESEncryptor()
	if err != nil {
		t.Fatalf("创建 AESEncryptor 失败: %v", err)
	}

	plaintext := []byte("base64-test-token")
	cipherB64, nonceB64, err := e.EncryptBase64(plaintext)
	if err != nil {
		t.Fatalf("EncryptBase64 失败: %v", err)
	}

	// 验证 Base64 编码有效
	_, err = base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		t.Fatalf("密文 Base64 解码失败: %v", err)
	}
	_, err = base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		t.Fatalf("Nonce Base64 解码失败: %v", err)
	}

	// 解密
	decrypted, err := e.DecryptBase64(cipherB64, nonceB64)
	if err != nil {
		t.Fatalf("DecryptBase64 失败: %v", err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("解密结果不匹配: got %s, want %s", decrypted, plaintext)
	}
}

func TestDecryptBase64_InvalidBase64(t *testing.T) {
	e, _ := NewAESEncryptor()

	_, err := e.DecryptBase64("not-valid-base64!!!", "")
	if err == nil {
		t.Fatal("无效 Base64 应返回错误")
	}
}

func TestGlobalEncryptor(t *testing.T) {
	// 重置全局加密器（通过反射或直接覆盖）
	globalMu.Lock()
	globalEncryptor = nil
	globalMu.Unlock()

	// 未初始化时应返回错误
	_, _, err := GlobalEncryptBase64([]byte("test"))
	if err == nil {
		t.Fatal("未初始化全局加密器时应返回错误")
	}
	_, err = GlobalDecryptBase64("", "")
	if err == nil {
		t.Fatal("未初始化全局加密器时应返回错误")
	}

	// 设置全局加密器
	e, _ := NewAESEncryptor()
	SetGlobalEncryptor(e)

	// 测试全局加密解密
	plaintext := []byte("global-test-token")
	cipherB64, nonceB64, err := GlobalEncryptBase64(plaintext)
	if err != nil {
		t.Fatalf("GlobalEncryptBase64 失败: %v", err)
	}

	decrypted, err := GlobalDecryptBase64(cipherB64, nonceB64)
	if err != nil {
		t.Fatalf("GlobalDecryptBase64 失败: %v", err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("全局加解密结果不匹配: got %s, want %s", decrypted, plaintext)
	}
}

func TestEncrypt_EmptyPlaintext(t *testing.T) {
	e, _ := NewAESEncryptor()

	plaintext := []byte("")
	ciphertext, nonce, err := e.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("加密空数据失败: %v", err)
	}

	decrypted, err := e.Decrypt(ciphertext, nonce)
	if err != nil {
		t.Fatalf("解密空数据失败: %v", err)
	}
	if len(decrypted) != 0 {
		t.Fatal("解密空数据应返回空切片")
	}
}

func TestEncrypt_LargePlaintext(t *testing.T) {
	e, _ := NewAESEncryptor()

	// 测试大文本（1KB）
	plaintext := bytes.Repeat([]byte("a"), 1024)
	ciphertext, nonce, err := e.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("加密大数据失败: %v", err)
	}

	decrypted, err := e.Decrypt(ciphertext, nonce)
	if err != nil {
		t.Fatalf("解密大数据失败: %v", err)
	}
	if !bytes.Equal(plaintext, decrypted) {
		t.Fatal("大文本加解密结果不匹配")
	}
}
