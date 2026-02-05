package httpx

import (
	"net/http"

	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "idsai-core-up/docs/swagger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(db *pgxpool.Pool, authz rbac.Authorizer) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	health := handlers.NewHealthHandler(db)
	r.GET("/health", health.Get)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"name": "IDSAI Core API"})
	})

	// Demo tasks routes with RBAC
	taskH := handlers.NewTaskDemoHandler()

	p := r.Group("/projects/:project_id")
	p.GET("/tasks",
		middleware.RequirePermission(authz, "task.view", middleware.ProjectScopeFromParam("project_id")),
		taskH.List,
	)
	p.POST("/tasks",
		middleware.RequirePermission(authz, "task.create", middleware.ProjectScopeFromParam("project_id")),
		taskH.Create,
	)

	return r
}
