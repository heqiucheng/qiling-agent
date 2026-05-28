# 企灵 Agent Go 后端架构设计

**状态**：Draft for backend scaffold  
**版本**：v0.1  
**日期**：2026-05-28  
**后端语言**：Go  
**依赖文档**：`AGENTS.md`、`docs/engineering/backend-engineering-rules.md`、`docs/specs/mvp-prd.md`、`docs/architecture/api-contracts.md`

## 1. 技术栈决策

后端主语言确定为 **Go**。该决策作为项目级技术栈约束，后续后端 API、企业微信集成、上传解析、Agent 编排、任务队列、统计复盘和向量库接入均按 Go 设计。

选择 Go 的原因：

- 并发和后台任务处理能力强，适合上传解析、企业微信同步、Agent 异步任务。
- 编译型语言，部署简单，运行时稳定。
- 标准库成熟，适合构建清晰的 HTTP API 和集成客户端。
- 类型系统足够约束 API、领域模型和 Agent 输出结构。
- 性能和资源占用适合企业级后台服务。

## 2. 架构模式

第一版采用 **模块化单体**，不是微服务。

原因：

- MVP 阶段边界还在验证，过早拆微服务会增加复杂度。
- 模块化单体可以保持部署简单，同时用清晰包结构避免屎山。
- 后续如果某些模块需要独立扩展，例如 Agent Worker 或企业微信同步，可以再拆服务。

默认分层：

```text
cmd/server                  启动入口
internal/http               路由、中间件、响应格式
internal/httpx              HTTP 通用响应、请求上下文、中间件工具
internal/config             配置加载
internal/domain             领域模型、枚举、业务规则
internal/service            业务用例编排
internal/store              数据库访问
internal/integration        第三方集成客户端
internal/agent              Agent 任务、Prompt、输出校验
internal/job                异步任务和 worker
internal/audit              审计日志
internal/apperror           统一错误码和错误结构
internal/testutil           测试辅助
```

控制器不能直接写复杂业务逻辑；业务服务不能直接拼 HTTP 响应；数据库查询不能散落在业务层。

## 3. 推荐目录结构

```text
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── app.go
│   ├── config/
│   │   └── config.go
│   ├── http/
│   │   ├── router.go
│   │   └── handler/
│   ├── httpx/
│   │   ├── middleware.go
│   │   └── response.go
│   ├── domain/
│   │   ├── customer.go
│   │   ├── conversation.go
│   │   ├── followup.go
│   │   └── agent.go
│   ├── service/
│   ├── store/
│   ├── integration/
│   │   ├── wecom/
│   │   └── llm/
│   ├── agent/
│   │   ├── prompt/
│   │   ├── schema/
│   │   └── runner.go
│   ├── job/
│   ├── audit/
│   ├── apperror/
│   └── testutil/
├── migrations/
├── go.mod
└── go.sum
```

## 4. HTTP 框架建议

MVP 推荐优先使用：

```text
Go 标准库 net/http + chi router
```

原因：

- 轻量、清晰、可测试。
- 避免过早绑定复杂框架。
- 中间件、路由分组、REST API 足够。

如果后续需要更完整生态，也可以评估 Gin，但第一版优先保持简单。

## 5. 数据库策略

MVP 推荐：

```text
PostgreSQL
```

原因：

- 关系型数据适合客户、会话、消息、任务、审计和复盘报表。
- 后续可用 `pgvector` 承接向量检索，减少早期多数据库复杂度。
- 事务、索引、JSONB、全文检索能力较完整。

本地开发可以先用 SQLite 或 Postgres Docker，但最终 schema 按 PostgreSQL 设计。

## 6. 核心模块

### 6.1 customers

职责：

```text
客户资料
客户阶段
意向等级
负责人
客户标签
客户画像摘要
```

### 6.2 conversations

职责：

```text
会话
消息
上传解析结果
聊天记录时间线
消息归属确认
```

### 6.3 uploads

职责：

```text
文件上传
粘贴文本导入
格式预检查
解析任务创建
解析结果确认
```

### 6.4 followup_tasks

职责：

```text
待确认话术任务
复制状态
跳过状态
标记不准反馈
换一种话术记录
```

### 6.5 agent

职责：

```text
客户画像生成
阶段识别
意向评分
话术生成
风险提示
结构化输出校验
Prompt 版本记录
AgentRun 记录
```

### 6.6 review_reports

