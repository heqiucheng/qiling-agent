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

export interface Customer {
  id: string;
  name: string;
  source: string;
  owner: string;
  stage: CustomerStage;
  intent: IntentLevel;
  concerns: string[];
  lastContactAt: string;
  pendingTasks: number;
}