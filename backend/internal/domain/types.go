package domain

type AuditAction string

const (
	AuditUploadConversationCreated AuditAction = "upload.conversation.created"
	AuditUploadConfirmed           AuditAction = "upload.confirmed"
	AuditFollowupTaskCopied        AuditAction = "followup_task.copied"
	AuditFollowupTaskSkipped       AuditAction = "followup_task.skipped"
	AuditFollowupTaskMarkedWrong   AuditAction = "followup_task.marked_wrong"
	AuditFollowupTaskRegenerated   AuditAction = "followup_task.regenerated"
	AuditMemoryFactRejected        AuditAction = "memory_fact.rejected"
	AuditMemoryFactCorrected       AuditAction = "memory_fact.corrected"
)

type AuditEvent struct {
	ID          string         `json:"id"`
	Action      AuditAction    `json:"action"`
	Actor       Actor          `json:"actor"`
	RequestID   string         `json:"request_id"`
	EntityType  string         `json:"entity_type"`
	EntityID    string         `json:"entity_id"`
	RelatedType string         `json:"related_type,omitempty"`
	RelatedID   string         `json:"related_id,omitempty"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   string         `json:"created_at"`
}

type IntentLevel string

const (
	IntentHigh   IntentLevel = "high"
	IntentMedium IntentLevel = "medium"
	IntentLow    IntentLevel = "low"
	IntentRisk   IntentLevel = "risk"
)

type CustomerStage string

const (
	StageNewLead           CustomerStage = "new_lead"
	StageOpened            CustomerStage = "opened"
	StageNeedsDiscovery    CustomerStage = "needs_discovery"
	StageProductInterested CustomerStage = "product_interested"
	StagePriceObjection    CustomerStage = "price_objection"
	StageHighIntent        CustomerStage = "high_intent"
	StageClosing           CustomerStage = "closing"
	StageWon               CustomerStage = "won"
	StageSilent            CustomerStage = "silent"
	StageChurnRisk         CustomerStage = "churn_risk"
)

type FollowupTaskStatus string

const (
	FollowupPending     FollowupTaskStatus = "pending"
	FollowupCopied      FollowupTaskStatus = "copied"
	FollowupSkipped     FollowupTaskStatus = "skipped"
	FollowupMarkedWrong FollowupTaskStatus = "marked_wrong"
)

type UploadStatus string

const (
	UploadUploaded          UploadStatus = "uploaded"
	UploadParsed            UploadStatus = "parsed"
	UploadNeedsConfirmation UploadStatus = "needs_confirmation"
	UploadConfirmed         UploadStatus = "confirmed"
	UploadFailed            UploadStatus = "failed"
)

type Actor struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type Owner struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Customer struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Source         string        `json:"source"`
	Owner          Owner         `json:"owner"`
	Stage          CustomerStage `json:"stage"`
	Intent         IntentLevel   `json:"intent"`
	Concerns       []string      `json:"concerns"`
	Tags           []string      `json:"tags"`
	ProfileSummary string        `json:"profile_summary"`
	LastContactAt  string        `json:"last_contact_at"`
	PendingTasks   int           `json:"pending_tasks"`
	RiskFlags      []string      `json:"risk_flags"`
}

type AgentRecommendation struct {
	CustomerStage     CustomerStage `json:"customer_stage"`
	IntentLevel       IntentLevel   `json:"intent_level"`
	MainConcerns      []string      `json:"main_concerns"`
	RecommendedAction string        `json:"recommended_action"`
	Script            string        `json:"script"`
	Reasoning         string        `json:"reasoning"`
	RiskFlags         []string      `json:"risk_flags"`
	NextFollowupTime  string        `json:"next_followup_time,omitempty"`
}

type FollowupTask struct {
	ID             string              `json:"id"`
	Customer       Customer            `json:"customer"`
	Type           string              `json:"type"`
	Status         FollowupTaskStatus  `json:"status"`
	GeneratedAt    string              `json:"generated_at"`
	Recommendation AgentRecommendation `json:"recommendation"`
	Feedback       *TaskFeedback       `json:"feedback"`
}

type TaskFeedback struct {
	Reason string `json:"reason"`
}

type ParsedCustomer struct {
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
}

type ConversationMessage struct {
	ID         string `json:"id"`
	SenderType string `json:"sender_type"`
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	SentAt     string `json:"sent_at"`
}

type UploadRecord struct {
	ID             string                `json:"id"`
	Status         UploadStatus          `json:"status"`
	SourceType     string                `json:"source_type"`
	ParsedCustomer ParsedCustomer        `json:"parsed_customer"`
	Messages       []ConversationMessage `json:"messages"`
	Warnings       []string              `json:"warnings"`
	CreatedAt      string                `json:"created_at"`
}

type UploadConversationResult struct {
	UploadID       string         `json:"upload_id"`
	Status         UploadStatus   `json:"status"`
	ParsedCustomer ParsedCustomer `json:"parsed_customer"`
	MessageCount   int            `json:"message_count"`
	Warnings       []string       `json:"warnings"`
	NextAction     string         `json:"next_action"`
}

type ConfirmUploadResult struct {
	CustomerID     string       `json:"customer_id"`
	ConversationID string       `json:"conversation_id"`
	AgentRunID     string       `json:"agent_run_id"`
	FollowupTaskID string       `json:"followup_task_id"`
	Status         UploadStatus `json:"status"`
}

type TaskCopyResult struct {
	TaskID   string             `json:"task_id"`
	Status   FollowupTaskStatus `json:"status"`
	CopiedAt string             `json:"copied_at"`
}

type TaskStatusResult struct {
	TaskID string             `json:"task_id"`
	Status FollowupTaskStatus `json:"status"`
}

type MarkWrongResult struct {
	TaskID     string             `json:"task_id"`
	Status     FollowupTaskStatus `json:"status"`
	FeedbackID string             `json:"feedback_id"`
}

type RegenerateTaskResult struct {
	TaskID         string              `json:"task_id"`
	AgentRunID     string              `json:"agent_run_id"`
	Recommendation AgentRecommendation `json:"recommendation"`
}

type AgentRun struct {
	ID               string              `json:"id"`
	CustomerID       string              `json:"customer_id,omitempty"`
	TaskType         string              `json:"task_type"`
	Status           string              `json:"status"`
	Model            string              `json:"model"`
	PromptVersion    string              `json:"prompt_version"`
	InputSummary     string              `json:"input_summary"`
	Output           AgentRecommendation `json:"output"`
	ValidationErrors []string            `json:"validation_errors"`
	RiskFlags        []string            `json:"risk_flags"`
	CreatedAt        string              `json:"created_at"`
	CompletedAt      string              `json:"completed_at"`
}

type Metric struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value any    `json:"value"`
	Hint  string `json:"hint"`
}

type DailyReview struct {
	Summary     string   `json:"summary"`
	Suggestions []string `json:"suggestions"`
}

type DashboardSummary struct {
	Metrics             []Metric       `json:"metrics"`
	PriorityTasks       []FollowupTask `json:"priority_tasks"`
	HighIntentCustomers []Customer     `json:"high_intent_customers"`
	SilentCustomers     []Customer     `json:"silent_customers"`
	RiskCustomers       []Customer     `json:"risk_customers"`
	DailyReview         DailyReview    `json:"daily_review"`
}

type CustomerDetail struct {
	Customer             Customer            `json:"customer"`
	LatestRecommendation AgentRecommendation `json:"latest_recommendation"`
	ProfileEvidence      []string            `json:"profile_evidence"`
	RecentTasks          []FollowupTask      `json:"recent_tasks"`
	RecentAgentRuns      []AgentRunSummary   `json:"recent_agent_runs"`
}

type AgentRunSummary struct {
	ID            string   `json:"id"`
	Status        string   `json:"status"`
	TaskType      string   `json:"task_type"`
	Model         string   `json:"model"`
	PromptVersion string   `json:"prompt_version"`
	InputSummary  string   `json:"input_summary"`
	RiskFlags     []string `json:"risk_flags"`
	CreatedAt     string   `json:"created_at"`
	CompletedAt   string   `json:"completed_at"`
}

type MemoryItem struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Summary   string `json:"summary"`
	Evidence  string `json:"evidence,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type ShortTermMemory struct {
	Customer               Customer     `json:"customer"`
	ConversationHighlights []MemoryItem `json:"conversation_highlights"`
	RecentTasks            []MemoryItem `json:"recent_tasks"`
	RecentAgentRuns        []MemoryItem `json:"recent_agent_runs"`
	RecentEvents           []MemoryItem `json:"recent_events"`
	PromptContext          string       `json:"prompt_context"`
	BuiltAt                string       `json:"built_at"`
}

