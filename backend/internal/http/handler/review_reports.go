package handler

import (
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
)

type ReviewReportsHandler struct {
	Service *service.QilingService
}

func (h ReviewReportsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, h.Service.ReviewSummary(httpx.ActorFromRequest(r)))
}
