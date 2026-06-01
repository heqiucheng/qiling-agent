package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/agent"
	"github.com/heqiucheng/qiling-agent/backend/internal/apperror"
	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/heqiucheng/qiling-agent/backend/internal/store"
)

type QilingService struct {
	store store.Repository
	agent agent.Runner
}

func NewQilingService(store store.Repository) *QilingService {
	return &QilingService{store: store, agent: agent.NewMockRunner()}
}

func (s *QilingService) DashboardSummary(actor domain.Actor) domain.DashboardSummary {
	customers := visibleCustomers(s.store.Customers(), actor)
	tasks := visibleTasks(s.store.FollowupTasks(), actor)

	return domain.DashboardSummary{
		Metrics: []domain.Metric{
			{Key: "pending_tasks", Label: "待确认话术", Value: len(tasks), Hint: "优先处理高意向客户"},
			{Key: "high_intent_customers", Label: "高意向客户", Value: countCustomersByIntent(customers, domain.IntentHigh), Hint: "建议今天完成跟进"},
			{Key: "risk_customers", Label: "风险客户", Value: countRiskCustomers(customers), Hint: "需要关注沉默和承诺风险"},
		},
		PriorityTasks:       tasks,
		HighIntentCustomers: filterCustomers(customers, func(c domain.Customer) bool { return c.Intent == domain.IntentHigh }),
		SilentCustomers:     filterCustomers(customers, func(c domain.Customer) bool { return c.Stage == domain.StageSilent }),
		RiskCustomers:       filterCustomers(customers, func(c domain.Customer) bool { return len(c.RiskFlags) > 0 }),
		DailyReview: domain.DailyReview{
			Summary:     "今天优先处理高意向客户和沉默客户，避免价格异议停留过久。",
			Suggestions: []string{"先跟进 2 小时内有互动的客户", "价格异议客户先给价值解释和案例，不直接承诺优惠"},
		},
	}
}

func (s *QilingService) Customers(r *http.Request, actor domain.Actor) PageResult[domain.Customer] {
	query := r.URL.Query()
	page := PageRequestFromQuery(r)
	ownerID := query.Get("owner_id")
	if actor.Role != "manager" {
		ownerID = actor.UserID
	}
	result := s.store.CustomerPage(store.CustomerFilter{
		Keyword: query.Get("keyword"),
		Stage:   query.Get("stage"),
		Intent:  query.Get("intent"),
		OwnerID: ownerID,
		Risk:    query.Get("risk"),
	}, store.PageRequest{Page: page.Page, PageSize: page.PageSize})

	return NewPageResultWithTotal(result.Items, page, result.Total)
}

func (s *QilingService) CustomerDetail(customerID string, actor domain.Actor) (domain.CustomerDetail, error) {
	customer, ok := s.store.Customer(customerID)
	if !ok {
		return domain.CustomerDetail{}, apperror.New("NOT_FOUND", "客户不存在", map[string]any{"customer_id": customerID})
	}
	if !canSeeCustomer(customer, actor) {
		return domain.CustomerDetail{}, apperror.New("FORBIDDEN", "无权查看该客户", map[string]any{"customer_id": customerID})
	}

	tasks := s.store.FollowupTasksByCustomer(customerID)
	tasks = visibleTasks(tasks, actor)
	var recommendation domain.AgentRecommendation
	if len(tasks) > 0 {
		recommendation = tasks[0].Recommendation
	}

	return domain.CustomerDetail{
		Customer:             customer,
		LatestRecommendation: recommendation,
		ProfileEvidence: []string{
			"客户主要顾虑：" + strings.Join(customer.Concerns, "、"),
			"当前阶段：" + string(customer.Stage),
			"最近沟通时间：" + customer.LastContactAt,
		},
		RecentTasks: tasks,
		RecentAgentRuns: []domain.AgentRunSummary{
			{
				ID:            "run_" + customer.ID,
				Status:        "succeeded",
				TaskType:      "generate_followup_script",
				Model:         "mock-local-v1",
				PromptVersion: "followup_v1",
				InputSummary:  customer.ProfileSummary,
				RiskFlags:     customer.RiskFlags,
				CreatedAt:     "2026-05-28T10:00:00Z",
				CompletedAt:   "2026-05-28T10:00:03Z",
			},
		},
	}, nil
}

