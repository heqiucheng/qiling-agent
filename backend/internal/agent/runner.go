package agent

import (
	"strings"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

type RunInput struct {
	CustomerName string
	OwnerID      string
	RawContent   string
	Instruction  string
	ExistingTask *domain.FollowupTask
	Now          time.Time
}

type RunResult struct {
	TaskType         string
	Model            string
	PromptVersion    string
	InputSummary     string
	Recommendation   domain.AgentRecommendation
	ValidationErrors []string
	RiskFlags        []string
}

type Runner interface {
	GenerateFollowup(input RunInput) RunResult
	RegenerateFollowup(input RunInput) RunResult
}

type MockRunner struct{}

func NewMockRunner() MockRunner {
	return MockRunner{}
}

func (r MockRunner) GenerateFollowup(input RunInput) RunResult {
	now := nonZeroTime(input.Now)
	recommendation := domain.AgentRecommendation{
		CustomerStage:     domain.StagePriceObjection,
		IntentLevel:       domain.IntentHigh,
		MainConcerns:      []string{"价格", "效果"},
		RecommendedAction: "解释价值并提供案例",
		Script:            "您好，您刚才提到价格和效果，我建议先结合您的使用场景看投入产出，我可以给您整理一个接近情况的案例。",
		Reasoning:         "上传内容显示客户关注价格和效果，需要先建立价值感再推动下一步。",
		RiskFlags:         []string{"避免直接承诺优惠或效果"},
		NextFollowupTime:  now.Add(6 * time.Hour).UTC().Format(time.RFC3339),
	}

	return RunResult{
		TaskType:         TaskGenerateFollowupScript,
		Model:            ModelMockLocalV1,
		PromptVersion:    PromptFollowupV1,
		InputSummary:     summarizeInput(input.RawContent, "上传聊天记录生成客户画像和跟进话术"),
		Recommendation:   recommendation,
		ValidationErrors: ValidateRecommendation(recommendation),
		RiskFlags:        recommendation.RiskFlags,
	}
}

func (r MockRunner) RegenerateFollowup(input RunInput) RunResult {
	recommendation := domain.AgentRecommendation{}
	if input.ExistingTask != nil {
		recommendation = input.ExistingTask.Recommendation
	}
	if strings.TrimSpace(input.Instruction) != "" {
		recommendation.Script = recommendation.Script + "（已按反馈调整语气）"
		recommendation.Reasoning = recommendation.Reasoning + " 本次换一种话术保留原客户上下文，仅调整表达方式。"
	}

	return RunResult{
		TaskType:         TaskRegenerateFollowup,
		Model:            ModelMockLocalV1,
		PromptVersion:    PromptRegenerateV1,
		InputSummary:     summarizeInput(input.Instruction, "基于用户反馈重新生成跟进话术"),
		Recommendation:   recommendation,
		ValidationErrors: ValidateRecommendation(recommendation),
		RiskFlags:        recommendation.RiskFlags,
	}
}

func ValidateRecommendation(recommendation domain.AgentRecommendation) []string {
	errors := make([]string, 0)
	if recommendation.CustomerStage == "" {
		errors = append(errors, "customer_stage is required")
	}
	if recommendation.IntentLevel == "" {
		errors = append(errors, "intent_level is required")
	}
	if strings.TrimSpace(recommendation.RecommendedAction) == "" {
		errors = append(errors, "recommended_action is required")
	}
	if strings.TrimSpace(recommendation.Script) == "" {
		errors = append(errors, "script is required")
	}
	if strings.TrimSpace(recommendation.Reasoning) == "" {
		errors = append(errors, "reasoning is required")
	}
	return errors
}

func summarizeInput(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	const maxRunes = 120
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes]) + "..."
}

func nonZeroTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}
