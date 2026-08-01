package a2a

import (
	"encoding/json"
	"strings"
	"testing"
)

// fixedNow 测试用的固定时间戳（脚本环境禁用 time.Now，测试直接传值）。
const fixedNow int64 = 1_700_000_000_000

// TestEnvelopeRoundTripSingleFrame 验证小消息单帧往返：编码为单帧，解码还原。
func TestEnvelopeRoundTripSingleFrame(t *testing.T) {
	payload := A2AStreamEventPayload{
		ConversationID: 42,
		EventType:      "text_delta",
		Content:        "hello",
	}
	env, err := NewEnvelope(fixedNow, EnvA2AStreamEvent, payload)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if env.Type != EnvA2AStreamEvent {
		t.Errorf("Type = %q, want %q", env.Type, EnvA2AStreamEvent)
	}
	if env.ID == "" {
		t.Error("ID should be set")
	}
	if env.Timestamp != fixedNow {
		t.Errorf("Timestamp = %d, want %d", env.Timestamp, fixedNow)
	}

	frames, err := EncodeEnvelopeFrames(env)
	if err != nil {
		t.Fatalf("EncodeEnvelopeFrames: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("expected single frame, got %d", len(frames))
	}

	var decoded Envelope
	if err := json.Unmarshal(frames[0], &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.ChunkID != "" {
		t.Errorf("single frame should not be chunked, got ChunkID=%q", decoded.ChunkID)
	}
	var got A2AStreamEventPayload
	if err := decoded.ParsePayload(&got); err != nil {
		t.Fatalf("ParsePayload: %v", err)
	}
	if got.ConversationID != 42 || got.EventType != "text_delta" || got.Content != "hello" {
		t.Errorf("payload mismatch: %+v", got)
	}
}

// TestEnvelopeChunkingReassembly 验证大消息分块后再重组还原。
func TestEnvelopeChunkingReassembly(t *testing.T) {
	// 构造一个超过 MaxEnvelopeBytes 的大 payload。
	big := strings.Repeat("A", MaxEnvelopeBytes+100)
	env, err := NewEnvelope(fixedNow, EnvA2AStreamEvent, A2AStreamEventPayload{
		ConversationID: 7,
		EventType:      "text_delta",
		Content:        big,
	})
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}

	frames, err := EncodeEnvelopeFrames(env)
	if err != nil {
		t.Fatalf("EncodeEnvelopeFrames: %v", err)
	}
	if len(frames) <= 1 {
		t.Fatalf("expected multiple chunks, got %d", len(frames))
	}

	// 解析所有分片帧。
	parsed := make([]*Envelope, 0, len(frames))
	for _, f := range frames {
		var e Envelope
		if err := json.Unmarshal(f, &e); err != nil {
			t.Fatalf("Unmarshal chunk: %v", err)
		}
		parsed = append(parsed, &e)
	}

	// 验证所有分片共享同一 ChunkID，且 ChunkTotal 一致。
	chunkID := parsed[0].ChunkID
	if chunkID == "" {
		t.Fatal("chunked frames must have ChunkID")
	}
	total := parsed[0].ChunkTotal
	for i, p := range parsed {
		if p.ChunkID != chunkID {
			t.Errorf("frame %d ChunkID=%q, want %q", i, p.ChunkID, chunkID)
		}
		if p.ChunkTotal != total {
			t.Errorf("frame %d ChunkTotal=%d, want %d", i, p.ChunkTotal, total)
		}
		if p.ChunkIndex != i {
			t.Errorf("frame %d ChunkIndex=%d, want %d", i, p.ChunkIndex, i)
		}
	}

	// 重组。
	complete, pending, err := DecodeEnvelopeFrames(parsed)
	if err != nil {
		t.Fatalf("DecodeEnvelopeFrames: %v", err)
	}
	if pending != 0 {
		t.Errorf("pending chunks = %d, want 0", pending)
	}
	if len(complete) != 1 {
		t.Fatalf("expected 1 complete envelope, got %d", len(complete))
	}

	var got A2AStreamEventPayload
	if err := complete[0].ParsePayload(&got); err != nil {
		t.Fatalf("ParsePayload: %v", err)
	}
	if got.Content != big {
		t.Errorf("reassembled content length = %d, want %d", len(got.Content), len(big))
	}
	if got.ConversationID != 7 {
		t.Errorf("ConversationID = %d, want 7", got.ConversationID)
	}
}

// TestDecodePartialChunkSet 验证未收齐的分块不产出完整信封。
func TestDecodePartialChunkSet(t *testing.T) {
	big := strings.Repeat("B", MaxEnvelopeBytes+100)
	env, err := NewEnvelope(fixedNow, EnvA2AStreamEvent, A2AStreamEventPayload{Content: big})
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	frames, err := EncodeEnvelopeFrames(env)
	if err != nil {
		t.Fatalf("EncodeEnvelopeFrames: %v", err)
	}

	// 只解析前一半分片。
	partial := make([]*Envelope, 0, len(frames)/2)
	for i := 0; i < len(frames)/2; i++ {
		var e Envelope
		if err := json.Unmarshal(frames[i], &e); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		partial = append(partial, &e)
	}

	complete, pending, err := DecodeEnvelopeFrames(partial)
	if err != nil {
		t.Fatalf("DecodeEnvelopeFrames: %v", err)
	}
	if len(complete) != 0 {
		t.Errorf("partial set should yield 0 complete, got %d", len(complete))
	}
	if pending != 1 {
		t.Errorf("pending = %d, want 1", pending)
	}
}

// TestDecodeMixedFrames 验证分块消息与非分块消息混合处理。
func TestDecodeMixedFrames(t *testing.T) {
	// 一条普通小消息 + 一条分块大消息。
	small, err := NewEnvelope(fixedNow, EnvPing, PingPayload{SentAt: fixedNow})
	if err != nil {
		t.Fatalf("NewEnvelope small: %v", err)
	}
	smallFrames, err := EncodeEnvelopeFrames(small)
	if err != nil {
		t.Fatalf("Encode small: %v", err)
	}

	big, err := NewEnvelope(fixedNow, EnvA2AStreamEvent, A2AStreamEventPayload{
		Content: strings.Repeat("C", MaxEnvelopeBytes+50),
	})
	if err != nil {
		t.Fatalf("NewEnvelope big: %v", err)
	}
	bigFrames, err := EncodeEnvelopeFrames(big)
	if err != nil {
		t.Fatalf("Encode big: %v", err)
	}

	// 交替解析。
	var parsed []*Envelope
	parseFrame := func(raw []byte) {
		var e Envelope
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		parsed = append(parsed, &e)
	}
	parseFrame(smallFrames[0])
	for _, f := range bigFrames {
		parseFrame(f)
	}

	complete, pending, err := DecodeEnvelopeFrames(parsed)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if pending != 0 {
		t.Errorf("pending = %d, want 0", pending)
	}
	if len(complete) != 2 {
		t.Fatalf("expected 2 complete, got %d", len(complete))
	}
	// 按 timestamp 排序后两条都在，类型正确。
	types := map[string]bool{}
	for _, c := range complete {
		types[c.Type] = true
	}
	if !types[EnvPing] || !types[EnvA2AStreamEvent] {
		t.Errorf("expected both ping and stream event, got %v", types)
	}
}

// TestParsePayloadEmpty 验证空 payload 的 ParsePayload 不修改目标。
func TestParsePayloadEmpty(t *testing.T) {
	env := &Envelope{Type: EnvPong}
	var p PongPayload
	if err := env.ParsePayload(&p); err != nil {
		t.Fatalf("ParsePayload on empty: %v", err)
	}
	if p.SentAt != 0 {
		t.Errorf("target should be unmodified, got SentAt=%d", p.SentAt)
	}
}

// TestNewEnvelopeWithRequest 验证 RequestID 透传。
func TestNewEnvelopeWithRequest(t *testing.T) {
	env, err := NewEnvelopeWithRequest(fixedNow, EnvA2AResponse, "req-123", A2AResponsePayload{Status: 200})
	if err != nil {
		t.Fatalf("NewEnvelopeWithRequest: %v", err)
	}
	if env.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want %q", env.RequestID, "req-123")
	}
}
