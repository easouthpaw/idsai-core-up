package httpx

import (
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(v2 *gin.RouterGroup, authMW gin.HandlerFunc, rbacSvc rbac.Authorizer, adminHandler *handlers.AdminHandler) {
	if adminHandler == nil {
		return
	}

	enforce := rbacFeatureEnabled("RBAC_ENFORCE_ADMIN_PERMS", true)
	admin := v2.Group("/admin")
	admin.Use(
		authMW,
		middleware.AdminRequired(),
		middleware.RequirePermissionIf(enforce && rbacSvc != nil, rbacSvc, "admin.manage_rbac", middleware.SystemScope()),
	)
	admin.GET("/users", adminHandler.ListUsers)
	admin.POST("/users/students", adminHandler.CreateStudent)
	admin.POST("/users/professors", adminHandler.CreateProfessor)
	admin.PATCH("/users/:user_id/status", adminHandler.SetStatus)
	admin.PATCH("/users/:user_id/role", adminHandler.SetRole)
	admin.PATCH("/users/:user_id/password", adminHandler.ResetPassword)
	admin.DELETE("/users/:user_id", adminHandler.DeleteUser)
	admin.GET("/projects", adminHandler.ListProjects)
	admin.GET("/projects/:project_id/observe", adminHandler.ObserveProject)
	admin.PATCH("/projects/:project_id/status", adminHandler.SetProjectStatus)
	admin.DELETE("/projects/:project_id", adminHandler.DeleteProject)
}
