package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/heqiucheng/qiling-agent/backend/internal/app"
	"github.com/heqiucheng/qiling-agent/backend/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: app.NewHTTPHandler(cfg),
	}

	logger.Info("starting qiling backend", "addr", cfg.Addr, "env", cfg.Env)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
