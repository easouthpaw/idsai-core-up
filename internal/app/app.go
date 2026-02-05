package app

import (
	"context"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/db"
	httpx "idsai-core-up/internal/http"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Cfg  config.Config
	DB   *pgxpool.Pool
	HTTP *gin.Engine
}

func New(ctx context.Context) (*App, error) {
	cfg := config.Load()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	rbacRepo := postgres.NewRBACRepo(pool)
	rbacSvc := rbac.NewService(rbacRepo)

	router := httpx.NewRouter(pool, rbacSvc)
	return &App{Cfg: cfg, DB: pool, HTTP: router}, nil
}
