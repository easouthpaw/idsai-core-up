package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/projects"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProjectsHandler struct {
	svc      *projects.Service
	notifier NotificationPublisher
}

func NewProjectsHandler(svc *projects.Service) *ProjectsHandler {
	return &ProjectsHandler{svc: svc}
}

func (h *ProjectsHandler) SetNotifier(pub NotificationPublisher) {
	h.notifier = pub
}

type createProjectRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Visibility  string  `json:"visibility"`         // PUBLIC | PRIVATE
	GroupID     *string `json:"group_id,omitempty"` // UUID string, optional alternative to group_code for PRIVATE
	GroupCode   *string `json:"group_code,omitempty"`
}
type projectResponse struct {
	ID                    string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title                 string    `json:"title" example:"Swagger Demo"`
	Description           string    `json:"description" example:"created from swagger"`
	Status                string    `json:"status" example:"DRAFT"`
	IsPublic              bool      `json:"is_public" example:"false"`
	CreatedBy             string    `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProfessorID           *string   `json:"professor_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProfessorReviewStatus string    `json:"professor_review_status" example:"PENDING"`
	FacultyID             string    `json:"faculty_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Visibility            string    `json:"visibility" example:"FACULTY"`
	GroupID               *string   `json:"group_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type groupOptionResponse struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Number     string `json:"number"`
}

func toUIVisibility(v string) string {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "GROUP", "FACULTY", "PRIVATE":
		return "PRIVATE"
	case "PUBLIC":
		return "PUBLIC"
	default:
		return strings.ToUpper(strings.TrimSpace(v))
	}
}

func splitGroupCode(code string) (department, number string) {
	parts := strings.SplitN(strings.TrimSpace(code), "-", 2)
	if len(parts) == 2 {
		return strings.ToUpper(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.ToUpper(strings.TrimSpace(code)), ""
}

func projectToResponse(p domain.Project) projectResponse {
	resp := projectResponse{
		ID:                    p.ID.String(),
		Title:                 p.Title,
		Description:           p.Description,
		Status:                string(p.Status),
		IsPublic:              p.IsPublic,
		CreatedBy:             p.CreatedBy.String(),
		ProfessorID:           nil,
		ProfessorReviewStatus: strings.ToUpper(strings.TrimSpace(p.ProfessorReviewStatus)),
		FacultyID:             p.FacultyID.String(),
		Visibility:            toUIVisibility(p.Visibility),
		GroupID:               nil,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}

	if p.ProfessorID != nil {
		s := p.ProfessorID.String()
		resp.ProfessorID = &s
	}
	if resp.ProfessorReviewStatus == "" {
		resp.ProfessorReviewStatus = "NONE"
	}
	if p.GroupID != nil {
		s := p.GroupID.String()
		resp.GroupID = &s
	}

	return resp
}

// CreateProject
// @Summary Create project
// @Tags Projects
// @Accept json
// @Produce json
// @Param body body createProjectRequest true "Project data"
// @Success 201 {object} map[string]string
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

	var req createProjectRequest
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
				if errors.Is(err, pgx.ErrNoRows) {
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

	c.JSON(http.StatusCreated, gin.H{"project_id": id.String()})
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

	c.JSON(http.StatusOK, projectToResponse(p))
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
