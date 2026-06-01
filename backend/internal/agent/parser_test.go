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

func TestParseRecommendationRejectsMissingRequiredFields(t *testing.T) {
	_, errors := ParseRecommendation(`{"script": "hello"}`)

	if len(errors) == 0 {
		t.Fatal("expected validation errors")
	}
}