func (s *QilingService) CustomerConversations(customerID string, r *http.Request, actor domain.Actor) (PageResult[domain.ConversationMessage], error) {
	customer, ok := s.store.Customer(customerID)
	if !ok {
		return PageResult[domain.ConversationMessage]{}, apperror.New("NOT_FOUND", "客户不存在", map[string]any{"customer_id": customerID})
	}
	if !canSeeCustomer(customer, actor) {
		return PageResult[domain.ConversationMessage]{}, apperror.New("FORBIDDEN", "无权查看该客户聊天记录", map[string]any{"customer_id": customerID})
	}
	page := PageRequestFromQuery(r)
	result := s.store.ConversationMessagePage(customerID, store.PageRequest{Page: page.Page, PageSize: page.PageSize})
	return NewPageResultWithTotal(result.Items, page, result.Total), nil
}

func (s *QilingService) CustomerShortTermMemory(customerID string, actor domain.Actor) (domain.ShortTermMemory, error) {
	customer, ok := s.store.Customer(customerID)
	if !ok {
		return domain.ShortTermMemory{}, apperror.New("NOT_FOUND", "customer not found", map[string]any{"customer_id": customerID})
	}
	if !canSeeCustomer(customer, actor) {
		return domain.ShortTermMemory{}, apperror.New("FORBIDDEN", "customer is not visible to current actor", map[string]any{"customer_id": customerID})
	}

	messages := s.store.ConversationMessagePage(customerID, store.PageRequest{Page: 1, PageSize: 5}).Items
	tasks := visibleTasks(s.store.FollowupTasksByCustomer(customerID), actor)
	runs := s.store.AgentRunsByCustomer(customerID, store.PageRequest{Page: 1, PageSize: 5}).Items
	events := s.customerAuditEvents(customer, tasks, actor)

	memory := domain.ShortTermMemory{
		Customer:               customer,
		ConversationHighlights: conversationMemoryItems(messages),
		RecentTasks:            taskMemoryItems(tasks, 5),
		RecentAgentRuns:        agentRunMemoryItems(runs),
		RecentEvents:           auditEventMemoryItems(events),
		BuiltAt:                time.Now().UTC().Format(time.RFC3339),
	}
	memory.PromptContext = buildPromptContext(memory)
	return memory, nil
}

func (s *QilingService) CustomerLongTermMemory(customerID string, actor domain.Actor) (domain.LongTermMemory, error) {
	customer, ok := s.store.Customer(customerID)
	if !ok {
		return domain.LongTermMemory{}, apperror.New("NOT_FOUND", "customer not found", map[string]any{"customer_id": customerID})
	}
	if !canSeeCustomer(customer, actor) {
		return domain.LongTermMemory{}, apperror.New("FORBIDDEN", "customer is not visible to current actor", map[string]any{"customer_id": customerID})
	}

	facts := s.store.LongTermMemoryFacts(customerID, store.PageRequest{Page: 1, PageSize: 50}).Items
	memory := domain.LongTermMemory{
		Customer: customer,
		Facts:    facts,
		BuiltAt:  time.Now().UTC().Format(time.RFC3339),
	}
	memory.PromptContext = buildLongTermPromptContext(memory)
	return memory, nil
}

func (s *QilingService) FollowupTasks(r *http.Request, actor domain.Actor) PageResult[domain.FollowupTask] {
	query := r.URL.Query()
	page := PageRequestFromQuery(r)
	ownerID := query.Get("owner_id")
	if actor.Role != "manager" {
		ownerID = actor.UserID
	}
	result := s.store.FollowupTaskPage(store.FollowupTaskFilter{
		Status:  query.Get("status"),
		Intent:  query.Get("intent"),
		OwnerID: ownerID,
	}, store.PageRequest{Page: page.Page, PageSize: page.PageSize})

	return NewPageResultWithTotal(result.Items, page, result.Total)
}

