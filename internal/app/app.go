package app

import (
	"context"
	"time"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/db"
	httpx "idsai-core-up/internal/http"
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/infra/alerts"

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
	publicContactHandler := handlers.NewPublicContactHandler(
		alerts.NewTelegramNotifier(
			cfg.TelegramBotToken,
			cfg.TelegramSuperadminChat,
			cfg.ServerName,
			5*time.Second,
			0,
		),
		cfg.ServerName,
	)

	router := httpx.NewRouter(
		pool,
		modules.rbacSvc,
		modules.projectsSvc,
		modules.projectFlowHandler,
		modules.authHandler,
		modules.adminHandler,
		modules.notificationsHandler,
		publicContactHandler,
		modules.notificationsSvc,
		cfg.JWTSecret,
	)

	return &App{Cfg: cfg, DB: pool, HTTP: router}, nil
}
