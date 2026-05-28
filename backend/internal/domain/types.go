package domain

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