type MemoryFactStatus string

const (
	MemoryFactActive     MemoryFactStatus = "active"
	MemoryFactSuperseded MemoryFactStatus = "superseded"
	MemoryFactRejected   MemoryFactStatus = "rejected"
)

type LongTermMemoryFact struct {
	ID         string           `json:"id"`
	CustomerID string           `json:"customer_id"`
	Category   string           `json:"category"`
	Key        string           `json:"key"`
	Value      string           `json:"value"`
	Confidence float64          `json:"confidence"`
	SourceType string           `json:"source_type"`
	SourceID   string           `json:"source_id"`
	Status     MemoryFactStatus `json:"status"`
	CreatedAt  string           `json:"created_at"`
	UpdatedAt  string           `json:"updated_at"`
}

type LongTermMemory struct {
	Customer      Customer             `json:"customer"`
	Facts         []LongTermMemoryFact `json:"facts"`
	PromptContext string               `json:"prompt_context"`
	BuiltAt       string               `json:"built_at"`
}

type MemoryFactStatusResult struct {
	FactID string           `json:"fact_id"`
	Status MemoryFactStatus `json:"status"`
}

type MemoryFactCorrectionResult struct {
	OldFactID string             `json:"old_fact_id"`
	OldStatus MemoryFactStatus   `json:"old_status"`
	NewFact   LongTermMemoryFact `json:"new_fact"`
}

