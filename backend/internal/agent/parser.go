package agent

import (
	"encoding/json"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

func ParseRecommendation(raw string) (domain.AgentRecommendation, []string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return domain.AgentRecommendation{}, []string{"llm output is empty"}
	}

	var recommendation domain.AgentRecommendation
	if err := json.Unmarshal([]byte(raw), &recommendation); err != nil {
		return domain.AgentRecommendation{}, []string{"llm output is not valid json: " + err.Error()}
	}

	return recommendation, ValidateRecommendation(recommendation)
}

func fallbackWithErrors(fallback RunResult, errors []string) RunResult {
	fallback.ValidationErrors = append([]string{}, errors...)
	if len(fallback.ValidationErrors) == 0 {
		fallback.ValidationErrors = []string{"llm output fallback used"}
	}
	return fallback
}
