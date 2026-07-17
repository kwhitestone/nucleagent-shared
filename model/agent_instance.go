package model

import "time"

// AgentInstance 用户雇佣的 Agent 实例（基于某个模板）。
// 对应表 agent_instances，字段对齐 docs/03-data-models.md。
type AgentInstance struct {
	ID         uint      `json:"id" gorm:"primaryKey;comment:主键ID"`
	UserID     uint      `json:"userId" gorm:"column:user_id;index;comment:所属用户ID"`
	TemplateID uint      `json:"templateId" gorm:"column:template_id;index;comment:来源模板ID"`
	Config     JSON      `json:"config" gorm:"column:config;type:json;comment:实例配置(nickname/override)"`
	I18n       JSON      `json:"i18n" gorm:"column:i18n;type:json;comment:多语言覆盖"`
	CreatedAt  time.Time `json:"createdAt" gorm:"column:created_at;comment:创建时间"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"column:updated_at;comment:更新时间"`
}