type ReviewInsight struct {
	Title      string `json:"title"`
	Evidence   string `json:"evidence"`
	Suggestion string `json:"suggestion"`
}

type StageDistribution struct {
	Stage CustomerStage `json:"stage"`
	Count int           `json:"count"`
}

type ReviewSummary struct {
	Metrics              []Metric            `json:"metrics"`
	StageDistribution    []StageDistribution `json:"stage_distribution"`
	OpportunityCustomers []Customer          `json:"opportunity_customers"`
	RiskCustomers        []Customer          `json:"risk_customers"`
	Insights             []ReviewInsight     `json:"insights"`
	SampleWarning        *string             `json:"sample_warning"`
}

type ReportType string

const (
	ReportCustomerIntent ReportType = "customer_intent"
)

type Report struct {
	ID          string             `json:"id"`
	Type        ReportType         `json:"type"`
	Title       string             `json:"title"`
	RangeLabel  string             `json:"range_label"`
	OwnerID     string             `json:"owner_id"`
	OwnerRole   string             `json:"owner_role"`
	Summary     string             `json:"summary"`
	Metrics     []Metric           `json:"metrics"`
	Sections    []ReportSection    `json:"sections"`
	ActionItems []ReportActionItem `json:"action_items"`
	Markdown    string             `json:"markdown"`
	GeneratedAt string             `json:"generated_at"`
}

type ReportSummary struct {
	ID              string     `json:"id"`
	Type            ReportType `json:"type"`
	Title           string     `json:"title"`
	RangeLabel      string     `json:"range_label"`
	Summary         string     `json:"summary"`
	OwnerID         string     `json:"owner_id"`
	OwnerRole       string     `json:"owner_role"`
	MetricCount     int        `json:"metric_count"`
	SectionCount    int        `json:"section_count"`
	ActionItemCount int        `json:"action_item_count"`
	GeneratedAt     string     `json:"generated_at"`
}

type ReportSection struct {
	Title    string               `json:"title"`
	Summary  string               `json:"summary"`
	Items    []ReportCustomerItem `json:"items"`
	Evidence []string             `json:"evidence"`
}

type ReportCustomerItem struct {
	CustomerID        string   `json:"customer_id"`
	CustomerName      string   `json:"customer_name"`
	Stage             string   `json:"stage"`
	Intent            string   `json:"intent"`
	RecommendedAction string   `json:"recommended_action"`
	Script            string   `json:"script"`
	Reasoning         string   `json:"reasoning"`
	Evidence          []string `json:"evidence"`
}

type ReportActionItem struct {
	CustomerID   string `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	Priority     string `json:"priority"`
	Action       string `json:"action"`
	DueHint      string `json:"due_hint"`
}

type ReportExportTaskStatus string

const (
	ReportExportQueued    ReportExportTaskStatus = "queued"
	ReportExportCompleted ReportExportTaskStatus = "completed"
	ReportExportFailed    ReportExportTaskStatus = "failed"
)

type ReportExportTask struct {
	ID          string                 `json:"id"`
	ReportID    string                 `json:"report_id"`
	Format      string                 `json:"format"`
	Status      ReportExportTaskStatus `json:"status"`
	OwnerID     string                 `json:"owner_id"`
	OwnerRole   string                 `json:"owner_role"`
	Filename    string                 `json:"filename"`
	ContentType string                 `json:"content_type"`
	SizeBytes   int                    `json:"size_bytes"`
	Error       string                 `json:"error"`
	CreatedAt   string                 `json:"created_at"`
	CompletedAt string                 `json:"completed_at"`
}
