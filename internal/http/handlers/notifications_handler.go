package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/notifications"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationsHandler struct {
	svc *notifications.Service
}

func NewNotificationsHandler(svc *notifications.Service) *NotificationsHandler {
	return &NotificationsHandler{svc: svc}
}

func (h *NotificationsHandler) List(c *gin.Context) {
	tenantID, userID, ok := h.ctxIDs(c)
	if !ok {
		return
	}

	limit := 50
	offset := 0
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			offset = parsed
		}
	}

	items, err := h.svc.List(c.Request.Context(), tenantID, userID, limit, offset)
	if err != nil {
		if errors.Is(err, notifications.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *NotificationsHandler) UnreadCount(c *gin.Context) {
	tenantID, userID, ok := h.ctxIDs(c)
	if !ok {
		return
	}

	count, err := h.svc.UnreadCount(c.Request.Context(), tenantID, userID)
	if err != nil {
		if errors.Is(err, notifications.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread": count})
}

func (h *NotificationsHandler) MarkRead(c *gin.Context) {
	tenantID, userID, ok := h.ctxIDs(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(strings.TrimSpace(c.Param("notification_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification_id"})
		return
	}

	if err := h.svc.MarkRead(c.Request.Context(), tenantID, userID, id); err != nil {
		if errors.Is(err, notifications.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, notifications.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notification"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NotificationsHandler) MarkAllRead(c *gin.Context) {
	tenantID, userID, ok := h.ctxIDs(c)
	if !ok {
		return
	}

	updated, err := h.svc.MarkAllRead(c.Request.Context(), tenantID, userID)
	if err != nil {
		if errors.Is(err, notifications.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

func (h *NotificationsHandler) Delete(c *gin.Context) {
	tenantID, userID, ok := h.ctxIDs(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(strings.TrimSpace(c.Param("notification_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification_id"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), tenantID, userID, id); err != nil {
		if errors.Is(err, notifications.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, notifications.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notification"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NotificationsHandler) Clear(c *gin.Context) {
	tenantID, userID, ok := h.ctxIDs(c)
	if !ok {
		return
	}
	deleted, err := h.svc.Clear(c.Request.Context(), tenantID, userID)
	if err != nil {
		if errors.Is(err, notifications.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": deleted})
}

func (h *NotificationsHandler) ctxIDs(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	tenantID, ok := middleware.TenantIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, uuid.Nil, false
	}
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, uuid.Nil, false
	}
	return tenantID, userID, true
}
