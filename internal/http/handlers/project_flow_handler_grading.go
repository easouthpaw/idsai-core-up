package handlers

import (
	"errors"
	"net/http"
	"strings"

	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projectflow"

	"github.com/gin-gonic/gin"
)

type createCriterionReq struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Weight      int    `json:"weight"`
}

type gradingItemReq struct {
	CriterionID string `json:"criterion_id" binding:"required"`
	IsMet       *bool  `json:"is_met"`
	Comment     string `json:"comment"`
}

type upsertGradingReq struct {
	Items []gradingItemReq `json:"items"`
}

func (h *ProjectFlowHandler) CreateCriterion(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req createCriterionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.CreateCriterion(c.Request.Context(), uid, pid, req.Title, req.Description, req.Weight)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *ProjectFlowHandler) ListCriteria(c *gin.Context) {
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	items, err := h.svc.ListCriteria(c.Request.Context(), pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ProjectFlowHandler) GetGrading(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	items, err := h.svc.GetGrading(c.Request.Context(), uid, pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *ProjectFlowHandler) UpsertGrading(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req upsertGradingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	payload := make([]projectflow.CriterionGrade, 0, len(req.Items))
	for _, item := range req.Items {
		payload = append(payload, projectflow.CriterionGrade{
			CriterionID: strings.TrimSpace(item.CriterionID),
			IsMet:       item.IsMet,
			Comment:     strings.TrimSpace(item.Comment),
		})
	}

	items, err := h.svc.UpsertGrading(c.Request.Context(), uid, pid, payload)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *ProjectFlowHandler) Readiness(c *gin.Context) {
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	item, err := h.svc.Readiness(c.Request.Context(), pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ProjectFlowHandler) ApproveProject(c *gin.Context) {
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
	p, ready, err := h.svc.ApproveProject(c.Request.Context(), uid, pid)
	if err != nil {
		if errors.Is(err, projectflow.ErrProjectNotReady) {
			c.JSON(http.StatusConflict, gin.H{
				"error":     err.Error(),
				"readiness": ready,
			})
			return
		}
		handleFlowErr(c, err)
		return
	}

	notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
		tenantID,
		uid,
		"project.activated",
		"Проект активирован",
		"Проект переведён в статус ACTIVE.",
		map[string]any{
			"project_id": pid.String(),
			"title":      p.Title,
		},
		true,
	))

	c.JSON(http.StatusOK, gin.H{"project": projectToResponse(p), "readiness": ready})
}

func (h *ProjectFlowHandler) SubmitProjectForGrading(c *gin.Context) {
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

	p, err := h.svc.SubmitProjectForGrading(c.Request.Context(), uid, pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
		tenantID,
		uid,
		"project.sent_to_grading",
		"Проект отправлен на оценивание",
		"Проект переведен в статус GRADING и отправлен преподавателю.",
		map[string]any{
			"project_id": pid.String(),
			"title":      p.Title,
			"status":     p.Status,
		},
		true,
	))

	if p.ProfessorID != nil {
		notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
			tenantID,
			*p.ProfessorID,
			"project.grading.requested",
			"Новый проект на оценивание",
			"Команда завершила проект и отправила его на оценивание.",
			map[string]any{
				"project_id": pid.String(),
				"title":      p.Title,
				"status":     p.Status,
			},
			true,
		))
	}

	c.JSON(http.StatusOK, gin.H{"project": projectToResponse(p)})
}

func (h *ProjectFlowHandler) PublishProjectGrading(c *gin.Context) {
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

	p, err := h.svc.PublishGrading(c.Request.Context(), uid, pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
		tenantID,
		uid,
		"project.grading.published",
		"Итоговая оценка опубликована",
		"Оценивание завершено, проект переведен в статус ARCHIVE.",
		map[string]any{
			"project_id": pid.String(),
			"title":      p.Title,
			"status":     p.Status,
		},
		false,
	))

	if p.CreatedBy != uid {
		notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
			tenantID,
			p.CreatedBy,
			"project.finished",
			"Проект завершен",
			"Преподаватель завершил оценивание проекта. Проект находится в статусе ARCHIVE.",
			map[string]any{
				"project_id": pid.String(),
				"title":      p.Title,
				"status":     p.Status,
			},
			true,
		))
	}

	c.JSON(http.StatusOK, gin.H{"project": projectToResponse(p)})
}
