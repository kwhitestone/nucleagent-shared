# nucleagent-shared

Nucleagent 共享类型库：GORM model、协议类型、API 契约定义。

被 nucleagent-core / nucleagent-executor / nucleagent-auth 共同依赖。

## 目录结构

```
nucleagent-shared/
├── model/           # GORM 数据模型 (与旧版 Agentia 100% 兼容)
├── a2a/             # A2A 协议类型
├── executor/        # Executor 协议类型 (ExecutionRequest, ExecutionResult, StreamReporter)
└── llm/             # LLM 协议类型 (TempLLMKey, APIFormat, AuthScheme)
```