func (s *QilingService) ReviewSummary(actor domain.Actor) domain.ReviewSummary {
	customers := visibleCustomers(s.store.Customers(), actor)
	tasks := visibleTasks(s.store.FollowupTasks(), actor)
	riskCustomers := filterCustomers(customers, func(c domain.Customer) bool {
		return len(c.RiskFlags) > 0 || c.Stage == domain.StageSilent || c.Stage == domain.StageChurnRisk
	})
	opportunityCustomers := filterCustomers(customers, func(c domain.Customer) bool { return c.Intent == domain.IntentHigh })

	var sampleWarning *string
	if len(customers) < 10 {
		warning := "当前样本不足 10 个客户，复盘建议仅供参考。"
		sampleWarning = &warning
	}

	return domain.ReviewSummary{
		Metrics: []domain.Metric{
			{Key: "customers_total", Label: "客户总数", Value: len(customers), Hint: "当前 mock 样本"},
			{Key: "pending_tasks", Label: "待确认话术", Value: len(tasks), Hint: "建议优先处理"},
			{Key: "high_intent_customers", Label: "高意向客户", Value: len(opportunityCustomers), Hint: "可推进成交"},
			{Key: "risk_customers", Label: "风险客户", Value: len(riskCustomers), Hint: "需要及时介入"},
		},
		StageDistribution:    buildStageDistribution(customers),
		OpportunityCustomers: opportunityCustomers,
		RiskCustomers:        riskCustomers,
		Insights: []domain.ReviewInsight{
			{Title: "价格异议集中", Evidence: "高意向客户中存在价格/方案异议。", Suggestion: "先补充案例证明和方案价值解释，不直接承诺优惠。"},
			{Title: "沉默客户需要唤醒", Evidence: "存在超过 72 小时未回复客户。", Suggestion: "使用轻触达案例话术，避免连续催促。"},
		},
		SampleWarning: sampleWarning,
	}
}

type UploadConversationRequest struct {
	SourceType string `json:"source_type"`
	Content    string `json:"content"`
	FileName   string `json:"file_name"`
	OwnerID    string `json:"owner_id"`
}

type ConfirmUploadRequest struct {
	CustomerName string `json:"customer_name"`
	OwnerID      string `json:"owner_id"`
}

type CopyTaskRequest struct {
	CopiedScript   string `json:"copied_script"`
	ClientCopiedAt string `json:"client_copied_at"`
}

type SkipTaskRequest struct {
	Reason string `json:"reason"`
}

type MarkWrongRequest struct {
	Reason      string   `json:"reason"`
	WrongFields []string `json:"wrong_fields"`
}

type RegenerateTaskRequest struct {
	Instruction string `json:"instruction"`
}

func (s *QilingService) UploadConversation(req UploadConversationRequest, actor domain.Actor, requestID string) (domain.UploadConversationResult, error) {
	if strings.TrimSpace(req.Content) == "" {
		return domain.UploadConversationResult{}, apperror.New("EMPTY_CONTENT", "聊天记录为空", map[string]any{"field": "content"})
	}
	if req.SourceType == "" {
		req.SourceType = "pasted_text"
	}
	if req.SourceType != "pasted_text" && req.SourceType != "txt_file" && req.SourceType != "csv_file" {
		return domain.UploadConversationResult{}, apperror.New("UNSUPPORTED_UPLOAD_TYPE", "暂不支持该上传格式", map[string]any{"source_type": req.SourceType})
	}
	if req.OwnerID == "" {
		req.OwnerID = "usr_001"
	}

	record, err := s.store.CreateUpload(req.SourceType, req.Content, req.OwnerID)
	if err != nil {
		return domain.UploadConversationResult{}, err
	}
	if err := s.recordAudit(domain.AuditEvent{
		Action:     domain.AuditUploadConversationCreated,
		Actor:      effectiveActor(actor, req.OwnerID),
		RequestID:  requestID,
		EntityType: "upload",
		EntityID:   record.ID,
		Metadata: map[string]any{
			"source_type":     record.SourceType,
			"message_count":   len(record.Messages),
			"parsed_customer": record.ParsedCustomer.Name,
		},
	}); err != nil {
		return domain.UploadConversationResult{}, err
	}

	return domain.UploadConversationResult{
		UploadID:       record.ID,
		Status:         record.Status,
		ParsedCustomer: record.ParsedCustomer,
		MessageCount:   len(record.Messages),
		Warnings:       record.Warnings,
		NextAction:     "confirm_parsed_result",
	}, nil
}

