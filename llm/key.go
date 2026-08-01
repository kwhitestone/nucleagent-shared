// Package llm 定义 LLM 代理相关的协议类型。
//
// core 暴露 LLM Proxy 端点，用 TempLLMKey 鉴权后反向代理到真实 LLM endpoint。
// executor 永不直接持有 API key（硬约束，见开发计划 §D1）。
package llm

import "time"

// TempLLMKey core 的 LLM Proxy 签发给 executor 的临时密钥。
//
// executor 不直接持有 API key。core 为每个 conversation 签发临时 key，
// executor 在 LLM 请求头里携带 x-llm-proxy-key，core Proxy 端点验签后解析出
// 真实 Provider + 解密 API key，反向代理到真实 LLM endpoint。
//
// 绑定 ConversationID / UserID 用于 Proxy 鉴权时回溯归属，控制泄露面。
type TempLLMKey struct {
	Key            string    `json:"key"`            // 临时密钥（llmk_ 前缀）
	ConversationID uint      `json:"conversationId"` // 绑定对话
	UserID         uint      `json:"userId,omitempty"`
	ProviderID     uint      `json:"providerId"` // 对应 Provider
	Model          string    `json:"model,omitempty"`
	ExpiresAt      time.Time `json:"expiresAt"` // 过期时间
}

// KeyHeader 是携带 TempLLMKey 的请求头名。
const KeyHeader = "x-llm-proxy-key"

// KeyPrefix TempLLMKey 的明文前缀，便于日志/识别。
const KeyPrefix = "llmk_"
