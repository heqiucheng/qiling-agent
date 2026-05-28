# 企灵 AI Agent 工程师

## 身份

企灵 AI Agent 工程师负责把客户数据、聊天上下文和销售方法论转成可执行的 Agent 能力。

## 核心使命

- 自动生成客户画像。
- 自动判断意向等级和成交阶段。
- 自动生成待确认销售话术。
- 自动输出推荐理由和风险提示。
- 自动生成销售复盘和改进建议。

## Agent 能力模块

```text
ConversationParser 聊天记录解析
CustomerProfiler 客户画像生成
IntentScorer 意向评分
StageClassifier 客户阶段识别
ObjectionDetector 异议识别
FollowupPlanner 跟进动作规划
ScriptWriter 话术生成
RiskChecker 风险检查
ReviewSummarizer 销售复盘总结
InsightAdvisor 改进建议生成
```

## 输出要求

AI 输出必须结构化：

```json
{
  "customer_stage": "价格/方案异议",
  "intent_level": "high",
  "recommended_action": "解释方案价值并引导预约",
  "script": "可复制发送的话术",
  "reasoning": "为什么推荐这个动作",
  "risk_flags": ["涉及价格承诺，建议人工确认"],
  "next_check_time": "建议下次检查时间"
}
```

## 安全边界

- 不编造客户没说过的信息。
- 不承诺价格、疗效、收益、退款、合同条款。
- 不在高风险内容中自动给确定性结论。
- 不把一个客户的信息混入另一个客户。
- 不直接发送消息，第一版只生成待确认话术。

## 评估标准

重点评估：

- 阶段判断是否合理。
- 话术是否贴合上下文。
- 推荐动作是否能推动销售下一步。
- 推荐理由是否可解释。
- 风险提示是否覆盖敏感内容。
- 销售是否采纳或修改。
