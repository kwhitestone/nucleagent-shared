// Package a2a 的 TaskSession 定义。
//
// 参考 agentia-executor/src/internal/executor/session.go：
// TaskSession 是单次执行任务的会话状态，executor-agnostic--backend 私有数据
// 放 BackendSessionID / BackendSessionDirectory 等字段，不入库，由 executor 内存 +
// JSON 文件持久化。
package a2a

// TaskSession 单次执行任务的会话状态（executor 内存 + JSON 文件持久化，不入库）。
//
// backend 私有字段（BackendSessionID/BackendSessionDirectory/BackendServerURL 等）
// 由各 backend 自行填充，runtime 不解释其语义，仅按 BackendSessionID 传给 Kill。
type TaskSession struct {
	ID             string `json:"id"`             // 会话 UUID
	ConversationID uint   `json:"conversationId"` // 关联对话
	StepID         string `json:"stepId"`         // 关联步骤
	Backend        string `json:"backend"`        // 执行后端 ID（= Capability()）
	Status         string `json:"status"`         // running / done / failed / killed
	Workdir        string `json:"workdir,omitempty"` // 沙箱工作目录

	// backend 私有数据（executor-agnostic：runtime 不解释，仅透传给 Kill）。
	BackendSessionID        string `json:"backendSessionId,omitempty"`
	BackendSessionDirectory string `json:"backendSessionDirectory,omitempty"`
	BackendServerURL        string `json:"backendServerUrl,omitempty"`

	// 生命周期时间戳（毫秒），由 executor 维护。
	StartedAt   int64  `json:"startedAt,omitempty"`
	UpdatedAt   int64  `json:"updatedAt,omitempty"`
	CompletedAt int64  `json:"completedAt,omitempty"`
	Error       string `json:"error,omitempty"`
}
