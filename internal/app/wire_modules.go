package app

import (
	"idsai-core-up/internal/config"
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/infra/cache"
	"idsai-core-up/internal/infra/storage"
	adminmodule "idsai-core-up/internal/modules/admin"
	authmodule "idsai-core-up/internal/modules/auth"
	kbmodule "idsai-core-up/internal/modules/kb"
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
	rbacAuthorizer       rbac.Authorizer
	projectsSvc          *projects.Service
	projectFlowHandler   *handlers.ProjectFlowHandler
	notificationsSvc     *notifications.Service
	notificationsHandler *handlers.NotificationsHandler
	notificationsRepo    *postgres.NotificationsRepo
	kbHandler            *handlers.KBHandler
	redisClient          *cache.RedisClient
}

func wireModules(pool *pgxpool.Pool, cfg config.Config) wiredModules {
	objectStorage := storage.NewFromConfig(cfg)

	// Redis (graceful — nil if unavailable)
	var redisClient *cache.RedisClient
	if cfg.RedisAddr != "" {
		redisClient = cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	}

	authModule := authmodule.New(pool, cfg)
	adminModule := adminmodule.New(pool)
	rbacModule := rbacmodule.New(pool, redisClient)
	projectsModule := projectsmodule.New(pool, rbacModule.Repo)
	projectFlowModule := projectflowmodule.New(pool, rbacModule.Authorizer, rbacModule.Repo)
	notificationsModule := notificationsmodule.New(pool)
	kbModule := kbmodule.New(pool)

	authModule.Service.SetNotifier(notificationsModule.Service)
	authModule.Service.SetStorage(objectStorage)
	authModule.Handler.SetAuthorizer(rbacModule.Authorizer)
	projectsModule.Service.SetStorage(objectStorage)
	adminModule.Handler.SetNotifier(notificationsModule.Service)
	projectFlowModule.Handler.SetNotifier(notificationsModule.Service)

	return wiredModules{
		authHandler:          authModule.Handler,
		adminHandler:         adminModule.Handler,
		rbacAuthorizer:       rbacModule.Authorizer,
		projectsSvc:          projectsModule.Service,
		projectFlowHandler:   projectFlowModule.Handler,
		notificationsSvc:     notificationsModule.Service,
		notificationsHandler: notificationsModule.Handler,
		notificationsRepo:    notificationsModule.Repo,
		kbHandler:            kbModule.Handler,
		redisClient:          redisClient,
	}
}
