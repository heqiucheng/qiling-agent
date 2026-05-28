# 企灵 Agent 前端架构设计

**状态**：Draft for scaffold  
**版本**：v0.1  
**日期**：2026-05-28  
**依赖文档**：`AGENTS.md`、`docs/design/ui-design-system.md`、`docs/design/page-wireframes.md`、`docs/specs/mvp-prd.md`

## 1. 目标

前端第一阶段要先跑通 MVP 演示主链路：

```text
上传/粘贴聊天记录
-> 查看解析结果
-> 查看客户画像和 Agent 推荐
-> 复制话术
-> 查看复盘摘要
```

架构目标：

```text
统一风格
组件复用
样式集中
类型清晰
Mock 先行
API Contract 先行
易于后续接真实后端
可测试
可维护
```

## 2. 技术选型

### 2.1 推荐选型

| 方向 | 选择 | 原因 |
| --- | --- | --- |
| 框架 | React | 生态成熟，组件边界清晰，适合后台和复杂状态界面 |
| 构建工具 | Vite | 启动快，配置简单，适合从 0 快速搭建 |
| 语言 | TypeScript | Agent 输出、API Contract、业务状态都需要类型约束 |
| 路由 | React Router | 页面结构清晰，适合 Web 后台 |
| 样式 | CSS Variables + CSS Modules 或普通 CSS 分层 | 符合设计 Token 和禁止页面内乱写样式的规则 |
| 图标 | lucide-react | 图标统一、轻量、适合企业后台 |
| 表单 | React Hook Form 可后置 | MVP 表单不复杂，可先不用或轻量接入 |
| 测试 | Vitest + React Testing Library | 适合组件和业务逻辑测试 |
| E2E | Playwright 后置 | 页面主链路稳定后补充 |
| Mock | 本地 mock service 模块，后续可接 MSW | 先跑通 UI 和流程，再替换真实 API |

### 2.2 不建议第一版引入

- 不先引入大型 UI 框架并重度覆盖样式，避免视觉系统被框架绑死。
- 不先引入复杂状态管理库，如 Redux。
- 不先做微前端。
- 不先做完整企业微信侧边栏适配，只预留布局能力。

## 3. 前端目录结构

推荐结构：

```text
src/
├── app/
│   ├── App.tsx
│   ├── routes.tsx
│   └── providers.tsx
├── assets/
├── components/
│   ├── ui/
│   ├── layout/
│   ├── agent/
│   ├── customer/
│   ├── followup/
│   ├── review/
│   └── upload/
├── features/
│   ├── dashboard/
│   ├── customers/
│   ├── followup-tasks/
│   ├── data-ingestion/
│   ├── review-center/
│   └── settings/
├── lib/
│   ├── api/
│   ├── mock/
│   ├── date.ts
│   ├── clipboard.ts
│   └── format.ts
├── styles/
│   ├── tokens.css
│   ├── base.css
│   ├── layout.css
│   ├── components.css
│   ├── states.css
│   └── themes.css
├── types/
│   ├── api.ts
│   ├── customer.ts
│   ├── agent.ts
│   ├── followup.ts
│   └── review.ts
└── main.tsx
```

## 4. 样式架构

### 4.1 硬性规则

必须遵守：

```text
不在页面文件写内联样式
不在页面文件硬编码颜色
不在页面文件硬编码字号
不在页面文件硬编码间距
不让每个页面自己定义 Badge 或按钮样式
```

所有样式来自：

```text
styles/tokens.css
styles/base.css
styles/layout.css
styles/components.css
styles/states.css
styles/themes.css
```

### 4.2 样式文件职责

| 文件 | 职责 |
| --- | --- |
| `tokens.css` | 颜色、字号、间距、圆角、阴影、z-index 等变量 |
| `base.css` | body、字体、链接、表单基础样式 |
| `layout.css` | AppShell、SideNav、TopBar、页面容器、网格 |
| `components.css` | 通用组件基础样式，如按钮、卡片、表格、Badge |
| `states.css` | loading、empty、error、success、AI analyzing 等状态样式 |
| `themes.css` | 后续主题扩展，MVP 可只提供默认主题 |

### 4.3 组件样式方式

MVP 推荐：

```text
基础样式走全局语义 class
复杂局部组件可用 CSS Modules
所有变量必须来自 tokens.css
```

示例原则：

```tsx
<Button variant="primary">复制话术</Button>
<IntentBadge level="high" />
<AgentRiskFlag level="warning" />
```

禁止：

```tsx
<button style={{ background: '#1677ff' }}>
<div className="text-red-500 rounded-[18px]">
```

## 5. 路由设计

```text
/login
/app/dashboard
/app/customers
/app/customers/:customerId
/app/followup-tasks
/app/data-ingestion
/app/review-center
/app/settings
```

`/app` 使用统一 `AppShell`，登录页不使用后台侧边栏。

## 6. 组件分层

### 6.1 基础 UI 组件

```text
Button
IconButton
Input
Select
Tabs
Table
Card
Drawer
Dialog
Toast
Tooltip
EmptyState
LoadingState
ErrorState
```

基础组件不包含业务含义。

### 6.2 业务组件

```text
StageBadge
IntentBadge
RiskBadge
TaskStatusBadge
MetricCard
CustomerTable
CustomerProfileSummary
ConversationTimeline
ScriptTaskCard
ScriptRecommendationPanel
AgentInsight
AgentReasoning
AgentRiskFlag
AgentRunStatus
ReviewInsightCard
UploadConversationCard
WeComIntegrationCard
```

