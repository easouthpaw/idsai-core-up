package app

import (
	"context"
	"log"
	"strings"
	"time"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/db"
	httpx "idsai-core-up/internal/http"
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/infra/alerts"
	"idsai-core-up/internal/infra/cache"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Cfg   config.Config
	DB    *pgxpool.Pool
	HTTP  *gin.Engine
	redis *cache.RedisClient
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

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
		modules.rbacAuthorizer,
		modules.projectsSvc,
		modules.projectFlowHandler,
		modules.authHandler,
		modules.adminHandler,
		modules.notificationsHandler,
		publicContactHandler,
		modules.kbHandler,
		modules.notificationsSvc,
		cfg.JWTSecret,
	)
	if dir := strings.TrimSpace(cfg.LocalStorageDir); dir != "" {
		router.StaticFS("/media", gin.Dir(dir, false))
	}

	return &App{Cfg: cfg, DB: pool, HTTP: router, redis: modules.redisClient}, nil
}

// Close performs graceful shutdown of resources.
func (a *App) Close() {
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			log.Printf("[WARN] redis: close error: %v", err)
		}
	}
	a.DB.Close()
}
