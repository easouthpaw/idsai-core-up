package httpx

import (
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projects"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
)

func registerProjectsRoutes(
	v2 *gin.RouterGroup,
	authMW gin.HandlerFunc,
	rbacSvc *rbac.Service,
	projectsSvc *projects.Service,
	notifier handlers.NotificationPublisher,
) {
	projectsH := handlers.NewProjectsHandler(projectsSvc)
	projectsH.SetNotifier(notifier)

	p := v2.Group("/projects")
	p.Use(authMW)
	p.GET("/my", projectsH.ListMine)
	p.GET("/public", projectsH.ListPublic)
	p.GET("/groups", projectsH.ListGroups)

	p.POST("",
		middleware.RequirePermission(rbacSvc, "project.create", middleware.FacultyScopeFromCtx()),
		projectsH.Create,
	)

	p.GET("/:project_id", projectsH.Get)
	p.POST("/:project_id/image", projectsH.UploadImage)
	p.DELETE("/:project_id/image", projectsH.DeleteImage)
}
