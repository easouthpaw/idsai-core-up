package httpx

import (
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(v2 *gin.RouterGroup, authMW gin.HandlerFunc, authHandler *handlers.AuthHandler) {
	if authHandler == nil {
		return
	}

	auth := v2.Group("/auth")
	auth.GET("/departments", authHandler.ListDepartments)
	auth.GET("/departments/:department_code/groups", authHandler.ListDepartmentGroups)
	auth.POST("/register", authHandler.RegisterStudent)
	auth.POST("/login", authHandler.Login)
	auth.POST("/logout", authHandler.Logout)
	auth.POST("/refresh", authHandler.Refresh)
	auth.GET("/verify-email", authHandler.VerifyEmail)
	auth.POST("/verify-email/resend", authHandler.ResendVerification)
	auth.POST("/password-reset/request", authHandler.RequestPasswordReset)
	auth.GET("/password-reset", authHandler.PasswordResetLanding)
	auth.POST("/password-reset/confirm", authHandler.PasswordResetConfirm)
	auth.GET("/settings/email/verify", authHandler.SettingsVerifyEmailChange)
	auth.POST("/settings/email/confirm", authMW, authHandler.SettingsConfirmEmailChange)
	auth.GET("/me", authMW, authHandler.Me)
	auth.GET("/settings", authMW, authHandler.SettingsGet)
	auth.PATCH("/settings/profile", authMW, authHandler.SettingsUpdateProfile)
	auth.POST("/settings/email/change", authMW, authHandler.SettingsStartEmailChange)
	auth.POST("/settings/email/resend", authMW, authHandler.SettingsResendEmailChange)
	auth.POST("/settings/avatar", authMW, authHandler.SettingsUploadAvatar)
	auth.DELETE("/settings/avatar", authMW, authHandler.SettingsDeleteAvatar)
	auth.POST("/settings/password", authMW, authHandler.SettingsChangePassword)
	auth.GET("/settings/group-change-requests", authMW, authHandler.SettingsListGroupChangeRequests)
	auth.POST("/settings/group-change-requests", authMW, authHandler.SettingsSubmitGroupChangeRequest)
	auth.GET("/groups/tree", authMW, authHandler.ListDepartmentGroupsTree)
	auth.GET("/admin/group-change-requests", authMW, middleware.AdminRequired(), authHandler.AdminListGroupChangeRequests)
	auth.POST("/admin/group-change-requests/:request_id/review", authMW, middleware.AdminRequired(), authHandler.AdminReviewGroupChangeRequest)
}
