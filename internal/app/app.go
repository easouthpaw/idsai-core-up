package app

import (
	"context"
	"time"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/db"
	httpx "idsai-core-up/internal/http"
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/infra/email"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/admin"
	"idsai-core-up/internal/services/auth"
	"idsai-core-up/internal/services/notifications"
	"idsai-core-up/internal/services/projectflow"
	"idsai-core-up/internal/services/projects"
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

	authRepo := postgres.NewAuthRepo(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(authSvc)

	adminRepo := postgres.NewAdminRepo(pool)
	adminSvc := admin.NewService(adminRepo)
	adminHandler := handlers.NewAdminHandler(adminSvc)

	rbacRepo := postgres.NewRBACRepo(pool)
	rbacSvc := rbac.NewService(rbacRepo)

	projectsRepo := postgres.NewProjectsRepo(pool)
	projectsSvc := projects.NewService(projectsRepo, rbacRepo)
	projectFlowSvc := projectflow.NewService(pool, rbacSvc, rbacRepo)
	projectFlowHandler := handlers.NewProjectFlowHandler(projectFlowSvc)

	notificationsRepo := postgres.NewNotificationsRepo(pool)
	notificationsSvc := notifications.NewService(notificationsRepo)
	notificationsHandler := handlers.NewNotificationsHandler(notificationsSvc)
	authSvc.SetNotifier(notificationsSvc)
	adminHandler.SetNotifier(notificationsSvc)
	projectFlowHandler.SetNotifier(notificationsSvc)

	if cfg.EmailEnable {
		emailSender := email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
		dispatcher := notifications.NewOutboxDispatcher(notificationsRepo, emailSender)
		pollEvery := time.Duration(cfg.OutboxPollS) * time.Second
		go dispatcher.Start(ctx, pollEvery)
	}

	router := httpx.NewRouter(pool, rbacSvc, projectsSvc, projectFlowHandler, authHandler, adminHandler, notificationsHandler, notificationsSvc, cfg.JWTSecret)

	return &App{Cfg: cfg, DB: pool, HTTP: router}, nil
}
