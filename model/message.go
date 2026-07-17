package model

import "time"

// Message 对话消息（用户/Agent/系统/工具产出，每条都要展示）。
// 对应表 messages，字段对齐 docs/03-data-models.md。
// 高频展示字段（sender_type / sender_name）独立成列，不放入 JSON。
type Message struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:主键ID"`
	ConversationID uint      `json:"conversationId" gorm:"column:conversation_id;index;comment:所属对话ID"`
	SenderType     string    `json:"senderType" gorm:"column:sender_type;size:16;comment:发送方类型(user/agent/system/tool)"`
	SenderName     string    `json:"senderName" gorm:"column:sender_name;size:64;comment:发送方展示名"`
	MsgType        string    `json:"msgType" gorm:"column:msg_type;size:32;comment:消息类型(text/plan/result/error/tool_call)"`
	Content        string    `json:"content" gorm:"column:content;type:text;comment:消息内容"`
	Metadata       JSON      `json:"metadata" gorm:"column:metadata;type:json;comment:元数据(step_id/attachments/history)"`
	CreatedAt      time.Time `json:"createdAt" gorm:"column:created_at;comment:创建时间"`
}
