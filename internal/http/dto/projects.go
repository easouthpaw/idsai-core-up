package dto

import (
	"strings"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projects"
)

type CreateProjectRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Visibility  string  `json:"visibility"`
	GroupID     *string `json:"group_id,omitempty"`
	GroupCode   *string `json:"group_code,omitempty"`
}

type CreateProjectResponse struct {
	ProjectID string `json:"project_id"`
}

type ProjectResponse struct {
	ID                    string                        `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title                 string                        `json:"title" example:"Project Demo"`
	Description           string                        `json:"description" example:"created from platform"`
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
	ImageKey              string                        `json:"image_key,omitempty"`
	ImageURL              string                        `json:"image_url,omitempty"`
	HasCustomImage        bool                          `json:"has_custom_image"`
	DefaultCoverVariant   int                           `json:"default_cover_variant"`
	RetakeCount           int                           `json:"retake_count"`
	RetakePenaltyPercent  int                           `json:"retake_penalty_percent"`
	CreatedAt             time.Time                     `json:"created_at"`
	UpdatedAt             time.Time                     `json:"updated_at"`
	ViewerAccess          *ProjectViewerAccessResponse  `json:"viewer_access,omitempty"`
	ReviewSummary         *ProjectReviewSummaryResponse `json:"review_summary,omitempty"`
}

type ProjectViewerAccessResponse struct {
	CanViewWorkspace      bool `json:"can_view_workspace"`
	CanViewProjectDetails bool `json:"can_view_project_details"`
	CanApply              bool `json:"can_apply"`
	CanViewFinalGrade     bool `json:"can_view_final_grade"`
}

type ProjectReviewSummaryResponse struct {
	Score       string     `json:"score"`
	PassPercent int        `json:"pass_percent"`
	Met         int        `json:"met"`
	Total       int        `json:"total"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	Reviewer    string     `json:"reviewer,omitempty"`
}

type GroupOptionResponse struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Number     string `json:"number"`
}

func splitGroupCode(code string) (department, number string) {
	parts := strings.SplitN(strings.TrimSpace(code), "-", 2)
	if len(parts) == 2 {
		return strings.ToUpper(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.ToUpper(strings.TrimSpace(code)), ""
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

func ProjectResponseFromDomain(p domain.Project) ProjectResponse {
	resp := ProjectResponse{
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
		ImageKey:              strings.TrimSpace(p.ImageKey),
		ImageURL:              strings.TrimSpace(p.ImageURL),
		HasCustomImage:        strings.TrimSpace(p.ImageKey) != "",
		DefaultCoverVariant:   p.DefaultCoverVariant,
		RetakeCount:           p.RetakeCount,
		RetakePenaltyPercent:  domain.RetakePenaltyPercent(p.RetakeCount),
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
	if resp.DefaultCoverVariant <= 0 {
		resp.DefaultCoverVariant = 1
	}

	return resp
}

func ProjectResponseFromView(view projects.ProjectView) ProjectResponse {
	resp := ProjectResponseFromDomain(view.Project)
	resp.ViewerAccess = &ProjectViewerAccessResponse{
		CanViewWorkspace:      view.Access.CanViewWorkspace,
		CanViewProjectDetails: view.Access.CanViewProjectDetails,
		CanApply:              view.Access.CanApply,
		CanViewFinalGrade:     view.Access.CanViewFinalGrade,
	}
	if view.ReviewSummary != nil {
		resp.ReviewSummary = &ProjectReviewSummaryResponse{
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

func GroupOptionResponsesFromService(items []projects.Group) []GroupOptionResponse {
	if items == nil {
		return nil
	}
	out := make([]GroupOptionResponse, 0, len(items))
	for _, item := range items {
		department, number := splitGroupCode(item.Code)
		out = append(out, GroupOptionResponse{
			ID:         item.ID.String(),
			Code:       item.Code,
			Name:       item.Name,
			Department: department,
			Number:     number,
		})
	}
	return out
}
