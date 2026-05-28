package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"os"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	"github.com/heqiucheng/qiling-agent/backend/internal/db"
)

var resetTables = []string{
	"agent_runs",
	"followup_tasks",
	"conversation_messages",
	"uploads",
	"customers",
	"users",
}

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

	if err := resetDatabase(database); err != nil {
		logger.Error("reset database failed", "error", err)
		os.Exit(1)
	}

	files, err := db.RunMigrations(database, *migrationsDir)
	if err != nil {
		logger.Error("rerun migrations failed", "error", err)
		os.Exit(1)
	}

	logger.Info("mysql reset ok", "tables", len(resetTables), "migration_files", len(files))
}

func resetDatabase(database *sql.DB) error {
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return err
	}
	defer database.Exec("SET FOREIGN_KEY_CHECKS = 1")

	for _, table := range resetTables {
		if _, err := database.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			return err
		}
	}
	return nil
}
