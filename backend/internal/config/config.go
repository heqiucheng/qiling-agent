package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr        string
	Env         string
	DatabaseURL string
}

func Load() Config {
	addr := strings.TrimSpace(os.Getenv("QILING_HTTP_ADDR"))
	if addr == "" {
		addr = ":8080"
	}

	env := strings.TrimSpace(os.Getenv("QILING_ENV"))
	if env == "" {
		env = "development"
	}

	databaseURL := strings.TrimSpace(os.Getenv("QILING_DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = "root:password@tcp(127.0.0.1:3306)/qiling_agent?parseTime=true&charset=utf8mb4&loc=Local"
	}

	return Config{Addr: addr, Env: env, DatabaseURL: databaseURL}
}
