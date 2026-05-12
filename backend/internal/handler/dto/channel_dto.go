package dto

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

// ChannelConfig is a map type that serializes to/from JSON for database storage.
// It generates `additionalProperties: true` in OpenAPI spec, which maps to `Record<string, unknown>` in TypeScript.
type ChannelConfig map[string]any

// Scan implements sql.Scanner interface for GORM database reading.
func (c *ChannelConfig) Scan(value any) error {
	if value == nil {
		*c = nil
		return nil
	}
	var jsonBytes []byte
	switch v := value.(type) {
	case []byte:
		jsonBytes = v
	case string:
		jsonBytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into ChannelConfig", value)
	}
	return json.Unmarshal(jsonBytes, c)
}

// Value implements driver.Valuer interface for GORM database writing.
func (c ChannelConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// ToJSON converts ChannelConfig to datatypes.JSON for internal use.
func (c ChannelConfig) ToJSON() datatypes.JSON {
	if c == nil {
		return datatypes.JSON("{}")
	}
	b, _ := json.Marshal(c)
	return datatypes.JSON(b)
}

// FromJSON converts datatypes.JSON to ChannelConfig.
func FromJSON(j datatypes.JSON) ChannelConfig {
	var cfg ChannelConfig
	if j == nil {
		return cfg
	}
	_ = json.Unmarshal(j, &cfg)
	return cfg
}

const (
	ChannelTypeWechat        = "wechat"
	ChannelTypeWechatClawbot = "wechat_clawbot"
	ChannelTypeTelegram      = "telegram"
	ChannelTypeDingtalk      = "dingtalk"
	ChannelTypeEmail         = "email"
	ChannelTypeWebhook       = "webhook"
	ChannelTypeFeishu        = "feishu"
)

var ValidChannelTypes = map[string]bool{
	ChannelTypeWechat:        true,
	ChannelTypeWechatClawbot: true,
	ChannelTypeTelegram:      true,
	ChannelTypeDingtalk:      true,
	ChannelTypeEmail:         true,
	ChannelTypeWebhook:       true,
	ChannelTypeFeishu:        true,
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
// 当前仅暴露已实现的渠道类型给前端（飞书、Telegram、邮件、钉钉）
var ChannelTypeMetas = []ChannelTypeMeta{
	{
		Type:        ChannelTypeWechatClawbot,
		Name:        "微信ClawBot",
		Description: "微信个人号AI助手（基于iLink协议）",
		ConfigFields: []ConfigField{
			{
				Name:        "bot_token_ciphertext",
				Label:       "Bot Token（加密）",
				Type:        "password",
				Required:    true,
				Placeholder: "扫码绑定后自动填充",
			},
			{
				Name:        "bot_token_nonce",
				Label:       "Token Nonce（加密）",
				Type:        "password",
				Required:    true,
				Placeholder: "扫码绑定后自动填充",
			},
			{
				Name:        "ilink_bot_id",
				Label:       "Bot ID",
				Type:        "string",
				Required:    true,
				Placeholder: "扫码绑定后自动填充",
			},
			{
				Name:        "ilink_user_id",
				Label:       "用户ID",
				Type:        "string",
				Required:    true,
				Placeholder: "扫码绑定后自动填充",
			},
		},
	},
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
	{
		Type:        ChannelTypeDingtalk,
		Name:        "钉钉",
		Description: "钉钉群机器人",
		ConfigFields: []ConfigField{
			{
				Name:        "webhook_url",
				Label:       "Webhook 地址",
				Type:        "url",
				Required:    true,
				Placeholder: "https://oapi.dingtalk.com/robot/send?access_token=xxx",
			},
			{
				Name:        "secret",
				Label:       "加签密钥",
				Type:        "password",
				Required:    true,
				Placeholder: "SECxxxxxxxxxxxxxxxxx",
			},
		},
	},
	{
		Type:        ChannelTypeTelegram,
		Name:        "Telegram",
		Description: "Telegram Bot 推送",
		ConfigFields: []ConfigField{
			{
				Name:        "bot_token",
				Label:       "Bot Token",
				Type:        "password",
				Required:    true,
				Placeholder: "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz",
			},
			{
				Name:        "chat_id",
				Label:       "Chat ID",
				Type:        "string",
				Required:    true,
				Placeholder: "私聊/群组填数字ID，频道填 @频道名",
			},
			{
				Name:        "proxy",
				Label:       "HTTP 代理（可选）",
				Type:        "url",
				Required:    false,
				Placeholder: "http://127.0.0.1:7890",
			},
		},
	},
	{
		Type:        ChannelTypeEmail,
		Name:        "邮件",
		Description: "邮件推送",
		ConfigFields: []ConfigField{
			{
				Name:        "smtp_host",
				Label:       "SMTP 服务器",
				Type:        "string",
				Required:    true,
				Placeholder: "smtp.example.com",
			},
			{
				Name:        "smtp_port",
				Label:       "SMTP 端口",
				Type:        "number",
				Required:    true,
				Placeholder: "587",
			},
			{
				Name:        "username",
				Label:       "用户名",
				Type:        "string",
				Required:    true,
				Placeholder: "user@example.com",
			},
			{
				Name:        "password",
				Label:       "密码/授权码",
				Type:        "password",
				Required:    true,
				Placeholder: "邮箱密码或授权码",
			},
			{
				Name:        "to",
				Label:       "收件人",
				Type:        "string",
				Required:    true,
				Placeholder: "recipient@example.com",
			},
			{
				Name:        "cc",
				Label:       "抄送人（可选）",
				Type:        "string",
				Required:    false,
				Placeholder: "cc1@example.com, cc2@example.com",
			},
			{
				Name:        "from_name",
				Label:       "发件人名称（可选）",
				Type:        "string",
				Required:    false,
				Placeholder: "系统通知",
			},
		},
	},
}

// CreateChannelReq 创建推送渠道请求
type CreateChannelReq struct {
	Type   string        `json:"type" binding:"required,channel_type" example:"dingtalk"` // 渠道类型，可选值：wechat, telegram, dingtalk, email, webhook
	Name   string        `json:"name" binding:"required,min=1,max=128" example:"钉钉-运维群"`  // 渠道名称，1-128 个字符
	Config ChannelConfig `json:"config" binding:"required"`                               // 渠道配置，JSON 对象，不同渠道类型配置不同
}

// UpdateChannelReq 编辑推送渠道请求
type UpdateChannelReq struct {
	Name   string        `json:"name" binding:"required,min=1,max=128" example:"钉钉-运维群"` // 渠道名称，1-128 个字符
	Config ChannelConfig `json:"config" binding:"required"`                              // 渠道配置，JSON 对象，不同渠道类型配置不同
}

// ChannelDTO 推送渠道响应
type ChannelDTO struct {
	ID         int64         `json:"id" example:"1"`                          // 渠道 ID
	UserID     int64         `json:"user_id" example:"1"`                     // 所属用户 ID
	Type       string        `json:"type" example:"dingtalk"`                 // 渠道类型：wechat, telegram, dingtalk, email, webhook
	Name       string        `json:"name" example:"钉钉-运维群"`                   // 渠道名称
	Config     ChannelConfig `json:"config"`                                  // 渠道配置，JSON 格式
	Status     int           `json:"status" example:"1"`                      // 状态：1-启用 2-停用 -1-已删除
	CreatedAt  int64         `json:"created_at_ts" example:"1714636800000"`   // 创建时间（Unix 毫秒时间戳）
	UpdatedAt  int64         `json:"updated_at_ts" example:"1714636800000"`   // 更新时间（Unix 毫秒时间戳）
	LastUsedAt int64         `json:"last_used_at_ts" example:"1714636800000"` // 最后使用时间（Unix 毫秒时间戳，0 表示未使用）
}
