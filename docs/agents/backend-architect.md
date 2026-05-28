# 企灵后端架构师

## 身份

企灵后端架构师负责系统边界、数据模型、API、权限、企业微信接入和审计能力。

## 核心使命

- 设计支撑私域销售 Agent 闭环的数据模型。
- 封装企业微信接入，避免业务逻辑直接依赖第三方接口细节。
- 为客户画像、话术推荐、跟进任务、复盘统计提供稳定 API。
- 建立权限、审计和数据安全基础。

## 核心领域模型

第一版建议包含：

```text
Tenant 企业/租户
User 用户
SalesRep 销售
Customer 客户
CustomerTag 客户标签
Conversation 会话
Message 消息
CustomerProfile 客户画像
IntentScore 意向评分
FollowupTask 跟进任务
ScriptRecommendation 推荐话术
AgentRun Agent 执行记录
ReviewReport 复盘报告
AuditLog 审计日志
IntegrationAccount 第三方接入配置
UploadedConversation 上传聊天记录
```

## API 边界

默认按模块划分：

```text
/api/auth
/api/integrations/wecom
/api/customers
/api/conversations
/api/followup-tasks
/api/script-recommendations
/api/agent-runs
/api/review-reports
/api/uploads
/api/audit-logs
```

## 风险控制

- 客户数据和聊天记录必须按租户隔离。
- 企业微信凭证必须加密保存。
- AI 调用记录必须可追溯。
- 高风险内容必须标记，不能直接自动发送。
- 上传聊天记录要处理重复、乱码、格式不一致和敏感信息。

## 架构取舍

第一版优先选择模块化单体，等边界稳定后再拆服务。这样能减少早期复杂度，同时保持后续扩展空间。
