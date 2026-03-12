package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projectflow"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectFlowHandler struct {
	svc      *projectflow.Service
	notifier NotificationPublisher
}

func NewProjectFlowHandler(svc *projectflow.Service) *ProjectFlowHandler {
	return &ProjectFlowHandler{svc: svc}
}

func (h *ProjectFlowHandler) SetNotifier(pub NotificationPublisher) {
	h.notifier = pub
}

func parseUserID(c *gin.Context) (uuid.UUID, bool) {
	id, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}
	return id, true
}

func parseProjectID(c *gin.Context) (uuid.UUID, bool) {
	raw := strings.TrimSpace(c.Param("project_id"))
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return uuid.Nil, false
	}
	return id, true
}

func parseUserIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	raw := strings.TrimSpace(c.Param(param))
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

func handleFlowErr(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, projectflow.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrRecruitmentOpen):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrProjectNotReady):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrProjectNotActive):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrPositionFull):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrInviteNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func parseLimit(raw string, def, max int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return def
	}
	if v > max {
		return max
	}
	return v
}
