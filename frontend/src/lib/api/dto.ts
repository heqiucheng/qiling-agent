import type { CustomerStage, IntentLevel } from "../../types/customer";
import type { FollowupTaskStatus } from "../../types/followup";

export interface ApiMetaDto {
  request_id: string;
  timestamp: string;
}

export interface ApiErrorDto {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ApiResponseDto<T> {
  data: T | null;
  error: ApiErrorDto | null;
  meta: ApiMetaDto;
}

export interface OwnerDto {
  id: string;
  name: string;
}

export interface CustomerDto {
  id: string;
  name: string;
  source: string;
  owner: OwnerDto;
  stage: CustomerStage;
  intent: IntentLevel;
  concerns: string[];
  tags: string[];
  profile_summary: string;
  last_contact_at: string;
  pending_tasks: number;
  risk_flags: string[];
}

export interface AgentRecommendationDto {
  customer_stage: CustomerStage;
  intent_level: IntentLevel;
  main_concerns: string[];
  recommended_action: string;
  script: string;
  reasoning: string;
  risk_flags: string[];
  next_followup_time?: string;
}

export interface FollowupTaskDto {
  id: string;
  customer: CustomerDto;
  type: string;
  status: FollowupTaskStatus;
  generated_at: string;
  recommendation: AgentRecommendationDto;
  feedback: { reason: string } | null;
}

export interface MetricDto {
  key: string;
  label: string;
  value: string | number;
  hint: string;
}

export interface DashboardSummaryDto {
  metrics: MetricDto[];
  priority_tasks: FollowupTaskDto[];
  high_intent_customers: CustomerDto[];
  silent_customers: CustomerDto[];
  risk_customers: CustomerDto[];
  daily_review: {
    summary: string;
    suggestions: string[];
  };
}
