package jwtx

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 测试辅助函数：生成 RSA 密钥对用于测试
func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(nil, 2048)
	if err != nil {
		t.Fatalf("failed to generate test key pair: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}

// ==================== NewJWTHelper 和 Option 相关测试 ====================

func TestNewJWTHelper_DefaultValues(t *testing.T) {
	helper := NewJWTHelper()

	if helper.signingMethod != jwt.SigningMethodRS256 {
		t.Errorf("expected default signing method RS256, got %v", helper.signingMethod)
	}

	if helper.expiredTime != time.Hour {
		t.Errorf("expected default expired time 1h, got %v", helper.expiredTime)
	}

	if helper.privateKey != nil {
		t.Error("expected default private key to be nil")
	}

	if helper.publicKey != nil {
		t.Error("expected default public key to be nil")
	}
}

func TestNewJWTHelper_WithAllOptions(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	customTTL := 2 * time.Hour
	customMethod := jwt.SigningMethodRS512

	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
		WithExpiredTime(customTTL),
		WithSigningMethod(customMethod),
	)

	if helper.privateKey != privateKey {
		t.Error("expected private key to be set")
	}

	if helper.publicKey != publicKey {
		t.Error("expected public key to be set")
	}

	if helper.expiredTime != customTTL {
		t.Errorf("expected expired time %v, got %v", customTTL, helper.expiredTime)
	}

	if helper.signingMethod != customMethod {
		t.Errorf("expected signing method %v, got %v", customMethod, helper.signingMethod)
	}
}

func TestNewJWTHelper_PartialOptions(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)

	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
	)

	if helper.privateKey != privateKey {
		t.Error("expected private key to be set")
	}

	if helper.expiredTime != time.Hour {
		t.Errorf("expected default expired time 1h, got %v", helper.expiredTime)
	}

	if helper.signingMethod != jwt.SigningMethodRS256 {
		t.Errorf("expected default signing method RS256, got %v", helper.signingMethod)
	}
}

func TestGetExpiredTime(t *testing.T) {
	customTTL := 30 * time.Minute
	helper := NewJWTHelper(WithExpiredTime(customTTL))

	if helper.GetExpiredTime() != customTTL {
		t.Errorf("expected expired time %v, got %v", customTTL, helper.GetExpiredTime())
	}
}

// ==================== GenerateToken 相关测试 ====================

func TestGenerateToken_AccessToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
		WithExpiredTime(time.Hour),
	)

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	token, err := helper.GenerateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Error("expected token to be non-empty")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected token to have 3 parts, got %d", len(parts))
	}
}

func TestGenerateToken_RefreshToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
		WithExpiredTime(time.Hour),
	)

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Refresh,
	}

	token, err := helper.GenerateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Error("expected token to be non-empty")
	}
}

func TestGenerateToken_PreserveExistingID(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
	)

	existingID := "my-custom-id"
	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: existingID,
		},
	}

	token, err := helper.GenerateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, parsedClaims, err := helper.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedClaims.ID != existingID {
		t.Errorf("expected ID %s, got %s", existingID, parsedClaims.ID)
	}
}

func TestGenerateToken_AutoGenerateID(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
	)

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	token, err := helper.GenerateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, parsedClaims, err := helper.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedClaims.ID == "" {
		t.Error("expected auto-generated ID to be non-empty")
	}
}

func TestGenerateToken_NilPrivateKey(t *testing.T) {
	helper := NewJWTHelper(WithExpiredTime(time.Hour))

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	_, err := helper.GenerateToken(claims)
	if err == nil {
		t.Fatal("expected error when private key is nil")
	}

	if !strings.Contains(err.Error(), "private key is not configured") {
		t.Errorf("expected error message about private key, got: %v", err)
	}
}

func TestGenerateToken_ClaimsFields(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
		WithExpiredTime(time.Hour),
	)

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	token, err := helper.GenerateToken(claims)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, parsedClaims, err := helper.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedClaims.UID != "user123" {
		t.Errorf("expected UID user123, got %s", parsedClaims.UID)
	}

	if parsedClaims.TokenType != Access {
		t.Errorf("expected token type %s, got %s", Access, parsedClaims.TokenType)
	}

	if parsedClaims.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set")
	}

	if parsedClaims.IssuedAt == nil {
		t.Fatal("expected IssuedAt to be set")
	}

	if parsedClaims.NotBefore == nil {
		t.Fatal("expected NotBefore to be set")
	}
}

// ==================== ValidateToken 相关测试 ====================

