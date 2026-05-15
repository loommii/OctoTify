package ctxutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWithRequestID_GetRequestID(t *testing.T) {
	ctx := context.Background()
	rid := "0196a3b2c4d50000a1b2c3d4e5f67890"

	ctx = WithRequestID(ctx, rid)
	got := GetRequestID(ctx)

	assert.Equal(t, rid, got)
}

func TestGetRequestID_EmptyContext(t *testing.T) {
	ctx := context.Background()

	got := GetRequestID(ctx)

	assert.Equal(t, "", got)
}

func TestGetRequestID_WrongType(t *testing.T) {
	type ctxKey string
	ctx := context.WithValue(context.Background(), requestIDKey, 12345)

	got := GetRequestID(ctx)

	assert.Equal(t, "", got)
}

func TestWithRequestID_Override(t *testing.T) {
	ctx := context.Background()

	ctx = WithRequestID(ctx, "first")
	ctx = WithRequestID(ctx, "second")

	assert.Equal(t, "second", GetRequestID(ctx))
}

func TestWithRequestID_PreservesOtherValues(t *testing.T) {
	type otherKey string
	ctx := context.WithValue(context.Background(), otherKey("foo"), "bar")

	ctx = WithRequestID(ctx, "rid-value")

	assert.Equal(t, "rid-value", GetRequestID(ctx))
	assert.Equal(t, "bar", ctx.Value(otherKey("foo")))
}

func TestGetRequestID_TodoContext(t *testing.T) {
	ctx := context.TODO()

	got := GetRequestID(ctx)

	assert.Equal(t, "", got)
}

// ============================================================================
// TestLoggerWithRequestID 测试带 request_id 的 logger 功能
// ============================================================================

func TestLoggerWithRequestID_NilLogger(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-rid-001")

	result := LoggerWithRequestID(ctx, nil)

	assert.Nil(t, result)
}

func TestLoggerWithRequestID_NoRequestID(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	result := LoggerWithRequestID(ctx, logger)

	assert.Equal(t, logger, result)
}

func TestLoggerWithRequestID_WithRequestID(t *testing.T) {
	ctx := context.Background()
	rid := "test-rid-002"
	ctx = WithRequestID(ctx, rid)

	logger, _ := zap.NewDevelopment()
	result := LoggerWithRequestID(ctx, logger)

	assert.NotNil(t, result)
	assert.NotEqual(t, logger, result)
}

func TestLoggerWithRequestID_NilLoggerWithRID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-rid-003")

	result := LoggerWithRequestID(ctx, nil)

	assert.Nil(t, result)
}
