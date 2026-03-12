package handlers

import (
	"net/http"

	"idsai-core-up/internal/http/middleware"

	"github.com/gin-gonic/gin"
)

type updateProjectReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

func (h *ProjectFlowHandler) UpdateProject(c *gin.Context) {
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

	var req updateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	p, err := h.svc.UpdateProject(c.Request.Context(), uid, pid, req.Title, req.Description)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
		tenantID,
		uid,
		"project.updated",
		"Изменения проекта сохранены",
		"Данные проекта успешно обновлены.",
		map[string]any{
			"project_id": pid.String(),
			"title":      p.Title,
		},
		true,
	))

	c.JSON(http.StatusOK, projectToResponse(p))
}

func (h *ProjectFlowHandler) DeleteProject(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	if err := h.svc.DeleteProject(c.Request.Context(), uid, pid); err != nil {
		handleFlowErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type setStacksReq struct {
	Stacks []string `json:"stacks"`
}

func (h *ProjectFlowHandler) SetStacks(c *gin.Context) {
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

	var req setStacksReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	items, err := h.svc.SetStacks(c.Request.Context(), uid, pid, req.Stacks)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
		tenantID,
		uid,
		"project.stacks.updated",
		"Стек проекта обновлён",
		"Технологический стек проекта успешно сохранён.",
		map[string]any{
			"project_id": pid.String(),
			"stacks":     req.Stacks,
		},
		false,
	))

	c.JSON(http.StatusOK, items)
}

func (h *ProjectFlowHandler) ListStacks(c *gin.Context) {
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	items, err := h.svc.ListStacks(c.Request.Context(), pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ProjectFlowHandler) OpenRecruitment(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	p, err := h.svc.OpenRecruitment(c.Request.Context(), uid, pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, projectToResponse(p))
}

type createPositionReq struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

func (h *ProjectFlowHandler) CreatePosition(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req createPositionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.CreatePosition(c.Request.Context(), uid, pid, req.Code, req.Name, req.Capacity)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *ProjectFlowHandler) ListPositions(c *gin.Context) {
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	items, err := h.svc.ListPositions(c.Request.Context(), pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ProjectFlowHandler) ListStudentCandidates(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	q := c.Query("q")
	limit := parseLimit(c.Query("limit"), 30, 100)

	items, err := h.svc.ListStudentCandidates(c.Request.Context(), uid, pid, q, limit)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}
