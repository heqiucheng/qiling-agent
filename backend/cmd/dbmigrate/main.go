package main

import (
	"log/slog"
	"os"
	"strings"

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

	content, err := os.ReadFile("migrations/001_init_mysql.sql")
	if err != nil {
		logger.Error("read migration failed", "error", err)
		os.Exit(1)
	}

	for _, statement := range splitSQLStatements(string(content)) {
		if _, err := database.Exec(statement); err != nil {
			logger.Error("execute migration statement failed", "error", err, "statement", statement)
			os.Exit(1)
		}
	}

	logger.Info("mysql migration ok")
}

func splitSQLStatements(content string) []string {
	parts := strings.Split(content, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		statement := strings.TrimSpace(part)
		if statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}
