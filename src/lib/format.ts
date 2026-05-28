import type { CustomerStage, IntentLevel } from "../types/customer";

export const intentLabels: Record<IntentLevel, string> = {
  high: "高意向",
  medium: "中意向",
  low: "低意向",
  risk: "风险"
};

export const stageLabels: Record<CustomerStage, string> = {
  new_lead: "新线索",
  opened: "已破冰",
  needs_discovery: "需求了解",
  product_interested: "产品了解",
  price_objection: "价格异议",
  high_intent: "强意向",
  closing: "待促单",
  won: "已成交",
  silent: "沉默",
  churn_risk: "流失风险"
};