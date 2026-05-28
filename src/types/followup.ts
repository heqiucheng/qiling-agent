import type { AgentRecommendation } from "./agent";
import type { Customer } from "./customer";

export type FollowupTaskStatus = "pending" | "copied" | "skipped" | "marked_wrong";

export interface FollowupTask {
  id: string;
  customer: Customer;
  type: string;
  status: FollowupTaskStatus;
  generatedAt: string;
  recommendation: AgentRecommendation;
}