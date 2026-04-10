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
	rbacSvc rbac.Authorizer,
	projectsSvc *projects.Service,
	notifier handlers.NotificationPublisher,
) {
	projectsH := handlers.NewProjectsHandler(projectsSvc)
	projectsH.SetNotifier(notifier)
	enforceProjects := rbacFeatureEnabled("RBAC_ENFORCE_PROJECTS_GET", true)

	p := v2.Group("/projects")
	p.Use(authMW)
	p.GET("/my", projectsH.ListMine)
	p.GET("/faculty",
		middleware.RequirePermissionIf(enforceProjects && rbacSvc != nil, rbacSvc, "project.view", middleware.FacultyScopeFromCtx()),
		projectsH.ListFaculty,
	)
	p.GET("/public", projectsH.ListPublic)
	p.GET("/groups", projectsH.ListGroups)

	p.POST("",
		middleware.RequirePermission(rbacSvc, "project.create", middleware.FacultyScopeFromCtx()),
		projectsH.Create,
	)

	p.GET("/:project_id",
		projectsH.Get,
	)
	p.GET("/:project_id/final-report.pdf",
		projectsH.DownloadFinalReport,
	)
	p.POST("/:project_id/image",
		middleware.RequirePermissionIf(enforceProjects && rbacSvc != nil, rbacSvc, "project.edit", middleware.ProjectScopeFromParam("project_id")),
		projectsH.UploadImage,
	)
	p.DELETE("/:project_id/image",
		middleware.RequirePermissionIf(enforceProjects && rbacSvc != nil, rbacSvc, "project.edit", middleware.ProjectScopeFromParam("project_id")),
		projectsH.DeleteImage,
	)
}
