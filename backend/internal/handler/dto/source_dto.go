package dto

// SourceBaseReq 消息来源公共请求字段
// @Description 创建和编辑消息来源时的公共字段
type SourceBaseReq struct {
	Name        string `json:"name" binding:"required,min=1,max=128" example:"CI Pipeline"`           // 来源名称，1-128 个字符
	Description string `json:"description" binding:"omitempty,max=512" example:"GitHub Actions 构建通知"` // 来源描述，最多 512 个字符
}

// CreateSourceReq 创建消息来源请求
// @Description 创建新的消息来源，系统自动生成 Source Token
type CreateSourceReq struct {
	SourceBaseReq
	ChannelIDs []int64 `json:"channel_ids" binding:"omitempty" example:"1,2,3"` // 关联的渠道ID列表
}

// UpdateSourceReq 编辑消息来源请求
// @Description 编辑已有消息来源的名称、描述和绑定渠道
type UpdateSourceReq struct {
	SourceBaseReq
	ChannelIDs []int64 `json:"channel_ids" binding:"omitempty" example:"1,2,3"` // 关联的渠道ID列表
}

// SourceDTO 消息来源信息响应
// @Description 消息来源的基本信息，不包含 Token
type SourceDTO struct {
	ID          int64  `json:"id" example:"1"`                                      // 来源 ID
	UserID      int64  `json:"user_id" example:"1"`                                 // 所属用户 ID
	Name        string `json:"name" example:"CI Pipeline"`                          // 来源名称
	Token       string `json:"token" example:"src0196a3b2c4d50000a1b2c3d4e5f67890"` // 推送 Token（仅创建时返回）
	Description string `json:"description" example:"GitHub Actions 构建通知"`           // 来源描述
	Status      int    `json:"status" example:"1"`                                  // 状态：1-启用 2-停用 -1-已删除
	CreatedAt   int64  `json:"created_at_ts" example:"1714636800000"`               // 创建时间（Unix 毫秒时间戳）
}

// SourceDetailDTO 消息来源详情响应（包含 Token）
// @Description 消息来源的详细信息，包含 Token 和使用时间
type SourceDetailDTO struct {
	ID          int64  `json:"id" example:"1"`                                      // 来源 ID
	UserID      int64  `json:"user_id" example:"1"`                                 // 所属用户 ID
	Name        string `json:"name" example:"CI Pipeline"`                          // 来源名称
	Token       string `json:"token" example:"src0196a3b2c4d50000a1b2c3d4e5f67890"` // 推送 Token
	Description string `json:"description" example:"GitHub Actions 构建通知"`           // 来源描述
	Status      int    `json:"status" example:"1"`                                  // 状态：1-启用 2-停用 -1-已删除
	CreatedAt   int64  `json:"created_at_ts" example:"1714636800000"`               // 创建时间（Unix 毫秒时间戳）
	UpdatedAt   int64  `json:"updated_at_ts" example:"1714636800000"`               // 更新时间（Unix 毫秒时间戳）
	LastUsedAt  int64  `json:"last_used_at_ts" example:"1714636800000"`             // 最后使用时间（Unix 毫秒时间戳，0 表示未使用）
}

// SourceDetailResponse 来源详情响应（包含渠道列表）
// @Description 来源详情及其已绑定的有效渠道列表
type SourceDetailResponse struct {
	Source   *SourceDetailDTO `json:"source"`   // 来源详情
	Channels []*ChannelDTO    `json:"channels"` // 已绑定渠道列表
}

// SourceTokenResponse 来源令牌响应
// @Description 查看或重置来源令牌时返回的 Token 值
type SourceTokenResponse struct {
	Token string `json:"token" example:"src0196a3b2c4d50000a1b2c3d4e5f67890"` // 推送 Token
}

// VerifyPasswordReq 密码验证请求（用于查看敏感数据）
// @Description 需要密码二次验证的接口（查看令牌、重置令牌、停用/启用/删除来源）
type VerifyPasswordReq struct {
	Password string `json:"password" binding:"required,password" example:"P@ssw0rd123"` // 用户密码
}

// PageReq 分页请求参数
// @Description 所有列表接口的通用分页参数
type PageReq struct {
	Page     int `form:"page" binding:"omitempty,min=1" example:"1"`               // 页码，从 1 开始，默认 1
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100" example:"20"` // 每页条数，默认 20，最大 100
}

// Normalize 规范化分页参数，确保 page 和pageSize 在有效范围内
func (p *PageReq) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}