func (s *QilingService) Upload(uploadID string) (domain.UploadRecord, error) {
	record, ok := s.store.Upload(uploadID)
	if !ok {
		return domain.UploadRecord{}, apperror.New("NOT_FOUND", "上传记录不存在", map[string]any{"upload_id": uploadID})
	}
	return record, nil
}

func (s *QilingService) AgentRun(runID string) (domain.AgentRun, error) {
	run, ok := s.store.AgentRun(runID)
	if !ok {
		return domain.AgentRun{}, apperror.New("NOT_FOUND", "AgentRun 不存在", map[string]any{"agent_run_id": runID})
	}
	return run, nil
}

func (s *QilingService) ConfirmUpload(uploadID string, req ConfirmUploadRequest, actor domain.Actor, requestID string) (domain.ConfirmUploadResult, error) {
	if req.OwnerID == "" {
		req.OwnerID = "usr_001"
	}
	upload, err := s.Upload(uploadID)
	if err != nil {
		return domain.ConfirmUploadResult{}, err
	}
	customerName := req.CustomerName
	if strings.TrimSpace(customerName) == "" {
		customerName = upload.ParsedCustomer.Name
	}
	memoryContext := s.memoryContextForUpload(customerName, upload, actor)
	agentRun := s.agent.GenerateFollowup(agent.RunInput{
		CustomerName:  customerName,
		OwnerID:       req.OwnerID,
		RawContent:    conversationContentSummary(upload.Messages),
		MemoryContext: memoryContext,
		Now:           time.Now().UTC(),
	})
	result, err := s.store.ConfirmUpload(uploadID, customerName, req.OwnerID, store.ConfirmUploadAgentRun{
		TaskType:         agentRun.TaskType,
		Model:            agentRun.Model,
		PromptVersion:    agentRun.PromptVersion,
		InputSummary:     agentRun.InputSummary,
		Recommendation:   agentRun.Recommendation,
		ValidationErrors: agentRun.ValidationErrors,
		RiskFlags:        agentRun.RiskFlags,
	})
	if err != nil {
		return domain.ConfirmUploadResult{}, err
	}
	if err := s.persistRecommendationMemory(result.CustomerID, result.AgentRunID, agentRun.Recommendation); err != nil {
		return domain.ConfirmUploadResult{}, err
	}
	if err := s.recordAudit(domain.AuditEvent{
		Action:      domain.AuditUploadConfirmed,
		Actor:       effectiveActor(actor, req.OwnerID),
		RequestID:   requestID,
		EntityType:  "upload",
		EntityID:    uploadID,
		RelatedType: "customer",
		RelatedID:   result.CustomerID,
		Metadata: map[string]any{
			"agent_run_id":      result.AgentRunID,
			"conversation_id":   result.ConversationID,
			"followup_task_id":  result.FollowupTaskID,
			"customer_name":     req.CustomerName,
			"confirmation_mode": "manual",
		},
	}); err != nil {
		return domain.ConfirmUploadResult{}, err
	}
	return result, nil
}

func (s *QilingService) CopyTask(taskID string, req CopyTaskRequest, actor domain.Actor, requestID string) (domain.TaskCopyResult, error) {
	copiedAt := req.ClientCopiedAt
	if copiedAt == "" {
		copiedAt = "2026-05-28T10:35:00Z"
	}
	result, err := s.store.CopyTask(taskID, copiedAt)
	if err != nil {
		return domain.TaskCopyResult{}, err
	}
	if err := s.recordAudit(domain.AuditEvent{
		Action:     domain.AuditFollowupTaskCopied,
		Actor:      actor,
		RequestID:  requestID,
		EntityType: "followup_task",
		EntityID:   taskID,
		Metadata: map[string]any{
			"client_copied_at": copiedAt,
			"has_script":       strings.TrimSpace(req.CopiedScript) != "",
		},
	}); err != nil {
		return domain.TaskCopyResult{}, err
	}
	return result, nil
}

