package ctxutil

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

func WithRequestID(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, requestIDKey, rid)
}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// LoggerWithRequestID 返回包含 request_id 的 logger
// 如果 context 中没有 request_id 或 logger 为 nil，返回原始 logger
func LoggerWithRequestID(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if logger == nil {
		return nil
	}
	if rid := GetRequestID(ctx); rid != "" {
		return logger.With(zap.String("request_id", rid))
	}
	return logger
}
