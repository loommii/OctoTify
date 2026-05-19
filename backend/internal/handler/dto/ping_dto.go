package dto

// PingResp 健康检查响应
type PingResp struct {
	ServerName string `json:"server_name"` // 服务名称
	ServerTime string `json:"server_time"` // 服务器时间，格式：YYYY-MM-DD HH:MM:SS
	Timestamp  int64  `json:"timestamp"`   // Unix 毫秒时间戳
}
