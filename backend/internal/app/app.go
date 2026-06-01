package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/agent"
	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	"github.com/heqiucheng/qiling-agent/backend/internal/db"
	httpx "github.com/heqiucheng/qiling-agent/backend/internal/http"
	"github.com/heqiucheng/qiling-agent/backend/internal/integration/llm"
	"github.com/heqiucheng/qiling-agent/backend/internal/store"
)

func NewHTTPHandler(cfg config.Config, logger *slog.Logger) (http.Handler, error) {
	repository, err := buildRepository(cfg)
	if err != nil {
		return nil, err
	}
	runner, err := buildAgentRunner(cfg)
	if err != nil {
		return nil, err
	}

	return httpx.NewRouterWithRepositoryAgentAndLogger(cfg, repository, runner, logger), nil
}

func buildRepository(cfg config.Config) (store.Repository, error) {
	if cfg.StoreDriver != "mysql" {
		return store.NewMockStore(), nil
	}

	database, err := db.OpenMySQL(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return store.NewMySQLRepository(database), nil
}

func buildAgentRunner(cfg config.Config) (agent.Runner, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.LLMProvider)) {
	case "", "mock":
		return agent.NewMockRunner(), nil
	case "openai_compatible", "codex_proxy":
		apiKey := strings.TrimSpace(os.Getenv(cfg.LLMAPIKeyEnv))
		client, err := llm.NewOpenAICompatibleClient(llm.OpenAICompatibleConfig{
			BaseURL: cfg.LLMBaseURL,
			APIKey:  apiKey,
			Timeout: time.Duration(cfg.LLMTimeoutSec) * time.Second,
		})
		if err != nil {
			return nil, err
		}
		return agent.NewLLMRunner(client, cfg.LLMModel), nil
	default:
		return nil, fmt.Errorf("unsupported llm provider %q", cfg.LLMProvider)
	}
}
