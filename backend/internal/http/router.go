package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	"github.com/heqiucheng/qiling-agent/backend/internal/http/handler"
	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
	"github.com/heqiucheng/qiling-agent/backend/internal/store"
)

func NewRouter(cfg config.Config) http.Handler {
	return NewRouterWithRepository(cfg, store.NewMockStore())
}

func NewRouterWithRepository(cfg config.Config, repository store.Repository) http.Handler {
	return NewRouterWithRepositoryAndLogger(cfg, repository, slog.Default())
}

func NewRouterWithRepositoryAndLogger(cfg config.Config, repository store.Repository, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	qilingService := service.NewQilingService(repository)
	dashboardHandler := handler.DashboardHandler{Service: qilingService}
	customersHandler := handler.CustomersHandler{Service: qilingService}
	followupTasksHandler := handler.FollowupTasksHandler{Service: qilingService}
	uploadsHandler := handler.UploadsHandler{Service: qilingService}
	reviewReportsHandler := handler.ReviewReportsHandler{Service: qilingService}
	auditEventsHandler := handler.AuditEventsHandler{Service: qilingService}
	agentRunsHandler := handler.AgentRunsHandler{Service: qilingService}

	mux.Handle("GET /api/health", handler.HealthHandler{Version: "0.1.0", Env: cfg.Env})
	mux.HandleFunc("GET /api/dashboard/summary", dashboardHandler.Summary)
	mux.HandleFunc("GET /api/customers", customersHandler.List)
	mux.HandleFunc("GET /api/customers/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/conversations") {
			customersHandler.Conversations(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/short-term-memory") {
			customersHandler.ShortTermMemory(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/long-term-memory") {
			customersHandler.LongTermMemory(w, r)
			return
		}
		customersHandler.Detail(w, r)
	})
	mux.HandleFunc("GET /api/followup-tasks", followupTasksHandler.List)
	mux.HandleFunc("GET /api/review-reports/summary", reviewReportsHandler.Summary)
	mux.HandleFunc("GET /api/audit-events", auditEventsHandler.List)
	mux.HandleFunc("GET /api/agent-runs/", agentRunsHandler.Get)
	mux.HandleFunc("POST /api/uploads/conversations", uploadsHandler.CreateConversation)
	mux.HandleFunc("GET /api/uploads/", uploadsHandler.Get)
	mux.HandleFunc("POST /api/uploads/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/confirm") {
			uploadsHandler.Confirm(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("POST /api/customers/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/long-term-memory/facts/") && strings.HasSuffix(r.URL.Path, "/reject") {
			customersHandler.RejectMemoryFact(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("POST /api/followup-tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/copy"):
			followupTasksHandler.Copy(w, r)
		case strings.HasSuffix(r.URL.Path, "/skip"):
			followupTasksHandler.Skip(w, r)
		case strings.HasSuffix(r.URL.Path, "/mark-wrong"):
			followupTasksHandler.MarkWrong(w, r)
		case strings.HasSuffix(r.URL.Path, "/regenerate"):
			followupTasksHandler.Regenerate(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	return httpx.WithRequestID(httpx.WithActor(httpx.WithAccessLog(logger, 500*time.Millisecond, mux)))
}
