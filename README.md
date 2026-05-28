# 企灵 Agent / Qiling Agent

企业微信私域销售自动跟进与复盘智能体。

Qiling Agent helps private-domain sales teams analyze customer conversations, generate follow-up scripts, build customer profiles, review sales performance, and improve follow-up strategy with a rule-first AI development workflow.

## 产品方向

企灵 Agent 的第一版聚焦企业微信私域销售场景：

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


## Tech Stack

```text
Frontend: React + Vite + TypeScript
Backend: Go
Architecture: modular monolith for MVP
Database direction: PostgreSQL, with pgvector evaluation reserved for later AI memory/RAG work
```

## Repository Structure

```text
frontend/   React + Vite + TypeScript frontend application
backend/    Go backend service
docs/       Product, agent, design, and architecture documents
```

## Development Commands

All checks:

```powershell
.\scripts\check.ps1
```

Frontend:

```powershell
cd frontend
npm run dev
npm run lint
npm run typecheck
npm run test
npm run build
```

Backend:

```powershell
cd backend
go test ./...
go run ./cmd/server
```

Local frontend development proxies `/api` requests to `http://127.0.0.1:8080`, so start the backend before checking API-backed pages.

## 当前文档

```text
AGENTS.md                     项目级 AI 协作规则
docs/product/product-vision.md 产品愿景和 MVP 范围
docs/agents/                  项目虚拟团队人设
```

## 第一版核心能力

- 企业微信客户和聊天记录接入
- 上传或粘贴聊天记录
- 客户画像自动生成
- 意向等级和客户阶段判断
- 推荐话术自动生成
- 待确认跟进任务
- 自动统计和销售复盘
- 改进建议和主管介入提醒

## License

TBD
