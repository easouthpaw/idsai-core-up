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
	auth.POST("/logout", authHandler.Logout)
	auth.POST("/refresh", authHandler.Refresh)
	auth.GET("/verify-email", authHandler.VerifyEmail)
	auth.POST("/verify-email/resend", authHandler.ResendVerification)
	auth.POST("/password-reset/request", authHandler.RequestPasswordReset)
	auth.GET("/password-reset", authHandler.PasswordResetLanding)
	auth.POST("/password-reset/confirm", authHandler.PasswordResetConfirm)
	auth.GET("/me", authMW, authHandler.Me)
}
