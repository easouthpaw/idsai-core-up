//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/notifications"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestNotificationsRepo_Integration_InboxAndOutboxFlow(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewNotificationsRepo(pool)
	tenantID, userID := seedKBAuthor(t, ctx, pool, "notifications")

	email, err := repo.UserEmail(ctx, tenantID, userID)
	require.NoError(t, err)
	require.Contains(t, email, "notifications-author-")

	empty, err := repo.FetchPendingOutbox(ctx, time.Now().UTC().Add(time.Hour), 0)
	require.NoError(t, err)
	require.Empty(t, empty)

	first, err := repo.Create(ctx, notifications.CreateInput{
		TenantID: tenantID,
		UserID:   userID,
		Type:     "project.created",
		Title:    "Project created",
		Body:     "Draft is ready",
	}, []byte(`{"project_id":"p1"}`))
	require.NoError(t, err)
	firstID := uuid.MustParse(first.ID)
	require.False(t, first.IsRead)
	require.JSONEq(t, `{"project_id":"p1"}`, string(first.Payload))

	second, err := repo.Create(ctx, notifications.CreateInput{
		TenantID:  tenantID,
		UserID:    userID,
		Type:      "project.reviewed",
		Title:     "Reviewed",
		Body:      "Professor reviewed your project",
		WithEmail: true,
		EmailTo:   email,
		EmailSubj: "Reviewed",
		EmailBody: "Professor reviewed your project",
	}, []byte(`{"project_id":"p2"}`))
	require.NoError(t, err)

	items, err := repo.List(ctx, tenantID, userID, 10, 0)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.ElementsMatch(t, []string{first.ID, second.ID}, []string{items[0].ID, items[1].ID})

	count, err := repo.UnreadCount(ctx, tenantID, userID)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	require.NoError(t, repo.MarkRead(ctx, tenantID, userID, firstID))
	err = repo.MarkRead(ctx, tenantID, userID, uuid.New())
	require.ErrorIs(t, err, notifications.ErrNotFound)

	count, err = repo.UnreadCount(ctx, tenantID, userID)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	updated, err := repo.MarkAllRead(ctx, tenantID, userID)
	require.NoError(t, err)
	require.Equal(t, 1, updated)

	pending, err := repo.FetchPendingOutbox(ctx, time.Now().UTC().Add(time.Hour), 1)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, email, pending[0].EmailTo)
	require.Equal(t, "Reviewed", pending[0].Subject)
	require.Equal(t, 0, pending[0].Attempts)
	requireOutboxStatus(t, ctx, pool, pending[0].ID, "RETRY", 0, "")

	retryAt := time.Now().UTC().Add(time.Minute).Truncate(time.Second)
	require.NoError(t, repo.MarkOutboxFailed(ctx, pending[0].ID, 1, retryAt, "smtp unavailable", 3))
	requireOutboxStatus(t, ctx, pool, pending[0].ID, "RETRY", 1, "smtp unavailable")

	require.NoError(t, repo.MarkOutboxSent(ctx, pending[0].ID))
	requireOutboxStatus(t, ctx, pool, pending[0].ID, "SENT", 1, "")

	require.NoError(t, repo.Delete(ctx, tenantID, userID, firstID))
	err = repo.Delete(ctx, tenantID, userID, firstID)
	require.ErrorIs(t, err, notifications.ErrNotFound)

	deleted, err := repo.Clear(ctx, tenantID, userID)
	require.NoError(t, err)
	require.Equal(t, 1, deleted)

	deadNotification, err := repo.Create(ctx, notifications.CreateInput{
		TenantID:  tenantID,
		UserID:    userID,
		Type:      "project.deadline",
		Title:     "Deadline",
		Body:      "Deadline is near",
		WithEmail: true,
		EmailTo:   email,
		EmailSubj: "Deadline",
		EmailBody: "Deadline is near",
	}, []byte(`{}`))
	require.NoError(t, err)
	require.NotEmpty(t, deadNotification.ID)

	pending, err = repo.FetchPendingOutbox(ctx, time.Now().UTC().Add(time.Hour), 0)
	require.NoError(t, err)
	require.Len(t, pending, 1)

	require.NoError(t, repo.MarkOutboxFailed(ctx, pending[0].ID, 3, retryAt, "permanent failure", 3))
	requireOutboxStatus(t, ctx, pool, pending[0].ID, "DEAD", 3, "permanent failure")

	err = repo.MarkOutboxFailed(ctx, uuid.New(), 1, retryAt, "missing", 3)
	require.Error(t, err)
}

func requireOutboxStatus(t *testing.T, ctx context.Context, pool *pgxpool.Pool, outboxID uuid.UUID, wantStatus string, wantAttempts int, wantErr string) {
	t.Helper()

	var status string
	var attempts int
	var lastErr *string
	err := pool.QueryRow(ctx, `
SELECT status, attempts, last_error
FROM notification_outbox
WHERE id = $1;
`, outboxID).Scan(&status, &attempts, &lastErr)
	require.NoError(t, err)
	require.Equal(t, wantStatus, status)
	require.Equal(t, wantAttempts, attempts)
	if wantErr == "" {
		require.Nil(t, lastErr)
		return
	}
	require.NotNil(t, lastErr)
	require.Equal(t, wantErr, *lastErr)
}
