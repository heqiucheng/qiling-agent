package agent

import (
	"errors"
	"testing"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
	"github.com/heqiucheng/qiling-agent/backend/internal/integration/llm"
)

func TestMockRunnerGenerateFollowupReturnsValidStructuredOutput(t *testing.T) {
	result := NewMockRunner().GenerateFollowup(RunInput{
		CustomerName: "Wang",
		RawContent:   "Wang 10:20 price and outcome need another look",
		Now:          time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
	})

	if result.TaskType != TaskGenerateFollowupScript {
		t.Fatalf("expected task type %s, got %s", TaskGenerateFollowupScript, result.TaskType)
	}
	if result.PromptVersion != PromptFollowupV1 {
		t.Fatalf("expected prompt version %s, got %s", PromptFollowupV1, result.PromptVersion)
	}
	if len(result.ValidationErrors) != 0 {
		t.Fatalf("expected no validation errors, got %#v", result.ValidationErrors)
	}
	if result.Recommendation.IntentLevel != domain.IntentHigh {
		t.Fatalf("expected high intent, got %s", result.Recommendation.IntentLevel)
	}
}

func TestBuildGenerateRequestIncludesPromptSchemaAndContext(t *testing.T) {
	template, ok := Template(PromptFollowupV1)
	if !ok {
		t.Fatal("expected followup template")
	}

	request := BuildGenerateRequest(RunInput{
		CustomerName: "Wang",
		OwnerID:      "usr_001",
		RawContent:   "price concern",
	}, template, ModelMockLocalV1)

	if request.Model != ModelMockLocalV1 {
		t.Fatalf("expected model %s, got %s", ModelMockLocalV1, request.Model)
	}
	if request.PromptVersion != PromptFollowupV1 {
		t.Fatalf("expected prompt version %s, got %s", PromptFollowupV1, request.PromptVersion)
	}
	if request.ResponseSchema == "" {
		t.Fatal("expected response schema")
	}
	if len(request.Messages) != 2 {
		t.Fatalf("expected system and user messages, got %d", len(request.Messages))
	}
}

func TestLLMRunnerUsesValidModelOutput(t *testing.T) {
	runner := NewLLMRunner(llm.MockClient{Response: llm.GenerateResponse{
		Model: "llm-test",
		Content: `{
			"customer_stage": "high_intent",
			"intent_level": "high",
			"main_concerns": ["delivery"],
			"recommended_action": "confirm delivery plan",
			"script": "Let's align the delivery plan.",
			"reasoning": "Customer is evaluating delivery timing.",
			"risk_flags": ["do not overpromise delivery"]
		}`,
		FinishReason: "stop",
	}}, "llm-test")

	result := runner.GenerateFollowup(RunInput{RawContent: "delivery timing"})

	if len(result.ValidationErrors) != 0 {
		t.Fatalf("expected no validation errors, got %#v", result.ValidationErrors)
	}
	if result.Model != "llm-test" {
		t.Fatalf("expected llm-test model, got %s", result.Model)
	}
	if result.Recommendation.Script != "Let's align the delivery plan." {
		t.Fatalf("expected model script, got %s", result.Recommendation.Script)
	}
}

func TestLLMRunnerFallsBackOnInvalidJSON(t *testing.T) {
	runner := NewLLMRunner(llm.MockClient{Response: llm.GenerateResponse{
		Model:   "llm-test",
		Content: `{"script":`,
	}}, "llm-test")

	result := runner.GenerateFollowup(RunInput{RawContent: "price"})

	if len(result.ValidationErrors) == 0 {
		t.Fatal("expected fallback validation errors")
	}
	if result.Recommendation.Script == "" {
		t.Fatal("expected fallback recommendation")
	}
}

func TestLLMRunnerFallsBackOnClientError(t *testing.T) {
	runner := NewLLMRunner(llm.MockClient{Err: errors.New("timeout")}, "llm-test")

	result := runner.GenerateFollowup(RunInput{RawContent: "price"})

	if len(result.ValidationErrors) == 0 {
		t.Fatal("expected fallback error")
	}
	if result.Recommendation.Script == "" {
		t.Fatal("expected fallback recommendation")
	}
}

func TestValidateRecommendationRejectsMissingRequiredFields(t *testing.T) {
	errors := ValidateRecommendation(domain.AgentRecommendation{})

	if len(errors) == 0 {
		t.Fatal("expected validation errors")
	}
}
