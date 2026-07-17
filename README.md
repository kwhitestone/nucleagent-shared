# nucleagent-shared

Nucleagent 共享类型库：GORM 数据模型、A2A 协议类型、LLM 协议类型。

被 nucleagent-core / nucleagent-executor 共同依赖（auth 不依赖，使用框架自带 User 表）。

## 目录结构

```
nucleagent-shared/
├── model/           # GORM 数据模型（11 张表，对齐 docs/03-data-models.md）
│   ├── json.go              # JSON 字段类型（json.RawMessage 封装 + 辅助函数）
│   ├── agent_template.go    # agent_templates
│   ├── agent_instance.go    # agent_instances
│   ├── conversation.go      # conversations
│   ├── message.go           # messages
│   ├── step.go              # steps
│   ├── skill.go             # skills
│   ├── skill_binding.go     # skill_bindings
│   ├── tool.go              # tools
│   ├── provider.go          # providers
│   ├── call_log.go          # call_logs
│   └── project.go           # projects
├── a2a/             # core ↔ executor 线协议类型（ExecutionRequest / ExecutionResult / StreamReporter / TaskSession）
└── llm/             # LLM 协议类型（TempLLMKey / APIFormat / AuthScheme）
```

## 约束

- 仅放数据模型定义与协议类型，禁止写业务逻辑
- model 只含 GORM tag + 基础方法，字段精简（能塞 JSON 的不拆列）
- 高频查询/更新字段必须在顶层列（不在 JSON 里）
- JSON 列统一用 `model.JSON` 类型
- 禁止 import 任何业务 repo（core / auth / executor）

## 构建

```bash
go build ./...
go vet ./...
```
