package store

import (
	"fmt"
	"strings"
	"sync"

	"github.com/heqiucheng/qiling-agent/backend/internal/agent"
	"github.com/heqiucheng/qiling-agent/backend/internal/apperror"
	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

type MockStore struct {
	mu        sync.Mutex
	nextID    int
	uploads   map[string]domain.UploadRecord
	customers []domain.Customer
	tasks     []domain.FollowupTask
	audit     []domain.AuditEvent
	runs      map[string]domain.AgentRun
}

func NewMockStore() *MockStore {
	customers := []domain.Customer{
		{
			ID:             "cus_001",
			Name:           "王女士",
			Source:         "企业微信",
			Owner:          domain.Owner{ID: "usr_001", Name: "销售A"},
			Stage:          domain.StagePriceObjection,
			Intent:         domain.IntentHigh,
			Concerns:       []string{"价格", "售后"},
			Tags:           []string{"价格敏感", "需要案例"},
			ProfileSummary: "关注价格和售后保障，近期购买意向较强。",
			LastContactAt:  "2026-05-28T09:30:00Z",
			PendingTasks:   1,
			RiskFlags:      []string{"涉及价格承诺，需人工确认"},
		},
		{
			ID:             "cus_002",
			Name:           "李先生",
			Source:         "上传聊天记录",
			Owner:          domain.Owner{ID: "usr_001", Name: "销售A"},
			Stage:          domain.StageSilent,
			Intent:         domain.IntentMedium,
			Concerns:       []string{"效果", "案例"},
			Tags:           []string{"需要案例"},
			ProfileSummary: "已了解产品但 3 天未回复，适合用案例轻触达。",
			LastContactAt:  "2026-05-25T18:20:00Z",
			PendingTasks:   1,
			RiskFlags:      []string{"客户沉默超过 72 小时"},
		},
		{
			ID:             "cus_003",
			Name:           "陈总",
			Source:         "企业微信",
			Owner:          domain.Owner{ID: "usr_002", Name: "销售B"},
			Stage:          domain.StageHighIntent,
			Intent:         domain.IntentHigh,
			Concerns:       []string{"交付周期", "售后"},
			Tags:           []string{"时间紧急", "关注售后"},
			ProfileSummary: "明确提出本周内确认方案，主管可关注推进。",
			LastContactAt:  "2026-05-28T08:50:00Z",
			PendingTasks:   1,
			RiskFlags:      []string{},
		},
	}

	return &MockStore{
		nextID:    4,
		uploads:   map[string]domain.UploadRecord{},
		customers: customers,
		audit:     []domain.AuditEvent{},
		runs:      map[string]domain.AgentRun{},
		tasks: []domain.FollowupTask{
			newTask("task_001", customers[0], "price_objection", "2026-05-28T10:00:00Z", domain.AgentRecommendation{
				CustomerStage:     domain.StagePriceObjection,
				IntentLevel:       domain.IntentHigh,
				MainConcerns:      []string{"价格", "效果", "售后"},
				RecommendedAction: "解释方案价值并引导预约",
				Script:            "您好，刚才您提到比较关注价格，我这边帮您整理了一下更适合您的方案，也可以结合售后保障一起看。",
				Reasoning:         "客户连续询问价格和售后，说明有购买兴趣但存在决策顾虑。",
				RiskFlags:         []string{"涉及价格承诺，建议人工确认"},
				NextFollowupTime:  "2026-05-28T16:00:00Z",
			}),
			newTask("task_002", customers[1], "silent_reactivation", "2026-05-28T10:10:00Z", domain.AgentRecommendation{
				CustomerStage:     domain.StageSilent,
				IntentLevel:       domain.IntentMedium,
				MainConcerns:      []string{"效果", "案例"},
				RecommendedAction: "用案例轻触达",
				Script:            "您好，前面您提到比较关注实际效果，我整理了一个和您情况接近的案例，您方便的话我发您看一下。",
				Reasoning:         "客户已表达兴趣但长时间未回复，直接促单风险较高，适合先提供案例降低压力。",
				RiskFlags:         []string{"避免连续催促造成反感"},
				NextFollowupTime:  "2026-05-28T15:00:00Z",
			}),
			newTask("task_003", customers[2], "closing", "2026-05-28T10:20:00Z", domain.AgentRecommendation{
				CustomerStage:     domain.StageHighIntent,
				IntentLevel:       domain.IntentHigh,
				MainConcerns:      []string{"交付周期", "售后"},
				RecommendedAction: "确认决策节点并推动方案评审",
				Script:            "陈总，您这边如果希望本周确认，我建议今天先把交付周期和售后边界对齐，我可以整理成一页方案给您内部评估。",
				Reasoning:         "客户给出明确时间窗口，需要推进到具体评审动作。",
				RiskFlags:         []string{"交付周期不能超出实际能力承诺"},
				NextFollowupTime:  "2026-05-28T14:00:00Z",
			}),
		},
	}
}

func (s *MockStore) Customers() []domain.Customer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.Customer(nil), s.customers...)
}

