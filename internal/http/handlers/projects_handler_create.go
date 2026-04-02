package handlers

import (
	"errors"
	"net/http"
	"strings"

	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projects"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateProject
// @Summary Create project
// @Tags Projects
// @Accept json
// @Produce json
// @Param body body dto.CreateProjectRequest true "Project data"
// @Success 201 {object} dto.CreateProjectResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /projects [post]
func (h *ProjectsHandler) Create(c *gin.Context) {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	tenantID, ok := middleware.TenantIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	facultyID, ok := middleware.FacultyIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	visibility := strings.ToUpper(strings.TrimSpace(req.Visibility))
	if visibility == "" {
		visibility = "PRIVATE"
	}
	if visibility != "PUBLIC" && visibility != "PRIVATE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visibility"})
		return
	}

	var groupID *uuid.UUID
	persistVisibility := visibility
	if visibility == "PRIVATE" {
		persistVisibility = "GROUP"

		if req.GroupID != nil && strings.TrimSpace(*req.GroupID) != "" {
			gid, err := uuid.Parse(strings.TrimSpace(*req.GroupID))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
				return
			}
			groupID = &gid
		} else if req.GroupCode != nil && strings.TrimSpace(*req.GroupCode) != "" {
			groupCode := strings.ToUpper(strings.TrimSpace(*req.GroupCode))
			gid, err := h.svc.ResolveGroupByCode(c.Request.Context(), facultyID, groupCode)
			if err != nil {
				if errors.Is(err, projects.ErrGroupNotFound) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "unknown group_code"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			groupID = &gid
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "group_code is required for visibility=PRIVATE"})
			return
		}
	} else {
		persistVisibility = "PUBLIC"
		groupID = nil
	}

	id, err := h.svc.CreateProject(c.Request.Context(), req.Title, req.Description, facultyID, persistVisibility, groupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	notifyBestEffort(h.notifier, c.Request.Context(), notifCreateInput(
		tenantID,
		userID,
		"project.created",
		"Проект создан",
		"Новый проект успешно создан и готов к настройке.",
		map[string]any{
			"project_id": id.String(),
			"title":      strings.TrimSpace(req.Title),
			"visibility": persistVisibility,
		},
		true,
	))

	c.JSON(http.StatusCreated, dto.CreateProjectResponse{ProjectID: id.String()})
}
