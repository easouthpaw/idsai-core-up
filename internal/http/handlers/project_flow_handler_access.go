package handlers

import (
	"net/http"

	"idsai-core-up/internal/http/dto"

	"github.com/gin-gonic/gin"
)

// GetAccessCatalog returns the catalog of assignable delegated project roles.
func (h *ProjectFlowHandler) GetAccessCatalog(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	items, err := h.svc.GetAccessCatalog(c.Request.Context(), userID, projectID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ListAccessCatalogResponse{Items: dto.ProjectFlowAccessCatalogItemResponsesFromService(items)})
}

// GetMemberAccess returns the access state for a specific member in the project.
func (h *ProjectFlowHandler) GetMemberAccess(c *gin.Context) {
	callerID, ok := parseUserID(c)
	if !ok {
		return
	}
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}
	targetUserID, ok := parseUserIDParam(c, "user_id")
	if !ok {
		return
	}

	access, err := h.svc.GetMemberAccess(c.Request.Context(), callerID, projectID, targetUserID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ProjectFlowMemberAccessResponseFromService(access))
}

// ReplaceMemberAccess atomically replaces assignable roles for a member.
func (h *ProjectFlowHandler) ReplaceMemberAccess(c *gin.Context) {
	callerID, ok := parseUserID(c)
	if !ok {
		return
	}
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}
	targetUserID, ok := parseUserIDParam(c, "user_id")
	if !ok {
		return
	}

	var req dto.ReplaceMemberAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	access, err := h.svc.ReplaceMemberAccess(c.Request.Context(), callerID, projectID, targetUserID, req.ManagedRoleCodes)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ProjectFlowMemberAccessResponseFromService(access))
}

// MyPermissions returns the current user's effective permission codes in the project.
func (h *ProjectFlowHandler) MyPermissions(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}
	projectID, ok := parseProjectID(c)
	if !ok {
		return
	}

	perms, err := h.svc.MyPermissions(c.Request.Context(), userID, projectID)
	if err != nil {
		handleFlowErr(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MyPermissionsResponse{Permissions: perms})
}
