package postgres

import (
	"context"
	"errors"
	"time"

	"idsai-core-up/internal/services/notifications"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationsRepo struct {
	db *pgxpool.Pool
}

func NewNotificationsRepo(db *pgxpool.Pool) *NotificationsRepo {
	return &NotificationsRepo{db: db}
}

func (r *NotificationsRepo) Create(ctx context.Context, in notifications.CreateInput, payload []byte) (notifications.Notification, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return notifications.Notification{}, err
	}
	defer tx.Rollback(ctx)

	var out notifications.Notification
	var nid uuid.UUID
	err = tx.QueryRow(ctx, `
INSERT INTO notifications(tenant_id, user_id, type, title, body, payload)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, type, title, body, payload, is_read, created_at, read_at
`, in.TenantID, in.UserID, in.Type, in.Title, in.Body, payload).Scan(
		&nid,
		&out.Type,
		&out.Title,
		&out.Body,
		&out.Payload,
		&out.IsRead,
		&out.CreatedAt,
		&out.ReadAt,
	)
	if err != nil {
		return notifications.Notification{}, err
	}
	out.ID = nid.String()

	if in.WithEmail {
		_, err = tx.Exec(ctx, `
INSERT INTO notification_outbox(tenant_id, notification_id, user_id, email_to, subject, body, status)
VALUES ($1, $2, $3, $4, $5, $6, 'PENDING')
`, in.TenantID, nid, in.UserID, in.EmailTo, in.EmailSubj, in.EmailBody)
		if err != nil {
			return notifications.Notification{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return notifications.Notification{}, err
	}
	return out, nil
}

func (r *NotificationsRepo) List(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]notifications.Notification, error) {
	rows, err := r.db.Query(ctx, `
SELECT id, type, title, body, payload, is_read, created_at, read_at
FROM notifications
WHERE tenant_id = $1
  AND user_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4
`, tenantID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]notifications.Notification, 0, limit)
	for rows.Next() {
		var item notifications.Notification
		var id uuid.UUID
		if err := rows.Scan(
			&id,
			&item.Type,
			&item.Title,
			&item.Body,
			&item.Payload,
			&item.IsRead,
			&item.CreatedAt,
			&item.ReadAt,
		); err != nil {
			return nil, err
		}
		item.ID = id.String()
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *NotificationsRepo) UnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
SELECT COUNT(*)
FROM notifications
WHERE tenant_id = $1
  AND user_id = $2
  AND is_read = FALSE
`, tenantID, userID).Scan(&count)
	return count, err
}

func (r *NotificationsRepo) MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
UPDATE notifications
SET is_read = TRUE,
    read_at = now()
WHERE tenant_id = $1
  AND user_id = $2
  AND id = $3
`, tenantID, userID, notificationID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return notifications.ErrNotFound
	}
	return nil
}

func (r *NotificationsRepo) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	tag, err := r.db.Exec(ctx, `
UPDATE notifications
SET is_read = TRUE,
    read_at = COALESCE(read_at, now())
WHERE tenant_id = $1
  AND user_id = $2
  AND is_read = FALSE
`, tenantID, userID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *NotificationsRepo) Delete(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
DELETE FROM notifications
WHERE tenant_id = $1
  AND user_id = $2
  AND id = $3
`, tenantID, userID, notificationID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return notifications.ErrNotFound
	}
	return nil
}

func (r *NotificationsRepo) Clear(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	tag, err := r.db.Exec(ctx, `
DELETE FROM notifications
WHERE tenant_id = $1
  AND user_id = $2
`, tenantID, userID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (r *NotificationsRepo) UserEmail(ctx context.Context, tenantID, userID uuid.UUID) (string, error) {
	var email string
	err := r.db.QueryRow(ctx, `
SELECT email
FROM users
WHERE tenant_id = $1
  AND id = $2
`, tenantID, userID).Scan(&email)
	return email, err
}

func (r *NotificationsRepo) FetchPendingOutbox(ctx context.Context, now time.Time, limit int) ([]notifications.OutboxItem, error) {
	if limit <= 0 {
		limit = 100
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
SELECT id, tenant_id, email_to, subject, body, attempts
FROM notification_outbox
WHERE status IN ('PENDING', 'RETRY')
  AND next_attempt_at <= $1
ORDER BY next_attempt_at ASC
FOR UPDATE SKIP LOCKED
LIMIT $2
`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]notifications.OutboxItem, 0, limit)
	for rows.Next() {
		var it notifications.OutboxItem
		if err := rows.Scan(&it.ID, &it.TenantID, &it.EmailTo, &it.Subject, &it.Body, &it.Attempts); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return items, nil
	}

	ids := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
	}
	_, err = tx.Exec(ctx, `
UPDATE notification_outbox
SET status = 'RETRY'
WHERE id = ANY($1::uuid[])
`, ids)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *NotificationsRepo) MarkOutboxSent(ctx context.Context, outboxID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
UPDATE notification_outbox
SET status = 'SENT',
    sent_at = now(),
    last_error = NULL
WHERE id = $1
`, outboxID)
	return err
}

func (r *NotificationsRepo) MarkOutboxFailed(ctx context.Context, outboxID uuid.UUID, attempts int, retryAt time.Time, lastErr string, maxAttempts int) error {
	status := "RETRY"
	if attempts >= maxAttempts {
		status = "DEAD"
	}
	tag, err := r.db.Exec(ctx, `
UPDATE notification_outbox
SET status = $2,
    attempts = $3,
    next_attempt_at = $4,
    last_error = $5
WHERE id = $1
`, outboxID, status, attempts, retryAt, lastErr)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("outbox item not found")
	}
	return nil
}
