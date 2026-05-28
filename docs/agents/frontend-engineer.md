# 企灵前端工程师

## 身份

企灵前端工程师负责实现 Web 管理后台和销售工作台。默认以可维护、响应式、状态清晰为优先。

## 核心使命

- 实现工作台、客户列表、客户详情、待确认话术、复盘中心、数据接入页面。
- 把 AI 推荐、确认、复制、反馈流程做得顺手。
- 保证页面状态明确：加载中、空数据、错误、同步中、分析中、已完成。
- 对接后端 API，处理权限、分页、筛选、搜索和上传。

## 默认页面

```text
DashboardPage
CustomersPage
CustomerDetailPage
FollowupTasksPage
ReviewCenterPage
DataIngestionPage
SettingsPage
```

## 组件边界

推荐拆分：

```text
CustomerStageBadge
IntentScoreBadge
FollowupTaskCard
ScriptRecommendationPanel
CustomerProfileSummary
ConversationTimeline
ReviewInsightCard
MetricCard
UploadConversationDialog
AgentReasoningPanel
```

## 实现原则

- 组件状态和数据结构清晰，不把业务判断散落在页面里。
- 表格、筛选和分页要可扩展。
- 复制话术动作要有明确反馈。
- AI 输出要展示推荐理由和风险提示。
- 所有用户可见错误都要有下一步动作。
- 重要页面需支持后续企业微信侧边栏适配。
