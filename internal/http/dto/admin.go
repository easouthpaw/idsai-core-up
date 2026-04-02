package dto

import (
	"time"

	"idsai-core-up/internal/services/admin"

	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	FullName       string `json:"full_name"`
	DepartmentCode string `json:"department_code"`
}

type SetStatusRequest struct {
	Status string `json:"status"`
}

type SetRoleRequest struct {
	Role string `json:"role"`
}

type ResetPasswordRequest struct {
	Password string `json:"password"`
}

type SetProjectStatusRequest struct {
	Status string `json:"status"`
}

type AdminUserResponse struct {
	ID             uuid.UUID `json:"id"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	RoleCode       string    `json:"role_code"`
	Status         string    `json:"status"`
	FacultyCode    string    `json:"faculty_code"`
	DepartmentCode string    `json:"department_code"`
}

type ListUsersResponse struct {
	Users []AdminUserResponse `json:"users"`
}

type AdminProjectResponse struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	Visibility     string    `json:"visibility"`
	IsPublic       bool      `json:"is_public"`
	CreatedBy      uuid.UUID `json:"created_by"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	FacultyCode    string    `json:"faculty_code"`
	DepartmentCode string    `json:"department_code"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ListProjectsResponse struct {
	Projects []AdminProjectResponse `json:"projects"`
}

type AdminProjectPositionResponse struct {
	ID       uuid.UUID `json:"id"`
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	Capacity int       `json:"capacity"`
}

type AdminProjectMemberResponse struct {
	UserID       uuid.UUID  `json:"user_id"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	RoleCode     string     `json:"role_code"`
	Status       string     `json:"status"`
	PositionCode string     `json:"position_code"`
	PositionName string     `json:"position_name"`
	JoinedAt     *time.Time `json:"joined_at,omitempty"`
	RespondedAt  *time.Time `json:"responded_at,omitempty"`
}

type AdminProjectTaskResponse struct {
	ID             uuid.UUID  `json:"id"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	PositionCode   string     `json:"position_code"`
	AssigneeUserID *uuid.UUID `json:"assignee_user_id,omitempty"`
	AssigneeName   string     `json:"assignee_name"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type AdminProjectCriterionResponse struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Weight    int        `json:"weight"`
	CreatedBy uuid.UUID  `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	IsMet     *bool      `json:"is_met,omitempty"`
	Comment   string     `json:"comment,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type AdminProjectObservationSummaryResponse struct {
	MembersTotal   int `json:"members_total"`
	MembersActive  int `json:"members_active"`
	MembersApplied int `json:"members_applied"`
	MembersInvited int `json:"members_invited"`
	TasksTotal     int `json:"tasks_total"`
	TasksDone      int `json:"tasks_done"`
	CriteriaTotal  int `json:"criteria_total"`
}

type AdminProjectObservationResponse struct {
	Project   AdminProjectResponse                   `json:"project"`
	Positions []AdminProjectPositionResponse         `json:"positions"`
	Members   []AdminProjectMemberResponse           `json:"members"`
	Tasks     []AdminProjectTaskResponse             `json:"tasks"`
	Criteria  []AdminProjectCriterionResponse        `json:"criteria"`
	Summary   AdminProjectObservationSummaryResponse `json:"summary"`
}

func AdminUserResponseFromService(item admin.User) AdminUserResponse {
	return AdminUserResponse{
		ID:             item.ID,
		FullName:       item.FullName,
		Email:          item.Email,
		RoleCode:       item.RoleCode,
		Status:         item.Status,
		FacultyCode:    item.FacultyCode,
		DepartmentCode: item.DepartmentCode,
	}
}

func AdminUserResponsesFromService(items []admin.User) []AdminUserResponse {
	if items == nil {
		return nil
	}
	out := make([]AdminUserResponse, 0, len(items))
	for _, item := range items {
		out = append(out, AdminUserResponseFromService(item))
	}
	return out
}

