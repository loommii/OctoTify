package dto

// SourceBaseReq 消息来源公共请求字段
type SourceBaseReq struct {
	Name        string `json:"name" binding:"required,min=1,max=128" example:"CI Pipeline"`           // 来源名称，1-128 个字符
	Description string `json:"description" binding:"omitempty,max=512" example:"GitHub Actions 构建通知"` // 来源描述，最多 512 个字符
}

// CreateSourceReq 创建消息来源请求
type CreateSourceReq struct {
	SourceBaseReq
}

// UpdateSourceReq 编辑消息来源请求
type UpdateSourceReq struct {
	SourceBaseReq
}

// SourceDTO 消息来源信息响应
type SourceDTO struct {
	ID          int64  `json:"id" example:"1"`                                      // 来源 ID
	UserID      int64  `json:"user_id" example:"1"`                                 // 所属用户 ID
	Name        string `json:"name" example:"CI Pipeline"`                          // 来源名称
	Token       string `json:"token" example:"src0196a3b2c4d50000a1b2c3d4e5f67890"` // 推送 Token
	Description string `json:"description" example:"GitHub Actions 构建通知"`           // 来源描述
	Status      int    `json:"status" example:"1"`                                  // 状态：1-启用 2-停用 -1-已删除
	CreatedAt   int64  `json:"created_at" example:"1714636800000"`                  // 创建时间（Unix 毫秒时间戳）
}

// PageReq 分页请求参数
type PageReq struct {
	Page     int `form:"page" binding:"omitempty,min=1" example:"1"`         // 页码，从 1 开始
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100" example:"20"` // 每页条数，最大 100
}

// Normalize 规范化分页参数，确保 page 和 pageSize 在有效范围内
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
