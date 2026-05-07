package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"octotify/pkg/jwtx"
	"octotify/pkg/response"
	"octotify/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestJWTAuth_MissingAuthHeader 测试缺少 Authorization 请求头
func TestJWTAuth_MissingAuthHeader(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
	assert.Equal(t, "未提供认证令牌", resp.Msg)
}

// TestJWTAuth_EmptyAuthHeader 测试 Authorization 请求头为空
func TestJWTAuth_EmptyAuthHeader(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
	assert.Equal(t, "未提供认证令牌", resp.Msg)
}

// TestJWTAuth_InvalidTokenFormat 测试 Token 格式错误（无法分割为 Bearer 和 token）
func TestJWTAuth_InvalidTokenFormat(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "invalid-format-no-space")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
	assert.Equal(t, "认证令牌格式错误", resp.Msg)
}

// TestJWTAuth_NonBearerScheme 测试非 Bearer 认证方案
func TestJWTAuth_NonBearerScheme(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
	assert.Equal(t, "认证令牌格式错误", resp.Msg)
}

// TestJWTAuth_EmptyTokenAfterBearer 测试 Bearer 后面没有 token
func TestJWTAuth_EmptyTokenAfterBearer(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJWTAuth_InvalidTokenSignature 测试无效的 Token 签名
func TestJWTAuth_InvalidTokenSignature(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	claims := jwtx.JWTClaims{
		UID:       "user123",
		TokenType: jwtx.Access,
	}
	token, err := helper.GenerateToken(claims)
	assert.NoError(t, err)

	tamperedToken := token[:len(token)-5] + "XXXXX"

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tamperedToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
	assert.Equal(t, "认证令牌无效或已过期", resp.Msg)
}

// TestJWTAuth_ExpiredToken 测试已过期的 Token
func TestJWTAuth_ExpiredToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	expiredHelper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(-time.Hour),
	)

	claims := jwtx.JWTClaims{
		UID:       "user123",
		TokenType: jwtx.Access,
	}
	token, err := expiredHelper.GenerateToken(claims)
	assert.NoError(t, err)

	validHelper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(validHelper))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
	assert.Equal(t, "认证令牌无效或已过期", resp.Msg)
}

// TestJWTAuth_RefreshTokenType 测试使用 Refresh Token 类型（不允许访问）
func TestJWTAuth_RefreshTokenType(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	claims := jwtx.JWTClaims{
		UID:       "user123",
		TokenType: jwtx.Refresh,
	}
	token, err := helper.GenerateToken(claims)
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
	assert.Equal(t, "无效的令牌类型", resp.Msg)
}

// TestJWTAuth_ValidAccessToken 测试有效的 Access Token 可以正常访问
func TestJWTAuth_ValidAccessToken(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	claims := jwtx.JWTClaims{
		UID:       "user123",
		TokenType: jwtx.Access,
	}
	token, err := helper.GenerateToken(claims)
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestJWTAuth_UserIDInContext 测试验证通过后用户 ID 被正确注入到 Context
func TestJWTAuth_UserIDInContext(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	claims := jwtx.JWTClaims{
		UID:       "user123",
		TokenType: jwtx.Access,
	}
	token, err := helper.GenerateToken(claims)
	assert.NoError(t, err)

	var capturedUID string

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		uid, exists := c.Get(ContextKeyUserID)
		if exists {
			capturedUID = uid.(string)
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, "user123", capturedUID)
}

// TestJWTAuth_CallNext 测试验证通过后后续的 Handler 会被执行
func TestJWTAuth_CallNext(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	helper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(time.Hour),
	)

	claims := jwtx.JWTClaims{
		UID:       "user456",
		TokenType: jwtx.Access,
	}
	token, err := helper.GenerateToken(claims)
	assert.NoError(t, err)

	handlerExecuted := false

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.Use(JWTAuth(helper))
	router.GET("/test", func(c *gin.Context) {
		handlerExecuted = true
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.True(t, handlerExecuted)
}
