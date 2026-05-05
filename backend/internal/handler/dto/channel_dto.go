package dto

const (
	ChannelTypeWechat   = "wechat"
	ChannelTypeTelegram = "telegram"
	ChannelTypeDingtalk = "dingtalk"
	ChannelTypeEmail    = "email"
	ChannelTypeWebhook  = "webhook"
	ChannelTypeFeishu   = "feishu"
)

var ValidChannelTypes = map[string]bool{
	ChannelTypeWechat:   true,
	ChannelTypeTelegram: true,
	ChannelTypeDingtalk: true,
	ChannelTypeEmail:    true,
	ChannelTypeWebhook:  true,
	ChannelTypeFeishu:   true,
}

// ConfigField 渠道配置字段定义
type ConfigField struct {
	Name        string `json:"name"`        // 字段名称
	Label       string `json:"label"`       // 字段标签
	Type        string `json:"type"`        // 字段类型：string, number, password, url
	Required    bool   `json:"required"`    // 是否必填
	Placeholder string `json:"placeholder"` // 占位符提示
}

// ChannelTypeMeta 渠道类型元数据
type ChannelTypeMeta struct {
	Type         string        `json:"type"`          // 渠道类型标识
	Name         string        `json:"name"`          // 渠道类型名称
	Description  string        `json:"description"`   // 描述
	ConfigFields []ConfigField `json:"config_fields"` // 配置字段定义
}

// ChannelTypeMetas 所有支持的渠道类型元数据
// TODO: 以下渠道待实现：企业微信、Telegram、钉钉、邮件、Webhook
var ChannelTypeMetas = []ChannelTypeMeta{
	{
		Type:        ChannelTypeFeishu,
		Name:        "飞书",
		Description: "飞书自定义机器人",
		ConfigFields: []ConfigField{
			{
				Name:        "webhook_url",
				Label:       "Webhook 地址",
				Type:        "url",
				Required:    true,
				Placeholder: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
			},
			{
				Name:        "secret",
				Label:       "签名密钥（可选）",
				Type:        "password",
				Required:    false,
				Placeholder: "用于签名校验的密钥",
			},
		},
	},
}

// TODO: 恢复以下渠道
// {
// 	Type:        ChannelTypeWechat,
// 	Name:        "企业微信",
// 	Description: "企业微信群机器人",
// 	ConfigFields: []ConfigField{
// 		{Name: "webhook", Label: "Webhook 地址", Type: "url", Required: true, Placeholder: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"},
// 	},
// },
// {
// 	Type:        ChannelTypeTelegram,
// 	Name:        "Telegram",
// 	Description: "Telegram Bot",
// 	ConfigFields: []ConfigField{
// 		{Name: "bot_token", Label: "Bot Token", Type: "password", Required: true, Placeholder: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"},
// 		{Name: "chat_id", Label: "Chat ID", Type: "string", Required: true, Placeholder: "-1001234567890"},
// 	},
// },
// {
// 	Type:        ChannelTypeDingtalk,
// 	Name:        "钉钉",
// 	Description: "钉钉群机器人",
// 	ConfigFields: []ConfigField{
// 		{Name: "webhook", Label: "Webhook 地址", Type: "url", Required: true, Placeholder: "https://oapi.dingtalk.com/robot/send?access_token=xxx"},
// 	},
// },
// {
// 	Type:        ChannelTypeEmail,
// 	Name:        "邮件",
// 	Description: "邮件推送",
// 	ConfigFields: []ConfigField{
// 		{Name: "smtp_host", Label: "SMTP 服务器", Type: "string", Required: true, Placeholder: "smtp.example.com"},
// 		{Name: "smtp_port", Label: "SMTP 端口", Type: "number", Required: true, Placeholder: "587"},
// 		{Name: "username", Label: "用户名", Type: "string", Required: true, Placeholder: "user@example.com"},
// 		{Name: "password", Label: "密码", Type: "password", Required: true, Placeholder: "邮箱密码"},
// 		{Name: "to", Label: "收件人", Type: "string", Required: true, Placeholder: "recipient@example.com"},
// 	},
// },
// {
// 	Type:        ChannelTypeWebhook,
// 	Name:        "Webhook",
// 	Description: "自定义 Webhook",
// 	ConfigFields: []ConfigField{
// 		{Name: "url", Label: "Webhook URL", Type: "url", Required: true, Placeholder: "https://example.com/webhook"},
// 		{Name: "method", Label: "HTTP 方法", Type: "string", Required: false, Placeholder: "POST"},
// 		{Name: "headers", Label: "自定义 Headers（JSON）", Type: "string", Required: false, Placeholder: `{"Content-Type": "application/json"}`},
// 	},
// },