func TestValidateToken_ValidToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
		WithExpiredTime(time.Hour),
	)

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	token, err := helper.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, parsedClaims, err := helper.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedClaims.UID != "user123" {
		t.Errorf("expected UID user123, got %s", parsedClaims.UID)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
		WithExpiredTime(-time.Hour),
	)

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	token, err := helper.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, _, err = helper.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateToken_TamperedToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
		WithExpiredTime(time.Hour),
	)

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	token, err := helper.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tamperedToken := token[:len(token)-5] + "XXXXX"

	_, _, err = helper.ValidateToken(tamperedToken)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
	)

	_, _, err := helper.ValidateToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
	)

	_, _, err := helper.ValidateToken("invalid-token-format")
	if err == nil {
		t.Fatal("expected error for invalid token format")
	}
}

func TestValidateToken_NilPublicKey(t *testing.T) {
	privateKey, _ := generateTestKeyPair(t)
	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithExpiredTime(time.Hour),
	)

	_, _, err := helper.ValidateToken("some.token.here")
	if err == nil {
		t.Fatal("expected error when public key is nil")
	}

	if !strings.Contains(err.Error(), "public key is not configured") {
		t.Errorf("expected error message about public key, got: %v", err)
	}
}

func TestValidateToken_WrongPublicKey(t *testing.T) {
	privateKey1, publicKey1 := generateTestKeyPair(t)
	_, publicKey2 := generateTestKeyPair(t)

	helper1 := NewJWTHelper(
		WithPrivateKey(privateKey1),
		WithPublicKey(publicKey1),
		WithExpiredTime(time.Hour),
	)

	helper2 := NewJWTHelper(
		WithPublicKey(publicKey2),
	)

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	token, err := helper1.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, _, err = helper2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error when using wrong public key")
	}
}

// ==================== JWTClaims 方法相关测试 ====================

func TestJWTClaims_IsAccessToken(t *testing.T) {
	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	if !claims.IsAccessToken() {
		t.Error("expected IsAccessToken to return true for access token")
	}
}

func TestJWTClaims_IsAccessToken_NotAccess(t *testing.T) {
	claims := JWTClaims{
		UID:       "user123",
		TokenType: Refresh,
	}

	if claims.IsAccessToken() {
		t.Error("expected IsAccessToken to return false for refresh token")
	}
}

func TestJWTClaims_IsRefreshToken(t *testing.T) {
	claims := JWTClaims{
		UID:       "user123",
		TokenType: Refresh,
	}

	if !claims.IsRefreshToken() {
		t.Error("expected IsRefreshToken to return true for refresh token")
	}
}

func TestJWTClaims_IsRefreshToken_NotRefresh(t *testing.T) {
	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	if claims.IsRefreshToken() {
		t.Error("expected IsRefreshToken to return false for access token")
	}
}

// ==================== EnsureRSAKeyPair 相关测试 ====================

func TestEnsureRSAKeyPair_BothKeysExist(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	privateKeyPEM, err := os.CreateTemp(tempDir, "private_*.pem")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	privateKeyPath = privateKeyPEM.Name()
	privateKeyPEM.Close()

	publicKeyPEM, err := os.CreateTemp(tempDir, "public_*.pem")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	publicKeyPath = publicKeyPEM.Name()
	publicKeyPEM.Close()

	err = EnsureRSAKeyPair(privateKeyPath, publicKeyPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureRSAKeyPair_AutoGenerate_BothMissing(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "keys", "private.pem")
	publicKeyPath := filepath.Join(tempDir, "keys", "public.pem")

	err := EnsureRSAKeyPair(privateKeyPath, publicKeyPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Error("expected private key file to be created")
	}

	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		t.Error("expected public key file to be created")
	}
}

func TestEnsureRSAKeyPair_NoAutoGenerate_BothMissing(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	err := EnsureRSAKeyPair(privateKeyPath, publicKeyPath, false)
	if err == nil {
		t.Fatal("expected error when keys are missing and autoGenerate is false")
	}
}

func TestEnsureRSAKeyPair_AutoGenerate_PrivateKeyExists_PublicKeyMissing(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	privateKeyPEM, _ := generatePEM(generateTestKeyPairHelper(), false)

	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		t.Fatalf("failed to write private key: %v", err)
	}

	err := EnsureRSAKeyPair(privateKeyPath, publicKeyPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		t.Error("expected public key file to be created")
	}
}

func TestEnsureRSAKeyPair_NoAutoGenerate_PrivateKeyExists_PublicKeyMissing(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	privateKeyPEM, _ := generatePEM(generateTestKeyPairHelper(), false)
	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		t.Fatalf("failed to write private key: %v", err)
	}

	err := EnsureRSAKeyPair(privateKeyPath, publicKeyPath, false)
	if err == nil {
		t.Fatal("expected error when public key is missing and autoGenerate is false")
	}

	if !strings.Contains(err.Error(), "public key not found") {
		t.Errorf("expected error message about public key, got: %v", err)
	}
}

