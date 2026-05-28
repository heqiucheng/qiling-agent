import type { AgentRecommendation } from "../../types/agent";
import type { PageResult } from "../../types/api";
import type { Customer } from "../../types/customer";
import type { DashboardSummary } from "../../types/dashboard";
import type { FollowupTask } from "../../types/followup";
import type { ReviewMetric } from "../../types/review";
import type { AgentRecommendationDto, CustomerDto, DashboardSummaryDto, FollowupTaskDto, MetricDto, PageDto } from "./dto";

export function mapCustomer(dto: CustomerDto): Customer {
  return {
    id: dto.id,
    name: dto.name,
    source: dto.source,
    owner: dto.owner.name,
    stage: dto.stage,
    intent: dto.intent,
    concerns: dto.concerns,
    lastContactAt: dto.last_contact_at,
    pendingTasks: dto.pending_tasks
  };
}

export function mapRecommendation(dto: AgentRecommendationDto): AgentRecommendation {
  return {
    customerStage: dto.customer_stage,
    intentLevel: dto.intent_level,
    mainConcerns: dto.main_concerns,
    recommendedAction: dto.recommended_action,
    script: dto.script,
    reasoning: dto.reasoning,
    riskFlags: dto.risk_flags,
    nextFollowupTime: dto.next_followup_time
  };
}

export function mapFollowupTask(dto: FollowupTaskDto): FollowupTask {
  return {
    id: dto.id,
    customer: mapCustomer(dto.customer),
    type: dto.type,
    status: dto.status,
    generatedAt: dto.generated_at,
    recommendation: mapRecommendation(dto.recommendation)
  };
}

export function mapMetric(dto: MetricDto): ReviewMetric {
  return {
    key: dto.key,
    label: dto.label,
    value: String(dto.value),
    hint: dto.hint
  };
}

export function mapDashboardSummary(dto: DashboardSummaryDto): DashboardSummary {
  return {
    metrics: dto.metrics.map(mapMetric),
    priorityTasks: dto.priority_tasks.map(mapFollowupTask),
    highIntentCustomers: dto.high_intent_customers.map(mapCustomer),
    silentCustomers: dto.silent_customers.map(mapCustomer),
    riskCustomers: dto.risk_customers.map(mapCustomer),
    dailyReview: {
      summary: dto.daily_review.summary,
      suggestions: dto.daily_review.suggestions
    }
  };
}

export function mapPage<TDto, TItem>(dto: PageDto<TDto>, mapItem: (item: TDto) => TItem): PageResult<TItem> {
  return {
    items: dto.items.map(mapItem),
    page: dto.page,
    pageSize: dto.page_size,
    total: dto.total
  };
}
