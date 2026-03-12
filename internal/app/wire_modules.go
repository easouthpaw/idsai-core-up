package app

import (
	"idsai-core-up/internal/config"
	"idsai-core-up/internal/http/handlers"
	adminmodule "idsai-core-up/internal/modules/admin"
	authmodule "idsai-core-up/internal/modules/auth"
	notificationsmodule "idsai-core-up/internal/modules/notifications"
	projectflowmodule "idsai-core-up/internal/modules/projectflow"
	projectsmodule "idsai-core-up/internal/modules/projects"
	rbacmodule "idsai-core-up/internal/modules/rbac"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/notifications"
	"idsai-core-up/internal/services/projects"
	"idsai-core-up/internal/services/rbac"

	"github.com/jackc/pgx/v5/pgxpool"
)

type wiredModules struct {
	authHandler          *handlers.AuthHandler
	adminHandler         *handlers.AdminHandler
	rbacSvc              *rbac.Service
	projectsSvc          *projects.Service
	projectFlowHandler   *handlers.ProjectFlowHandler
	notificationsSvc     *notifications.Service
	notificationsHandler *handlers.NotificationsHandler
	notificationsRepo    *postgres.NotificationsRepo
}

func wireModules(pool *pgxpool.Pool, cfg config.Config) wiredModules {
	authModule := authmodule.New(pool, cfg.JWTSecret)
	adminModule := adminmodule.New(pool)
	rbacModule := rbacmodule.New(pool)
	projectsModule := projectsmodule.New(pool, rbacModule.Repo)
	projectFlowModule := projectflowmodule.New(pool, rbacModule.Service, rbacModule.Repo)
	notificationsModule := notificationsmodule.New(pool)

	authModule.Service.SetNotifier(notificationsModule.Service)
	adminModule.Handler.SetNotifier(notificationsModule.Service)
	projectFlowModule.Handler.SetNotifier(notificationsModule.Service)

	return wiredModules{
		authHandler:          authModule.Handler,
		adminHandler:         adminModule.Handler,
		rbacSvc:              rbacModule.Service,
		projectsSvc:          projectsModule.Service,
		projectFlowHandler:   projectFlowModule.Handler,
		notificationsSvc:     notificationsModule.Service,
		notificationsHandler: notificationsModule.Handler,
		notificationsRepo:    notificationsModule.Repo,
	}
}
