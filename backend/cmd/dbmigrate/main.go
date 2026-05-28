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
		logger.Error("mysql connect failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	files, err := db.RunMigrations(database, "migrations")
	if err != nil {
		logger.Error("run migration failed", "error", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		logger.Error("no migration files found")
		os.Exit(1)
	}

	logger.Info("mysql migration ok", "files", len(files))
}
