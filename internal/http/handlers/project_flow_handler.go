package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projectflow"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProjectFlowHandler struct {
	svc      *projectflow.Service
	notifier NotificationPublisher
}

func NewProjectFlowHandler(svc *projectflow.Service) *ProjectFlowHandler {
	return &ProjectFlowHandler{svc: svc}
}

func (h *ProjectFlowHandler) SetNotifier(pub NotificationPublisher) {
	h.notifier = pub
}

func parseUserID(c *gin.Context) (uuid.UUID, bool) {
	id, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}
	return id, true
}

func parseProjectID(c *gin.Context) (uuid.UUID, bool) {
	raw := strings.TrimSpace(c.Param("project_id"))
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return uuid.Nil, false
	}
	return id, true
}

func handleFlowErr(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, projectflow.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrRecruitmentOpen):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrProjectNotReady):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrProjectNotActive):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, projectflow.ErrPositionFull):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, pgx.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

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

func (h *ProjectFlowHandler) ApplyMember(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	item, err := h.svc.ApplyMember(c.Request.Context(), uid, pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *ProjectFlowHandler) ListMembers(c *gin.Context) {
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	items, err := h.svc.ListMembers(c.Request.Context(), pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

type approveMemberReq struct {
	PositionID string `json:"position_id" binding:"required"`
}

func parseUserIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	raw := strings.TrimSpace(c.Param(param))
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

func (h *ProjectFlowHandler) ApproveMember(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	memberID, ok := parseUserIDParam(c, "user_id")
	if !ok {
		return
	}

	var req approveMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	positionID, err := uuid.Parse(strings.TrimSpace(req.PositionID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position_id"})
		return
	}

	item, err := h.svc.ApproveMember(c.Request.Context(), uid, pid, memberID, positionID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ProjectFlowHandler) SetMemberPosition(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	memberID, ok := parseUserIDParam(c, "user_id")
	if !ok {
		return
	}

	var req approveMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	positionID, err := uuid.Parse(strings.TrimSpace(req.PositionID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position_id"})
		return
	}

	item, err := h.svc.SetMemberPosition(c.Request.Context(), uid, pid, memberID, positionID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

type assignProfessorReq struct {
	ProfessorID string `json:"professor_id" binding:"required"`
}

func (h *ProjectFlowHandler) AssignProfessor(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req assignProfessorReq
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
	c.JSON(http.StatusOK, projectToResponse(p))
}

type createCriterionReq struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Weight      int    `json:"weight"`
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

type createTaskReq struct {
	Title          string  `json:"title" binding:"required"`
	Description    string  `json:"description"`
	PositionID     string  `json:"position_id" binding:"required"`
	AssigneeUserID *string `json:"assignee_user_id,omitempty"`
	DueAt          *string `json:"due_at,omitempty"` // RFC3339
}

func (h *ProjectFlowHandler) CreateTask(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req createTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	positionID, err := uuid.Parse(strings.TrimSpace(req.PositionID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position_id"})
		return
	}

	var assignee *uuid.UUID
	if req.AssigneeUserID != nil && strings.TrimSpace(*req.AssigneeUserID) != "" {
		x, err := uuid.Parse(strings.TrimSpace(*req.AssigneeUserID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee_user_id"})
			return
		}
		assignee = &x
	}

	var dueAt *time.Time
	if req.DueAt != nil && strings.TrimSpace(*req.DueAt) != "" {
		tm, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.DueAt))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_at (RFC3339 expected)"})
			return
		}
		dueAt = &tm
	}

	item, err := h.svc.CreateTask(c.Request.Context(), uid, pid, req.Title, req.Description, positionID, assignee, dueAt)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *ProjectFlowHandler) ListTasks(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	items, err := h.svc.ListTasks(c.Request.Context(), uid, pid)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

type updateTaskStatusReq struct {
	Status string `json:"status" binding:"required"`
}

func (h *ProjectFlowHandler) UpdateTaskStatus(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	tid, ok := parseUserIDParam(c, "task_id")
	if !ok {
		return
	}

	var req updateTaskStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.UpdateTaskStatus(c.Request.Context(), uid, pid, tid, req.Status)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

type assignTaskReq struct {
	AssigneeUserID string `json:"assignee_user_id" binding:"required"`
}

func (h *ProjectFlowHandler) AssignTask(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	tid, ok := parseUserIDParam(c, "task_id")
	if !ok {
		return
	}

	var req assignTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	assigneeID, err := uuid.Parse(strings.TrimSpace(req.AssigneeUserID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee_user_id"})
		return
	}

	item, err := h.svc.AssignTask(c.Request.Context(), uid, pid, tid, assigneeID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *ProjectFlowHandler) ClaimTask(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	tid, ok := parseUserIDParam(c, "task_id")
	if !ok {
		return
	}

	if err := h.svc.ClaimTask(c.Request.Context(), uid, pid, tid); err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "claimed"})
}