func (s *MockStore) CustomerPage(filter CustomerFilter, page PageRequest) CustomerPage {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := make([]domain.Customer, 0, len(s.customers))
	for _, customer := range s.customers {
		if MatchCustomer(customer, filter.Keyword, filter.Stage, filter.Intent, filter.OwnerID, filter.Risk) {
			filtered = append(filtered, customer)
		}
	}

	return CustomerPage{Items: paginate(filtered, page), Total: len(filtered)}
}

func (s *MockStore) Customer(id string) (domain.Customer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, customer := range s.customers {
		if customer.ID == id {
			return customer, true
		}
	}
	return domain.Customer{}, false
}

func (s *MockStore) FollowupTasks() []domain.FollowupTask {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.FollowupTask(nil), s.tasks...)
}

func (s *MockStore) FollowupTaskPage(filter FollowupTaskFilter, page PageRequest) FollowupTaskPage {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := make([]domain.FollowupTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		if filter.Status != "" && string(task.Status) != filter.Status {
			continue
		}
		if filter.Intent != "" && string(task.Customer.Intent) != filter.Intent {
			continue
		}
		if filter.OwnerID != "" && task.Customer.Owner.ID != filter.OwnerID {
			continue
		}
		filtered = append(filtered, task)
	}

	return FollowupTaskPage{Items: paginate(filtered, page), Total: len(filtered)}
}

