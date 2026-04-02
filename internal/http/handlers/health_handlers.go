package handlers

import (
	"context"
	"net/http"
	"time"

	"idsai-core-up/internal/http/dto"

	"github.com/gin-gonic/gin"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db DBPinger
}

func NewHealthHandler(db DBPinger) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Get(c *gin.Context) {
	if h.db != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		if err := h.db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, dto.HealthStatusResponse{Status: "db_down", Error: err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, dto.HealthStatusResponse{Status: "ok"})
}
