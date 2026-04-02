package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/services/notifications"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type notificationsHandlerRepo struct {
	items  []notifications.Notification
	unread int
}

func (f *notificationsHandlerRepo) Create(ctx context.Context, in notifications.CreateInput, payload []byte) (notifications.Notification, error) {
	return notifications.Notification{}, nil
}

func (f *notificationsHandlerRepo) List(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]notifications.Notification, error) {
	return f.items, nil
}

func (f *notificationsHandlerRepo) UnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return f.unread, nil
}

func (f *notificationsHandlerRepo) MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	return nil
}

func (f *notificationsHandlerRepo) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (f *notificationsHandlerRepo) Delete(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	return nil
}

func (f *notificationsHandlerRepo) Clear(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (f *notificationsHandlerRepo) UserEmail(ctx context.Context, tenantID, userID uuid.UUID) (string, error) {
	return "", nil
}

func withNotificationActor(tenantID, userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("tenantID", tenantID)
		c.Set("userID", userID)
		c.Next()
	}
}

func TestNotificationsHandlerList_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 1, 13, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"project_id":"p1"}`)
	repo := &notificationsHandlerRepo{
		items: []notifications.Notification{
			{
				ID:        "notif-1",
				Type:      "project.created",
				Title:     "Created",
				Body:      "Project created",
				Payload:   payload,
				IsRead:    false,
				CreatedAt: now,
			},
		},
	}

	handler := NewNotificationsHandler(notifications.NewService(repo))
	router := gin.New()
	tenantID := uuid.New()
	userID := uuid.New()
	router.Use(withNotificationActor(tenantID, userID))
	router.GET("/notifications", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ListNotificationsResponse{Items: dto.NotificationResponsesFromService(repo.items)}), rec.Body.String())
}

func TestNotificationsHandlerUnreadCount_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &notificationsHandlerRepo{unread: 7}
	handler := NewNotificationsHandler(notifications.NewService(repo))
	router := gin.New()
	router.Use(withNotificationActor(uuid.New(), uuid.New()))
	router.GET("/notifications/unread-count", handler.UnreadCount)

	req := httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.UnreadCountResponse{Unread: 7}), rec.Body.String())
}
