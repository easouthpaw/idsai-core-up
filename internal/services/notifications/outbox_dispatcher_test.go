package notifications

import (
	"bytes"
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeOutboxRepo struct {
	items         []OutboxItem
	markFailedID  uuid.UUID
	markAttempts  int
	markRetryAt   time.Time
	markLastErr   string
	markMaxReties int
}

func (f *fakeOutboxRepo) FetchPendingOutbox(ctx context.Context, now time.Time, limit int) ([]OutboxItem, error) {
	return f.items, nil
}

func (f *fakeOutboxRepo) MarkOutboxSent(ctx context.Context, outboxID uuid.UUID) error {
	return nil
}

func (f *fakeOutboxRepo) MarkOutboxFailed(ctx context.Context, outboxID uuid.UUID, attempts int, retryAt time.Time, lastErr string, maxAttempts int) error {
	f.markFailedID = outboxID
	f.markAttempts = attempts
	f.markRetryAt = retryAt
	f.markLastErr = lastErr
	f.markMaxReties = maxAttempts
	return nil
}

type fakeEmailSender struct {
	err error
}

func (f fakeEmailSender) Send(ctx context.Context, to, subject, body string) error {
	return f.err
}

func captureNotificationLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	prevWriter := log.Writer()
	prevFlags := log.Flags()
	prevPrefix := log.Prefix()
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(prevWriter)
		log.SetFlags(prevFlags)
		log.SetPrefix(prevPrefix)
	})
	return &buf
}

func TestOutboxDispatcherLogsSendFailure(t *testing.T) {
	itemID := uuid.New()
	repo := &fakeOutboxRepo{
		items: []OutboxItem{{
			ID:       itemID,
			TenantID: uuid.New(),
			EmailTo:  "student@example.edu",
			Subject:  "Reset password",
			Body:     "Code: 123456",
		}},
	}
	dispatcher := NewOutboxDispatcher(repo, fakeEmailSender{
		err: errors.New("smtp auth failed"),
	})
	logs := captureNotificationLogs(t)

	dispatcher.runBatch(context.Background())

	require.Equal(t, itemID, repo.markFailedID)
	require.Equal(t, 1, repo.markAttempts)
	require.Equal(t, "smtp auth failed", repo.markLastErr)
	require.Contains(t, logs.String(), "notifications outbox send failed")
	require.Contains(t, logs.String(), "student@example.edu")
	require.Contains(t, logs.String(), "smtp auth failed")
}
