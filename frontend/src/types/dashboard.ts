import type { Customer } from "./customer";
import type { FollowupTask } from "./followup";
import type { ReviewMetric } from "./review";

export interface DashboardSummary {
  metrics: ReviewMetric[];
  priorityTasks: FollowupTask[];
  highIntentCustomers: Customer[];
  silentCustomers: Customer[];
  riskCustomers: Customer[];
  dailyReview: {
    summary: string;
    suggestions: string[];
  };
}
