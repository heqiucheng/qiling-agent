package app

import (
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	"github.com/heqiucheng/qiling-agent/backend/internal/db"
	httpx "github.com/heqiucheng/qiling-agent/backend/internal/http"
	"github.com/heqiucheng/qiling-agent/backend/internal/store"
)

func NewHTTPHandler(cfg config.Config) (http.Handler, error) {
	repository, err := buildRepository(cfg)
	if err != nil {
		return nil, err
	}

	return httpx.NewRouterWithRepository(cfg, repository), nil
}

func buildRepository(cfg config.Config) (store.Repository, error) {
	if cfg.StoreDriver != "mysql" {
		return store.NewMockStore(), nil
	}

	database, err := db.OpenMySQL(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return store.NewMySQLRepository(database), nil
}
