package notifications

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

type OutboxItem struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	EmailTo  string
	Subject  string
	Body     string
	Attempts int
}

type OutboxRepository interface {
	FetchPendingOutbox(ctx context.Context, now time.Time, limit int) ([]OutboxItem, error)
	MarkOutboxSent(ctx context.Context, outboxID uuid.UUID) error
	MarkOutboxFailed(ctx context.Context, outboxID uuid.UUID, attempts int, retryAt time.Time, lastErr string, maxAttempts int) error
}

type OutboxDispatcher struct {
	repo          OutboxRepository
	email         EmailSender
	maxAttempts   int
	batchSize     int
	baseRetryWait time.Duration
	sendTimeout   time.Duration
}

func NewOutboxDispatcher(repo OutboxRepository, email EmailSender) *OutboxDispatcher {
	return &OutboxDispatcher{
		repo:          repo,
		email:         email,
		maxAttempts:   8,
		batchSize:     100,
		baseRetryWait: 15 * time.Second,
		sendTimeout:   15 * time.Second,
	}
}

func (d *OutboxDispatcher) SetSendTimeout(timeout time.Duration) {
	d.sendTimeout = timeout
}

func (d *OutboxDispatcher) Start(ctx context.Context, poll time.Duration) {
	if d.repo == nil || d.email == nil || poll <= 0 {
		return
	}
	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	d.runBatch(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.runBatch(ctx)
		}
	}
}

func (d *OutboxDispatcher) runBatch(ctx context.Context) {
	items, err := d.repo.FetchPendingOutbox(ctx, time.Now(), d.batchSize)
	if err != nil {
		log.Printf("notifications outbox fetch failed: %v", err)
		return
	}
	for _, it := range items {
		sendCtx := ctx
		cancel := func() {}
		if d.sendTimeout > 0 {
			sendCtx, cancel = context.WithTimeout(ctx, d.sendTimeout)
		}
		sendErr := d.email.Send(sendCtx, it.EmailTo, it.Subject, it.Body)
		cancel()
		if sendErr == nil {
			if err := d.repo.MarkOutboxSent(ctx, it.ID); err != nil {
				log.Printf("notifications outbox mark sent failed: id=%s err=%v", it.ID, err)
			}
			continue
		}

		attempts := it.Attempts + 1
		wait := d.retryDelay(attempts)
		retryAt := time.Now().Add(wait)
		log.Printf("notifications outbox send failed: id=%s email_to=%s attempts=%d err=%v", it.ID, it.EmailTo, attempts, sendErr)
		if err := d.repo.MarkOutboxFailed(ctx, it.ID, attempts, retryAt, sendErr.Error(), d.maxAttempts); err != nil {
			log.Printf("notifications outbox mark failed failed: id=%s err=%v", it.ID, err)
		}
	}
}

func (d *OutboxDispatcher) retryDelay(attempt int) time.Duration {
	if attempt <= 1 {
		return d.baseRetryWait
	}
	wait := d.baseRetryWait
	for i := 1; i < attempt; i++ {
		wait *= 2
		if wait >= 10*time.Minute {
			return 10 * time.Minute
		}
	}
	return wait
}
