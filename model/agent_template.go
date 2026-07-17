package model

import "time"

// AgentTemplate Agent 模板（对话/编排用的预设角色）。
// 对应表 agent_templates，字段对齐 docs/03-data-models.md。
type AgentTemplate struct {
	ID        uint      `json:"id" gorm:"primaryKey;comment:主键ID"`
	Name      string    `json:"name" gorm:"column:name;size:128;comment:模板名称"`
	Slug      string    `json:"slug" gorm:"column:slug;size:128;uniqueIndex;comment:模板唯一标识"`
	Config    JSON      `json:"config" gorm:"column:config;type:json;comment:模板配置(category/role/personality/bio/prompt/avatar/color/sort_order)"`
	I18n      JSON      `json:"i18n" gorm:"column:i18n;type:json;comment:多语言文案"`
	IsActive  bool      `json:"isActive" gorm:"column:is_active;comment:是否启用"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;comment:创建时间"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;comment:更新时间"`
}
