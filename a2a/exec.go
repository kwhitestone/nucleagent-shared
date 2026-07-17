// Package a2a 定义 core ↔ executor 之间的 A2A 线协议类型。
// 仅包含协议数据结构与接口，不含业务逻辑。
package a2a

import "encoding/json"

// ExecutionRequest core 下发给 executor 的执行请求。
type ExecutionRequest struct {
	ConversationID uint              `json:"conversationId"`       // 目标对话
	StepID         string            `json:"stepId"`               // 本次执行步骤 UUID
	Mode           string            `json:"mode"`                 // a2a / a2a_agent / a2a_employee
	AgentID        *uint             `json:"agentId,omitempty"`    // 执行 Agent 实例
	ProviderID     *uint             `json:"providerId,omitempty"` // LLM 提供商
	Model          string            `json:"model,omitempty"`      // 模型
	Input          string            `json:"input"`                // 用户输入 / 上行指令
	Context        json.RawMessage   `json:"context,omitempty"`    // 编排上下文(plan/history/skills)
	Headers        map[string]string `json:"headers,omitempty"`    // S2S 头(x-llm-proxy-key 等)
}

// ExecutionResult executor 执行完成后的同步结果摘要（详细内容走流式上报）。
type ExecutionResult struct {
	StepID string `json:"stepId"`
	Status string `json:"status"`           // completed / failed / cancelled
	Output string `json:"output,omitempty"` // 摘要输出
	Error  string `json:"error,omitempty"`
}

// StreamReporter executor 向 core 回报流式事件的接口。
// 实现方负责把事件桥接到 core 的 WebSocket（并最终经 Redis pub/sub 转 SSE）。
type StreamReporter interface {
	TextDelta(content string)     // 流式文本增量
	ThinkingDelta(content string) // 思考过程增量
	Progress(content string)      // 进度更新
	ToolUse(tool, content string) // 工具调用事件
	Flush()                       // 刷新缓冲
}
