package service

import (
	"net/http"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/apperror"
	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/heqiucheng/qiling-agent/backend/internal/store"
)

type QilingService struct {
	store *store.MockStore
}

func NewQilingService(store *store.MockStore) *QilingService {
	return &QilingService{store: store}
}

func (s *QilingService) DashboardSummary() domain.DashboardSummary {
	customers := s.store.Customers()
	tasks := s.store.FollowupTasks()

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

func (s *QilingService) Customers(r *http.Request) PageResult[domain.Customer] {
	query := r.URL.Query()
	customers := s.store.Customers()
	filtered := make([]domain.Customer, 0, len(customers))
	for _, customer := range customers {
		if store.MatchCustomer(
			customer,
			query.Get("keyword"),
			query.Get("stage"),
			query.Get("intent"),
			query.Get("owner_id"),
			query.Get("risk"),
		) {
			filtered = append(filtered, customer)
		}
	}
	return NewPageResult(filtered, PageRequestFromQuery(r))
}

func (s *QilingService) CustomerDetail(customerID string) (domain.CustomerDetail, error) {
	customer, ok := s.store.Customer(customerID)
	if !ok {
		return domain.CustomerDetail{}, apperror.New("NOT_FOUND", "客户不存在", map[string]any{"customer_id": customerID})
	}

	tasks := s.store.FollowupTasksByCustomer(customerID)
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

func (s *QilingService) CustomerConversations(customerID string, r *http.Request) PageResult[domain.ConversationMessage] {
	return NewPageResult(s.store.ConversationMessages(customerID), PageRequestFromQuery(r))
}

func (s *QilingService) FollowupTasks(r *http.Request) PageResult[domain.FollowupTask] {
	query := r.URL.Query()
	tasks := s.store.FollowupTasks()
	filtered := make([]domain.FollowupTask, 0, len(tasks))
	for _, task := range tasks {
		if query.Get("status") != "" && string(task.Status) != query.Get("status") {
			continue
		}
		if query.Get("intent") != "" && string(task.Customer.Intent) != query.Get("intent") {
			continue
		}
		if query.Get("owner_id") != "" && task.Customer.Owner.ID != query.Get("owner_id") {
			continue
		}
		filtered = append(filtered, task)
	}
	return NewPageResult(filtered, PageRequestFromQuery(r))
}

func (s *QilingService) ReviewSummary() domain.ReviewSummary {
	customers := s.store.Customers()
	tasks := s.store.FollowupTasks()
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

func (s *QilingService) UploadConversation(req UploadConversationRequest) (domain.UploadConversationResult, error) {
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

	record := s.store.CreateUpload(req.SourceType, req.Content, req.OwnerID)
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

func (s *QilingService) ConfirmUpload(uploadID string, req ConfirmUploadRequest) (domain.ConfirmUploadResult, error) {
	if req.OwnerID == "" {
		req.OwnerID = "usr_001"
	}
	return s.store.ConfirmUpload(uploadID, req.CustomerName, req.OwnerID)
}

func (s *QilingService) CopyTask(taskID string, req CopyTaskRequest) (domain.TaskCopyResult, error) {
	copiedAt := req.ClientCopiedAt
	if copiedAt == "" {
		copiedAt = "2026-05-28T10:35:00Z"
	}
	return s.store.CopyTask(taskID, copiedAt)
}

func (s *QilingService) SkipTask(taskID string, req SkipTaskRequest) (domain.TaskStatusResult, error) {
	return s.store.SkipTask(taskID, req.Reason)
}

func (s *QilingService) MarkTaskWrong(taskID string, req MarkWrongRequest) (domain.MarkWrongResult, error) {
	if strings.TrimSpace(req.Reason) == "" {
		return domain.MarkWrongResult{}, apperror.New("VALIDATION_ERROR", "标记不准必须填写原因", map[string]any{"field": "reason"})
	}
	return s.store.MarkTaskWrong(taskID, req.Reason)
}

func (s *QilingService) RegenerateTask(taskID string, req RegenerateTaskRequest) (domain.RegenerateTaskResult, error) {
	return s.store.RegenerateTask(taskID, req.Instruction)
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
