// Package a2a 定义 core ↔ executor 之间的 A2A 线协议类型。
//
// 文件组织：
//   - backend.go    : Backend 接口、ExecutionRequest/ExecutionResult/StreamReporter
//   - envelope.go   : WebSocket 传输层 Envelope 与分块机制
//   - payloads.go   : 各 Envelope 类型对应的业务负载
//   - session.go    : TaskSession 与 SessionStore 协议
//
// 仅包含协议数据结构与接口，不含业务逻辑。core 与 executor 共用此包，
// 保证两端协议一致。
package a2a
