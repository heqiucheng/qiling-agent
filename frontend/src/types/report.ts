import type { ReviewMetric } from "./review";

export interface ReportCustomerItem {
  customerId: string;
  customerName: string;
  stage: string;
  intent: string;
  recommendedAction: string;
  script: string;
  reasoning: string;
  evidence: string[];
}

export interface ReportSection {
  title: string;
  summary: string;
  items: ReportCustomerItem[];
  evidence: string[];
}

export interface ReportActionItem {
  customerId: string;
  customerName: string;
  priority: string;
  action: string;
  dueHint: string;
}

export interface Report {
  id: string;
  type: "customer_intent";
  title: string;
  rangeLabel: string;
  summary: string;
  ownerId: string;
  ownerRole: string;
  metrics: ReviewMetric[];
  sections: ReportSection[];
  actionItems: ReportActionItem[];
  markdown: string;
  generatedAt: string;
}

export interface ReportSummary {
  id: string;
  type: "customer_intent";
  title: string;
  rangeLabel: string;
  summary: string;
  ownerId: string;
  ownerRole: string;
  metricCount: number;
  sectionCount: number;
  actionItemCount: number;
  generatedAt: string;
}
