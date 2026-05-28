import type { FollowupTask } from "../../types/followup";
import type { ReviewInsight, ReviewMetric } from "../../types/review";

const customer = {
  id: "cus_001",
  name: "张女士",
  source: "企业微信",
  owner: "李销售",
  stage: "price_objection" as const,
  intent: "high" as const,
  concerns: ["价格", "售后"],
  lastContactAt: "2 小时前",
  pendingTasks: 1
};

export const mockMetrics: ReviewMetric[] = [
  { label: "待确认话术", value: "12", hint: "较昨日 +4" },
  { label: "高意向客户", value: "8", hint: "建议优先跟进" },
  { label: "沉默客户", value: "17", hint: "5 个超过 3 天" },
  { label: "流失风险", value: "3", hint: "需要主管介入" }
];

export const mockTasks: FollowupTask[] = [
  {
    id: "task_001",
    customer,
    type: "价格异议",
    status: "pending",
    generatedAt: "今天 10:20",
    recommendation: {
      customerStage: "price_objection",
      intentLevel: "high",
      mainConcerns: ["价格", "售后"],
      recommendedAction: "解释方案价值并引导预约",
      script: "您好，刚才您提到比较关注价格。我这边先按您的需求整理一个更适合的方案，里面会把费用、服务和后续保障说明清楚，您看我下午发您确认可以吗？",
      reasoning: "客户连续两次询问价格和售后，说明有购买兴趣但存在决策顾虑。",
      riskFlags: ["涉及价格承诺，建议人工确认后复制发送。"],
      nextFollowupTime: "今天 16:00 前"
    }
  }
];

export const mockInsights: ReviewInsight[] = [
  {
    title: "价格异议客户需要优先处理",
    evidence: "高意向客户中有 4 位停留在价格异议阶段。",
    suggestion: "优先补充案例证明和方案价值解释，不要直接降价。"
  }
];