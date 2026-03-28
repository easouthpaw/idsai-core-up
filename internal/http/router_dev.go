package httpx

import (
	"io/fs"
	"net/http"
	pathpkg "path"
	"strings"

	"idsai-core-up/internal/http/frontend"
	"idsai-core-up/internal/http/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
	r.GET("/dev/profile", handlers.DevProfilePage)
	r.GET("/dev/groups", handlers.DevGroupsPage)
	r.GET("/dev/kb", handlers.DevKBPage)
	r.GET("/dev/kb/article", handlers.DevKBArticlePage)
	r.GET("/dev/404", handlers.DevNotFoundPage)
	r.GET("/dev/tester", handlers.DevLoginPage)

	r.GET("/", handlers.DevLandingPage)
	r.GET("/author", handlers.DevAuthorPage)
	r.GET("/404", handlers.DevNotFoundPage)

	r.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		switch {
		case strings.HasPrefix(requestPath, "/v2/"):
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case strings.HasPrefix(requestPath, "/dev/static/"), pathpkg.Ext(requestPath) != "":
			c.Status(http.StatusNotFound)
		case c.Request.Method == http.MethodGet:
			handlers.DevNotFoundPage(c)
		default:
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})
}
