package handler

import (
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
)

type DashboardHandler struct {
	Service *service.QilingService
}

func (h DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, h.Service.DashboardSummary(httpx.ActorFromRequest(r)))
}
