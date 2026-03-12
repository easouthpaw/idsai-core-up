package httpx

import (
	"idsai-core-up/internal/http/handlers"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(v2 *gin.RouterGroup, authMW gin.HandlerFunc, authHandler *handlers.AuthHandler) {
	if authHandler == nil {
		return
	}

	auth := v2.Group("/auth")
	auth.POST("/register", authHandler.RegisterStudent)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.GET("/me", authMW, authHandler.Me)
}
