package app

import (
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	httpx "github.com/heqiucheng/qiling-agent/backend/internal/http"
)

func NewHTTPHandler(cfg config.Config) http.Handler {
	return httpx.NewRouter(cfg)
}
