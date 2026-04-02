package dto

import (
	"encoding/json"
	"time"

	"idsai-core-up/internal/services/notifications"
)

type NotificationResponse struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Payload   json.RawMessage `json:"payload"`
	IsRead    bool            `json:"is_read"`
	CreatedAt time.Time       `json:"created_at"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
}

type ListNotificationsResponse struct {
	Items []NotificationResponse `json:"items"`
}

type UnreadCountResponse struct {
	Unread int `json:"unread"`
}

type NotificationsUpdatedResponse struct {
	Updated int `json:"updated"`
}

type NotificationsDeletedResponse struct {
	Deleted int `json:"deleted"`
}

func NotificationResponseFromService(item notifications.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        item.ID,
		Type:      item.Type,
		Title:     item.Title,
		Body:      item.Body,
		Payload:   item.Payload,
		IsRead:    item.IsRead,
		CreatedAt: item.CreatedAt,
		ReadAt:    item.ReadAt,
	}
}

func NotificationResponsesFromService(items []notifications.Notification) []NotificationResponse {
	if items == nil {
		return nil
	}
	out := make([]NotificationResponse, 0, len(items))
	for _, item := range items {
		out = append(out, NotificationResponseFromService(item))
	}
	return out
}