业务组件只能接收语义化 props。例如 `intent="high"`，不能接收任意颜色。

### 6.3 页面组件

页面只负责：

```text
组合业务组件
读取数据
处理页面级事件
控制筛选、分页、弹窗、抽屉状态
```

页面不负责：

```text
定义视觉样式
拼接复杂业务规则
散落 AI 状态文案
直接访问后端 SDK 或第三方 SDK
```

## 7. 类型设计

### 7.1 核心类型

```ts
export type IntentLevel = "high" | "medium" | "low" | "risk";

export type CustomerStage =
  | "new_lead"
  | "opened"
  | "needs_discovery"
  | "product_interested"
  | "price_objection"
  | "high_intent"
  | "closing"
  | "won"
  | "silent"
  | "churn_risk";

export interface AgentRecommendation {
  customerStage: CustomerStage;
  intentLevel: IntentLevel;
  mainConcerns: string[];
  recommendedAction: string;
  script: string;
  reasoning: string;
  riskFlags: string[];
  nextFollowupTime?: string;
}
```

### 7.2 类型原则

- 阶段、意向、任务状态必须用枚举或联合类型。
- 后端返回结构进入页面前必须转成前端领域类型。
- AI 输出必须有结构化类型，不能在页面里直接解析任意文本。

## 8. API 和 Mock 策略

### 8.1 Mock 先行

前端第一阶段用 Mock 数据跑通流程：

```text
客户列表
客户详情
聊天记录
Agent 推荐
待确认任务
复盘指标
上传解析结果
```

Mock 数据不应散落在页面里，应集中在：

```text
src/lib/mock/
```

### 8.2 API Contract 先行

对接真实后端前必须先定义：

```text
请求参数
响应结构
错误结构
分页结构
状态码
异常情况
```

Contract 建议放在：

```text
docs/architecture/api-contracts.md
src/types/api.ts
```

### 8.3 API Client

所有请求通过统一 API Client：

```text
src/lib/api/client.ts
src/lib/api/customers.ts
src/lib/api/followupTasks.ts
src/lib/api/uploads.ts
src/lib/api/review.ts
```

页面不能直接 `fetch`。

### 8.4 接口对接硬规则

后续对接企业微信或后端真实接口时：

```text
先有本地验证结果
先有 mock 或 contract
先跑通 API smoke test
再替换页面数据源
```

不能盲写功能后靠一点点改 bug 对齐接口。

## 9. 状态管理

MVP 不引入复杂全局状态库。建议：

```text
页面内局部状态：useState/useReducer
跨页面用户信息：轻量 Context
服务端数据：先封装 hooks，后续可接 TanStack Query
```

可预留：

```text
src/features/customers/useCustomers.ts
src/features/followup-tasks/useFollowupTasks.ts
src/features/review-center/useReviewReport.ts
```

如果后续数据缓存、重试、失效刷新变复杂，再引入 TanStack Query。

## 10. 错误和状态处理

统一状态：

```text
loading
empty
error
success
analyzing
uploading
parsing
permission_denied
sync_failed
```

每个错误必须给用户下一步动作：

```text
上传失败 -> 查看支持格式 / 重新上传
解析失败 -> 检查时间、发送人、消息内容
同步失败 -> 检查企业微信配置
权限不足 -> 联系主管或管理员
复制失败 -> 手动选择文本复制
```

## 11. 测试策略

### 11.1 MVP 必须覆盖

```text
Button / Badge 等基础组件
ScriptTaskCard 操作状态
复制话术工具函数
客户阶段和意向映射
上传解析结果展示
空状态和错误状态
```

### 11.2 测试工具

```text
Vitest
React Testing Library
Playwright 后置用于主流程 E2E
```

### 11.3 必跑命令

前端代码提交前应运行：

```bash
npm run lint
npm run typecheck
npm run test
npm run build
```

如果某个命令暂未建立，必须在交付说明里说清楚。

## 12. 实现顺序

推荐实现顺序：

```text
1. 项目脚手架：Vite + React + TypeScript
2. 样式系统：tokens/base/layout/components/states/themes
3. AppShell：TopBar + SideNav + 页面容器
4. 基础 UI 组件：Button、Card、Badge、Table、EmptyState、Toast
5. 业务类型和 Mock 数据
6. 业务组件：StageBadge、IntentBadge、AgentInsight、ScriptTaskCard
7. 数据接入页
8. 客户详情页
9. 工作台
10. 客户列表
11. 待确认话术
12. 复盘中心
13. 登录页和设置页
14. 测试补齐和构建验证
```

原因：MVP 演示主链路最依赖数据接入和客户详情，先实现这部分可以最快验证价值。

## 13. 风险和取舍

### 13.1 不立即使用大型 UI 库

收益：视觉系统更可控，不容易被组件库风格绑架。  
代价：基础组件需要自己实现一部分。  
缓解：MVP 只做必要组件，不追求完整组件库。

### 13.2 Mock 先行

收益：前端可以先跑通主流程，不被后端阻塞。  
代价：后续需要严格对齐真实 API。  
缓解：提前定义 API Contract，统一 API Client。

### 13.3 不立即引入复杂状态库

收益：降低早期复杂度。  
代价：后续服务端状态复杂时可能要迁移。  
缓解：通过 hooks 封装数据访问，后续可替换为 TanStack Query。

## 14. 阶段说明要求

进入前端实现后，每完成一个关键阶段都要说明：

```text
这一步做了什么
用了哪些技术点
为什么这么选
难点是什么
风险在哪里
跑了哪些验证命令
下一步是什么
```

这是项目协作规则，不是可选项。