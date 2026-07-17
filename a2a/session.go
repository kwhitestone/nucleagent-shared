package a2a

// TaskSession 单次执行任务的会话状态（executor 内存 + JSON 文件持久化，不入库）。
type TaskSession struct {
	ID             string `json:"id"`                  // 会话 UUID
	ConversationID uint   `json:"conversationId"`      // 关联对话
	StepID         string `json:"stepId"`              // 关联步骤
	Backend        string `json:"backend"`             // 执行后端(opencode / hermes)
	Status         string `json:"status"`              // running / done / failed / killed
	Workdir        string `json:"workdir,omitempty"`   // 沙箱工作目录
}
