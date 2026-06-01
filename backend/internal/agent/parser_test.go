package agent

import "testing"

func TestParseRecommendationAcceptsValidJSON(t *testing.T) {
	recommendation, errors := ParseRecommendation(`{
		"customer_stage": "price_objection",
		"intent_level": "high",
		"main_concerns": ["price"],
		"recommended_action": "explain value",
		"script": "Use a similar case.",
		"reasoning": "Customer asked about price.",
		"risk_flags": ["avoid promises"]
	}`)

	if len(errors) != 0 {
		t.Fatalf("expected no errors, got %#v", errors)
	}
	if recommendation.Script == "" {
		t.Fatal("expected parsed script")
	}
}

func TestParseRecommendationRejectsInvalidJSON(t *testing.T) {
	_, errors := ParseRecommendation(`{"script":`)

	if len(errors) == 0 {
		t.Fatal("expected invalid json error")
	}
}

func TestParseRecommendationNormalizesReasoningArray(t *testing.T) {
	recommendation, errors := ParseRecommendation(`{
		"customer_stage": "needs_discovery",
		"intent_level": "medium",
		"main_concerns": ["pending_confirmation"],
		"recommended_action": "send a light reminder",
		"script": "Please confirm when convenient.",
		"reasoning": ["The chat is short.", "The customer promised to confirm later."],
		"risk_flags": ["avoid repeated urging"]
	}`)

	if len(errors) != 0 {
		t.Fatalf("expected no errors, got %#v", errors)
	}
	if recommendation.Reasoning != "The chat is short. The customer promised to confirm later." {
		t.Fatalf("expected normalized reasoning, got %q", recommendation.Reasoning)
	}
}

func TestParseRecommendationRejectsMissingRequiredFields(t *testing.T) {
	_, errors := ParseRecommendation(`{"script": "hello"}`)

	if len(errors) == 0 {
		t.Fatal("expected validation errors")
	}
}
