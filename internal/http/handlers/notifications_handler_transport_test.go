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
	items          []notifications.Notification
	listErr        error
	unread         int
	unreadErr      error
	markReadErr    error
	markAllUpdated int
	markAllErr     error
	deleteErr      error
	clearDeleted   int
	clearErr       error
}

func (f *notificationsHandlerRepo) Create(ctx context.Context, in notifications.CreateInput, payload []byte) (notifications.Notification, error) {
	return notifications.Notification{}, nil
}

func (f *notificationsHandlerRepo) List(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]notifications.Notification, error) {
	return f.items, f.listErr
}

func (f *notificationsHandlerRepo) UnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return f.unread, f.unreadErr
}

func (f *notificationsHandlerRepo) MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	return f.markReadErr
}

func (f *notificationsHandlerRepo) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return f.markAllUpdated, f.markAllErr
}

func (f *notificationsHandlerRepo) Delete(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	return f.deleteErr
}

func (f *notificationsHandlerRepo) Clear(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return f.clearDeleted, f.clearErr
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

func TestNotificationsHandlerMutatingRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	notificationID := uuid.New()
	repo := &notificationsHandlerRepo{
		markAllUpdated: 3,
		clearDeleted:   2,
	}
	handler := NewNotificationsHandler(notifications.NewService(repo))
	router := gin.New()
	router.Use(withNotificationActor(uuid.New(), uuid.New()))
	router.PATCH("/notifications/:notification_id/read", handler.MarkRead)
	router.PATCH("/notifications/read-all", handler.MarkAllRead)
	router.DELETE("/notifications/:notification_id", handler.Delete)
	router.DELETE("/notifications", handler.Clear)

	requireStatus(t, router, http.MethodPatch, "/notifications/"+notificationID.String()+"/read", "", http.StatusNoContent)
	requireStatus(t, router, http.MethodPatch, "/notifications/read-all", "", http.StatusOK)
	requireStatus(t, router, http.MethodDelete, "/notifications/"+notificationID.String(), "", http.StatusNoContent)
	requireStatus(t, router, http.MethodDelete, "/notifications", "", http.StatusOK)
}

func TestNotificationsHandlerErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	notificationID := uuid.New()
	handler := NewNotificationsHandler(notifications.NewService(&notificationsHandlerRepo{
		markReadErr: notifications.ErrNotFound,
		deleteErr:   notifications.ErrNotFound,
	}))
	router := gin.New()
	router.Use(withNotificationActor(uuid.New(), uuid.New()))
	router.GET("/notifications", handler.List)
	router.PATCH("/notifications/:notification_id/read", handler.MarkRead)
	router.DELETE("/notifications/:notification_id", handler.Delete)

	requireStatus(t, router, http.MethodGet, "/notifications", "", http.StatusOK)
	requireStatus(t, router, http.MethodPatch, "/notifications/not-a-uuid/read", "", http.StatusBadRequest)
	requireStatus(t, router, http.MethodPatch, "/notifications/"+notificationID.String()+"/read", "", http.StatusNotFound)
	requireStatus(t, router, http.MethodDelete, "/notifications/not-a-uuid", "", http.StatusBadRequest)
	requireStatus(t, router, http.MethodDelete, "/notifications/"+notificationID.String(), "", http.StatusNotFound)

	noActorRouter := gin.New()
	noActorRouter.GET("/notifications", handler.List)
	requireStatus(t, noActorRouter, http.MethodGet, "/notifications", "", http.StatusUnauthorized)

	internalHandler := NewNotificationsHandler(notifications.NewService(&notificationsHandlerRepo{
		listErr:    context.Canceled,
		unreadErr:  context.Canceled,
		markAllErr: context.Canceled,
		clearErr:   context.Canceled,
	}))
	internalRouter := gin.New()
	internalRouter.Use(withNotificationActor(uuid.New(), uuid.New()))
	internalRouter.GET("/notifications", internalHandler.List)
	internalRouter.GET("/notifications/unread-count", internalHandler.UnreadCount)
	internalRouter.PATCH("/notifications/read-all", internalHandler.MarkAllRead)
	internalRouter.DELETE("/notifications", internalHandler.Clear)
	requireStatus(t, internalRouter, http.MethodGet, "/notifications", "", http.StatusInternalServerError)
	requireStatus(t, internalRouter, http.MethodGet, "/notifications/unread-count", "", http.StatusInternalServerError)
	requireStatus(t, internalRouter, http.MethodPatch, "/notifications/read-all", "", http.StatusInternalServerError)
	requireStatus(t, internalRouter, http.MethodDelete, "/notifications", "", http.StatusInternalServerError)
}
