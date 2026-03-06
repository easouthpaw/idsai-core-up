package httpx

import (
	"io/fs"
	"net/http"

	"idsai-core-up/internal/http/frontend"
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projects"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "idsai-core-up/docs/swagger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(
	pool *pgxpool.Pool,
	rbacSvc *rbac.Service,
	projectsSvc *projects.Service,
	projectFlowH *handlers.ProjectFlowHandler,
	authHandler *handlers.AuthHandler,
	adminHandler *handlers.AdminHandler,
	jwtSecret string,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	if authHandler != nil {
		r.POST("/auth/register", authHandler.RegisterStudent)
		r.POST("/auth/login", authHandler.Login)
		r.POST("/auth/refresh", authHandler.Refresh)

		authMW := middleware.AuthRequired(jwtSecret)
		r.GET("/auth/me", authMW, authHandler.Me)
	}

	if adminHandler != nil {
		adminMW := middleware.AuthRequired(jwtSecret)
		admin := r.Group("/admin")
		admin.Use(adminMW, middleware.AdminRequired())
		admin.GET("/users", adminHandler.ListUsers)
		admin.POST("/users/students", adminHandler.CreateStudent)
		admin.POST("/users/professors", adminHandler.CreateProfessor)
		admin.PATCH("/users/:user_id/status", adminHandler.SetStatus)
		admin.GET("/projects", adminHandler.ListProjects)
		admin.PATCH("/projects/:project_id/status", adminHandler.SetProjectStatus)
	}

	projectsH := handlers.NewProjectsHandler(projectsSvc)
	r.GET("/projects/my", projectsH.ListMine)
	r.GET("/projects/public", projectsH.ListPublic)
	r.GET("/projects/groups", projectsH.ListGroups)

	r.POST("/projects",
		middleware.RequirePermission(rbacSvc, "project.create", middleware.FacultyScopeFromHeader("X-Faculty-ID")),
		projectsH.Create,
	)

	r.GET("/projects/:project_id",
		middleware.RequirePermission(rbacSvc, "project.view", middleware.ProjectScopeFromParam("project_id")),
		projectsH.Get,
	)

	health := handlers.NewHealthHandler(pool)
	r.GET("/health", health.Get)
	staticFS, err := fs.Sub(frontend.Files, ".")
	if err == nil {
		r.StaticFS("/dev/static", http.FS(staticFS))
	}
	r.GET("/dev/login", handlers.DevLoginPage)
	r.GET("/dev/admin", handlers.DevAdminPage)
	r.GET("/dev/projects", handlers.DevProjectsPage)
	r.GET("/dev/projects/:project_id", handlers.DevProjectPage)
	r.GET("/dev/tester", handlers.DevLoginPage)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"name": "IDSAI Core API"})
	})

	if projectFlowH != nil {
		p := r.Group("/projects/:project_id")
		p.PATCH("", projectFlowH.UpdateProject)
		p.PUT("/stacks", projectFlowH.SetStacks)
		p.GET("/stacks", projectFlowH.ListStacks)
		p.POST("/recruitment/open", projectFlowH.OpenRecruitment)
		p.POST("/positions", projectFlowH.CreatePosition)
		p.GET("/positions", projectFlowH.ListPositions)
		p.POST("/members/apply", projectFlowH.ApplyMember)
		p.GET("/members", projectFlowH.ListMembers)
		p.POST("/members/:user_id/approve", projectFlowH.ApproveMember)
		p.PATCH("/members/:user_id/position", projectFlowH.SetMemberPosition)
		p.POST("/professor", projectFlowH.AssignProfessor)
		p.POST("/criteria", projectFlowH.CreateCriterion)
		p.GET("/criteria", projectFlowH.ListCriteria)
		p.GET("/readiness", projectFlowH.Readiness)
		p.POST("/approve", projectFlowH.ApproveProject)
		p.GET("/tasks", projectFlowH.ListTasks)
		p.POST("/tasks", projectFlowH.CreateTask)
		p.PATCH("/tasks/:task_id/status", projectFlowH.UpdateTaskStatus)
		p.PATCH("/tasks/:task_id/assignee", projectFlowH.AssignTask)
		p.POST("/tasks/:task_id/claim", projectFlowH.ClaimTask)
	}

	return r
}
