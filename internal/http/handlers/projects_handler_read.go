package handlers

import (
	"errors"
	"net/http"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projects"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type groupOptionResponse struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Number     string `json:"number"`
}

// GetProject
// @Summary Get project by id
// @Tags Projects
// @Produce json
// @Param project_id path string true "Project UUID"
// @Success 200 {object} projectResponse
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /projects/{project_id} [get]
func (h *ProjectsHandler) Get(c *gin.Context) {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	facultyID, ok := middleware.FacultyIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	raw := c.Param("project_id")
	projectID, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	view, err := h.svc.GetProjectViewForViewer(c.Request.Context(), projectID, userID, facultyID)
	if err != nil {
		if errors.Is(err, projects.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		if errors.Is(err, domain.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projectViewToResponse(view))
}

// ListMyProjects
// @Summary List my projects
// @Tags Projects
// @Produce json
// @Success 200 {array} projectResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects/my [get]
func (h *ProjectsHandler) ListMine(c *gin.Context) {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	items, err := h.svc.ListProjectsByCreator(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]projectResponse, 0, len(items))
	for _, p := range items {
		resp = append(resp, projectToResponse(p))
	}

	c.JSON(http.StatusOK, resp)
}

// ListPublicProjects
// @Summary List all public projects
// @Tags Projects
// @Produce json
// @Success 200 {array} projectResponse
// @Failure 500 {object} map[string]string
// @Router /projects/public [get]
func (h *ProjectsHandler) ListPublic(c *gin.Context) {
	items, err := h.svc.ListPublicProjects(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]projectResponse, 0, len(items))
	for _, p := range items {
		resp = append(resp, projectToResponse(p))
	}

	c.JSON(http.StatusOK, resp)
}

// ListGroups
// @Summary List predefined groups for faculty
// @Tags Projects
// @Produce json
// @Success 200 {array} groupOptionResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /projects/groups [get]
func (h *ProjectsHandler) ListGroups(c *gin.Context) {
	facultyID, ok := middleware.FacultyIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	items, err := h.svc.ListGroupsByFaculty(c.Request.Context(), facultyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]groupOptionResponse, 0, len(items))
	for _, g := range items {
		dept, num := splitGroupCode(g.Code)
		resp = append(resp, groupOptionResponse{
			ID:         g.ID.String(),
			Code:       g.Code,
			Name:       g.Name,
			Department: dept,
			Number:     num,
		})
	}

	c.JSON(http.StatusOK, resp)
}
