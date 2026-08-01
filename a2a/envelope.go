// Package a2a 定义 core ↔ executor 之间的 A2A 线协议类型。
//
// 本文件定义 WebSocket 传输层的 Envelope（信封）与分块机制。
// 参考 agentia-executor/src/internal/protocol/protocol.go：
//   - Envelope 是所有 WS 消息的统一外层结构，携带 type/id/request_id/trace_id/ts/payload。
//   - 超过 MaxEnvelopeBytes 的消息按 base64 切片分块发送，接收方按 chunk_id 重组。
//
// 仅包含协议数据结构与编解码函数，不含业务逻辑。
package a2a

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// 协议版本与分块阈值。
const (
	EnvelopeVersion = 1

	// MaxEnvelopeBytes 单个 WS 帧的最大字节数，超过则触发分块。
	// LLM 长输出 / 大 artifact 必须走分块，否则会被 WS 帧大小限制截断。
	MaxEnvelopeBytes = 512 * 1024

	// chunkPayloadBytes 分块时每片的 base64 负载字节数。
	chunkPayloadBytes = 128 * 1024
)

// Envelope 类型常量。
const (
	EnvHandshake         = "handshake"          // executor -> core：握手（携带 capability/descriptor）
	EnvHandshakeAck      = "handshake_ack"      // core -> executor：握手确认
	EnvA2ARequest        = "a2a_request"        // core -> executor：下发执行任务
	EnvA2AResponse       = "a2a_response"       // executor -> core：任务 ACK（working）
	EnvA2AStreamEvent    = "a2a_stream_event"   // executor -> core：流式事件（text_delta/tool_use/...）
	EnvA2ATaskResult     = "a2a_task_result"    // executor -> core：任务最终结果（done/error）
	EnvA2ATaskResultAck  = "a2a_task_result_ack" // core -> executor：结果确认
	EnvA2AHeartbeatBatch = "a2a_heartbeat_batch" // executor -> core：心跳批量上报
	EnvTaskKill          = "task_kill"          // core -> executor：取消运行中任务
	EnvPing              = "ping"               // 双向：心跳探测
	EnvPong              = "pong"               // 双向：心跳响应
	EnvError             = "error"              // 双向：错误
)

// Envelope 是 core ↔ executor 之间所有 WebSocket 消息的统一信封。
//
// 业务数据序列化为 JSON 放在 Payload；接收方按 Type 选择对应 payload 类型解析。
// 分块消息（ChunkID != ""）的 Payload 形如 {"_chunk": "<base64 片段>"}，
// 接收方需用 DecodeEnvelopeFrames 重组后再解析原始 payload。
type Envelope struct {
	Version    int             `json:"v"`                 // 协议版本（EnvelopeVersion）
	Type       string          `json:"type"`              // 消息类型（Env* 常量）
	ID         string          `json:"id"`                // 消息 ID（UUID，用于请求/响应配对）
	RequestID  string          `json:"request_id,omitempty"` // 关联的请求 ID（响应/事件回指请求）
	TraceID    string          `json:"trace_id,omitempty"` // 链路追踪 ID
	Timestamp  int64           `json:"ts"`                // 毫秒时间戳
	Payload    json.RawMessage `json:"payload,omitempty"` // 业务负载
	ChunkID    string          `json:"chunk_id,omitempty"`    // 分块组 ID（同一原始消息的各片相同）
	ChunkIndex int             `json:"chunk_index,omitempty"` // 当前片序号（0-based）
	ChunkTotal int             `json:"chunk_total,omitempty"` // 总片数
}

// NewEnvelope 构造一个带新 UUID 和当前时间戳的信封。ts 由调用方传入（脚本环境禁用 Date.now）。
func NewEnvelope(nowMillis int64, kind string, payload any) (*Envelope, error) {
	var raw json.RawMessage
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		raw = data
	}
	return &Envelope{
		Version:   EnvelopeVersion,
		Type:      kind,
		ID:        uuid.NewString(),
		Timestamp: nowMillis,
		Payload:   raw,
	}, nil
}

// NewEnvelopeNow 构造一个用 time.Now() 作时间戳的信封（普通 Go 进程用）。
func NewEnvelopeNow(kind string, payload any) (*Envelope, error) {
	return NewEnvelope(time.Now().UnixMilli(), kind, payload)
}

// NewEnvelopeWithRequest 构造一个带 RequestID 的信封（用于响应/事件回指请求）。
func NewEnvelopeWithRequest(nowMillis int64, kind, requestID string, payload any) (*Envelope, error) {
	env, err := NewEnvelope(nowMillis, kind, payload)
	if err != nil {
		return nil, err
	}
	env.RequestID = requestID
	return env, nil
}

