# nucleagent-shared

共享类型库：GORM model + 协议类型。被 core / executor 依赖。

## 构建

```bash
go build ./...
go vet ./...
```

## 架构约束

- 这个 repo 只放数据模型定义和协议类型，禁止写业务逻辑
- GORM model 定义参考 docs/03-data-models.md（在 nucleagent-docs repo）
- 11 张表，每张表字段精简，能塞 JSON 的不拆列
- 高频查询/更新字段必须在顶层列（不在 JSON 里）
- JSON 类型用 `json.RawMessage` 封装，带辅助函数

## 目录结构

```
model/           # GORM model（agent_templates, agent_instances, conversations, messages, steps, skills, skill_bindings, tools, providers, call_logs, projects）
a2a/             # A2A 协议类型（ExecutionRequest, ExecutionResult, StreamReporter, TaskSession）
llm/             # LLM 协议类型（TempLLMKey, APIFormat, AuthScheme）
```

## 边界

- **Always**: 新增表必须同步更新 docs/03-data-models.md
- **Never**: 禁止 import 任何业务 repo（core/auth/executor）
- **Never**: 禁止在 model 里写业务方法（只有 GORM tag + 基础方法）
