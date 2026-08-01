// Package a2a 的 WS 业务负载类型。
//
// 对应 agentia-executor/src/internal/protocol/protocol.go 的各 *Payload 结构，
// 按 nucleagent 简化后的协议（见 docs/04-api-contracts.md §6）。
package a2a

import "encoding/json"

// HandshakePayload executor 连接 core WS 后的首个消息负载。
//
// executor 上报自身身份与能力，core 据此把它纳入灰度池。
type HandshakePayload struct {
	DeviceID     string             `json:"deviceId"`          // 逻辑设备 ID（灰度池分组名，多容器可共用）
	InstanceID   string             `json:"instanceId,omitempty"` // 实例 ID（单容器稳定复用）
	DeviceName   string             `json:"deviceName,omitempty"` // 展示名
	AppVersion   string             `json:"appVersion,omitempty"` // executor 应用版本
	BackendType  string             `json:"backendType,omitempty"` // 默认执行后端类型（opencode/hermes/...）
	Capabilities []DesktopCapability `json:"capabilities,omitempty"` // 支持的能力列表
	Executors    []DesktopExecutor  `json:"executors,omitempty"`    // 可用执行后端描述符
	Capacity     *ExecutorCapacity  `json:"capacity,omitempty"`     // 容量声明
	OS           string             `json:"os,omitempty"`
}

// HandshakeAckPayload core 对握手的确认。
type HandshakeAckPayload struct {
	Status string `json:"status"` // "ok" / "error"
}

// DesktopCapability executor 支持的某项能力（如 opencode 执行）。
type DesktopCapability struct {
	Name        string `json:"name"`        // 能力 ID（与 Backend.Capability() 对应）
	DisplayName string `json:"displayName"` // 展示名
	Streaming   bool   `json:"streaming"`   // 是否支持流式
}

// DesktopExecutor 执行后端自描述，由 Backend.Descriptor() 产生。
type DesktopExecutor struct {
	ID          string `json:"id"`          // 后端 ID（= Capability()）
	Type        string `json:"type"`        // 后端类型（opencode/hermes/...）
	DisplayName string `json:"displayName"`
	Streaming   bool   `json:"streaming"`
}

// ExecutorCapacity executor 的并发容量声明。
type ExecutorCapacity struct {
	MaxConcurrency int `json:"maxConcurrency,omitempty"` // 最大并发 session 数
}

// A2ARequestPayload core 下发给 executor 的执行请求负载。
//
// 嵌入 ExecutionRequest 的核心字段，外加 method/headers 用于 LLM Proxy 鉴权。
// Body 是完整的 ExecutionRequest JSON（按 nucleagent-shared/a2a/exec.go 结构）。
type A2ARequestPayload struct {
	Method     string            `json:"method"`     // 固定 "message/send"（JSON-RPC 风格）
	Capability string            `json:"capability"` // 目标后端能力 ID
	Headers    map[string]string `json:"headers,omitempty"` // S2S 头（含 x-llm-proxy-key）
	Body       json.RawMessage   `json:"body"`      // ExecutionRequest 序列化
	Stream     bool              `json:"stream,omitempty"` // 是否流式（nucleagent 恒 true）
}

// A2AResponsePayload executor 对 a2a_request 的即时 ACK。
type A2AResponsePayload struct {
	Status  int               `json:"status"`  // HTTP 风格状态码（200=接受执行）
	Headers map[string]string `json:"headers,omitempty"`
	Body    json.RawMessage   `json:"body,omitempty"` // 可携带早期错误信息
}

// A2AStreamEventPayload executor 执行过程中的流式事件。
//
// EventType 决定 Content/Tool 等字段的语义：
//   - text_delta      : Content = 流式文本增量（最终回答）
//   - thinking_delta  : Content = 用户可读的中间思考/子任务输出
//   - progress        : Content = 短状态文案，Progress 携带结构化进度
//   - tool_use        : Tool = 工具名，Content = 工具输出摘要
//
// StatusKey / ToolKind 用于前端国际化（展示端优先用 key，未知时用 fallback 文案）。
type A2AStreamEventPayload struct {
	ConversationID uint              `json:"conversationId"`
	StepID         string            `json:"stepId,omitempty"`
	EventType      string            `json:"eventType"` // text_delta/thinking_delta/progress/tool_use
	Content        string            `json:"content,omitempty"`
	Tool           string            `json:"tool,omitempty"`
	ToolCallID     string            `json:"toolCallId,omitempty"`
	StatusKey      string            `json:"statusKey,omitempty"`  // 进度文案 i18n key
	ToolKind       string            `json:"toolKind,omitempty"`   // 工具分类 i18n key
	Progress       *A2AStreamProgress `json:"progress,omitempty"`
}

// A2AStreamProgress 结构化进度（phase/计数/百分比）。
type A2AStreamProgress struct {
	Phase   string `json:"phase,omitempty"`
	Current int64  `json:"current,omitempty"`
	Total   int64  `json:"total,omitempty"`
	Percent int    `json:"percent,omitempty"`
	Unit    string `json:"unit,omitempty"`
}

// A2ATaskResultPayload executor 执行完成后的最终结果。
//
// Status = completed / failed / cancelled。
// Body 携带 ExecutionResult 序列化（含摘要输出）。
type A2ATaskResultPayload struct {
	ConversationID uint            `json:"conversationId"`
	StepID         string          `json:"stepId,omitempty"`
	Status         string          `json:"status"` // completed/failed/cancelled
	Body           json.RawMessage `json:"body"`   // ExecutionResult 序列化
}

// A2ATaskResultAckPayload core 对最终结果的确认（用于 executor 侧幂等清理）。
type A2ATaskResultAckPayload struct {
	ConversationID uint   `json:"conversationId"`
	StepID         string `json:"stepId,omitempty"`
	Status         string `json:"status"`  // "accepted" / "duplicate"（去重命中）
	Message        string `json:"message,omitempty"`
}

// A2AHeartbeatBatchPayload executor 批量上报运行中 session 的心跳与状态。
//
// core 据此更新 conversations.heartbeat_at / lease_until（高频独立列）。
type A2AHeartbeatBatchPayload struct {
	Items  []A2AHeartbeatBatchItem `json:"items"`
	SentAt int64                   `json:"sentAt,omitempty"` // 毫秒时间戳
}

// A2AHeartbeatBatchItem 单个 session 的心跳项。
type A2AHeartbeatBatchItem struct {
	ConversationID uint   `json:"conversationId"`
	StepID         string `json:"stepId,omitempty"`
	Status         string `json:"status,omitempty"` // running/done/failed/killed
	StartedAt      int64  `json:"startedAt,omitempty"` // 毫秒时间戳
	UpdatedAt      int64  `json:"updatedAt,omitempty"` // 毫秒时间戳
}

// TaskKillPayload core 取消运行中任务，按 conversation_id 列表下发。
type TaskKillPayload struct {
	ConversationIDs []uint `json:"conversationIds"`
}

// ErrorPayload 通用错误负载。
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Retry   bool   `json:"retry,omitempty"` // 是否可重试
}

// PingPayload / PongPayload 心跳探测负载。
type PingPayload struct {
	SentAt int64 `json:"sentAt"`
}

type PongPayload struct {
	SentAt     int64 `json:"sentAt"`
	ReceivedAt int64 `json:"receivedAt,omitempty"`
}
