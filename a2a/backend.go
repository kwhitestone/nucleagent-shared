// Package a2a 的执行后端契约。
//
// 参考 agentia-executor/src/internal/executor/backend.go：
// Backend 是每个执行后端（opencode/hermes/mock-llm/...）实现的契约。
// runtime 按 ExecutionRequest.Capability 路由到对应 Backend，新增后端不改 runtime/main。
package a2a

import "context"

// Backend 是每个执行后端实现的契约。
//
// executor runtime 收到 a2a_request 后，按 req.Capability 选 Backend，
// 调用 Run 执行；取消时调用 Kill。
//
// 实现方必须：
//   - 响应 ctx 取消，完整退出旧 run 后再释放私有资源；
//   - 通过 reporter 回报流式事件，通过返回值给出最终结果；
//   - 不解释 core 的 plan/candidate/业务字段（executor-agnostic 边界）。
type Backend interface {
	// Capability 返回后端能力 ID，与 ExecutionRequest.Capability 匹配。
	Capability() string

	// Descriptor 返回后端自描述，用于握手时上报给 core。
	Descriptor() DesktopExecutor

	// Run 执行请求，流式回报到 reporter，返回最终结果。
	Run(ctx context.Context, req *ExecutionRequest, reporter StreamReporter) ExecutionResult

	// Kill 取消运行中的 session（按 BackendSessionID 定位）。
	Kill(ctx context.Context, session TaskSession) error
}

// ExecutionRequest core 下发给 executor 的执行请求。
type ExecutionRequest struct {
	ConversationID uint              `json:"conversationId"`       // 目标对话
	StepID         string            `json:"stepId"`               // 本次执行步骤 UUID
	Mode           string            `json:"mode"`                 // a2a / a2a_agent / a2a_employee
	AgentID        *uint             `json:"agentId,omitempty"`    // 执行 Agent 实例
	ProviderID     *uint             `json:"providerId,omitempty"` // LLM 提供商
	Model          string            `json:"model,omitempty"`      // 模型
	Input          string            `json:"input"`                // 用户输入 / 上行指令
	Context        []byte            `json:"context,omitempty"`    // 编排上下文 JSON（plan/history/skills）
	Headers        map[string]string `json:"headers,omitempty"`    // S2S 头（x-llm-proxy-key 等）
}

// ExecutionResult executor 执行完成后的同步结果摘要（详细内容走流式上报）。
type ExecutionResult struct {
	StepID string `json:"stepId"`
	Status string `json:"status"`           // completed / failed / cancelled
	Output string `json:"output,omitempty"` // 摘要输出
	Error  string `json:"error,omitempty"`
}

// StreamReporter executor 向 core 回报流式事件的接口。
//
// 实现方负责把事件桥接到 core 的 WebSocket（并最终经 Redis pub/sub 转 SSE）。
// 参考附录 §7.1：text_delta 走就地 upsert + 幂等键，不是每次 append 新消息。
type StreamReporter interface {
	TextDelta(content string)     // 流式文本增量（最终回答）
	ThinkingDelta(content string) // 思考过程增量（用户可读中间输出）
	Progress(content string)      // 进度更新（短状态文案）
	ToolUse(tool, content string) // 工具调用事件
	Flush()                       // 刷新缓冲（确保已产出的 delta 落库/上报）
}
