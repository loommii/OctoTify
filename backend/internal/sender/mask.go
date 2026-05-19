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

// maskIdentifier 对标识符进行脱敏处理（如ilink_bot_id、ilink_user_id等）
// 仅显示前 4 位 + "..." + 后 4 位
func maskIdentifier(identifier string) string {
	if len(identifier) <= 8 {
		return "****"
	}
	return identifier[:4] + "..." + identifier[len(identifier)-4:]
}

// maskURL 对 URL 进行脱敏处理，隐藏 query 参数中的敏感信息（如 access_token）
func maskURL(urlStr string) string {
	// 简单处理：如果包含 access_token，则将其值替换为 ***
	if idx := strings.Index(urlStr, "access_token="); idx != -1 {
		tokenStart := idx + len("access_token=")
		// 找到 token 结束位置（& 或字符串末尾）
		tokenEnd := strings.Index(urlStr[tokenStart:], "&")
		if tokenEnd == -1 {
			return urlStr[:tokenStart] + "***"
		}
		return urlStr[:tokenStart] + "***" + urlStr[tokenStart+tokenEnd:]
	}
	return urlStr
}
