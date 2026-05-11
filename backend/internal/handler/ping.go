package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"octotify/internal/handler/dto"
	"octotify/pkg/response"
)

// Ping 健康检查，返回服务名、服务器时间和时间戳
//
//	@Summary		健康检查
//	@Description	检查服务是否正常运行，返回服务名称、服务器时间和 Unix 毫秒时间戳。无需认证。
//	@Tags			系统
//	@Produce		json
//	@Success		200	{object}	response.Response{data=dto.PingResp}	"服务正常"
//	@Router			/ping [get]
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