func TestEnsureRSAKeyPair_NoAutoGenerate_PrivateKeyMissing_PublicKeyExists(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	publicKeyPEM, _ := generatePEM(generateTestKeyPairHelper(), true)
	if err := os.WriteFile(publicKeyPath, publicKeyPEM, 0644); err != nil {
		t.Fatalf("failed to write public key: %v", err)
	}

	err := EnsureRSAKeyPair(privateKeyPath, publicKeyPath, false)
	if err == nil {
		t.Fatal("expected error when private key is missing and autoGenerate is false")
	}

	if !strings.Contains(err.Error(), "private key not found") {
		t.Errorf("expected error message about private key, got: %v", err)
	}
}

func TestEnsureRSAKeyPair_AutoCreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "nested", "dir", "private.pem")
	publicKeyPath := filepath.Join(tempDir, "nested", "dir", "public.pem")

	err := EnsureRSAKeyPair(privateKeyPath, publicKeyPath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dirInfo, err := os.Stat(filepath.Dir(privateKeyPath))
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}

	if !dirInfo.IsDir() {
		t.Error("expected directory to be created")
	}
}

func TestEnsureRSAKeyPair_GeneratedKeysMatch(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	err := EnsureRSAKeyPair(privateKeyPath, publicKeyPath, true)
	if err != nil {
		t.Fatalf("failed to ensure key pair: %v", err)
	}

	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		t.Fatalf("failed to read private key: %v", err)
	}

	publicKeyPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		t.Fatalf("failed to read public key: %v", err)
	}

	privateKey, err := ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	publicKey, err := ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	if privateKey.N.Cmp(publicKey.N) != 0 {
		t.Error("expected private and public key modulus to match")
	}
}

// ==================== PEM 解析相关测试 ====================

func TestParseRSAPrivateKeyFromPEM_ValidKey(t *testing.T) {
	privateKey, _ := generateTestKeyPair(t)
	pemBytes, _ := generatePEM(privateKey, false)

	parsedKey, err := ParseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedKey.N.Cmp(privateKey.N) != 0 {
		t.Error("expected parsed key to match original key")
	}
}

func TestParseRSAPublicKeyFromPEM_ValidKey(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)
	pemBytes, _ := generatePEM(publicKey, true)

	parsedKey, err := ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedKey.N.Cmp(publicKey.N) != 0 {
		t.Error("expected parsed key to match original key")
	}
}

func TestParseRSAPrivateKeyFromPEM_InvalidData(t *testing.T) {
	_, err := ParseRSAPrivateKeyFromPEM([]byte("invalid data"))
	if err == nil {
		t.Fatal("expected error for invalid PEM data")
	}
}

func TestParseRSAPublicKeyFromPEM_EmptyData(t *testing.T) {
	_, err := ParseRSAPublicKeyFromPEM([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestParseRSAPrivateKeyFromPEM_PublicKeyAsPrivate(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)
	pemBytes, _ := generatePEM(publicKey, true)

	_, err := ParseRSAPrivateKeyFromPEM(pemBytes)
	if err == nil {
		t.Fatal("expected error when parsing public key as private key")
	}
}

// ==================== 集成测试 ====================

func TestFullWorkflow_GenerateAndValidate(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	err := EnsureRSAKeyPair(privateKeyPath, publicKeyPath, true)
	if err != nil {
		t.Fatalf("failed to ensure key pair: %v", err)
	}

	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		t.Fatalf("failed to read private key: %v", err)
	}

	publicKeyPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		t.Fatalf("failed to read public key: %v", err)
	}

	privateKey, err := ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	publicKey, err := ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	helper := NewJWTHelper(
		WithPrivateKey(privateKey),
		WithPublicKey(publicKey),
		WithExpiredTime(time.Hour),
	)

	claims := JWTClaims{
		UID:       "user123",
		TokenType: Access,
	}

	token, err := helper.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, parsedClaims, err := helper.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if parsedClaims.UID != "user123" {
		t.Errorf("expected UID user123, got %s", parsedClaims.UID)
	}

	if parsedClaims.TokenType != Access {
		t.Errorf("expected token type %s, got %s", Access, parsedClaims.TokenType)
	}

	if !parsedClaims.IsAccessToken() {
		t.Error("expected IsAccessToken to return true")
	}
}

// 测试辅助函数：生成简单的 RSA 私钥
func generateTestKeyPairHelper() *rsa.PrivateKey {
	privateKey, _ := rsa.GenerateKey(nil, 2048)
	return privateKey
}

// 生成 PEM 格式的密钥
func generatePEM(key any, isPublic bool) ([]byte, error) {
	if isPublic {
		pubKey, ok := key.(*rsa.PublicKey)
		if !ok {
			privateKey := key.(*rsa.PrivateKey)
			pubKey = &privateKey.PublicKey
		}
		return pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pubKey),
		}), nil
	}

	privateKey := key.(*rsa.PrivateKey)
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}), nil
}
