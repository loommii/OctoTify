package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"octotify/pkg/response"
)

// Ping 健康检查，返回服务名、服务器时间和时间戳
func Ping(serverName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Success(c, gin.H{
			"server_name": serverName,
			"server_time": time.Now().Format("2006-01-02 15:04:05"),
			"timestamp":   time.Now().Unix(),
		})
	}
}
