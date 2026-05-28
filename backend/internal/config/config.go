package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr string
	Env  string
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

	return Config{Addr: addr, Env: env}
}
