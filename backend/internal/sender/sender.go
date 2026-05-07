package sender

import (
	"context"
	"gorm.io/datatypes"
)

// Sender 消息推送发送器接口
// 所有渠道发送器必须实现此接口
type Sender interface {
	Send(ctx context.Context, config datatypes.JSON, title string, content string) error
}
