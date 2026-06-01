package agent

import (
	"context"
	"strings"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/heqiucheng/qiling-agent/backend/internal/integration/llm"
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

type LLMRunner struct {
	client llm.Client
	model  string
}

func NewLLMRunner(client llm.Client, model string) LLMRunner {
	if model == "" {
		model = ModelMockLocalV1
	}
	return LLMRunner{client: client, model: model}
}

func (r LLMRunner) GenerateFollowup(input RunInput) RunResult {
	template, _ := Template(PromptFollowupV1)
	_, _ = r.client.Generate(context.Background(), BuildGenerateRequest(input, template, r.model))
	return NewMockRunner().GenerateFollowup(input)
}

func (r LLMRunner) RegenerateFollowup(input RunInput) RunResult {
	template, _ := Template(PromptRegenerateV1)
	_, _ = r.client.Generate(context.Background(), BuildGenerateRequest(input, template, r.model))
	return NewMockRunner().RegenerateFollowup(input)
}

func BuildGenerateRequest(input RunInput, template PromptTemplate, model string) llm.GenerateRequest {
	return llm.GenerateRequest{
		Model:          model,
		PromptVersion:  template.Version,
		ResponseSchema: template.OutputSchema,
		Messages: []llm.Message{
			{Role: "system", Content: template.SystemPrompt},
			{Role: "user", Content: buildUserPrompt(input, template)},
		},
		Metadata: map[string]any{
			"task_type":     template.TaskType,
			"customer_name": input.CustomerName,
			"owner_id":      input.OwnerID,
		},
	}
}

func (r MockRunner) GenerateFollowup(input RunInput) RunResult {
	now := nonZeroTime(input.Now)
	recommendation := domain.AgentRecommendation{
		CustomerStage:     domain.StagePriceObjection,
		IntentLevel:       domain.IntentHigh,
		MainConcerns:      []string{"price", "outcome"},
		RecommendedAction: "explain value and provide a similar case",
		Script:            "I can first explain the expected value with a similar case, then discuss whether the plan fits your scenario.",
		Reasoning:         "The uploaded conversation shows price and outcome concerns. Establish value before pushing for commitment.",
		RiskFlags:         []string{"avoid direct discount or outcome promises"},
		NextFollowupTime:  now.Add(6 * time.Hour).UTC().Format(time.RFC3339),
	}

	return RunResult{
		TaskType:         TaskGenerateFollowupScript,
		Model:            ModelMockLocalV1,
		PromptVersion:    PromptFollowupV1,
		InputSummary:     summarizeInput(input.RawContent, "generate profile and followup script from uploaded conversation"),
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
		recommendation.Script = recommendation.Script + " Adjusted according to the latest feedback."
		recommendation.Reasoning = recommendation.Reasoning + " The regenerated script preserves customer context and only changes expression."
	}

	return RunResult{
		TaskType:         TaskRegenerateFollowup,
		Model:            ModelMockLocalV1,
		PromptVersion:    PromptRegenerateV1,
		InputSummary:     summarizeInput(input.Instruction, "regenerate followup script from user feedback"),
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

func buildUserPrompt(input RunInput, template PromptTemplate) string {
	parts := []string{
		template.UserPromptHint,
		"Customer name: " + input.CustomerName,
		"Owner ID: " + input.OwnerID,
	}
	if strings.TrimSpace(input.RawContent) != "" {
		parts = append(parts, "Conversation:\n"+strings.TrimSpace(input.RawContent))
	}
	if strings.TrimSpace(input.Instruction) != "" {
		parts = append(parts, "User feedback:\n"+strings.TrimSpace(input.Instruction))
	}
	return strings.Join(parts, "\n\n")
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
