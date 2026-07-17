// Package llm 定义 LLM 代理相关的协议类型。
package llm

import "time"

// TempLLMKey core 的 LLM Proxy 签发给 executor 的临时密钥（executor 不直接持有 API key）。
type TempLLMKey struct {
	Key        string    `json:"key"`        // 临时密钥
	ProviderID uint      `json:"providerId"` // 对应 Provider
	Model      string    `json:"model,omitempty"`
	ExpiresAt  time.Time `json:"expiresAt"`  // 过期时间
}
