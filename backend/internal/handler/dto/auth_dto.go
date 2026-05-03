package dto

// RegisterReq 用户注册请求
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"octotify"`                                             // 用户名，3-50 个字符
	Password string `json:"password" binding:"required,min=8,max=128,regexp=^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)" example:"P@ssw0rd123"` // 密码，8-128 个字符，需包含大小写字母和数字
}

// UserDTO 用户信息响应
type UserDTO struct {
	ID        int64  `json:"id" example:"1"`                  // 用户 ID
	Username  string `json:"username" example:"octotify"`     // 用户名
	CreatedAt int64  `json:"created_at" example:"1714636800"` // 创建时间（Unix 时间戳）
}

// AuthResp 认证响应（登录/注册/刷新令牌）
type AuthResp struct {
	AccessToken  string  `json:"access_token" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`  // 访问令牌
	RefreshToken string  `json:"refresh_token" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."` // 刷新令牌
	User         UserDTO `json:"user"`                                                            // 用户信息
}
