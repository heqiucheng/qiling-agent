package app

import (
	"testing"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
)

func TestBuildAgentRunnerUsesMockByDefault(t *testing.T) {
	runner, err := buildAgentRunner(config.Config{LLMProvider: "mock", LLMModel: "mock-local-v1"})
	if err != nil {
		t.Fatalf("build runner: %v", err)
	}
	if runner == nil {
		t.Fatal("expected runner")
	}
}

func TestBuildAgentRunnerRequiresConfiguredAPIKey(t *testing.T) {
	t.Setenv("AICODEMIRROR_API_KEY", "")
	_, err := buildAgentRunner(config.Config{
		LLMProvider:   "openai_compatible",
		LLMModel:      "gpt-5.4",
		LLMBaseURL:    "https://api.claudecode.net.cn/api/codex/backend-api/codex",
		LLMAPIKeyEnv:  "AICODEMIRROR_API_KEY",
		LLMTimeoutSec: 45,
	})
	if err == nil {
		t.Fatal("expected missing api key error")
	}
}
