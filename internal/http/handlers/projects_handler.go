package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"idsai-core-up/internal/services/projects"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProjectsHandler struct {
	svc *projects.Service
}

func NewProjectsHandler(svc *projects.Service) *ProjectsHandler {
	return &ProjectsHandler{svc: svc}
}

type createProjectRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Visibility  string  `json:"visibility"`         // PUBLIC | FACULTY | GROUP | PRIVATE
	GroupID     *string `json:"group_id,omitempty"` // UUID string, required if visibility=GROUP
}
type projectResponse struct {
	ID          string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title       string    `json:"title" example:"Swagger Demo"`
	Description string    `json:"description" example:"created from swagger"`
	Status      string    `json:"status" example:"DRAFT"`
	IsPublic    bool      `json:"is_public" example:"false"`
	CreatedBy   string    `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProfessorID *string   `json:"professor_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	FacultyID   string    `json:"faculty_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Visibility  string    `json:"visibility" example:"FACULTY"`
	GroupID     *string   `json:"group_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProject
// @Summary Create project
// @Tags Projects
// @Accept json
// @Produce json
// @Param X-User-ID header string true "User UUID"
// @Param X-Faculty-ID header string true "Faculty UUID"
// @Param body body createProjectRequest true "Project data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /projects [post]
func (h *ProjectsHandler) Create(c *gin.Context) {
	// 1) user
	userRaw := c.GetHeader("X-User-ID")
	if userRaw == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-User-ID"})
		return
	}
	userID, err := uuid.Parse(userRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid X-User-ID"})
		return
	}

	// 2) faculty
	facultyRaw := c.GetHeader("X-Faculty-ID")
	if facultyRaw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Faculty-ID"})
		return
	}
	facultyID, err := uuid.Parse(facultyRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid X-Faculty-ID"})
		return
	}

	// 3) body
	var req createProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	visibility := strings.ToUpper(strings.TrimSpace(req.Visibility))
	if visibility == "" {
		visibility = "FACULTY"
	}
	if visibility != "PUBLIC" && visibility != "FACULTY" && visibility != "GROUP" && visibility != "PRIVATE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visibility"})
		return
	}

	var groupID *uuid.UUID
	if visibility == "GROUP" {
		if req.GroupID == nil || strings.TrimSpace(*req.GroupID) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required for visibility=GROUP"})
			return
		}
		gid, err := uuid.Parse(strings.TrimSpace(*req.GroupID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
			return
		}
		groupID = &gid
	} else {
		groupID = nil
	}

	// 4) create
	id, err := h.svc.CreateProject(c.Request.Context(), req.Title, req.Description, facultyID, visibility, groupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"project_id": id.String()})
}

// GetProject
// @Summary Get project by id
// @Tags Projects
// @Produce json
// @Param X-User-ID header string true "User UUID"
// @Param project_id path string true "Project UUID"
// @Success 200 {object} projectResponse
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /projects/{project_id} [get]
func (h *ProjectsHandler) Get(c *gin.Context) {
	raw := c.Param("project_id")
	projectID, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	p, err := h.svc.GetProject(c.Request.Context(), projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := projectResponse{
		ID:          p.ID.String(),
		Title:       p.Title,
		Description: p.Description,
		Status:      string(p.Status),
		IsPublic:    p.IsPublic,
		CreatedBy:   p.CreatedBy.String(),
		ProfessorID: nil,
		FacultyID:   p.FacultyID.String(),
		Visibility:  p.Visibility,
		GroupID:     nil,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if p.ProfessorID != nil {
		s := p.ProfessorID.String()
		resp.ProfessorID = &s
	}
	if p.GroupID != nil {
		s := p.GroupID.String()
		resp.GroupID = &s
	}

	c.JSON(http.StatusOK, resp)
}
