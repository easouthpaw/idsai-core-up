package handlers

import (
	"context"
	"log"
	"strings"

	notifsvc "idsai-core-up/internal/services/notifications"

	"github.com/google/uuid"
)

// NotificationPublisher is a narrow interface used by HTTP handlers
// to publish in-app/email notifications after successful actions.
type NotificationPublisher interface {
	Notify(ctx context.Context, in notifsvc.CreateInput) (notifsvc.Notification, error)
}

func notifyBestEffort(pub NotificationPublisher, ctx context.Context, in notifsvc.CreateInput) {
	if pub == nil {
		return
	}
	if in.TenantID == uuid.Nil || in.UserID == uuid.Nil {
		return
	}
	in.Type = strings.TrimSpace(in.Type)
	in.Title = strings.TrimSpace(in.Title)
	in.Body = strings.TrimSpace(in.Body)
	if in.Type == "" || in.Title == "" || in.Body == "" {
		return
	}
	if in.Payload == nil {
		in.Payload = map[string]any{}
	}
	if in.WithEmail {
		if strings.TrimSpace(in.EmailSubj) == "" {
			in.EmailSubj = in.Title
		}
		if strings.TrimSpace(in.EmailBody) == "" {
			in.EmailBody = in.Body
		}
	}
	if _, err := pub.Notify(ctx, in); err != nil {
		log.Printf("notification publish failed: type=%s user=%s err=%v", in.Type, in.UserID, err)
	}
}

func notifCreateInput(tenantID, userID uuid.UUID, typ, title, body string, payload map[string]any, withEmail bool) notifsvc.CreateInput {
	return notifsvc.CreateInput{
		TenantID:  tenantID,
		UserID:    userID,
		Type:      typ,
		Title:     title,
		Body:      body,
		Payload:   payload,
		WithEmail: withEmail,
		EmailSubj: title,
		EmailBody: body,
	}
}
