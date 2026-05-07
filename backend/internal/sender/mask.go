package sender

import "strings"

// maskToken 对 Token 进行脱敏处理
// 仅显示前 4 位 + "..." + 后 4 位
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

// maskEmailUsername 对邮箱用户名部分进行脱敏处理
func maskEmailUsername(email string) string {
	atIdx := strings.Index(email, "@")
	if atIdx <= 0 {
		return "***"
	}
	return "***" + email[atIdx:]
}
