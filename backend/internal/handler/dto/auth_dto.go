package dto

// AuthCredentials 认证凭据（注册/登录通用）
// @Description 用户名和密码认证信息
type AuthCredentials struct {
	Username string `json:"username" binding:"required,username" example:"octotify"`    // 用户名，3-64 个字符，仅允许字母、数字和下划线
	Password string `json:"password" binding:"required,password" example:"P@ssw0rd123"` // 密码，8-128 个字符，需包含大小写字母和数字
}

// RegisterReq 用户注册请求
// @Description 注册新用户时使用的请求参数
type RegisterReq struct {
	AuthCredentials
}

// LoginReq 用户登录请求
// @Description 用户登录时使用的请求参数
type LoginReq struct {
	AuthCredentials
}

// UserDTO 用户信息响应
// @Description 用户基本信息，用于登录/注册/刷新令牌等接口返回
type UserDTO struct {
	ID        int64  `json:"id" example:"1"`                  // 用户 ID
	Username  string `json:"username" example:"octotify"`     // 用户名
	CreatedAt int64  `json:"created_at" example:"1714636800"` // 创建时间（Unix 时间戳）
}

// RefreshReq 刷新令牌请求
// @Description 使用 Refresh Token 获取新的 Access Token
type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."` // Refresh Token，登录时获取
}

// ChangePasswordReq 修改密码请求
// @Description 修改用户登录密码
type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required,password" example:"OldP@ssw0rd123"` // 旧密码
	NewPassword string `json:"new_password" binding:"required,password" example:"NewP@ssw0rd123"` // 新密码，8-128 个字符，需包含大小写字母和数字
}

// AuthResp 认证响应（登录/注册/刷新令牌）
// @Description 认证成功后返回的 Access Token、Refresh Token 和用户信息
type AuthResp struct {
	AccessToken  string  `json:"access_token" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`  // 访问令牌，有效期 1 小时
	RefreshToken string  `json:"refresh_token" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."` // 刷新令牌，有效期 7 天
	User         UserDTO `json:"user"`                                                            // 用户信息
}

// UserProfileResp 用户信息响应
// @Description 查询用户个人信息时返回的响应结构
type UserProfileResp struct {
	User UserDTO `json:"user"` // 用户信息
}
