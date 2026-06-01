package handler

import (
	"net/http"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
)

type AgentRunsHandler struct {
	Service *service.QilingService
}

func (h AgentRunsHandler) Get(w http.ResponseWriter, r *http.Request) {
	runID := strings.TrimPrefix(r.URL.Path, "/api/agent-runs/")
	result, err := h.Service.AgentRun(runID)
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}
