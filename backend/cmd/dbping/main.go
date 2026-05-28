package main

import (
	"log/slog"
	"os"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	"github.com/heqiucheng/qiling-agent/backend/internal/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	database, err := db.OpenMySQL(cfg.DatabaseURL)
	if err != nil {
		logger.Error("mysql ping failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	logger.Info("mysql ping ok")
}
