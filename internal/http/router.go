package httpx

import (
	"strings"

	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projects"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(
	pool *pgxpool.Pool,
	rbacSvc *rbac.Service,
	projectsSvc *projects.Service,
	projectFlowH *handlers.ProjectFlowHandler,
	authHandler *handlers.AuthHandler,
	adminHandler *handlers.AdminHandler,
	notificationsH *handlers.NotificationsHandler,
	notifier handlers.NotificationPublisher,
	jwtSecret string,
) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestLogger(), gin.Recovery())
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/dev") {
			c.Header("Cache-Control", "no-store, max-age=0")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	v2 := r.Group("/v2")
	authMW := middleware.AuthRequired(jwtSecret)

	registerAuthRoutes(v2, authMW, authHandler)
	registerAdminRoutes(v2, authMW, adminHandler)
	registerProjectsRoutes(v2, authMW, rbacSvc, projectsSvc, notifier)
	registerNotificationRoutes(v2, authMW, notificationsH)
	registerProjectFlowRoutes(v2, authMW, rbacSvc, projectFlowH)
	registerDevAndDocsRoutes(r, pool)

	return r
}