职责：

```text
客户分层统计
风险客户列表
话术采纳率
回复率和阶段推进
销售复盘建议
主管介入建议
```

### 6.7 integrations/wecom

职责：

```text
企业微信配置
客户同步
标签同步
聊天记录或消息存档集成预留
接口本地验证
错误映射
限流和重试
```

## 7. API 响应格式

统一成功响应：

```json
{
  "data": {},
  "error": null,
  "meta": {
    "request_id": "req_xxx",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

统一错误响应：

```json
{
  "data": null,
  "error": {
    "code": "CUSTOMER_NOT_FOUND",
    "message": "客户不存在或无权访问",
    "details": {}
  },
  "meta": {
    "request_id": "req_xxx",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

Go 中应集中实现：

```text
internal/httpx/response.go
internal/apperror
```

## 8. API 初版边界

```text
GET    /api/health
POST   /api/uploads/conversations
GET    /api/uploads/{id}
POST   /api/uploads/{id}/confirm
GET    /api/customers
GET    /api/customers/{id}
GET    /api/customers/{id}/conversations
GET    /api/followup-tasks
POST   /api/followup-tasks/{id}/copy
POST   /api/followup-tasks/{id}/skip
POST   /api/followup-tasks/{id}/mark-wrong
POST   /api/followup-tasks/{id}/regenerate
GET    /api/review-reports/summary
GET    /api/agent-runs/{id}
```

对接真实前端前，必须遵守 `docs/architecture/api-contracts.md`。接口实现、前端 API client、Mock 数据和测试断言都以该契约为准。

## 9. Agent 和 AI 集成

第一版后端对 AI 的处理原则：

- AI 调用封装在 `internal/integration/llm`。
- Prompt 和输出 schema 放在 `internal/agent`。
- AI 输出必须结构化校验，不允许直接把自由文本写入业务状态。
- 每次执行记录 `AgentRun`，包含输入摘要、Prompt 版本、模型、输出校验结果、风险标记和失败原因。
- 高风险内容只生成待确认话术，不直接触发外部发送。

## 10. 企业微信集成规则

后续接企业微信时必须遵守：

```text
先读官方接口文档
先写本地 smoke test 或 mock
先拿到脱敏请求/响应样例
先封装 integration client
先映射错误码和限流策略
再写业务功能
```

不能盲写企业微信业务功能后再一点点改。

## 11. 异步任务

以下任务默认异步化：

```text
聊天记录解析
客户画像生成
话术生成
复盘报告生成
企业微信同步
向量化入库
统计聚合
```

MVP 可先使用 Go goroutine + 简单任务表/状态机；如果任务量上来，再引入队列，例如 Redis Stream、Asynq 或消息队列。

## 12. 向量库和记忆预留

后续进入 AI 记忆阶段时再详细设计，但当前预留方向：

```text
长期记忆：客户画像、历史沟通、成交偏好、销售反馈
短期记忆：当前会话上下文、最近 N 条消息、当前任务状态
RAG 数据：产品知识、案例素材、销售 SOP、历史有效话术
向量存储：优先评估 PostgreSQL + pgvector，必要时再评估专用向量库
```

当前阶段不直接引入向量库，避免过早复杂化。

## 13. 测试策略

Go 后端必须建立：

```text
go test ./...
handler 层接口测试
service 层业务规则测试
store 层数据库测试
integration 层 mock 测试
agent 输出 schema 测试
上传解析测试
权限和租户隔离测试
```

后续压力测试：

```text
客户列表分页
聊天记录读取
批量上传解析
批量话术生成
复盘统计
企业微信同步
```

## 14. 阶段实现顺序

推荐顺序：

```text
1. Go module 和 backend 目录
2. 配置、日志、错误、响应格式
3. health API
4. 领域模型和 Mock store
5. 上传聊天记录 API contract
6. 客户和任务 API contract
7. 前端替换 Mock 为本地 API
8. 数据库 schema 和 store
9. Agent mock runner
10. LLM 集成验证
11. 企业微信集成验证
12. 异步任务和复盘统计
```

## 15. 当前阶段结论

- 后端语言：Go。
- 架构：模块化单体。
- HTTP：优先 `net/http + chi`。
- 数据库：按 PostgreSQL 设计，后续可评估 pgvector。
- AI：结构化输出 + AgentRun 可追溯。
- 企业微信：必须先本地验证接口，再写业务功能。
