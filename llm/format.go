package llm

// APIFormat Provider 的 API 协议格式。
type APIFormat string

const (
	APIFormatOpenAI    APIFormat = "openai"    // OpenAI 兼容
	APIFormatAnthropic APIFormat = "anthropic" // Anthropic Messages API
)

// AuthScheme Provider 的鉴权方案。
type AuthScheme string

const (
	AuthSchemeBearer AuthScheme = "bearer"  // Authorization: Bearer <key>
	AuthSchemeAPIKey AuthScheme = "api_key" // 自定义 header 携带 key
)
