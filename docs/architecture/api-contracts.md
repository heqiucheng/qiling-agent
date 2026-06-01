# 企灵 Agent API Contracts

**状态**：Draft for MVP implementation  
**版本**：v0.1  
**日期**：2026-05-28  
**适用范围**：MVP 上传聊天记录、客户画像、跟进任务、复盘统计主链路  
**依赖文档**：`docs/specs/mvp-prd.md`、`docs/architecture/backend-architecture.md`

## 1. 设计原则

- API 路径统一使用 `/api` 前缀。
- JSON 字段统一使用 `snake_case`，前端 API client 负责映射为 TypeScript `camelCase`。
- 所有列表接口必须分页，禁止无边界返回全量数据。
- 所有 AI 结果必须结构化返回，不能只返回一段自然语言。
- 所有销售动作必须可追溯，复制、跳过、换一种、标记不准都要记录。
- 第一版只生成待确认话术，不提供自动发送客户消息接口。
- 企业微信真实接口暂不接入本契约，只保留配置和集成预留。

## 2. 统一响应

成功响应：

```json
{
  "data": {},
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

错误响应：

```json
{
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数不正确",
    "details": {
      "field": "content"
    }
  },
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

分页响应：

```json
{
  "data": {
    "items": [],
    "page": 1,
    "page_size": 20,
    "total": 0
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

## 3. 枚举

客户阶段：

```text
new_lead
opened
needs_discovery
product_interested
price_objection
high_intent
closing
won
silent
churn_risk
```

意向等级：

```text
high
medium
low
risk
```

跟进任务状态：

```text
pending
copied
skipped
marked_wrong
```

上传状态：

```text
uploaded
parsed
needs_confirmation
confirmed
failed
```

Agent 执行状态：

```text
queued
running
succeeded
failed
needs_review
```

## 4. 核心对象

### 4.1 Customer

```json
{
  "id": "cus_001",
  "name": "王女士",
  "source": "企业微信",
  "owner": {
    "id": "usr_001",
    "name": "销售A"
  },
  "stage": "price_objection",
  "intent": "high",
  "concerns": ["价格", "售后"],
  "tags": ["价格敏感", "需要案例"],
  "profile_summary": "关注价格和售后保障，近期购买意向较强。",
  "last_contact_at": "2026-05-28T09:30:00Z",
  "pending_tasks": 1,
  "risk_flags": ["涉及价格承诺，需人工确认"]
}
```

### 4.2 ConversationMessage

```json
{
  "id": "msg_001",
  "sender_type": "customer",
  "sender_name": "王女士",
  "content": "这个价格还能优惠吗？",
  "sent_at": "2026-05-28T09:20:00Z"
}
```

### 4.3 AgentRecommendation

```json
{
  "customer_stage": "price_objection",
  "intent_level": "high",
  "main_concerns": ["价格", "效果", "售后"],
  "recommended_action": "解释方案价值并引导预约",
  "script": "您好，刚才您提到比较关注价格，我这边帮您整理了一下更适合您的方案。",
  "reasoning": "客户连续询问价格和售后，说明有购买兴趣但存在决策顾虑。",
  "risk_flags": ["涉及价格承诺，建议人工确认"],
  "next_followup_time": "2026-05-28T16:00:00Z"
}
```

### 4.4 FollowupTask

```json
{
  "id": "task_001",
  "customer": {},
  "type": "price_objection",
  "status": "pending",
  "generated_at": "2026-05-28T10:00:00Z",
  "recommendation": {},
  "feedback": null
}
```

## 5. Health

### GET `/api/health`

用途：服务健康检查。

响应：

```json
{
  "data": {
    "status": "ok",
    "service": "qiling-agent-backend",
    "version": "0.1.0",
    "env": "development"
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

## Audit Events

### GET `/api/audit-events`

Purpose: read the structured business event trail for follow-up review, debugging, future memory generation, and Agent explainability.

Query:

```text
action
actor_id
entity_type
entity_id
page
page_size
```

Sales users only read their own events. Managers can read all events or filter by `actor_id`.

Item shape:

```json
{
  "id": "audit_123",
  "action": "followup_task.copied",
  "actor": {
    "user_id": "usr_001",
    "role": "sales"
  },
  "request_id": "req_123",
  "entity_type": "followup_task",
  "entity_id": "task_001",
  "related_type": "agent_run",
  "related_id": "run_001",
  "metadata": {
    "has_script": true
  },
  "created_at": "2026-06-01T10:00:00Z"
}
```

## 6. 上传和解析聊天记录

### POST `/api/uploads/conversations`

用途：上传或粘贴聊天记录，创建解析任务。MVP 优先支持 `text` 和 `.txt/.csv` 文件。

请求：

```json
{
  "source_type": "pasted_text",
  "content": "王女士 09:20 这个价格还能优惠吗？",
  "file_name": null,
  "owner_id": "usr_001"
}
```

字段规则：

- `source_type`：`pasted_text`、`txt_file`、`csv_file`。
- `content`：粘贴文本或已读取的文件文本内容。
- `owner_id`：MVP 可由本地 mock 登录态提供。

响应：

```json
{
  "data": {
    "upload_id": "upl_001",
    "status": "needs_confirmation",
    "parsed_customer": {
      "name": "王女士",
      "owner_name": "销售A"
    },
    "message_count": 12,
    "warnings": ["有 1 条消息缺少明确时间，已按相邻消息估算"],
    "next_action": "confirm_parsed_result"
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

失败场景：

- `EMPTY_CONTENT`：聊天记录为空。
- `UNSUPPORTED_UPLOAD_TYPE`：暂不支持该格式。
- `PARSE_FAILED`：无法识别客户或消息结构。

### GET `/api/uploads/{upload_id}`

用途：查看上传解析结果。

响应：

```json
{
  "data": {
    "id": "upl_001",
    "status": "needs_confirmation",
    "source_type": "pasted_text",
    "parsed_customer": {
      "name": "王女士",
      "owner_name": "销售A"
    },
    "messages": [],
    "warnings": [],
    "created_at": "2026-05-28T10:00:00Z"
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

### POST `/api/uploads/{upload_id}/confirm`

用途：销售确认或修正解析结果，触发客户画像、意向判断和话术生成。

请求：

```json
{
  "customer_name": "王女士",
  "owner_id": "usr_001",
  "message_fixes": [
    {
      "message_id": "msg_001",
      "sender_type": "customer"
    }
  ]
}
```

响应：

```json
{
  "data": {
    "customer_id": "cus_001",
    "conversation_id": "conv_001",
    "agent_run_id": "run_001",
    "followup_task_id": "task_001",
    "status": "confirmed"
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

## 7. 客户

### GET `/api/customers`

用途：客户列表，支持搜索、阶段、意向和风险筛选。

查询参数：

```text
page=1
page_size=20
keyword=王
stage=price_objection
intent=high
risk=1
owner_id=usr_001
```

响应：

```json
{
  "data": {
    "items": [],
    "page": 1,
    "page_size": 20,
    "total": 0
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

### GET `/api/customers/{customer_id}`

用途：客户详情。

响应：

```json
{
  "data": {
    "customer": {},
    "latest_recommendation": {},
    "profile_evidence": [
      "客户连续询问价格和售后",
      "最近一次沟通发生在 2 小时内"
    ],
    "recent_tasks": [],
    "recent_agent_runs": []
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

### GET `/api/customers/{customer_id}/conversations`

用途：客户聊天记录时间线。

查询参数：

```text
page=1
page_size=50
```

响应：

```json
{
  "data": {
    "items": [],
    "page": 1,
    "page_size": 50,
    "total": 0
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

## 8. 跟进任务

### GET `/api/followup-tasks`

用途：工作台和待确认任务页读取任务。

查询参数：

```text
page=1
page_size=20
status=pending
intent=high
owner_id=usr_001
```

响应：

```json
{
  "data": {
    "items": [],
    "page": 1,
    "page_size": 20,
    "total": 0
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

### POST `/api/followup-tasks/{task_id}/copy`

用途：记录销售已复制话术。注意：该接口不发送客户消息。

请求：

```json
{
  "copied_script": "您好，刚才您提到比较关注价格...",
  "client_copied_at": "2026-05-28T10:05:00Z"
}
```

响应：

```json
{
  "data": {
    "task_id": "task_001",
    "status": "copied",
    "copied_at": "2026-05-28T10:05:00Z"
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

### POST `/api/followup-tasks/{task_id}/skip`

用途：跳过任务并记录原因。

请求：

```json
{
  "reason": "客户刚刚已经回复，暂不需要跟进"
}
```

响应：

```json
{
  "data": {
    "task_id": "task_001",
    "status": "skipped"
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

### POST `/api/followup-tasks/{task_id}/mark-wrong`

用途：标记 Agent 判断不准，沉淀反馈用于复盘和后续优化。

请求：

```json
{
  "reason": "客户不是价格异议，是在问付款方式",
  "wrong_fields": ["customer_stage", "recommended_action"]
}
```

响应：

```json
{
  "data": {
    "task_id": "task_001",
    "status": "marked_wrong",
    "feedback_id": "fb_001"
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

### POST `/api/followup-tasks/{task_id}/regenerate`

用途：基于原上下文换一种话术，不丢失原任务链路。

请求：

```json
{
  "instruction": "语气更自然一点，不要太像营销话术"
}
```

响应：

```json
{
  "data": {
    "task_id": "task_001",
    "agent_run_id": "run_002",
    "recommendation": {}
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

## 9. 工作台

### GET `/api/dashboard/summary`

用途：销售打开系统后第一屏看到今天该做什么。

响应：

```json
{
  "data": {
    "metrics": [
      {
        "key": "pending_tasks",
        "label": "待确认话术",
        "value": 8,
        "hint": "优先处理高意向客户"
      }
    ],
    "priority_tasks": [],
    "high_intent_customers": [],
    "silent_customers": [],
    "risk_customers": [],
    "daily_review": {
      "summary": "今天优先跟进 3 个价格异议客户。",
      "suggestions": ["先处理 2 小时内有互动的高意向客户"]
    }
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

## 10. 复盘中心

### GET `/api/review-reports/summary`

用途：主管查看客户机会、风险和销售改进建议。

查询参数：

```text
range=today
owner_id=usr_001
```

响应：

```json
{
  "data": {
    "metrics": [
      {
        "key": "script_copy_rate",
        "label": "话术复制率",
        "value": "36%",
        "hint": "较昨日提升 5%"
      }
    ],
    "stage_distribution": [
      {
        "stage": "price_objection",
        "count": 12
      }
    ],
    "opportunity_customers": [],
    "risk_customers": [],
    "insights": [
      {
        "title": "价格异议集中",
        "evidence": "12 个客户停留在价格/方案异议阶段。",
        "suggestion": "补充案例证明和售后保障话术。"
      }
    ],
    "sample_warning": null
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

数据不足时：

```json
{
  "sample_warning": "当前样本不足 10 个客户，复盘建议仅供参考。"
}
```

## 11. Agent 执行记录

### GET `/api/agent-runs/{agent_run_id}`

用途：查看 Agent 执行输入摘要、Prompt 版本、结构化输出和失败原因。

响应：

```json
{
  "data": {
    "id": "run_001",
    "status": "succeeded",
    "task_type": "generate_followup_script",
    "model": "mock-local-v1",
    "prompt_version": "followup_v1",
    "input_summary": "12 条聊天消息，客户主要询问价格和售后。",
    "output": {},
    "validation_errors": [],
    "risk_flags": ["涉及价格承诺，建议人工确认"],
    "created_at": "2026-05-28T10:00:00Z",
    "completed_at": "2026-05-28T10:00:03Z"
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-05-28T10:00:00Z"
  }
}
```

## 12. 错误码

```text
VALIDATION_ERROR
UNAUTHORIZED
FORBIDDEN
NOT_FOUND
CONFLICT
EMPTY_CONTENT
UNSUPPORTED_UPLOAD_TYPE
PARSE_FAILED
AGENT_OUTPUT_INVALID
AGENT_RUN_FAILED
TASK_ALREADY_FINALIZED
RATE_LIMITED
INTERNAL_ERROR
```

## 13. MVP 实现顺序

```text
1. 在 Go 后端创建领域类型和 Mock store。
2. 实现 /api/dashboard/summary、/api/customers、/api/followup-tasks 的 mock-backed 读取接口。
3. 实现上传聊天记录接口和确认接口。
4. 实现复制、跳过、标记不准、换一种的任务动作接口。
5. 前端新增 API client，把现有 mock 页面逐步替换为后端接口。
6. 增加 handler/service 测试，锁定响应结构和错误码。
```

## 14. 当前阶段约束

- 当前契约不调用真实 LLM。
- 当前契约不调用真实企业微信接口。
- 当前契约不接收真实生产客户数据。
- 当前契约允许使用本地 Mock Agent 生成稳定可测的结构化结果。
- 后续接入 LLM、企业微信、数据库和向量库时，必须分别补充集成契约和本地验证记录。
## Agent Runs

### GET `/api/agent-runs/{id}`

Purpose: inspect one Agent execution trace, including model, prompt version, input summary, structured output, validation errors, and risk flags.

Response:

```json
{
  "data": {
    "id": "run_001",
    "customer_id": "cus_001",
    "task_type": "generate_followup_script",
    "status": "succeeded",
    "model": "mock-local-v1",
    "prompt_version": "followup_v1",
    "input_summary": "上传聊天记录生成客户画像和跟进话术",
    "output": {
      "customer_stage": "price_objection",
      "intent_level": "high",
      "main_concerns": ["价格", "效果"],
      "recommended_action": "解释价值并提供案例",
      "script": "您好，您刚才提到价格和效果...",
      "reasoning": "上传内容显示客户关注价格和效果...",
      "risk_flags": ["避免直接承诺优惠或效果"]
    },
    "validation_errors": [],
    "risk_flags": ["避免直接承诺优惠或效果"],
    "created_at": "2026-06-01T10:00:00Z",
    "completed_at": "2026-06-01T10:00:00Z"
  },
  "error": null,
  "meta": {
    "request_id": "req_123",
    "timestamp": "2026-06-01T10:00:00Z"
  }
}
```
