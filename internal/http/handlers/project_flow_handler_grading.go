package handlers

import (
	"errors"
	"net/http"

	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projectflow"

	"github.com/gin-gonic/gin"
)

func (h *ProjectFlowHandler) CreateCriterion(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req dto.CreateCriterionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.CreateCriterion(c.Request.Context(), uid, pid, req.Title, req.Description, req.Weight)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ProjectFlowCriterionResponseFromService(item))
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
	c.JSON(http.StatusOK, dto.ProjectFlowCriterionResponsesFromService(items))
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
	c.JSON(http.StatusOK, dto.GradingItemsResponse{Items: dto.ProjectFlowCriterionGradeResponsesFromService(items)})
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

	var req dto.UpsertGradingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	payload := dto.CriterionGradesFromRequest(req.Items)

	items, err := h.svc.UpsertGrading(c.Request.Context(), uid, pid, payload)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.GradingItemsResponse{Items: dto.ProjectFlowCriterionGradeResponsesFromService(items)})
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
	c.JSON(http.StatusOK, dto.ProjectFlowReadinessResponseFromService(item))
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
			c.JSON(http.StatusConflict, dto.ProjectReadinessConflictResponse{
				Error:     err.Error(),
				Readiness: dto.ProjectFlowReadinessResponseFromService(ready),
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

	c.JSON(http.StatusOK, dto.ApproveProjectResponse{
		Project:   dto.ProjectResponseFromDomain(p),
		Readiness: dto.ProjectFlowReadinessResponseFromService(ready),
	})
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

	c.JSON(http.StatusOK, dto.ProjectEnvelopeResponse{Project: dto.ProjectResponseFromDomain(p)})
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
		"Оценивание завершено, проект переведен в статус COMPLETED.",
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
			"Преподаватель завершил оценивание проекта. Проект находится в статусе COMPLETED.",
			map[string]any{
				"project_id": pid.String(),
				"title":      p.Title,
				"status":     p.Status,
			},
			true,
		))
	}

	c.JSON(http.StatusOK, dto.ProjectEnvelopeResponse{Project: dto.ProjectResponseFromDomain(p)})
}
