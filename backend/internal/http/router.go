package http

import (
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	"github.com/heqiucheng/qiling-agent/backend/internal/http/handler"
	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
	"github.com/heqiucheng/qiling-agent/backend/internal/store"
)

func NewRouter(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	qilingService := service.NewQilingService(store.NewMockStore())
	dashboardHandler := handler.DashboardHandler{Service: qilingService}
	customersHandler := handler.CustomersHandler{Service: qilingService}
	followupTasksHandler := handler.FollowupTasksHandler{Service: qilingService}

	mux.Handle("GET /api/health", handler.HealthHandler{Version: "0.1.0", Env: cfg.Env})
	mux.HandleFunc("GET /api/dashboard/summary", dashboardHandler.Summary)
	mux.HandleFunc("GET /api/customers", customersHandler.List)
	mux.HandleFunc("GET /api/followup-tasks", followupTasksHandler.List)

	return httpx.WithRequestID(mux)
}
