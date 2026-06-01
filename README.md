# 企灵 Agent / Qiling Agent

企灵 Agent 是面向企业微信私域销售场景的自动跟进与复盘智能体。

Qiling Agent helps private-domain sales teams analyze customer conversations, generate follow-up scripts, build customer profiles, review sales performance, and improve follow-up strategy with a rule-first AI development workflow.

## 产品方向

企灵 Agent 第一版聚焦企业微信私域销售场景：

```text
企业微信或上传聊天记录
-> Agent 自动分析客户上下文
-> 自动生成客户画像和意向判断
-> 自动生成下一步跟进话术
-> 销售确认后复制发送
-> 系统记录采纳和反馈
-> Agent 自动统计、复盘并提出改进建议
```

第一版不直接自动发送客户消息，而是采用“自动准备，人工确认，复制发送”的方式，降低企业微信接口、权限和客户信任风险。

## 工程原则

本项目采用规则先行的 AI 协作方式，避免失控式生成和不可维护代码：

- 使用 `AGENTS.md` 为 AI 提供清晰、恒定、可执行的项目规则。
- 将产品、销售、UX、UI、前端、后端、AI、测试、代码审查等人设固化为项目文档。
- 将评审、测试、Lint 和构建作为 AI 工作流的最终约束。
- 通过测试反馈和项目复盘持续优化 `AGENTS.md`。

## 技术栈

```text
Frontend: React + Vite + TypeScript
Backend: Go
Architecture: modular monolith for MVP
Database: MySQL for MVP persistence
LLM: mock/local runner first, real provider later
```

当前实现默认不调用外部付费 LLM。真实模型接入前必须先明确供应商、模型、密钥注入方式、成本边界和降级策略。

## 仓库结构

```text
frontend/   React + Vite + TypeScript frontend application
backend/    Go backend service
docs/       Product, agent, design, and architecture documents
scripts/    Local quality and load-test scripts
```

## 快速启动

### 1. 前端

```powershell
cd frontend
npm install
npm run dev
```

前端默认把 `/api` 代理到 `http://127.0.0.1:8080`。

### 2. 后端 Mock 模式

Mock 模式不需要 MySQL，适合快速看页面和接口结构：

```powershell
cd backend
go run ./cmd/server
```

### 3. 后端 MySQL 模式

MySQL 模式用于验证真实数据库读写：

```powershell
cd backend
$env:QILING_STORE_DRIVER="mysql"
$env:QILING_DATABASE_URL="root:your_password@tcp(127.0.0.1:3306)/qiling_agent?parseTime=true&charset=utf8mb4&loc=Local"
go run ./cmd/dbping
go run ./cmd/dbmigrate
go run ./cmd/server
```

本地开发数据库名建议为：

```text
qiling_agent
```

不要把真实数据库密码提交到仓库。

## 真实链路验证

Stage 19 已验证以下链路：

```text
上传中文多行聊天记录
-> 解析客户名和多条消息
-> 确认上传
-> 写入 customers / conversation_messages / followup_tasks / agent_runs
-> 生成长期记忆 facts
-> 前端可进入客户详情、待确认话术和长期记忆管理
```

关键验证点：

- 中文客户名不会变成乱码或 `???`。
- 多行聊天记录会拆成多条 `conversation_messages`。
- 包含“销售”的发送人会识别为 `sales`，其他发送人默认识别为 `customer`。
- 确认上传后会生成 `customer_id`、`followup_task_id`、`agent_run_id`。
- 长期记忆只返回 active facts，拒绝和修正会影响后续 Agent prompt。

## 常用命令

全量质量检查：

```powershell
.\scripts\check.ps1
```

前端：

```powershell
cd frontend
npm run lint
npm run typecheck
npm run test
npm run build
```

后端：

```powershell
cd backend
go test ./...
go run ./cmd/dbping
go run ./cmd/dbmigrate
go run ./cmd/dbreset -confirm qiling_agent
go run ./cmd/loadtest -base-url http://127.0.0.1:8080 -duration 30s -concurrency 16 -scenario read
```

MySQL 集成测试会重置配置的数据库，只能用于本地开发库：

```powershell
cd backend
$env:QILING_INTEGRATION_DATABASE_URL="root:your_password@tcp(127.0.0.1:3306)/qiling_agent?parseTime=true&charset=utf8mb4&loc=Local"
go test ./internal/store -run MySQLRepository -count=1 -v
```

`go run ./cmd/dbreset -confirm qiling_agent` 会删除本地开发表并重新迁移、写入演示数据。不要对生产或共享数据库执行。

## 当前文档

```text
AGENTS.md                                项目级 AI 协作规则
docs/product/product-vision.md            产品愿景和 MVP 范围
docs/agents/                             项目虚拟团队人设
docs/architecture/api-contracts.md        API 契约
docs/architecture/backend-architecture.md 后端架构
docs/architecture/frontend-architecture.md 前端架构
docs/architecture/short-term-memory.md    短期记忆
docs/architecture/long-term-memory.md     长期记忆
docs/architecture/memory-retrieval.md     向量召回与记忆检索
docs/engineering/performance-and-load-testing.md 性能和压测规则
```

## 第一版核心能力

- 企业微信客户和聊天记录接入。
- 上传或粘贴聊天记录。
- 客户画像自动生成。
- 意向等级和客户阶段判断。
- 推荐话术自动生成。
- 待确认跟进任务。
- 长期记忆事实管理、拒绝和修正。
- 自动统计和销售复盘。
- 改进建议和主管介入提醒。

## License

MIT
