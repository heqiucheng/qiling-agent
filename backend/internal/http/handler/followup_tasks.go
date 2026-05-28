package handler

import (
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
)

type FollowupTasksHandler struct {
	Service *service.QilingService
}

func (h FollowupTasksHandler) List(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, h.Service.FollowupTasks(r))
}
