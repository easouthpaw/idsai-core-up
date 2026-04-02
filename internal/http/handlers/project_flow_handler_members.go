package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

	var req dto.InviteMemberRequest
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

	c.JSON(http.StatusOK, dto.ProjectFlowMemberResponseFromService(item))
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

	var req dto.ApplyMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.ApplyMember(c.Request.Context(), uid, pid, req.Comment)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ProjectFlowMemberResponseFromService(item))
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

	var req dto.RespondInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	item, err := h.svc.RespondMemberInvite(c.Request.Context(), uid, pid, req.Accept)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ProjectFlowMemberResponseFromService(item))
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
	c.JSON(http.StatusOK, dto.ProjectFlowMemberResponsesFromService(items))
}

func (h *ProjectFlowHandler) ApproveMember(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	tenantID, hasTenant := middleware.TenantIDFromCtx(c)
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	memberID, ok := parseUserIDParam(c, "user_id")
	if !ok {
		return
	}

	var req dto.ApproveMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	var positionID *uuid.UUID
	if raw := strings.TrimSpace(req.PositionID); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position_id"})
			return
		}
		positionID = &parsed
	}

	item, err := h.svc.ApproveMember(c.Request.Context(), uid, pid, memberID, positionID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	if hasTenant && memberID != uid {
		notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
			tenantID,
			memberID,
			"project.member.application.accepted",
			"Заявка в проект принята",
			"Тимлид принял вашу заявку. Вы добавлены в команду проекта.",
			map[string]any{
				"project_id": pid.String(),
			},
			true,
		))
	}

	c.JSON(http.StatusOK, dto.ProjectFlowMemberResponseFromService(item))
}

func (h *ProjectFlowHandler) RejectMember(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	tenantID, hasTenant := middleware.TenantIDFromCtx(c)
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	memberID, ok := parseUserIDParam(c, "user_id")
	if !ok {
		return
	}

	item, err := h.svc.RejectMemberApplication(c.Request.Context(), uid, pid, memberID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	if hasTenant && memberID != uid {
		notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
			tenantID,
			memberID,
			"project.member.application.rejected",
			"Заявка в проект отклонена",
			"Тимлид отклонил вашу заявку в проект.",
			map[string]any{
				"project_id": pid.String(),
			},
			true,
		))
	}

	c.JSON(http.StatusOK, dto.ProjectFlowMemberResponseFromService(item))
}

func (h *ProjectFlowHandler) RemoveMember(c *gin.Context) {
	uid, ok := parseUserID(c)
	if !ok {
		return
	}
	tenantID, hasTenant := middleware.TenantIDFromCtx(c)
	pid, ok := parseProjectID(c)
	if !ok {
		return
	}
	memberID, ok := parseUserIDParam(c, "user_id")
	if !ok {
		return
	}

	item, err := h.svc.RemoveMember(c.Request.Context(), uid, pid, memberID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}
	if hasTenant && memberID != uid {
		notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
			tenantID,
			memberID,
			"project.member.removed",
			"Вас убрали из команды проекта",
			"Доступ к проекту и командным задачам был отозван.",
			map[string]any{
				"project_id": pid.String(),
			},
			true,
		))
	}

	c.JSON(http.StatusOK, dto.ProjectFlowMemberResponseFromService(item))
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

	var req dto.ApproveMemberRequest
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
	c.JSON(http.StatusOK, dto.ProjectFlowMemberResponseFromService(item))
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
	c.JSON(http.StatusOK, dto.ListIncomingInvitesResponse{Items: dto.ProjectFlowIncomingInviteResponsesFromService(items)})
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
	c.JSON(http.StatusOK, dto.ListOutgoingApplicationsResponse{Items: dto.ProjectFlowOutgoingApplicationResponsesFromService(items)})
}