func (s *MockStore) FollowupTasksByCustomer(customerID string) []domain.FollowupTask {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks := make([]domain.FollowupTask, 0)
	for _, task := range s.tasks {
		if task.Customer.ID == customerID {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (s *MockStore) ConversationMessages(customerID string) []domain.ConversationMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, upload := range s.uploads {
		for _, customer := range s.customers {
			if customer.ID == customerID && customer.Name == upload.ParsedCustomer.Name {
				return append([]domain.ConversationMessage(nil), upload.Messages...)
			}
		}
	}

	customerName := "客户"
	for _, customer := range s.customers {
		if customer.ID == customerID {
			customerName = customer.Name
			break
		}
	}
	return []domain.ConversationMessage{
		{ID: "msg_" + customerID + "_001", SenderType: "customer", SenderName: customerName, Content: "这个价格还能优惠吗？", SentAt: "2026-05-28T09:20:00Z"},
		{ID: "msg_" + customerID + "_002", SenderType: "sales", SenderName: "销售A", Content: "我先结合您的需求整理一个更适合的方案。", SentAt: "2026-05-28T09:22:00Z"},
	}
}

func (s *MockStore) ConversationMessagePage(customerID string, page PageRequest) ConversationMessagePage {
	messages := s.ConversationMessages(customerID)
	return ConversationMessagePage{Items: paginate(messages, page), Total: len(messages)}
}

func (s *MockStore) CreateUpload(sourceType string, content string, ownerID string) (domain.UploadRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("upl_%03d", s.nextID)
	s.nextID++

	record := domain.UploadRecord{
		ID:         id,
		Status:     domain.UploadNeedsConfirmation,
		SourceType: sourceType,
		ParsedCustomer: domain.ParsedCustomer{
			Name:      inferCustomerName(content),
			OwnerName: ownerName(ownerID),
		},
		Messages: []domain.ConversationMessage{
			{
				ID:         "msg_" + id,
				SenderType: "customer",
				SenderName: inferCustomerName(content),
				Content:    strings.TrimSpace(content),
				SentAt:     "2026-05-28T10:30:00Z",
			},
		},
		Warnings:  []string{},
		CreatedAt: "2026-05-28T10:30:00Z",
	}
	s.uploads[id] = record
	return record, nil
}

func (s *MockStore) Upload(id string) (domain.UploadRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.uploads[id]
	return record, ok
}

func (s *MockStore) ConfirmUpload(uploadID string, customerName string, ownerID string) (domain.ConfirmUploadResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.uploads[uploadID]
	if !ok {
		return domain.ConfirmUploadResult{}, apperror.New("NOT_FOUND", "上传记录不存在", map[string]any{"upload_id": uploadID})
	}

	if strings.TrimSpace(customerName) == "" {
		customerName = record.ParsedCustomer.Name
	}

	customerID := fmt.Sprintf("cus_%03d", s.nextID)
	taskID := fmt.Sprintf("task_%03d", s.nextID)
	agentRunID := fmt.Sprintf("run_%03d", s.nextID)
	conversationID := fmt.Sprintf("conv_%03d", s.nextID)
	s.nextID++

	customer := domain.Customer{
		ID:             customerID,
		Name:           customerName,
		Source:         "上传聊天记录",
		Owner:          domain.Owner{ID: ownerID, Name: ownerName(ownerID)},
		Stage:          domain.StagePriceObjection,
		Intent:         domain.IntentHigh,
		Concerns:       []string{"价格", "效果"},
		Tags:           []string{"价格敏感"},
		ProfileSummary: "由上传聊天记录生成的 mock 客户画像，客户正在比较价格和效果。",
		LastContactAt:  "2026-05-28T10:30:00Z",
		PendingTasks:   1,
		RiskFlags:      []string{"涉及价格承诺，需人工确认"},
	}

	task := newTask(taskID, customer, "price_objection", "2026-05-28T10:31:00Z", domain.AgentRecommendation{
		CustomerStage:     domain.StagePriceObjection,
		IntentLevel:       domain.IntentHigh,
		MainConcerns:      []string{"价格", "效果"},
		RecommendedAction: "解释价值并提供案例",
		Script:            "您好，您刚才提到价格和效果，我建议先结合您的使用场景看投入产出，我可以给您整理一个接近情况的案例。",
		Reasoning:         "上传内容显示客户关注价格和效果，需要先建立价值感再推动下一步。",
		RiskFlags:         []string{"避免直接承诺优惠或效果"},
		NextFollowupTime:  "2026-05-28T16:00:00Z",
	})

	record.Status = domain.UploadConfirmed
	s.uploads[uploadID] = record
	s.customers = append(s.customers, customer)
	s.tasks = append(s.tasks, task)
	s.runs[agentRunID] = domain.AgentRun{
		ID:               agentRunID,
		CustomerID:       customerID,
		TaskType:         agent.TaskGenerateFollowupScript,
		Status:           "succeeded",
		Model:            agent.ModelMockLocalV1,
		PromptVersion:    agent.PromptFollowupV1,
		InputSummary:     "上传聊天记录生成客户画像和跟进话术",
		Output:           task.Recommendation,
		ValidationErrors: agent.ValidateRecommendation(task.Recommendation),
		RiskFlags:        task.Recommendation.RiskFlags,
		CreatedAt:        "2026-05-28T10:31:00Z",
		CompletedAt:      "2026-05-28T10:31:00Z",
	}

	return domain.ConfirmUploadResult{
		CustomerID:     customerID,
		ConversationID: conversationID,
		AgentRunID:     agentRunID,
		FollowupTaskID: taskID,
		Status:         domain.UploadConfirmed,
	}, nil
}

func (s *MockStore) CopyTask(taskID string, copiedAt string) (domain.TaskCopyResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.taskIndex(taskID)
	if err != nil {
		return domain.TaskCopyResult{}, err
	}
	if s.tasks[index].Status != domain.FollowupPending {
		return domain.TaskCopyResult{}, apperror.New("TASK_ALREADY_FINALIZED", "任务已经处理，不能重复操作", map[string]any{"task_id": taskID})
	}
	s.tasks[index].Status = domain.FollowupCopied
	return domain.TaskCopyResult{TaskID: taskID, Status: domain.FollowupCopied, CopiedAt: copiedAt}, nil
}

func (s *MockStore) SkipTask(taskID string, reason string) (domain.TaskStatusResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.taskIndex(taskID)
	if err != nil {
		return domain.TaskStatusResult{}, err
	}
	if s.tasks[index].Status != domain.FollowupPending {
		return domain.TaskStatusResult{}, apperror.New("TASK_ALREADY_FINALIZED", "任务已经处理，不能重复操作", map[string]any{"task_id": taskID})
	}
	s.tasks[index].Status = domain.FollowupSkipped
	s.tasks[index].Feedback = &domain.TaskFeedback{Reason: reason}
	return domain.TaskStatusResult{TaskID: taskID, Status: domain.FollowupSkipped}, nil
}

func (s *MockStore) MarkTaskWrong(taskID string, reason string) (domain.MarkWrongResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.taskIndex(taskID)
	if err != nil {
		return domain.MarkWrongResult{}, err
	}
	if s.tasks[index].Status != domain.FollowupPending {
		return domain.MarkWrongResult{}, apperror.New("TASK_ALREADY_FINALIZED", "任务已经处理，不能重复操作", map[string]any{"task_id": taskID})
	}
	s.tasks[index].Status = domain.FollowupMarkedWrong
	s.tasks[index].Feedback = &domain.TaskFeedback{Reason: reason}
	return domain.MarkWrongResult{TaskID: taskID, Status: domain.FollowupMarkedWrong, FeedbackID: "fb_" + taskID}, nil
}

func (s *MockStore) RegenerateTask(taskID string, instruction string) (domain.RegenerateTaskResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.taskIndex(taskID)
	if err != nil {
		return domain.RegenerateTaskResult{}, err
	}

	recommendation := s.tasks[index].Recommendation
	if strings.TrimSpace(instruction) != "" {
		recommendation.Script = recommendation.Script + "（已按反馈调整语气）"
		recommendation.Reasoning = recommendation.Reasoning + " 本次换一种话术保留原客户上下文，仅调整表达方式。"
	}
	s.tasks[index].Recommendation = recommendation
	agentRunID := "run_" + taskID
	s.runs[agentRunID] = domain.AgentRun{
		ID:               agentRunID,
		CustomerID:       s.tasks[index].Customer.ID,
		TaskType:         agent.TaskRegenerateFollowup,
		Status:           "succeeded",
		Model:            agent.ModelMockLocalV1,
		PromptVersion:    agent.PromptRegenerateV1,
		InputSummary:     "基于用户反馈重新生成跟进话术",
		Output:           recommendation,
		ValidationErrors: agent.ValidateRecommendation(recommendation),
		RiskFlags:        recommendation.RiskFlags,
		CreatedAt:        "2026-05-28T10:35:00Z",
		CompletedAt:      "2026-05-28T10:35:00Z",
	}

	return domain.RegenerateTaskResult{
		TaskID:         taskID,
		AgentRunID:     agentRunID,
		Recommendation: recommendation,
	}, nil
}

func (s *MockStore) AgentRun(id string) (domain.AgentRun, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, ok := s.runs[id]
	return run, ok
}

func (s *MockStore) CreateAuditEvent(event domain.AuditEvent) (domain.AuditEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.ID == "" {
		event.ID = fmt.Sprintf("audit_%03d", s.nextID)
		s.nextID++
	}
	if event.CreatedAt == "" {
		event.CreatedAt = "2026-05-28T10:35:00Z"
	}
	if event.Metadata == nil {
		event.Metadata = map[string]any{}
	}
	s.audit = append(s.audit, event)
	return event, nil
}

func (s *MockStore) AuditEventPage(filter AuditEventFilter, page PageRequest) AuditEventPage {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := make([]domain.AuditEvent, 0, len(s.audit))
	for _, event := range s.audit {
		if filter.Action != "" && string(event.Action) != filter.Action {
			continue
		}
		if filter.ActorID != "" && event.Actor.UserID != filter.ActorID {
			continue
		}
		if filter.EntityType != "" && event.EntityType != filter.EntityType {
			continue
		}
		if filter.EntityID != "" && event.EntityID != filter.EntityID {
			continue
		}
		filtered = append(filtered, event)
	}

	return AuditEventPage{Items: paginate(filtered, page), Total: len(filtered)}
}

func (s *MockStore) taskIndex(taskID string) (int, error) {
	for index, task := range s.tasks {
		if task.ID == taskID {
			return index, nil
		}
	}
	return 0, apperror.New("NOT_FOUND", "跟进任务不存在", map[string]any{"task_id": taskID})
}

func newTask(id string, customer domain.Customer, taskType string, generatedAt string, recommendation domain.AgentRecommendation) domain.FollowupTask {
	return domain.FollowupTask{
		ID:             id,
		Customer:       customer,
		Type:           taskType,
		Status:         domain.FollowupPending,
		GeneratedAt:    generatedAt,
		Recommendation: recommendation,
		Feedback:       nil,
	}
}

func MatchCustomer(customer domain.Customer, keyword string, stage string, intent string, ownerID string, risk string) bool {
	if keyword != "" && !strings.Contains(customer.Name, keyword) {
		return false
	}
	if stage != "" && string(customer.Stage) != stage {
		return false
	}
	if intent != "" && string(customer.Intent) != intent {
		return false
	}
	if ownerID != "" && customer.Owner.ID != ownerID {
		return false
	}
	if risk == "1" && len(customer.RiskFlags) == 0 {
		return false
	}
	return true
}

func inferCustomerName(content string) string {
	content = strings.TrimSpace(content)
	fields := strings.Fields(content)
	if len(fields) > 0 {
		first := strings.Trim(fields[0], "：:，,。 ")
		if first != "" {
			return first
		}
	}
	return "新客户"
}

func ownerName(ownerID string) string {
	switch ownerID {
	case "usr_002":
		return "销售B"
	default:
		return "销售A"
	}
}

func paginate[T any](items []T, page PageRequest) []T {
	start := (page.Page - 1) * page.PageSize
	if start >= len(items) {
		return []T{}
	}
	end := start + page.PageSize
	if end > len(items) {
		end = len(items)
	}
	return append([]T(nil), items[start:end]...)
}
