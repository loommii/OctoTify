package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"octotify/pkg/response"
	"octotify/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func mockVerifyFn(_ context.Context, _ int64, password string) error {
	if password == "correct" {
		return nil
	}
	return errors.New("wrong password")
}

func TestStepUpAuth_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, "123")
		c.Next()
	})
	r.POST("/test", StepUpAuth(mockVerifyFn), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"password":"correct"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
}

func TestStepUpAuth_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	r.POST("/test", StepUpAuth(mockVerifyFn), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"password":"correct"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
}

func TestStepUpAuth_WrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, "123")
		c.Next()
	})
	r.POST("/test", StepUpAuth(mockVerifyFn), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.CodeLoginInvalidCredentials, resp.Code)
}

func TestStepUpAuth_MissingPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, "123")
		c.Next()
	})
	r.POST("/test", StepUpAuth(mockVerifyFn), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, resp.Code)
}

func TestStepUpAuth_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, "123")
		c.Next()
	})
	r.POST("/test", StepUpAuth(mockVerifyFn), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{invalid}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, resp.Code)
}

func TestStepUpAuth_InvalidUserIDType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, int64(123))
		c.Next()
	})
	r.POST("/test", StepUpAuth(mockVerifyFn), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"password":"correct"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
}

func TestStepUpAuth_InvalidUserIDFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set(ContextKeyUserID, "abc")
		c.Next()
	})
	r.POST("/test", StepUpAuth(mockVerifyFn), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	req, _ := http.NewRequest("POST", "/test", strings.NewReader(`{"password":"correct"}`))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, xerr.ErrUnauthorized.Code, resp.Code)
}
