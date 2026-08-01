// Package llm 的 Proxy 透传协议类型。
//
// 参考 agentia-engine/src/service/llm_proxy.go：
// core 的 LLM Proxy 是一个反向代理端点，验 TempLLMKey 后把请求透传到真实 LLM，
// 透传流式响应，并写 CallLog。executor 调 LLM 时统一走此端点。
package llm

// ResolvedProvider Proxy 验签后解析出的真实 Provider 信息（不含 API key 明文，
// APIKey 仅在 Proxy 进程内存中短暂持有用于转发，不落日志）。
type ResolvedProvider struct {
	ProviderID uint   `json:"providerId"`
	BaseURL    string `json:"baseUrl"`    // 真实 LLM endpoint base URL
	APIKey     string `json:"-"`          // 解密后的 API key（永不序列化/落日志）
	APIFormat  string `json:"apiFormat"`  // openai / anthropic（见 format.go）
	AuthScheme string `json:"authScheme"` // bearer / api_key（见 format.go）
	Model      string `json:"model"`      // 实际请求的模型
}

// ProxyRequestMeta Proxy 在写 CallLog 时记录的调用元数据。
type ProxyRequestMeta struct {
	ConversationID uint   `json:"conversationId"`
	StepID         string `json:"stepId,omitempty"`
	ProviderID     uint   `json:"providerId"`
	Model          string `json:"model"`
	LatencyMs      int64  `json:"latencyMs"`
	PromptTokens   int    `json:"promptTokens,omitempty"`
	CompletionTokens int  `json:"completionTokens,omitempty"`
	TotalTokens    int    `json:"totalTokens,omitempty"`
	Error          string `json:"error,omitempty"`
}

// CallType LLM 调用类型常量（对齐 agentia model/llm_call_log.go 的 CallType）。
const (
	CallTypeLLM          = "llm"          // 常规 LLM 调用
	CallTypeSummarization = "summarization" // 上下文压缩
	CallTypeEmbedding    = "embedding"    // 向量 embedding
)
