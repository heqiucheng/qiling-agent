package agent

import (
	"testing"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

func TestMockRunnerGenerateFollowupReturnsValidStructuredOutput(t *testing.T) {
	result := NewMockRunner().GenerateFollowup(RunInput{
		CustomerName: "王女士",
		RawContent:   "王女士 10:20 价格和效果需要再看看",
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

func TestValidateRecommendationRejectsMissingRequiredFields(t *testing.T) {
	errors := ValidateRecommendation(domain.AgentRecommendation{})

	if len(errors) == 0 {
		t.Fatal("expected validation errors")
	}
}
