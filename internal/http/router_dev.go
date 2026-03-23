package httpx

import (
	"io/fs"
	"net/http"

	"idsai-core-up/internal/http/frontend"
	"idsai-core-up/internal/http/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "idsai-core-up/docs/swagger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func registerDevAndDocsRoutes(r *gin.Engine, pool *pgxpool.Pool) {
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
	r.GET("/dev/invites", handlers.DevInvitesPage)
	r.GET("/dev/professor", handlers.DevProfessorPage)
	r.GET("/dev/professor/reviews", handlers.DevProfessorReviewsPage)
	r.GET("/dev/professor/criteria", handlers.DevProfessorCriteriaPage)
	r.GET("/dev/professor/grading", handlers.DevProfessorGradingPage)
	r.GET("/dev/settings", handlers.DevSettingsPage)
	r.GET("/dev/tester", handlers.DevLoginPage)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", handlers.DevLandingPage)
	r.GET("/author", handlers.DevAuthorPage)
}