func (s *QilingService) SkipTask(taskID string, req SkipTaskRequest, actor domain.Actor, requestID string) (domain.TaskStatusResult, error) {
	result, err := s.store.SkipTask(taskID, req.Reason)
	if err != nil {
		return domain.TaskStatusResult{}, err
	}
	if err := s.recordAudit(domain.AuditEvent{
		Action:     domain.AuditFollowupTaskSkipped,
		Actor:      actor,
		RequestID:  requestID,
		EntityType: "followup_task",
		EntityID:   taskID,
		Metadata: map[string]any{
			"reason": req.Reason,
		},
	}); err != nil {
		return domain.TaskStatusResult{}, err
	}
	return result, nil
}

func (s *QilingService) MarkTaskWrong(taskID string, req MarkWrongRequest, actor domain.Actor, requestID string) (domain.MarkWrongResult, error) {
	if strings.TrimSpace(req.Reason) == "" {
		return domain.MarkWrongResult{}, apperror.New("VALIDATION_ERROR", "标记不准必须填写原因", map[string]any{"field": "reason"})
	}
	result, err := s.store.MarkTaskWrong(taskID, req.Reason)
	if err != nil {
		return domain.MarkWrongResult{}, err
	}
	if err := s.recordAudit(domain.AuditEvent{
		Action:     domain.AuditFollowupTaskMarkedWrong,
		Actor:      actor,
		RequestID:  requestID,
		EntityType: "followup_task",
		EntityID:   taskID,
		Metadata: map[string]any{
			"feedback_id":  result.FeedbackID,
			"reason":       req.Reason,
			"wrong_fields": req.WrongFields,
		},
	}); err != nil {
		return domain.MarkWrongResult{}, err
	}
	return result, nil
}

func (s *QilingService) RegenerateTask(taskID string, req RegenerateTaskRequest, actor domain.Actor, requestID string) (domain.RegenerateTaskResult, error) {
	result, err := s.store.RegenerateTask(taskID, req.Instruction)
	if err != nil {
		return domain.RegenerateTaskResult{}, err
	}
	if err := s.recordAudit(domain.AuditEvent{
		Action:      domain.AuditFollowupTaskRegenerated,
		Actor:       actor,
		RequestID:   requestID,
		EntityType:  "followup_task",
		EntityID:    taskID,
		RelatedType: "agent_run",
		RelatedID:   result.AgentRunID,
		Metadata: map[string]any{
			"has_instruction": strings.TrimSpace(req.Instruction) != "",
			"intent_level":    result.Recommendation.IntentLevel,
			"customer_stage":  result.Recommendation.CustomerStage,
		},
	}); err != nil {
		return domain.RegenerateTaskResult{}, err
	}
	return result, nil
}

func (s *QilingService) AuditEvents(r *http.Request, actor domain.Actor) PageResult[domain.AuditEvent] {
	query := r.URL.Query()
	page := PageRequestFromQuery(r)
	filter := store.AuditEventFilter{
		Action:     query.Get("action"),
		EntityType: query.Get("entity_type"),
		EntityID:   query.Get("entity_id"),
	}
	if actor.Role == "manager" {
		filter.ActorID = query.Get("actor_id")
	} else {
		filter.ActorID = actor.UserID
	}
	result := s.store.AuditEventPage(filter, store.PageRequest{Page: page.Page, PageSize: page.PageSize})
	return NewPageResultWithTotal(result.Items, page, result.Total)
}

func (s *QilingService) recordAudit(event domain.AuditEvent) error {
	now := time.Now().UTC()
	event.ID = fmt.Sprintf("audit_%d", now.UnixNano())
	event.CreatedAt = now.Format(time.RFC3339)
	_, err := s.store.CreateAuditEvent(event)
	return err
}

