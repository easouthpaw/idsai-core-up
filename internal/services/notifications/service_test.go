package notifications_test

import (
	"context"
	"testing"
	"time"

	"idsai-core-up/internal/services/notifications"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	createIn     notifications.CreateInput
	createBytes  []byte
	created      notifications.Notification
	createErr    error
	userEmail    string
	userEmailErr error
	listLimit    int
	listOffset   int
	listOut      []notifications.Notification
	listErr      error
	unread       int
	unreadErr    error
	markErr      error
	markAll      int
	markAllErr   error
	deleteErr    error
	clearCount   int
	clearErr     error
}

func (f *fakeRepo) Create(ctx context.Context, in notifications.CreateInput, payload []byte) (notifications.Notification, error) {
	f.createIn = in
	f.createBytes = payload
	return f.created, f.createErr
}

func (f *fakeRepo) List(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]notifications.Notification, error) {
	f.listLimit = limit
	f.listOffset = offset
	return f.listOut, f.listErr
}

func (f *fakeRepo) UnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return f.unread, f.unreadErr
}

func (f *fakeRepo) MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	return f.markErr
}

func (f *fakeRepo) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return f.markAll, f.markAllErr
}

func (f *fakeRepo) Delete(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	return f.deleteErr
}

func (f *fakeRepo) Clear(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	return f.clearCount, f.clearErr
}

func (f *fakeRepo) UserEmail(ctx context.Context, tenantID, userID uuid.UUID) (string, error) {
	return f.userEmail, f.userEmailErr
}

func TestNotify_WithEmailRequiresFields(t *testing.T) {
	repo := &fakeRepo{}
	svc := notifications.NewService(repo)

	_, err := svc.Notify(context.Background(), notifications.CreateInput{
		TenantID:  uuid.New(),
		UserID:    uuid.New(),
		Type:      "project.approved",
		Title:     "approved",
		Body:      "done",
		WithEmail: true,
	})
	require.ErrorIs(t, err, notifications.ErrInvalidInput)
}

func TestNotify_WithEmailAutoResolvesRecipient(t *testing.T) {
	repo := &fakeRepo{
		created:   notifications.Notification{ID: uuid.NewString()},
		userEmail: "student@example.com",
	}
	svc := notifications.NewService(repo)

	_, err := svc.Notify(context.Background(), notifications.CreateInput{
		TenantID:  uuid.New(),
		UserID:    uuid.New(),
		Type:      "project.updated",
		Title:     "updated",
		Body:      "ok",
		WithEmail: true,
		EmailSubj: "Updated",
		EmailBody: "Project updated",
	})
	require.NoError(t, err)
	require.Equal(t, "student@example.com", repo.createIn.EmailTo)
}

func TestNotify_Success(t *testing.T) {
	now := time.Now()
	want := notifications.Notification{
		ID:        uuid.NewString(),
		Type:      "project.approved",
		Title:     "approved",
		Body:      "done",
		CreatedAt: now,
	}
	repo := &fakeRepo{created: want}
	svc := notifications.NewService(repo)

	got, err := svc.Notify(context.Background(), notifications.CreateInput{
		TenantID:  uuid.New(),
		UserID:    uuid.New(),
		Type:      "project.approved",
		Title:     "approved",
		Body:      "done",
		Payload:   map[string]any{"project_id": uuid.NewString()},
		WithEmail: true,
		EmailTo:   "student@example.com",
		EmailSubj: "Project approved",
		EmailBody: "Congrats",
	})
	require.NoError(t, err)
	require.Equal(t, want.ID, got.ID)
	require.NotEmpty(t, repo.createBytes)
}

func TestList_ClampsLimit(t *testing.T) {
	repo := &fakeRepo{listOut: []notifications.Notification{}}
	svc := notifications.NewService(repo)

	_, err := svc.List(context.Background(), uuid.New(), uuid.New(), 9999, -10)
	require.NoError(t, err)
	require.Equal(t, 200, repo.listLimit)
	require.Equal(t, 0, repo.listOffset)
}
