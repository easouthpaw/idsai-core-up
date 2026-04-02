package handlers

import (
	"idsai-core-up/internal/services/projects"
)

type ProjectsHandler struct {
	svc      *projects.Service
	notifier NotificationPublisher
}

func NewProjectsHandler(svc *projects.Service) *ProjectsHandler {
	return &ProjectsHandler{svc: svc}
}

func (h *ProjectsHandler) SetNotifier(pub NotificationPublisher) {
	h.notifier = pub
}
