package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidInput = errors.New("invalid input")

type Notification struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Payload   json.RawMessage `json:"payload"`
	IsRead    bool            `json:"is_read"`
	CreatedAt time.Time       `json:"created_at"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
}

type CreateInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	Type       string
	Title      string
	Body       string
	Payload    map[string]any
	EmailTo    string
	EmailSubj  string
	EmailBody  string
	WithEmail  bool
	ForceInbox bool
}

type Repository interface {
	Create(ctx context.Context, in CreateInput, payload []byte) (Notification, error)
	List(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]Notification, error)
	UnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int, error)
	MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error
	MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) (int, error)
	Delete(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error
	Clear(ctx context.Context, tenantID, userID uuid.UUID) (int, error)
	UserEmail(ctx context.Context, tenantID, userID uuid.UUID) (string, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Notify(ctx context.Context, in CreateInput) (Notification, error) {
	if in.TenantID == uuid.Nil || in.UserID == uuid.Nil {
		return Notification{}, ErrInvalidInput
	}
	in.Type = strings.TrimSpace(in.Type)
	in.Title = strings.TrimSpace(in.Title)
	in.Body = strings.TrimSpace(in.Body)
	in.EmailTo = strings.TrimSpace(in.EmailTo)
	in.EmailSubj = strings.TrimSpace(in.EmailSubj)
	in.EmailBody = strings.TrimSpace(in.EmailBody)
	if in.Type == "" || in.Title == "" || in.Body == "" {
		return Notification{}, ErrInvalidInput
	}
	if in.WithEmail && in.EmailTo == "" {
		email, err := s.repo.UserEmail(ctx, in.TenantID, in.UserID)
		if err != nil {
			return Notification{}, err
		}
		in.EmailTo = strings.TrimSpace(email)
	}
	if in.WithEmail && (in.EmailTo == "" || in.EmailSubj == "" || in.EmailBody == "") {
		return Notification{}, ErrInvalidInput
	}
	if in.Payload == nil {
		in.Payload = map[string]any{}
	}
	payload, err := json.Marshal(in.Payload)
	if err != nil {
		return Notification{}, err
	}
	return s.repo.Create(ctx, in, payload)
}

func (s *Service) List(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]Notification, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, tenantID, userID, limit, offset)
}

func (s *Service) UnreadCount(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return 0, ErrInvalidInput
	}
	return s.repo.UnreadCount(ctx, tenantID, userID)
}

func (s *Service) MarkRead(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	if tenantID == uuid.Nil || userID == uuid.Nil || notificationID == uuid.Nil {
		return ErrInvalidInput
	}
	return s.repo.MarkRead(ctx, tenantID, userID, notificationID)
}

func (s *Service) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return 0, ErrInvalidInput
	}
	return s.repo.MarkAllRead(ctx, tenantID, userID)
}

func (s *Service) Delete(ctx context.Context, tenantID, userID, notificationID uuid.UUID) error {
	if tenantID == uuid.Nil || userID == uuid.Nil || notificationID == uuid.Nil {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, tenantID, userID, notificationID)
}

func (s *Service) Clear(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return 0, ErrInvalidInput
	}
	return s.repo.Clear(ctx, tenantID, userID)
}
