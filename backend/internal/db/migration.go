package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
)

var ResetTables = []string{
	"customer_memory_facts",
	"audit_events",
	"agent_runs",
	"followup_tasks",
	"conversation_messages",
	"uploads",
	"customers",
	"users",
}

func RunMigrations(database *sql.DB, dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}

	applied := make([]string, 0, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		for _, statement := range SplitSQLStatements(string(content)) {
			if _, err := database.Exec(statement); err != nil {
				return nil, err
			}
		}
		applied = append(applied, file)
	}
	return applied, nil
}

func Reset(database *sql.DB, migrationsDir string) ([]string, error) {
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return nil, err
	}
	defer database.Exec("SET FOREIGN_KEY_CHECKS = 1")

	for _, table := range ResetTables {
		if _, err := database.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			return nil, err
		}
	}

	return RunMigrations(database, migrationsDir)
}

func SplitSQLStatements(content string) []string {
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
