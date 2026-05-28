package http

import (
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	"github.com/heqiucheng/qiling-agent/backend/internal/http/handler"
	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
)

func NewRouter(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /api/health", handler.HealthHandler{Version: "0.1.0", Env: cfg.Env})

	return httpx.WithRequestID(mux)
}
