package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"idsai-core-up/internal/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type listDepartmentsResp struct {
	Departments any `json:"departments"`
}

type listGroupsResp struct {
	Groups any `json:"groups"`
}

type submitGroupChangeRequestReq struct {
	DepartmentCode string `json:"department_code" binding:"required"`
	GroupCode      string `json:"group_code" binding:"required"`
}

type listGroupChangeRequestsResp struct {
	Requests any `json:"requests"`
}

type reviewGroupChangeRequestReq struct {
	Action  string `json:"action" binding:"required"`
	Comment string `json:"comment"`
}

func (h *AuthHandler) ListDepartments(c *gin.Context) {
	authResponseNoStore(c)

	items, err := h.svc.ListDepartments(c.Request.Context(), tenantCodeFromHeader(c))
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, listDepartmentsResp{Departments: items})
}

func (h *AuthHandler) ListDepartmentGroups(c *gin.Context) {
	authResponseNoStore(c)

	departmentCode := strings.ToUpper(strings.TrimSpace(c.Param("department_code")))
	items, err := h.svc.ListGroupsByDepartmentCode(c.Request.Context(), tenantCodeFromHeader(c), departmentCode)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, listGroupsResp{Groups: items})
}

func (h *AuthHandler) SettingsSubmitGroupChangeRequest(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}
	isAdmin, _ := middleware.IsAdminFromCtx(c)
	isProfessor, _ := middleware.IsProfessorFromCtx(c)
	if isAdmin || isProfessor {
		c.JSON(http.StatusForbidden, gin.H{"error": "only students can submit group change requests"})
		return
	}

	var req submitGroupChangeRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	item, err := h.svc.SubmitGroupChangeRequest(
		c.Request.Context(),
		tenantID,
		userID,
		req.DepartmentCode,
		req.GroupCode,
	)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *AuthHandler) SettingsListGroupChangeRequests(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, userID, ok := settingsActorIDs(c)
	if !ok {
		return
	}
	isAdmin, _ := middleware.IsAdminFromCtx(c)
	isProfessor, _ := middleware.IsProfessorFromCtx(c)
	if isAdmin || isProfessor {
		c.JSON(http.StatusForbidden, gin.H{"error": "only students can view personal group change requests"})
		return
	}

	limit := 50
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	items, err := h.svc.ListOwnGroupChangeRequests(c.Request.Context(), tenantID, userID, limit)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, listGroupChangeRequestsResp{Requests: items})
}

func (h *AuthHandler) ListDepartmentGroupsTree(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, ok := middleware.TenantIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	isAdmin, _ := middleware.IsAdminFromCtx(c)
	isProfessor, _ := middleware.IsProfessorFromCtx(c)
	if !isAdmin && !isProfessor {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	departmentCode := strings.ToUpper(strings.TrimSpace(c.Query("department_code")))
	search := strings.TrimSpace(c.Query("q"))
	items, err := h.svc.ListDepartmentGroupsTree(c.Request.Context(), tenantID, departmentCode, search)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, listDepartmentsResp{Departments: items})
}

func (h *AuthHandler) AdminListGroupChangeRequests(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, ok := middleware.TenantIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	search := strings.TrimSpace(c.Query("q"))
	limit := 100
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	items, err := h.svc.ListGroupChangeRequests(c.Request.Context(), tenantID, status, search, limit)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, listGroupChangeRequestsResp{Requests: items})
}

func (h *AuthHandler) AdminReviewGroupChangeRequest(c *gin.Context) {
	authResponseNoStore(c)

	tenantID, ok := middleware.TenantIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	reviewerID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	requestID, err := uuid.Parse(strings.TrimSpace(c.Param("request_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	var req reviewGroupChangeRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	item, err := h.svc.ReviewGroupChangeRequest(
		c.Request.Context(),
		tenantID,
		reviewerID,
		requestID,
		req.Action,
		req.Comment,
	)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}