func effectiveActor(actor domain.Actor, fallbackUserID string) domain.Actor {
	if actor.UserID != "" {
		return actor
	}
	role := actor.Role
	if role == "" {
		role = "sales"
	}
	return domain.Actor{UserID: fallbackUserID, Role: role}
}

func conversationContentSummary(messages []domain.ConversationMessage) string {
	parts := make([]string, 0, len(messages))
	for _, message := range messages {
		content := strings.TrimSpace(message.Content)
		if content != "" {
			parts = append(parts, content)
		}
	}
	return strings.Join(parts, "\n")
}

func (s *QilingService) persistRecommendationMemory(customerID string, agentRunID string, recommendation domain.AgentRecommendation) error {
	facts := recommendationMemoryFacts(customerID, agentRunID, recommendation)
	for _, fact := range facts {
		if _, err := s.store.UpsertLongTermMemoryFact(fact); err != nil {
			return err
		}
	}
	return nil
}

func recommendationMemoryFacts(customerID string, agentRunID string, recommendation domain.AgentRecommendation) []domain.LongTermMemoryFact {
	now := time.Now().UTC().Format(time.RFC3339)
	facts := make([]domain.LongTermMemoryFact, 0)
	appendFact := func(category string, key string, value string, confidence float64) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		facts = append(facts, domain.LongTermMemoryFact{
			ID:         "",
			CustomerID: customerID,
			Category:   category,
			Key:        key,
			Value:      value,
			Confidence: confidence,
			SourceType: "agent_run",
			SourceID:   agentRunID,
			Status:     domain.MemoryFactActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}

	appendFact("profile", "customer_stage", string(recommendation.CustomerStage), 0.82)
	appendFact("profile", "intent_level", string(recommendation.IntentLevel), 0.82)
	appendFact("sales", "recommended_action", recommendation.RecommendedAction, 0.72)
	for _, concern := range recommendation.MainConcerns {
		appendFact("concern", normalizeMemoryKey(concern), concern, 0.78)
	}
	for _, risk := range recommendation.RiskFlags {
		appendFact("risk", normalizeMemoryKey(risk), risk, 0.68)
	}
	return facts
}

func normalizeMemoryKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(" ", "_", "\t", "_", "\n", "_", "/", "_", "\\", "_", "|", "_")
	value = replacer.Replace(value)
	runes := []rune(value)
	if len(runes) > 80 {
		return string(runes[:80])
	}
	return value
}

func (s *QilingService) memoryContextForUpload(customerName string, upload domain.UploadRecord, actor domain.Actor) string {
	if customer, ok := s.customerByName(customerName, actor); ok {
		memory, err := s.CustomerShortTermMemory(customer.ID, actor)
		if err == nil && strings.TrimSpace(memory.PromptContext) != "" {
			return memory.PromptContext
		}
	}

	return buildUploadMemoryContext(customerName, upload)
}

func (s *QilingService) customerByName(customerName string, actor domain.Actor) (domain.Customer, bool) {
	customerName = strings.TrimSpace(customerName)
	if customerName == "" {
		return domain.Customer{}, false
	}
	for _, customer := range visibleCustomers(s.store.Customers(), actor) {
		if strings.EqualFold(strings.TrimSpace(customer.Name), customerName) {
			return customer, true
		}
	}
	return domain.Customer{}, false
}

func buildUploadMemoryContext(customerName string, upload domain.UploadRecord) string {
	parts := []string{
		"Customer: " + strings.TrimSpace(customerName),
		"Source: " + upload.SourceType,
		"Upload status: " + string(upload.Status),
	}
	if upload.ParsedCustomer.OwnerName != "" {
		parts = append(parts, "Parsed owner: "+upload.ParsedCustomer.OwnerName)
	}
	if len(upload.Warnings) > 0 {
		parts = append(parts, "Upload warnings: "+strings.Join(upload.Warnings, ", "))
	}
	if len(upload.Messages) > 0 {
		items := conversationMemoryItems(upload.Messages)
		parts = appendMemorySection(parts, "Uploaded conversation", items)
	}
	return strings.Join(parts, "\n")
}