func AdminProjectResponseFromService(item admin.Project) AdminProjectResponse {
	return AdminProjectResponse{
		ID:             item.ID,
		Title:          item.Title,
		Description:    item.Description,
		Status:         item.Status,
		Visibility:     item.Visibility,
		IsPublic:       item.IsPublic,
		CreatedBy:      item.CreatedBy,
		AuthorName:     item.AuthorName,
		AuthorEmail:    item.AuthorEmail,
		FacultyCode:    item.FacultyCode,
		DepartmentCode: item.DepartmentCode,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func AdminProjectResponsesFromService(items []admin.Project) []AdminProjectResponse {
	if items == nil {
		return nil
	}
	out := make([]AdminProjectResponse, 0, len(items))
	for _, item := range items {
		out = append(out, AdminProjectResponseFromService(item))
	}
	return out
}

func AdminProjectPositionResponsesFromService(items []admin.ProjectPosition) []AdminProjectPositionResponse {
	if items == nil {
		return nil
	}
	out := make([]AdminProjectPositionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, AdminProjectPositionResponse{
			ID:       item.ID,
			Code:     item.Code,
			Name:     item.Name,
			Capacity: item.Capacity,
		})
	}
	return out
}

func AdminProjectMemberResponsesFromService(items []admin.ProjectMember) []AdminProjectMemberResponse {
	if items == nil {
		return nil
	}
	out := make([]AdminProjectMemberResponse, 0, len(items))
	for _, item := range items {
		out = append(out, AdminProjectMemberResponse{
			UserID:       item.UserID,
			FullName:     item.FullName,
			Email:        item.Email,
			RoleCode:     item.RoleCode,
			Status:       item.Status,
			PositionCode: item.PositionCode,
			PositionName: item.PositionName,
			JoinedAt:     item.JoinedAt,
			RespondedAt:  item.RespondedAt,
		})
	}
	return out
}

func AdminProjectTaskResponsesFromService(items []admin.ProjectTask) []AdminProjectTaskResponse {
	if items == nil {
		return nil
	}
	out := make([]AdminProjectTaskResponse, 0, len(items))
	for _, item := range items {
		out = append(out, AdminProjectTaskResponse{
			ID:             item.ID,
			Title:          item.Title,
			Status:         item.Status,
			PositionCode:   item.PositionCode,
			AssigneeUserID: item.AssigneeUserID,
			AssigneeName:   item.AssigneeName,
			DueAt:          item.DueAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return out
}

func AdminProjectCriterionResponsesFromService(items []admin.ProjectCriterion) []AdminProjectCriterionResponse {
	if items == nil {
		return nil
	}
	out := make([]AdminProjectCriterionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, AdminProjectCriterionResponse{
			ID:        item.ID,
			Title:     item.Title,
			Weight:    item.Weight,
			CreatedBy: item.CreatedBy,
			CreatedAt: item.CreatedAt,
			IsMet:     item.IsMet,
			Comment:   item.Comment,
			UpdatedAt: item.UpdatedAt,
		})
	}
	return out
}

func AdminProjectObservationResponseFromService(item admin.ProjectObservation) AdminProjectObservationResponse {
	return AdminProjectObservationResponse{
		Project:   AdminProjectResponseFromService(item.Project),
		Positions: AdminProjectPositionResponsesFromService(item.Positions),
		Members:   AdminProjectMemberResponsesFromService(item.Members),
		Tasks:     AdminProjectTaskResponsesFromService(item.Tasks),
		Criteria:  AdminProjectCriterionResponsesFromService(item.Criteria),
		Summary: AdminProjectObservationSummaryResponse{
			MembersTotal:   item.Summary.MembersTotal,
			MembersActive:  item.Summary.MembersActive,
			MembersApplied: item.Summary.MembersApplied,
			MembersInvited: item.Summary.MembersInvited,
			TasksTotal:     item.Summary.TasksTotal,
			TasksDone:      item.Summary.TasksDone,
			CriteriaTotal:  item.Summary.CriteriaTotal,
		},
	}
}
