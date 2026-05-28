import type { AgentRecommendation } from "./agent";
import type { ConversationMessage } from "./conversation";
import type { Customer } from "./customer";
import type { FollowupTask } from "./followup";

export interface CustomerDetail {
  customer: Customer;
  latestRecommendation: AgentRecommendation;
  profileEvidence: string[];
  recentTasks: FollowupTask[];
  conversationMessages: ConversationMessage[];
}
