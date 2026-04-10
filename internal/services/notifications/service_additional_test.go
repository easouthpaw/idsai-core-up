package notifications_test

import (
	"context"
	"testing"

	"idsai-core-up/internal/services/notifications"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnreadCountMarkReadMarkAllDeleteAndClear(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	notificationID := uuid.New()
	repo := &fakeRepo{
		unread:     3,
		markAll:    7,
		clearCount: 4,
	}
	svc := notifications.NewService(repo)

	_, err := svc.UnreadCount(context.Background(), uuid.Nil, userID)
	require.ErrorIs(t, err, notifications.ErrInvalidInput)

	unread, err := svc.UnreadCount(context.Background(), tenantID, userID)
	require.NoError(t, err)
	require.Equal(t, 3, unread)

	err = svc.MarkRead(context.Background(), uuid.Nil, userID, notificationID)
	require.ErrorIs(t, err, notifications.ErrInvalidInput)
	err = svc.MarkRead(context.Background(), tenantID, userID, notificationID)
	require.NoError(t, err)

	updated, err := svc.MarkAllRead(context.Background(), tenantID, userID)
	require.NoError(t, err)
	require.Equal(t, 7, updated)
	_, err = svc.MarkAllRead(context.Background(), tenantID, uuid.Nil)
	require.ErrorIs(t, err, notifications.ErrInvalidInput)

	err = svc.Delete(context.Background(), tenantID, userID, notificationID)
	require.NoError(t, err)
	err = svc.Delete(context.Background(), tenantID, userID, uuid.Nil)
	require.ErrorIs(t, err, notifications.ErrInvalidInput)

	deleted, err := svc.Clear(context.Background(), tenantID, userID)
	require.NoError(t, err)
	require.Equal(t, 4, deleted)
	_, err = svc.Clear(context.Background(), uuid.Nil, userID)
	require.ErrorIs(t, err, notifications.ErrInvalidInput)
}
