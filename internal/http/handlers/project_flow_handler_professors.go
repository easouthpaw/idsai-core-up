package handlers

import (
	"net/http"
	"strings"

	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *ProjectFlowHandler) SearchProfessors(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	q := c.Query("q")
	limit := parseLimit(c.Query("limit"), 20, 50)

	items, err := h.svc.SearchProfessors(c.Request.Context(), uid, pid, q, limit)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ProjectFlowProfessorCandidateResponsesFromService(items))
}

func (h *ProjectFlowHandler) GetAssignedProfessor(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	item, err := h.svc.GetAssignedProfessor(c.Request.Context(), uid, pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.AssignedProfessorResponseFromService(item))
}

func (h *ProjectFlowHandler) AssignProfessor(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	tenantID, ok := middleware.TenantIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req dto.AssignProfessorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	profID, err := uuid.Parse(strings.TrimSpace(req.ProfessorID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid professor_id"})
		return
	}

	p, err := h.svc.AssignProfessor(c.Request.Context(), uid, pid, profID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
		tenantID,
		profID,
		"project.professor.invited",
		"Запрос на ревью проекта",
		"Вас пригласили преподавателем-ревьюером. Подтвердите приглашение.",
		map[string]any{
			"project_id": pid.String(),
			"title":      p.Title,
		},
		true,
	))

	c.JSON(http.StatusOK, dto.ProjectResponseFromDomain(p))
}

func (h *ProjectFlowHandler) RespondProfessorInvite(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	tenantID, ok := middleware.TenantIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req dto.RespondInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	p, err := h.svc.RespondProfessorInvite(c.Request.Context(), uid, pid, req.Accept)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	title := "Преподаватель отклонил приглашение"
	body := "Приглашение преподавателя в проект отклонено."
	typ := "project.professor.rejected"
	if req.Accept {
		title = "Преподаватель подтвердил участие"
		body = "Преподаватель принял приглашение и подключился к ревью."
		typ = "project.professor.accepted"
	}

	notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
		tenantID,
		p.CreatedBy,
		typ,
		title,
		body,
		map[string]any{
			"project_id": pid.String(),
			"title":      p.Title,
		},
		true,
	))

	c.JSON(http.StatusOK, dto.ProjectResponseFromDomain(p))
}

func (h *ProjectFlowHandler) ListProfessorReviewInvites(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	q := c.Query("q")
	limit := parseLimit(c.Query("limit"), 100, 100)

	items, err := h.svc.ListProfessorReviewInvites(c.Request.Context(), uid, q, limit)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	out := make([]dto.ProjectResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.ProjectResponseFromDomain(item))
	}
	c.JSON(http.StatusOK, out)
}
