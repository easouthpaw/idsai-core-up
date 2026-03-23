package handlers

import (
	"strings"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projects"
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

type projectResponse struct {
	ID                    string                        `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title                 string                        `json:"title" example:"Swagger Demo"`
	Description           string                        `json:"description" example:"created from swagger"`
	Status                string                        `json:"status" example:"DRAFT"`
	IsPublic              bool                          `json:"is_public" example:"false"`
	CreatedBy             string                        `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedByName         string                        `json:"created_by_name,omitempty" example:"Айболат Ермекбай"`
	CreatedByEmail        string                        `json:"created_by_email,omitempty" example:"aibolat.student@idsai.local"`
	ProfessorID           *string                       `json:"professor_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProfessorReviewStatus string                        `json:"professor_review_status" example:"PENDING"`
	FacultyID             string                        `json:"faculty_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Visibility            string                        `json:"visibility" example:"FACULTY"`
	GroupID               *string                       `json:"group_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatedAt             time.Time                     `json:"created_at"`
	UpdatedAt             time.Time                     `json:"updated_at"`
	ViewerAccess          *projectViewerAccessResponse  `json:"viewer_access,omitempty"`
	ReviewSummary         *projectReviewSummaryResponse `json:"review_summary,omitempty"`
}

type projectViewerAccessResponse struct {
	CanViewWorkspace  bool `json:"can_view_workspace"`
	CanApply          bool `json:"can_apply"`
	CanViewFinalGrade bool `json:"can_view_final_grade"`
}

type projectReviewSummaryResponse struct {
	Score       string     `json:"score"`
	PassPercent int        `json:"pass_percent"`
	Met         int        `json:"met"`
	Total       int        `json:"total"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	Reviewer    string     `json:"reviewer,omitempty"`
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
		CreatedByName:         strings.TrimSpace(p.CreatedByName),
		CreatedByEmail:        strings.TrimSpace(p.CreatedByEmail),
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
	if resp.CreatedByName == "" {
		resp.CreatedByName = resp.CreatedBy
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

func projectViewToResponse(view projects.ProjectView) projectResponse {
	resp := projectToResponse(view.Project)
	resp.ViewerAccess = &projectViewerAccessResponse{
		CanViewWorkspace:  view.Access.CanViewWorkspace,
		CanApply:          view.Access.CanApply,
		CanViewFinalGrade: view.Access.CanViewFinalGrade,
	}
	if view.ReviewSummary != nil {
		resp.ReviewSummary = &projectReviewSummaryResponse{
			Score:       view.ReviewSummary.Score,
			PassPercent: view.ReviewSummary.PassPercent,
			Met:         view.ReviewSummary.Met,
			Total:       view.ReviewSummary.Total,
			ReviewedAt:  view.ReviewSummary.ReviewedAt,
			Reviewer:    view.ReviewSummary.Reviewer,
		}
	}
	return resp
}
