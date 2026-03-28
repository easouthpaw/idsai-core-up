package httpx

import (
	"idsai-core-up/internal/http/handlers"

	"github.com/gin-gonic/gin"
)

func registerKBRoutes(api *gin.RouterGroup, authMW gin.HandlerFunc, kbH *handlers.KBHandler) {
	kb := api.Group("/kb")
	kb.Use(authMW)

	// Categories
	kb.GET("/categories", kbH.ListCategories)
	kb.POST("/categories", kbH.CreateCategory)
	kb.PATCH("/categories/:id", kbH.UpdateCategory)
	kb.DELETE("/categories/:id", kbH.DeleteCategory)

	// Articles
	kb.GET("/articles", kbH.ListArticles)
	kb.POST("/articles", kbH.CreateArticle)
	kb.POST("/articles/upload", kbH.UploadArticle)
	kb.GET("/articles/:id", kbH.GetArticle)
	kb.PATCH("/articles/:id", kbH.UpdateArticle)
	kb.DELETE("/articles/:id", kbH.DeleteArticle)

	// Tags
	kb.GET("/tags", kbH.ListTags)
}