func (s *QilingService) customerAuditEvents(customer domain.Customer, tasks []domain.FollowupTask, actor domain.Actor) []domain.AuditEvent {
	filter := store.AuditEventFilter{EntityType: "customer", EntityID: customer.ID}
	if actor.Role != "manager" {
		filter.ActorID = actor.UserID
	}
	events := s.store.AuditEventPage(filter, store.PageRequest{Page: 1, PageSize: 5}).Items

	seen := map[string]bool{}
	for _, event := range events {
		seen[event.ID] = true
	}
	for _, task := range tasks {
		if len(events) >= 10 {
			break
		}
		filter := store.AuditEventFilter{EntityType: "followup_task", EntityID: task.ID}
		if actor.Role != "manager" {
			filter.ActorID = actor.UserID
		}
		taskEvents := s.store.AuditEventPage(filter, store.PageRequest{Page: 1, PageSize: 3}).Items
		for _, event := range taskEvents {
			if seen[event.ID] {
				continue
			}
			events = append(events, event)
			seen[event.ID] = true
			if len(events) >= 10 {
				break
			}
		}
	}
	return events
}

func conversationMemoryItems(messages []domain.ConversationMessage) []domain.MemoryItem {
	items := make([]domain.MemoryItem, 0, len(messages))
	for _, message := range messages {
		content := compactText(message.Content, 120)
		if content == "" {
			continue
		}
		items = append(items, domain.MemoryItem{
			Type:      "conversation_message",
			ID:        message.ID,
			Summary:   message.SenderName + ": " + content,
			Evidence:  message.SenderType,
			CreatedAt: message.SentAt,
		})
	}
	return items
}

func taskMemoryItems(tasks []domain.FollowupTask, limit int) []domain.MemoryItem {
	if len(tasks) > limit {
		tasks = tasks[:limit]
	}
	items := make([]domain.MemoryItem, 0, len(tasks))
	for _, task := range tasks {
		summary := strings.TrimSpace(task.Recommendation.RecommendedAction)
		if summary == "" {
			summary = task.Type
		}
		items = append(items, domain.MemoryItem{
			Type:      "followup_task",
			ID:        task.ID,
			Summary:   string(task.Status) + " | " + summary,
			Evidence:  compactText(task.Recommendation.Reasoning, 160),
			CreatedAt: task.GeneratedAt,
		})
	}
	return items
}

func agentRunMemoryItems(runs []domain.AgentRun) []domain.MemoryItem {
	items := make([]domain.MemoryItem, 0, len(runs))
	for _, run := range runs {
		evidence := compactText(run.InputSummary, 160)
		if len(run.ValidationErrors) > 0 {
			evidence = compactText(evidence+" | validation: "+strings.Join(run.ValidationErrors, "; "), 220)
		}
		items = append(items, domain.MemoryItem{
			Type:      "agent_run",
			ID:        run.ID,
			Summary:   run.TaskType + " | " + run.Status + " | " + run.PromptVersion,
			Evidence:  evidence,
			CreatedAt: run.CreatedAt,
		})
	}
	return items
}

func auditEventMemoryItems(events []domain.AuditEvent) []domain.MemoryItem {
	items := make([]domain.MemoryItem, 0, len(events))
	for _, event := range events {
		evidence := ""
		if event.RelatedType != "" || event.RelatedID != "" {
			evidence = event.RelatedType + "/" + event.RelatedID
		}
		items = append(items, domain.MemoryItem{
			Type:      "audit_event",
			ID:        event.ID,
			Summary:   string(event.Action) + " on " + event.EntityType + "/" + event.EntityID,
			Evidence:  evidence,
			CreatedAt: event.CreatedAt,
		})
	}
	return items
}

