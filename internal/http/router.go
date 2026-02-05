package httpx

import (
	"net/http"

	"idsai-core-up/internal/http/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "idsai-core-up/docs/swagger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(db *pgxpool.Pool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	health := handlers.NewHealthHandler(db)
	r.GET("/health", health.Get)

	// Swagger UI: /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"name": "IDSAI Core API"})
	})

	return r
}
