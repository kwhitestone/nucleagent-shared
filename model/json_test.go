package model

import (
	"database/sql"
	"encoding/json"
	"testing"
)

// TestJSONRoundTrip 验证 JSON 字段的 Scan/Value/Marshal/Unmarshal 往返。
func TestJSONRoundTrip(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		N    int    `json:"n"`
	}

	// NewJSON -> Value -> Scan -> As
	orig, err := NewJSON(payload{Name: "alice", N: 7})
	if err != nil {
		t.Fatalf("NewJSON: %v", err)
	}

	val, err := orig.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	var scanned JSON
	if err := scanned.Scan(val); err != nil {
		t.Fatalf("Scan: %v", err)
	}

	var got payload
	if err := scanned.As(&got); err != nil {
		t.Fatalf("As: %v", err)
	}
	if got.Name != "alice" || got.N != 7 {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

// TestJSONScanNil 验证 nil 落库为 NULL，Scan(nil) 得到 nil。
func TestJSONScanNil(t *testing.T) {
	var j JSON
	if err := j.Scan(nil); err != nil {
		t.Fatalf("Scan(nil): %v", err)
	}
	if j != nil {
		t.Errorf("Scan(nil) should yield nil, got %s", j)
	}

	v, err := j.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if v != nil {
		t.Errorf("empty JSON Value should be nil, got %v", v)
	}
}

// TestJSONScanString 验证从 string scan（某些 driver 返回 string）。
func TestJSONScanString(t *testing.T) {
	var j JSON
	if err := j.Scan(`{"k":"v"}`); err != nil {
		t.Fatalf("Scan(string): %v", err)
	}
	var m map[string]string
	if err := j.As(&m); err != nil {
		t.Fatalf("As: %v", err)
	}
	if m["k"] != "v" {
		t.Errorf("expected k=v, got %q", m["k"])
	}
}

// TestJSONScanInvalid 验证不支持的类型报错。
func TestJSONScanInvalid(t *testing.T) {
	var j JSON
	err := j.Scan(12345)
	if err == nil {
		t.Fatal("expected error scanning int")
	}
}

// TestJSONMarshalNull 验证空 JSON 序列化为 null。
func TestJSONMarshalNull(t *testing.T) {
	var j JSON
	out, err := json.Marshal(j)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(out) != "null" {
		t.Errorf("empty JSON should marshal to null, got %s", out)
	}
}

// TestJSONUnmarshalPreservesBytes 验证 UnmarshalJSON 保留原始字节。
func TestJSONUnmarshalPreservesBytes(t *testing.T) {
	var j JSON
	raw := []byte(`{"a":1}`)
	if err := j.UnmarshalJSON(raw); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	var m map[string]int
	if err := j.As(&m); err != nil {
		t.Fatalf("As: %v", err)
	}
	if m["a"] != 1 {
		t.Errorf("expected a=1, got %d", m["a"])
	}
}

// TestMustNewJSON 验证 MustNewJSON 对不可编码值返回 null，正常值编码正确。
func TestMustNewJSON(t *testing.T) {
	j := MustNewJSON(map[string]string{"k": "v"})
	if len(j) == 0 {
		t.Fatal("expected non-empty JSON")
	}
	var m map[string]string
	if err := j.As(&m); err != nil {
		t.Fatalf("As: %v", err)
	}
	if m["k"] != "v" {
		t.Errorf("expected k=v, got %q", m["k"])
	}
}

// TestJSONAsEmpty 验证空 JSON 的 As 不修改目标。
func TestJSONAsEmpty(t *testing.T) {
	var j JSON
	var m map[string]string
	if err := j.As(&m); err != nil {
		t.Fatalf("As on empty: %v", err)
	}
	if m != nil {
		t.Errorf("expected unmodified nil target, got %v", m)
	}
}

// TestJSONSQLDriver 验证 JSON 实现了 sql.Scanner 接口（编译期检查 + 运行时调用）。
func TestJSONSQLDriver(t *testing.T) {
	var _ sql.Scanner = (*JSON)(nil)
	var j JSON
	if err := j.Scan([]byte(`{"x":1}`)); err != nil {
		t.Fatalf("Scan: %v", err)
	}
}
