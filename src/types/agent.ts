import type { CustomerStage, IntentLevel } from "./customer";

export interface AgentRecommendation {
  customerStage: CustomerStage;
  intentLevel: IntentLevel;
  mainConcerns: string[];
  recommendedAction: string;
  script: string;
  reasoning: string;
  riskFlags: string[];
  nextFollowupTime?: string;
}