func buildPromptContext(memory domain.ShortTermMemory) string {
	sections := []string{
		"Customer: " + memory.Customer.Name,
		"Stage: " + string(memory.Customer.Stage) + ", intent: " + string(memory.Customer.Intent),
		"Owner: " + memory.Customer.Owner.Name + " (" + memory.Customer.Owner.ID + ")",
		"Profile: " + compactText(memory.Customer.ProfileSummary, 240),
	}
	if len(memory.Customer.Concerns) > 0 {
		sections = append(sections, "Concerns: "+strings.Join(memory.Customer.Concerns, ", "))
	}
	if len(memory.Customer.RiskFlags) > 0 {
		sections = append(sections, "Risk flags: "+strings.Join(memory.Customer.RiskFlags, ", "))
	}
	sections = appendMemorySection(sections, "Recent conversation", memory.ConversationHighlights)
	sections = appendMemorySection(sections, "Recent tasks", memory.RecentTasks)
	sections = appendMemorySection(sections, "Recent agent runs", memory.RecentAgentRuns)
	sections = appendMemorySection(sections, "Recent events", memory.RecentEvents)
	return strings.Join(sections, "\n")
}

func buildLongTermPromptContext(memory domain.LongTermMemory) string {
	sections := []string{
		"Long-term memory for customer: " + memory.Customer.Name,
	}
	if len(memory.Facts) == 0 {
		sections = append(sections, "No active long-term facts yet.")
		return strings.Join(sections, "\n")
	}
	for _, fact := range memory.Facts {
		sections = append(sections, "- "+fact.Category+"."+fact.Key+": "+fact.Value+" (confidence "+formatConfidence(fact.Confidence)+", source "+fact.SourceType+"/"+fact.SourceID+")")
	}
	return strings.Join(sections, "\n")
}

func formatConfidence(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", value), "0"), ".")
}

func appendMemorySection(sections []string, title string, items []domain.MemoryItem) []string {
	if len(items) == 0 {
		return sections
	}
	sections = append(sections, title+":")
	for _, item := range items {
		line := "- " + item.Summary
		if item.Evidence != "" {
			line += " (" + item.Evidence + ")"
		}
		sections = append(sections, line)
	}
	return sections
}

func compactText(value string, maxRunes int) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes]) + "..."
}

func countCustomersByIntent(customers []domain.Customer, intent domain.IntentLevel) int {
	count := 0
	for _, customer := range customers {
		if customer.Intent == intent {
			count++
		}
	}
	return count
}

func countRiskCustomers(customers []domain.Customer) int {
	count := 0
	for _, customer := range customers {
		if len(customer.RiskFlags) > 0 {
			count++
		}
	}
	return count
}

func filterCustomers(customers []domain.Customer, keep func(domain.Customer) bool) []domain.Customer {
	filtered := make([]domain.Customer, 0, len(customers))
	for _, customer := range customers {
		if keep(customer) {
			filtered = append(filtered, customer)
		}
	}
	return filtered
}

func visibleCustomers(customers []domain.Customer, actor domain.Actor) []domain.Customer {
	if actor.Role == "manager" {
		return customers
	}
	return filterCustomers(customers, func(customer domain.Customer) bool {
		return customer.Owner.ID == actor.UserID
	})
}

func visibleTasks(tasks []domain.FollowupTask, actor domain.Actor) []domain.FollowupTask {
	if actor.Role == "manager" {
		return tasks
	}
	filtered := make([]domain.FollowupTask, 0, len(tasks))
	for _, task := range tasks {
		if task.Customer.Owner.ID == actor.UserID {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

func canSeeCustomer(customer domain.Customer, actor domain.Actor) bool {
	return actor.Role == "manager" || customer.Owner.ID == actor.UserID
}

func buildStageDistribution(customers []domain.Customer) []domain.StageDistribution {
	counts := map[domain.CustomerStage]int{}
	order := make([]domain.CustomerStage, 0)
	for _, customer := range customers {
		if _, exists := counts[customer.Stage]; !exists {
			order = append(order, customer.Stage)
		}
		counts[customer.Stage]++
	}

	distribution := make([]domain.StageDistribution, 0, len(order))
	for _, stage := range order {
		distribution = append(distribution, domain.StageDistribution{Stage: stage, Count: counts[stage]})
	}
	return distribution
}
