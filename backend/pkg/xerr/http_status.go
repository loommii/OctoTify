// http_status.go 定义错误码到 HTTP 状态码的映射逻辑
//
// 提供以下功能：
//   - commonHTTPStatus: 特殊错误码的 HTTP 状态码映射
//   - HTTPStatusFromCode(): 根据业务错误码获取对应的 HTTP 状态码
//
// 映射原则：
//   - 除 JWT 鉴权失败外，所有业务错误均返回 HTTP 200
//   - 通过响应体中的 code 字段区分错误类型
//   - 详见 API 规范文档：docs/03-API-Specification.md
package xerr

import "net/http"

// commonHTTPStatus 特殊错误码的 HTTP 状态码映射
// 仅记录非 200 的错误码，未列出的默认返回 200
var commonHTTPStatus = map[int]int{
	100001: http.StatusUnauthorized, // JWT 鉴权失败（通用）
	111000: http.StatusUnauthorized, // JWT 未提供认证令牌
	111001: http.StatusUnauthorized, // JWT 令牌格式错误
	111002: http.StatusUnauthorized, // JWT 令牌无效或已过期
	111003: http.StatusUnauthorized, // JWT 无效的令牌类型
}

// HTTPStatusFromCode 根据业务错误码返回对应的 HTTP 状态码
//
// 根据项目 API 规范：
//   - 除 JWT 鉴权失败（100001）返回 401 外
//   - 所有业务错误均返回 HTTP 200
//   - 通过响应体中的 code 字段区分错误类型
func HTTPStatusFromCode(code int) int {
	if code == 0 {
		return http.StatusOK
	}

	if status, ok := commonHTTPStatus[code]; ok {
		return status
	}

	return http.StatusOK
}
