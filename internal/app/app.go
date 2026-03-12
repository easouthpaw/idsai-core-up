package app

import (
	"context"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/db"
	httpx "idsai-core-up/internal/http"

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

	modules := wireModules(pool, cfg)
	startEmailOutboxDispatcher(ctx, cfg, modules.notificationsRepo)

	router := httpx.NewRouter(
		pool,
		modules.rbacSvc,
		modules.projectsSvc,
		modules.projectFlowHandler,
		modules.authHandler,
		modules.adminHandler,
		modules.notificationsHandler,
		modules.notificationsSvc,
		cfg.JWTSecret,
	)

	return &App{Cfg: cfg, DB: pool, HTTP: router}, nil
}