// ParsePayload 把 Payload 反序列化到目标结构。空 Payload 不修改目标。
func (e *Envelope) ParsePayload(target any) error {
	if len(e.Payload) == 0 {
		return nil
	}
	return json.Unmarshal(e.Payload, target)
}

// EncodeEnvelopeFrames 把一个信封编码为若干 WS 帧。
//
// 不超过 MaxEnvelopeBytes 时返回单帧；超过则把 payload base64 编码后切片，
// 每片包装成一个独立 Envelope（ChunkID 指向原始消息 ID），接收方用
// DecodeEnvelopeFrames 重组。
func EncodeEnvelopeFrames(env *Envelope) ([][]byte, error) {
	data, err := json.Marshal(env)
	if err != nil {
		return nil, err
	}
	if len(data) <= MaxEnvelopeBytes {
		return [][]byte{data}, nil
	}

	// 分块：payload 转 base64 后按 chunkPayloadBytes 切片。
	payload := base64.StdEncoding.EncodeToString(env.Payload)
	total := (len(payload) + chunkPayloadBytes - 1) / chunkPayloadBytes
	frames := make([][]byte, 0, total)
	chunkID := env.ID
	for i := 0; i < total; i++ {
		start := i * chunkPayloadBytes
		end := start + chunkPayloadBytes
		if end > len(payload) {
			end = len(payload)
		}
		chunkPayload, err := json.Marshal(map[string]string{"_chunk": payload[start:end]})
		if err != nil {
			return nil, err
		}
		chunk := Envelope{
			Version:    env.Version,
			Type:       env.Type,
			ID:         uuid.NewString(),
			RequestID:  env.RequestID,
			TraceID:    env.TraceID,
			Timestamp:  env.Timestamp,
			Payload:    chunkPayload,
			ChunkID:    chunkID,
			ChunkIndex: i,
			ChunkTotal: total,
		}
		frame, err := json.Marshal(&chunk)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)
	}
	return frames, nil
}

// chunkFragment 是分片信封的 payload 结构。
type chunkFragment struct {
	Chunk string `json:"_chunk"`
}

// DecodeEnvelopeFrames 把接收到的若干帧重组为原始信封。
//
// 输入帧按到达顺序传入。非分块帧（ChunkID == ""）直接返回；分块帧按
// ChunkID 分组，收集齐 ChunkTotal 片后按 ChunkIndex 拼接 base64，解码出
// 原始 payload 注入回信封。
//
// 返回 (重组完成的信封列表, 仍在等待的 chunk 组数, error)。
// 重组完成的信封按其原始 Timestamp 排序返回。
func DecodeEnvelopeFrames(frames []*Envelope) ([]*Envelope, int, error) {
	byID := make(map[string][]*Envelope) // chunkID -> 片段
	var complete []*Envelope
	for _, f := range frames {
		if f == nil {
			continue
		}
		if f.ChunkID == "" {
			// 非分块帧，直接完成。
			complete = append(complete, f)
			continue
		}
		byID[f.ChunkID] = append(byID[f.ChunkID], f)
	}

	for id, parts := range byID {
		if len(parts) == 0 {
			continue
		}
		total := parts[0].ChunkTotal
		if len(parts) < total {
			// 尚未收齐，保留等待。
			continue
		}
		// 按 ChunkIndex 排序。
		ordered := make([]*Envelope, total)
		for _, p := range parts {
			if p.ChunkIndex < 0 || p.ChunkIndex >= total {
				continue
			}
			ordered[p.ChunkIndex] = p
		}
		// 拼接 base64。
		concat := make([]byte, 0, total*chunkPayloadBytes)
		ok := true
		for _, p := range ordered {
			if p == nil {
				ok = false
				break
			}
			var frag chunkFragment
			if err := json.Unmarshal(p.Payload, &frag); err != nil {
				return nil, 0, err
			}
			concat = append(concat, frag.Chunk...)
		}
		if !ok {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(string(concat))
		if err != nil {
			return nil, 0, err
		}
		// 用第一片作为基础信封，注入重组后的 payload。
		base := *ordered[0]
		base.ID = id
		base.Payload = decoded
		base.ChunkID = ""
		base.ChunkIndex = 0
		base.ChunkTotal = 0
		complete = append(complete, &base)
		delete(byID, id)
	}

	// 按 Timestamp 排序，保证消息顺序。
	sortEnvelopesByTimestamp(complete)
	return complete, len(byID), nil
}

// sortEnvelopesByTimestamp 原地按 Timestamp 升序排序（稳定）。
func sortEnvelopesByTimestamp(envs []*Envelope) {
	// 简单插入排序：消息数量小，避免引入 sort 包的额外依赖开销。
	for i := 1; i < len(envs); i++ {
		for j := i; j > 0 && envs[j-1].Timestamp > envs[j].Timestamp; j-- {
			envs[j-1], envs[j] = envs[j], envs[j-1]
		}
	}
}
