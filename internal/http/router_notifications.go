package httpx

import (
	"idsai-core-up/internal/http/handlers"

	"github.com/gin-gonic/gin"
)

func registerNotificationRoutes(v2 *gin.RouterGroup, authMW gin.HandlerFunc, notificationsH *handlers.NotificationsHandler) {
	if notificationsH == nil {
		return
	}

	n := v2.Group("/notifications")
	n.Use(authMW)
	n.GET("", notificationsH.List)
	n.GET("/unread-count", notificationsH.UnreadCount)
	n.POST("/read-all", notificationsH.MarkAllRead)
	n.POST("/:notification_id/read", notificationsH.MarkRead)
	n.DELETE("/:notification_id", notificationsH.Delete)
	n.DELETE("", notificationsH.Clear)
}
