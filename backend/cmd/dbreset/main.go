package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	"github.com/heqiucheng/qiling-agent/backend/internal/db"
)

func main() {
	confirm := flag.String("confirm", "", "required confirmation value: qiling_agent")
	migrationsDir := flag.String("migrations", "migrations", "migration directory")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if *confirm != "qiling_agent" {
		logger.Error("refusing to reset database without explicit confirmation", "required", "-confirm qiling_agent")
		os.Exit(1)
	}

	cfg := config.Load()
	database, err := db.OpenMySQL(cfg.DatabaseURL)
	if err != nil {
		logger.Error("mysql connect failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	files, err := db.Reset(database, *migrationsDir)
	if err != nil {
		logger.Error("reset database failed", "error", err)
		os.Exit(1)
	}

	logger.Info("mysql reset ok", "tables", len(db.ResetTables), "migration_files", len(files))
}
