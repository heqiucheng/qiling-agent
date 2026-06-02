package store

import "github.com/heqiucheng/qiling-agent/backend/internal/domain"

type PageRequest struct {
	Page     int
	PageSize int
}

type CustomerFilter struct {
	Keyword string
	Stage   string
	Intent  string
	OwnerID string
	Risk    string
}

type CustomerPage struct {
	Items []domain.Customer
	Total int
}

type FollowupTaskFilter struct {
	Status  string
	Intent  string
	OwnerID string
}

type FollowupTaskPage struct {
	Items []domain.FollowupTask
	Total int
}

type ConversationMessagePage struct {
	Items []domain.ConversationMessage
	Total int
}

type ConfirmUploadAgentRun struct {
	TaskType         string
	Model            string
	PromptVersion    string
	InputSummary     string
	Recommendation   domain.AgentRecommendation
	ValidationErrors []string
	RiskFlags        []string
}

type AuditEventFilter struct {
	Action     string
	ActorID    string
	EntityType string
	EntityID   string
}

type AuditEventPage struct {
	Items []domain.AuditEvent
	Total int
}

type AgentRunPage struct {
	Items []domain.AgentRun
	Total int
}

type LongTermMemoryFactPage struct {
	Items []domain.LongTermMemoryFact
	Total int
}

type ReportPage struct {
	Items []domain.ReportSummary
	Total int
}

func reportSummary(report domain.Report) domain.ReportSummary {
	return domain.ReportSummary{
		ID:              report.ID,
		Type:            report.Type,
		Title:           report.Title,
		RangeLabel:      report.RangeLabel,
		Summary:         report.Summary,
		OwnerID:         report.OwnerID,
		OwnerRole:       report.OwnerRole,
		MetricCount:     len(report.Metrics),
		SectionCount:    len(report.Sections),
		ActionItemCount: len(report.ActionItems),
		GeneratedAt:     report.GeneratedAt,
	}
}

type Repository interface {
	Customers() []domain.Customer
	CustomerPage(filter CustomerFilter, page PageRequest) CustomerPage
	Customer(id string) (domain.Customer, bool)
	FollowupTasks() []domain.FollowupTask
	FollowupTaskPage(filter FollowupTaskFilter, page PageRequest) FollowupTaskPage
	FollowupTasksByCustomer(customerID string) []domain.FollowupTask
	ConversationMessages(customerID string) []domain.ConversationMessage
	ConversationMessagePage(customerID string, page PageRequest) ConversationMessagePage
	CreateUpload(sourceType string, content string, ownerID string) (domain.UploadRecord, error)
	Upload(id string) (domain.UploadRecord, bool)
	ConfirmUpload(uploadID string, customerName string, ownerID string, agentRun ConfirmUploadAgentRun) (domain.ConfirmUploadResult, error)
	CopyTask(taskID string, copiedAt string) (domain.TaskCopyResult, error)
	SkipTask(taskID string, reason string) (domain.TaskStatusResult, error)
	MarkTaskWrong(taskID string, reason string) (domain.MarkWrongResult, error)
	RegenerateTask(taskID string, instruction string) (domain.RegenerateTaskResult, error)
	AgentRun(id string) (domain.AgentRun, bool)
	AgentRunsByCustomer(customerID string, page PageRequest) AgentRunPage
	LongTermMemoryFacts(customerID string, page PageRequest) LongTermMemoryFactPage
	UpsertLongTermMemoryFact(fact domain.LongTermMemoryFact) (domain.LongTermMemoryFact, error)
	UpdateLongTermMemoryFactStatus(customerID string, factID string, status domain.MemoryFactStatus) (domain.MemoryFactStatusResult, error)
	CorrectLongTermMemoryFact(customerID string, factID string, corrected domain.LongTermMemoryFact) (domain.MemoryFactCorrectionResult, error)
	SaveReport(report domain.Report) (domain.Report, error)
	Report(id string) (domain.Report, bool)
	ReportPage(ownerID string, ownerRole string, page PageRequest) ReportPage
	CreateAuditEvent(event domain.AuditEvent) (domain.AuditEvent, error)
	AuditEventPage(filter AuditEventFilter, page PageRequest) AuditEventPage
}
