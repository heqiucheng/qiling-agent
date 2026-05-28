package handler

import (
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
)

type HealthHandler struct {
	Version string
	Env     string
}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "qiling-agent-backend",
		"version": h.Version,
		"env":     h.Env,
	})
}
