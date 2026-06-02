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

export interface PageDto<T> {
  items: T[];
  page: number;
  page_size: number;
  total: number;
}

export interface ParsedCustomerDto {
  name: string;
  owner_name: string;
}

export interface UploadConversationResultDto {
  upload_id: string;
  status: "uploaded" | "parsed" | "needs_confirmation" | "confirmed" | "failed";
  parsed_customer: ParsedCustomerDto;
  message_count: number;
  warnings: string[];
  next_action: string;
}

export interface ConfirmUploadResultDto {
  customer_id: string;
  conversation_id: string;
  agent_run_id: string;
  followup_task_id: string;
  status: "confirmed";
}

export interface TaskCopyResultDto {
  task_id: string;
  status: FollowupTaskStatus;
  copied_at: string;
}

export interface TaskStatusResultDto {
  task_id: string;
  status: FollowupTaskStatus;
}

export interface MarkWrongResultDto {
  task_id: string;
  status: FollowupTaskStatus;
  feedback_id: string;
}

export interface RegenerateTaskResultDto {
  task_id: string;
  agent_run_id: string;
  recommendation: AgentRecommendationDto;
}

export interface ConversationMessageDto {
  id: string;
  sender_type: "customer" | "sales";
  sender_name: string;
  content: string;
  sent_at: string;
}

export interface CustomerDetailDto {
  customer: CustomerDto;
  latest_recommendation: AgentRecommendationDto;
  profile_evidence: string[];
  recent_tasks: FollowupTaskDto[];
  recent_agent_runs: Array<Record<string, unknown>>;
}

export type MemoryFactStatusDto = "active" | "superseded" | "rejected";

export interface LongTermMemoryFactDto {
  id: string;
  customer_id: string;
  category: string;
  key: string;
  value: string;
  confidence: number;
  source_type: string;
  source_id: string;
  status: MemoryFactStatusDto;
  created_at: string;
  updated_at: string;
}

export interface LongTermMemoryDto {
  customer: CustomerDto;
  facts: LongTermMemoryFactDto[];
  prompt_context: string;
  built_at: string;
}

export interface MemoryFactStatusResultDto {
  fact_id: string;
  status: MemoryFactStatusDto;
}

export interface MemoryFactCorrectionResultDto {
  old_fact_id: string;
  old_status: MemoryFactStatusDto;
  new_fact: LongTermMemoryFactDto;
}

export interface ReviewInsightDto {
  title: string;
  evidence: string;
  suggestion: string;
}

export interface ReviewSummaryDto {
  metrics: MetricDto[];
  stage_distribution: Array<{ stage: string; count: number }>;
  opportunity_customers: CustomerDto[];
  risk_customers: CustomerDto[];
  insights: ReviewInsightDto[];
  sample_warning: string | null;
}

export interface ReportCustomerItemDto {
  customer_id: string;
  customer_name: string;
  stage: string;
  intent: string;
  recommended_action: string;
  script: string;
  reasoning: string;
  evidence: string[];
}

export interface ReportSectionDto {
  title: string;
  summary: string;
  items: ReportCustomerItemDto[];
  evidence: string[];
}

export interface ReportActionItemDto {
  customer_id: string;
  customer_name: string;
  priority: string;
  action: string;
  due_hint: string;
}

export interface ReportDto {
  id: string;
  type: "customer_intent";
  title: string;
  range_label: string;
  summary: string;
  owner_id: string;
  owner_role: string;
  metrics: MetricDto[];
  sections: ReportSectionDto[];
  action_items: ReportActionItemDto[];
  markdown: string;
  generated_at: string;
}

export interface ReportSummaryDto {
  id: string;
  type: "customer_intent";
  title: string;
  range_label: string;
  summary: string;
  owner_id: string;
  owner_role: string;
  metric_count: number;
  section_count: number;
  action_item_count: number;
  generated_at: string;
}