// CreateChannelReq 创建推送渠道请求
// @Description 创建新的推送渠道，需要指定渠道类型、名称和配置信息
// @Description ## 渠道配置示例
// @Description - wechat: {"webhook": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"}
// @Description - telegram: {"bot_token": "xxx", "chat_id": "xxx"}
// @Description - dingtalk: {"webhook": "https://oapi.dingtalk.com/robot/send?access_token=xxx"}
// @Description - email: {"smtp_host": "smtp.example.com", "smtp_port": 587, "username": "xxx", "password": "xxx", "to": "xxx@example.com"}
// @Description - webhook: {"url": "https://example.com/webhook", "method": "POST", "headers": {"Content-Type": "application/json"}}
// @Description - feishu: {"webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx", "secret": ""}
type CreateChannelReq struct {
	Type   string `json:"type" binding:"required,channel_type" example:"dingtalk"` // 渠道类型，可选值：wechat, telegram, dingtalk, email, webhook
	Name   string `json:"name" binding:"required,min=1,max=128" example:"钉钉-运维群"`  // 渠道名称，1-128 个字符
	Config string `json:"config" binding:"required" swaggertype:"object"`          // 渠道配置，JSON 格式，不同渠道类型配置不同
}

// UpdateChannelReq 编辑推送渠道请求
// @Description 编辑已有推送渠道的名称和配置信息
// @Description ## 渠道配置示例
// @Description - wechat: {"webhook": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"}
// @Description - telegram: {"bot_token": "xxx", "chat_id": "xxx"}
// @Description - dingtalk: {"webhook": "https://oapi.dingtalk.com/robot/send?access_token=xxx"}
// @Description - email: {"smtp_host": "smtp.example.com", "smtp_port": 587, "username": "xxx", "password": "xxx", "to": "xxx@example.com"}
// @Description - webhook: {"url": "https://example.com/webhook", "method": "POST", "headers": {"Content-Type": "application/json"}}
type UpdateChannelReq struct {
	Name   string `json:"name" binding:"required,min=1,max=128" example:"钉钉-运维群"` // 渠道名称，1-128 个字符
	Config string `json:"config" binding:"required" swaggertype:"object"`         // 渠道配置，JSON 格式，不同渠道类型配置不同
}

// ChannelDTO 推送渠道响应
// @Description 推送渠道的详细信息
type ChannelDTO struct {
	ID         int64  `json:"id" example:"1"`                       // 渠道 ID
	UserID     int64  `json:"user_id" example:"1"`                  // 所属用户 ID
	Type       string `json:"type" example:"dingtalk"`              // 渠道类型：wechat, telegram, dingtalk, email, webhook
	Name       string `json:"name" example:"钉钉-运维群"`                // 渠道名称
	Config     string `json:"config" swaggertype:"object"`          // 渠道配置，JSON 格式
	Status     int    `json:"status" example:"1"`                   // 状态：1-启用 2-停用 -1-已删除
	CreatedAt  int64  `json:"created_at" example:"1714636800000"`   // 创建时间（Unix 毫秒时间戳）
	UpdatedAt  int64  `json:"updated_at" example:"1714636800000"`   // 更新时间（Unix 毫秒时间戳）
	LastUsedAt int64  `json:"last_used_at" example:"1714636800000"` // 最后使用时间（Unix 毫秒时间戳，0 表示未使用）
}
