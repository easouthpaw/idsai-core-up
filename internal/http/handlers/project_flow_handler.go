package handlers

import (
	"errors"
	"net/http"
	"strconv"
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
	case errors.Is(err, projectflow.ErrInviteNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, pgx.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func parseLimit(raw string, def, max int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return def
	}
	if v > max {
		return max
	}
	return v
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

type inviteMemberReq struct {
	UserID  string `json:"user_id" binding:"required"`
	Comment string `json:"comment"`
}

func (h *ProjectFlowHandler) InviteMember(c *gin.Context) {
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

	var req inviteMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	memberID, err := uuid.Parse(strings.TrimSpace(req.UserID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	item, err := h.svc.InviteStudent(c.Request.Context(), uid, pid, memberID, req.Comment)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
		tenantID,
		memberID,
		"project.member.invited",
		"Новое приглашение в команду",
		"Вас пригласили в проект. Вы можете принять или отклонить приглашение.",
		map[string]any{
			"project_id": pid.String(),
			"comment":    strings.TrimSpace(req.Comment),
		},
		true,
	))

	c.JSON(http.StatusOK, item)
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

type respondInviteReq struct {
	Accept bool `json:"accept"`
}

func (h *ProjectFlowHandler) RespondMemberInvite(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	var req respondInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.RespondMemberInvite(c.Request.Context(), uid, pid, req.Accept)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
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
	c.JSON(http.StatusOK, items)
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
	c.JSON(http.StatusOK, gin.H{"professor": item})
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

	c.JSON(http.StatusOK, projectToResponse(p))
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

	var req respondInviteReq
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

	c.JSON(http.StatusOK, projectToResponse(p))
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

	out := make([]projectResponse, 0, len(items))
	for _, item := range items {
		out = append(out, projectToResponse(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *ProjectFlowHandler) ListIncomingInvites(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	limit := parseLimit(c.Query("limit"), 50, 100)

	items, err := h.svc.ListIncomingInvites(c.Request.Context(), uid, limit)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *ProjectFlowHandler) ListOutgoingApplications(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	limit := parseLimit(c.Query("limit"), 50, 100)

	items, err := h.svc.ListOutgoingApplications(c.Request.Context(), uid, limit)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

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

func (h *ProjectFlowHandler) ListTaskActivities(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}

	rawTaskID := strings.TrimSpace(c.Query("task_id"))
	var taskID *uuid.UUID
	if rawTaskID != "" {
		tid, err := uuid.Parse(rawTaskID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_id"})
			return
		}
		taskID = &tid
	}

	items, err := h.svc.ListTaskActivities(c.Request.Context(), uid, pid, taskID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
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

type completeTaskReq struct {
	Comment     string   `json:"comment"`
	Attachments []string `json:"attachments"`
}

func (h *ProjectFlowHandler) CompleteTask(c *gin.Context) {
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

	var req completeTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.CompleteTask(c.Request.Context(), uid, pid, tid, req.Comment, req.Attachments)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}
