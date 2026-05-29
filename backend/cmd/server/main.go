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
	handler, err := app.NewHTTPHandler(cfg, logger)
	if err != nil {
		logger.Error("initialize backend failed", "error", err, "store_driver", cfg.StoreDriver)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}

	logger.Info("starting qiling backend", "addr", cfg.Addr, "env", cfg.Env, "store_driver", cfg.StoreDriver)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
