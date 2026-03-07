package httpx

import (
	"io/fs"
	"net/http"
	"strings"

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

	if authHandler != nil {
		auth := v2.Group("/auth")
		auth.POST("/register", authHandler.RegisterStudent)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)

		auth.GET("/me", authMW, authHandler.Me)
	}

	if adminHandler != nil {
		admin := v2.Group("/admin")
		admin.Use(authMW, middleware.AdminRequired())
		admin.GET("/users", adminHandler.ListUsers)
		admin.POST("/users/students", adminHandler.CreateStudent)
		admin.POST("/users/professors", adminHandler.CreateProfessor)
		admin.PATCH("/users/:user_id/status", adminHandler.SetStatus)
		admin.DELETE("/users/:user_id", adminHandler.DeleteUser)
		admin.GET("/projects", adminHandler.ListProjects)
		admin.PATCH("/projects/:project_id/status", adminHandler.SetProjectStatus)
		admin.DELETE("/projects/:project_id", adminHandler.DeleteProject)
	}

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

	p.GET("/:project_id",
		middleware.RequirePermission(rbacSvc, "project.view", middleware.ProjectScopeFromParam("project_id")),
		projectsH.Get,
	)

	if notificationsH != nil {
		n := v2.Group("/notifications")
		n.Use(authMW)
		n.GET("", notificationsH.List)
		n.GET("/unread-count", notificationsH.UnreadCount)
		n.POST("/read-all", notificationsH.MarkAllRead)
		n.POST("/:notification_id/read", notificationsH.MarkRead)
		n.DELETE("/:notification_id", notificationsH.Delete)
		n.DELETE("", notificationsH.Clear)
	}

	health := handlers.NewHealthHandler(pool)
	r.GET("/health", health.Get)
	staticFS, err := fs.Sub(frontend.Files, ".")
	if err == nil {
		r.StaticFS("/dev/static", http.FS(staticFS))
	}
	r.GET("/dev/login", handlers.DevLoginPage)
	r.GET("/dev/author", handlers.DevAuthorPage)
	r.GET("/dev/admin", handlers.DevAdminPage)
	r.GET("/dev/projects", handlers.DevProjectsPage)
	r.GET("/dev/projects/:project_id", handlers.DevProjectPage)
	r.GET("/dev/professor", handlers.DevProfessorPage)
	r.GET("/dev/professor/reviews", handlers.DevProfessorReviewsPage)
	r.GET("/dev/professor/criteria", handlers.DevProfessorCriteriaPage)
	r.GET("/dev/professor/grading", handlers.DevProfessorGradingPage)
	r.GET("/dev/tester", handlers.DevLoginPage)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", handlers.DevLandingPage)
	r.GET("/author", handlers.DevAuthorPage)

	if projectFlowH != nil {
		projectFlow := v2.Group("/projects/:project_id")
		projectFlow.Use(authMW)
		projectFlow.PATCH("", projectFlowH.UpdateProject)
		projectFlow.PUT("/stacks", projectFlowH.SetStacks)
		projectFlow.GET("/stacks", projectFlowH.ListStacks)
		projectFlow.POST("/recruitment/open", projectFlowH.OpenRecruitment)
		projectFlow.POST("/positions", projectFlowH.CreatePosition)
		projectFlow.GET("/positions", projectFlowH.ListPositions)
		projectFlow.GET("/candidates/students", projectFlowH.ListStudentCandidates)
		projectFlow.GET("/candidates/professors", projectFlowH.SearchProfessors)
		projectFlow.POST("/members/apply", projectFlowH.ApplyMember)
		projectFlow.POST("/members/invite", projectFlowH.InviteMember)
		projectFlow.POST("/members/respond", projectFlowH.RespondMemberInvite)
		projectFlow.GET("/members", projectFlowH.ListMembers)
		projectFlow.POST("/members/:user_id/approve", projectFlowH.ApproveMember)
		projectFlow.PATCH("/members/:user_id/position", projectFlowH.SetMemberPosition)
		projectFlow.GET("/professor", projectFlowH.GetAssignedProfessor)
		projectFlow.POST("/professor", projectFlowH.AssignProfessor)
		projectFlow.POST("/professor/respond", projectFlowH.RespondProfessorInvite)
		projectFlow.POST("/criteria", projectFlowH.CreateCriterion)
		projectFlow.GET("/criteria", projectFlowH.ListCriteria)
		projectFlow.GET("/grading", projectFlowH.GetGrading)
		projectFlow.PUT("/grading", projectFlowH.UpsertGrading)
		projectFlow.GET("/readiness", projectFlowH.Readiness)
		projectFlow.POST("/approve", projectFlowH.ApproveProject)
		projectFlow.GET("/tasks", projectFlowH.ListTasks)
		projectFlow.POST("/tasks", projectFlowH.CreateTask)
		projectFlow.PATCH("/tasks/:task_id/status", projectFlowH.UpdateTaskStatus)
		projectFlow.PATCH("/tasks/:task_id/assignee", projectFlowH.AssignTask)
		projectFlow.POST("/tasks/:task_id/claim", projectFlowH.ClaimTask)

		professor := v2.Group("/professor")
		professor.Use(authMW)
		professor.GET("/review-invites", projectFlowH.ListProfessorReviewInvites)
	}

	return r
}
