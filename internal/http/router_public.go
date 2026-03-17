package httpx

import (
	"idsai-core-up/internal/http/handlers"

	"github.com/gin-gonic/gin"
)

func registerPublicRoutes(v2 *gin.RouterGroup, publicContactH *handlers.PublicContactHandler) {
	if publicContactH == nil {
		return
	}

	v2.POST("/contact", publicContactH.Submit)
}
