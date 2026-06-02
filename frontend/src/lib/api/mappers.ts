import type { AgentRecommendation } from "../../types/agent";
import type { PageResult } from "../../types/api";
import type { Customer } from "../../types/customer";
import type { DashboardSummary } from "../../types/dashboard";
import type { FollowupTask } from "../../types/followup";
import type { ReviewMetric } from "../../types/review";
import type { Report, ReportExportTask, ReportSummary } from "../../types/report";
import type { ConfirmUploadResult, UploadConversationResult } from "../../types/upload";
import type { ConversationMessage } from "../../types/conversation";
import type { CustomerDetail } from "../../types/customerDetail";
import type { LongTermMemory, LongTermMemoryFact, MemoryFactCorrectionResult, MemoryFactStatusResult } from "../../types/memory";
import type { ReviewInsight, ReviewSummary } from "../../types/review";
import type { AgentRecommendationDto, ConversationMessageDto, CustomerDetailDto, CustomerDto, DashboardSummaryDto, FollowupTaskDto, LongTermMemoryDto, LongTermMemoryFactDto, MemoryFactCorrectionResultDto, MemoryFactStatusResultDto, MetricDto, PageDto, ReportDto, ReportExportTaskDto, ReportSummaryDto, ReviewInsightDto, ReviewSummaryDto } from "./dto";
import type { ConfirmUploadResultDto, UploadConversationResultDto } from "./dto";

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

export function mapUploadConversationResult(dto: UploadConversationResultDto): UploadConversationResult {
  return {
    uploadId: dto.upload_id,
    status: dto.status,
    parsedCustomer: {
      name: dto.parsed_customer.name,
      ownerName: dto.parsed_customer.owner_name
    },
    messageCount: dto.message_count,
    warnings: dto.warnings,
    nextAction: dto.next_action
  };
}

export function mapConfirmUploadResult(dto: ConfirmUploadResultDto): ConfirmUploadResult {
  return {
    customerId: dto.customer_id,
    conversationId: dto.conversation_id,
    agentRunId: dto.agent_run_id,
    followupTaskId: dto.followup_task_id,
    status: dto.status
  };
}

export function mapConversationMessage(dto: ConversationMessageDto): ConversationMessage {
  return {
    id: dto.id,
    senderType: dto.sender_type,
    senderName: dto.sender_name,
    content: dto.content,
    sentAt: dto.sent_at
  };
}

export function mapCustomerDetail(dto: CustomerDetailDto, messages: ConversationMessage[]): CustomerDetail {
  return {
    customer: mapCustomer(dto.customer),
    latestRecommendation: mapRecommendation(dto.latest_recommendation),
    profileEvidence: dto.profile_evidence,
    recentTasks: dto.recent_tasks.map(mapFollowupTask),
    conversationMessages: messages
  };
}

export function mapLongTermMemoryFact(dto: LongTermMemoryFactDto): LongTermMemoryFact {
  return {
    id: dto.id,
    customerId: dto.customer_id,
    category: dto.category,
    key: dto.key,
    value: dto.value,
    confidence: dto.confidence,
    sourceType: dto.source_type,
    sourceId: dto.source_id,
    status: dto.status,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at
  };
}

export function mapLongTermMemory(dto: LongTermMemoryDto): LongTermMemory {
  return {
    customer: mapCustomer(dto.customer),
    facts: dto.facts.map(mapLongTermMemoryFact),
    promptContext: dto.prompt_context,
    builtAt: dto.built_at
  };
}

export function mapMemoryFactStatusResult(dto: MemoryFactStatusResultDto): MemoryFactStatusResult {
  return {
    factId: dto.fact_id,
    status: dto.status
  };
}

export function mapMemoryFactCorrectionResult(dto: MemoryFactCorrectionResultDto): MemoryFactCorrectionResult {
  return {
    oldFactId: dto.old_fact_id,
    oldStatus: dto.old_status,
    newFact: mapLongTermMemoryFact(dto.new_fact)
  };
}

export function mapReviewInsight(dto: ReviewInsightDto): ReviewInsight {
  return {
    title: dto.title,
    evidence: dto.evidence,
    suggestion: dto.suggestion
  };
}

export function mapReviewSummary(dto: ReviewSummaryDto): ReviewSummary {
  return {
    metrics: dto.metrics.map(mapMetric),
    insights: dto.insights.map(mapReviewInsight),
    sampleWarning: dto.sample_warning ?? undefined
  };
}

export function mapReport(dto: ReportDto): Report {
  return {
    id: dto.id,
    type: dto.type,
    title: dto.title,
    rangeLabel: dto.range_label,
    summary: dto.summary,
    ownerId: dto.owner_id,
    ownerRole: dto.owner_role,
    metrics: dto.metrics.map(mapMetric),
    sections: dto.sections.map((section) => ({
      title: section.title,
      summary: section.summary,
      evidence: section.evidence,
      items: section.items.map((item) => ({
        customerId: item.customer_id,
        customerName: item.customer_name,
        stage: item.stage,
        intent: item.intent,
        recommendedAction: item.recommended_action,
        script: item.script,
        reasoning: item.reasoning,
        evidence: item.evidence
      }))
    })),
    actionItems: dto.action_items.map((item) => ({
      customerId: item.customer_id,
      customerName: item.customer_name,
      priority: item.priority,
      action: item.action,
      dueHint: item.due_hint
    })),
    markdown: dto.markdown,
    generatedAt: dto.generated_at
  };
}

export function mapReportSummary(dto: ReportSummaryDto): ReportSummary {
  return {
    id: dto.id,
    type: dto.type,
    title: dto.title,
    rangeLabel: dto.range_label,
    summary: dto.summary,
    ownerId: dto.owner_id,
    ownerRole: dto.owner_role,
    metricCount: dto.metric_count,
    sectionCount: dto.section_count,
    actionItemCount: dto.action_item_count,
    generatedAt: dto.generated_at
  };
}

export function mapReportExportTask(dto: ReportExportTaskDto): ReportExportTask {
  return {
    id: dto.id,
    reportId: dto.report_id,
    format: dto.format,
    status: dto.status,
    ownerId: dto.owner_id,
    ownerRole: dto.owner_role,
    filename: dto.filename,
    contentType: dto.content_type,
    sizeBytes: dto.size_bytes,
    error: dto.error,
    createdAt: dto.created_at,
    completedAt: dto.completed_at
  };
}
