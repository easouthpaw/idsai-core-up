package handlers

import (
	"idsai-core-up/internal/services/admin"
)

type AdminHandler struct {
	svc      *admin.Service
	notifier NotificationPublisher
}

func NewAdminHandler(svc *admin.Service) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) SetNotifier(pub NotificationPublisher) {
	h.notifier = pub
}
