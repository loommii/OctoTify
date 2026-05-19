package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"octotify/internal/handler/dto"
	"octotify/pkg/response"
)

// Ping 健康检查，返回服务名、服务器时间和时间戳
func Ping(serverName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		response.Success(c, dto.PingResp{
			ServerName: serverName,
			ServerTime: now.Format(time.DateTime),
			Timestamp:  now.UnixMilli(),
		})
	}
}
