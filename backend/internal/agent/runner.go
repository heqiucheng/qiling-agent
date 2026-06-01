package agent

import (
	"context"
	"strings"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/heqiucheng/qiling-agent/backend/internal/integration/llm"
)

type RunInput struct {
	CustomerName           string
	OwnerID                string
	RawContent             string
	MemoryContext          string
	ShortTermMemoryContext string
	LongTermMemoryContext  string
	Instruction            string
	ExistingTask           *domain.FollowupTask
	Now                    time.Time
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
	response, err := r.client.Generate(context.Background(), BuildGenerateRequest(input, template, r.model))
	fallback := NewMockRunner().GenerateFollowup(input)
	if err != nil {
		return fallbackWithErrors(fallback, []string{"llm generate failed: " + err.Error()})
	}
	recommendation, validationErrors := ParseRecommendation(response.Content)
	if len(validationErrors) > 0 {
		return fallbackWithErrors(fallback, validationErrors)
	}
	return RunResult{
		TaskType:         TaskGenerateFollowupScript,
		Model:            responseModel(response.Model, r.model),
		PromptVersion:    PromptFollowupV1,
		InputSummary:     summarizeRunInput(input, input.RawContent, "generate profile and followup script from uploaded conversation"),
		Recommendation:   recommendation,
		ValidationErrors: []string{},
		RiskFlags:        recommendation.RiskFlags,
	}
}

func (r LLMRunner) RegenerateFollowup(input RunInput) RunResult {
	template, _ := Template(PromptRegenerateV1)
	response, err := r.client.Generate(context.Background(), BuildGenerateRequest(input, template, r.model))
	fallback := NewMockRunner().RegenerateFollowup(input)
	if err != nil {
		return fallbackWithErrors(fallback, []string{"llm regenerate failed: " + err.Error()})
	}
	recommendation, validationErrors := ParseRecommendation(response.Content)
	if len(validationErrors) > 0 {
		return fallbackWithErrors(fallback, validationErrors)
	}
	return RunResult{
		TaskType:         TaskRegenerateFollowup,
		Model:            responseModel(response.Model, r.model),
		PromptVersion:    PromptRegenerateV1,
		InputSummary:     summarizeRunInput(input, input.Instruction, "regenerate followup script from user feedback"),
		Recommendation:   recommendation,
		ValidationErrors: []string{},
		RiskFlags:        recommendation.RiskFlags,
	}
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
			"task_type":                     template.TaskType,
			"customer_name":                 input.CustomerName,
			"owner_id":                      input.OwnerID,
			"has_memory_context":            hasAnyMemoryContext(input),
			"has_short_term_memory_context": strings.TrimSpace(shortTermMemoryContext(input)) != "",
			"has_long_term_memory_context":  strings.TrimSpace(input.LongTermMemoryContext) != "",
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
		InputSummary:     summarizeRunInput(input, input.RawContent, "generate profile and followup script from uploaded conversation"),
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
		InputSummary:     summarizeRunInput(input, input.Instruction, "regenerate followup script from user feedback"),
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
	if strings.TrimSpace(shortTermMemoryContext(input)) != "" {
		parts = append(parts, "Short-term memory:\n"+strings.TrimSpace(shortTermMemoryContext(input)))
	}
	if strings.TrimSpace(input.LongTermMemoryContext) != "" {
		parts = append(parts, "Long-term memory:\n"+strings.TrimSpace(input.LongTermMemoryContext))
	}
	if strings.TrimSpace(input.Instruction) != "" {
		parts = append(parts, "User feedback:\n"+strings.TrimSpace(input.Instruction))
	}
	return strings.Join(parts, "\n\n")
}

func summarizeRunInput(input RunInput, primary string, fallback string) string {
	primary = strings.TrimSpace(primary)
	memory := combinedMemoryContext(input)
	if strings.TrimSpace(memory) == "" {
		return summarizeInput(primary, fallback)
	}
	header := "memory context:"
	if strings.TrimSpace(shortTermMemoryContext(input)) != "" {
		header += " short-term: present;"
	}
	if strings.TrimSpace(input.LongTermMemoryContext) != "" {
		header += " long-term: present;"
	}
	if primary == "" {
		return summarizeInput(header+"\n"+memory, fallback)
	}
	return summarizeInput(header+"\n"+memory+"\ncurrent input: "+primary, fallback)
}

func hasAnyMemoryContext(input RunInput) bool {
	return strings.TrimSpace(combinedMemoryContext(input)) != ""
}

func shortTermMemoryContext(input RunInput) string {
	if strings.TrimSpace(input.ShortTermMemoryContext) != "" {
		return input.ShortTermMemoryContext
	}
	return input.MemoryContext
}

func combinedMemoryContext(input RunInput) string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(shortTermMemoryContext(input)) != "" {
		parts = append(parts, "short-term: "+strings.TrimSpace(shortTermMemoryContext(input)))
	}
	if strings.TrimSpace(input.LongTermMemoryContext) != "" {
		parts = append(parts, "long-term: "+strings.TrimSpace(input.LongTermMemoryContext))
	}
	return strings.Join(parts, "\n")
}

func summarizeInput(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	const maxRunes = 240
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

func responseModel(responseModel string, fallback string) string {
	if strings.TrimSpace(responseModel) == "" {
		return fallback
	}
	return responseModel
}
