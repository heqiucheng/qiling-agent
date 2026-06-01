package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

func ParseRecommendation(raw string) (domain.AgentRecommendation, []string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return domain.AgentRecommendation{}, []string{"llm output is empty"}
	}

	recommendation, err := decodeRecommendation([]byte(raw))
	if err != nil {
		return domain.AgentRecommendation{}, []string{"llm output is not valid json: " + err.Error()}
	}

	return recommendation, ValidateRecommendation(recommendation)
}

func decodeRecommendation(raw []byte) (domain.AgentRecommendation, error) {
	var payload struct {
		CustomerStage     domain.CustomerStage `json:"customer_stage"`
		IntentLevel       domain.IntentLevel   `json:"intent_level"`
		MainConcerns      []string             `json:"main_concerns"`
		RecommendedAction string               `json:"recommended_action"`
		Script            string               `json:"script"`
		Reasoning         any                  `json:"reasoning"`
		RiskFlags         []string             `json:"risk_flags"`
		NextFollowupTime  string               `json:"next_followup_time"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return domain.AgentRecommendation{}, err
	}
	return domain.AgentRecommendation{
		CustomerStage:     payload.CustomerStage,
		IntentLevel:       payload.IntentLevel,
		MainConcerns:      payload.MainConcerns,
		RecommendedAction: payload.RecommendedAction,
		Script:            payload.Script,
		Reasoning:         normalizeReasoning(payload.Reasoning),
		RiskFlags:         payload.RiskFlags,
		NextFollowupTime:  payload.NextFollowupTime,
	}, nil
}

func normalizeReasoning(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(fmt.Sprint(item))
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, " ")
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func fallbackWithErrors(fallback RunResult, errors []string) RunResult {
	fallback.ValidationErrors = append([]string{}, errors...)
	if len(fallback.ValidationErrors) == 0 {
		fallback.ValidationErrors = []string{"llm output fallback used"}
	}
	return fallback
}
