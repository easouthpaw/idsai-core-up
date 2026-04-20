package dto

import (
	"strings"
	"time"

	"idsai-core-up/internal/services/projectflow"
)

type InviteMemberRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	Comment string `json:"comment"`
}

type ApplyMemberRequest struct {
	Comment string `json:"comment"`
}

type RespondInviteRequest struct {
	Accept bool `json:"accept"`
}

type ApproveMemberRequest struct {
	PositionID string `json:"position_id"`
}

type ReplaceMemberAccessRequest struct {
	ManagedRoleCodes []string `json:"managed_role_codes"`
}

type CreateProjectAccessRoleRequest struct {
	Code            string   `json:"code" binding:"required"`
	Name            string   `json:"name" binding:"required"`
	Description     string   `json:"description"`
	PermissionCodes []string `json:"permission_codes"`
}

type UpdateProjectRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type SetStacksRequest struct {
	Stacks []string `json:"stacks"`
}

type CreatePositionRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}

type CreateTaskRequest struct {
	Title          string  `json:"title" binding:"required"`
	Description    string  `json:"description"`
	PositionID     string  `json:"position_id" binding:"required"`
	AssigneeUserID *string `json:"assignee_user_id,omitempty"`
	DueAt          *string `json:"due_at,omitempty"`
}

type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type AssignTaskRequest struct {
	AssigneeUserID string `json:"assignee_user_id" binding:"required"`
}

type CompleteTaskRequest struct {
	Comment     string   `json:"comment"`
	Attachments []string `json:"attachments"`
}

type CreateCriterionRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Weight      int    `json:"weight"`
}

type GradingItemRequest struct {
	CriterionID string `json:"criterion_id" binding:"required"`
	IsMet       *bool  `json:"is_met"`
	Comment     string `json:"comment"`
}

type UpsertGradingRequest struct {
	Items []GradingItemRequest `json:"items"`
}

type AssignProfessorRequest struct {
	ProfessorID string `json:"professor_id" binding:"required"`
}

type ProjectStackResponse struct {
	Code string `json:"code"`
}

type ProjectFlowCriterionResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Weight      int       `json:"weight"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectFlowCriterionGradeResponse struct {
	CriterionID string     `json:"criterion_id"`
	IsMet       *bool      `json:"is_met,omitempty"`
	Comment     string     `json:"comment,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type GradingItemsResponse struct {
	Items []ProjectFlowCriterionGradeResponse `json:"items"`
}

type ProjectFlowPositionResponse struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Capacity  int       `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectFlowMemberResponse struct {
	ID                  string     `json:"id"`
	ProjectID           string     `json:"project_id"`
	UserID              string     `json:"user_id"`
	FullName            string     `json:"full_name,omitempty"`
	Email               string     `json:"email,omitempty"`
	PositionID          *string    `json:"position_id,omitempty"`
	PositionCode        *string    `json:"position_code,omitempty"`
	PositionName        *string    `json:"position_name,omitempty"`
	AccessRoleName      *string    `json:"access_role_name,omitempty"`
	AccessRoleCode      *string    `json:"access_role_code,omitempty"`
	Status              string     `json:"status"`
	InviteComment       string     `json:"invite_comment,omitempty"`
	InvitedBy           *string    `json:"invited_by,omitempty"`
	RespondedAt         *time.Time `json:"responded_at,omitempty"`
	JoinedAt            *time.Time `json:"joined_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

type ProjectFlowTaskResponse struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"project_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	PositionID     string     `json:"position_id"`
	PositionCode   string     `json:"position_code"`
	PositionName   string     `json:"position_name"`
	AssigneeUserID *string    `json:"assignee_user_id,omitempty"`
	Status         string     `json:"status"`
	CreatedBy      string     `json:"created_by"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ProjectFlowTaskActivityResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	TaskID      string    `json:"task_id"`
	ActorUserID *string   `json:"actor_user_id,omitempty"`
	ActorName   string    `json:"actor_name,omitempty"`
	ActorEmail  string    `json:"actor_email,omitempty"`
	EventType   string    `json:"event_type"`
	FromStatus  string    `json:"from_status,omitempty"`
	ToStatus    string    `json:"to_status,omitempty"`
	Title       string    `json:"title,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	Attachments []string  `json:"attachments,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListTaskActivitiesResponse struct {
	Items []ProjectFlowTaskActivityResponse `json:"items"`
}

type ClaimTaskResponse struct {
	Status string `json:"status"`
}

type ProjectFlowStudentCandidateResponse struct {
	UserID         string `json:"user_id"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	DepartmentCode string `json:"department_code"`
}

type ProjectFlowProfessorCandidateResponse struct {
	UserID         string `json:"user_id"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	DepartmentCode string `json:"department_code"`
}

type AssignedProfessorResponse struct {
	Professor *ProjectFlowProfessorCandidateResponse `json:"professor"`
}

type ProjectFlowIncomingInviteResponse struct {
	ProjectID     string     `json:"project_id"`
	ProjectTitle  string     `json:"project_title"`
	ProjectStatus string     `json:"project_status"`
	Status        string     `json:"status"`
	UserID        string     `json:"user_id,omitempty"`
	InviteComment string     `json:"invite_comment,omitempty"`
	InvitedBy     *string    `json:"invited_by,omitempty"`
	InviterName   string     `json:"inviter_name,omitempty"`
	InviterEmail  string     `json:"inviter_email,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	RespondedAt   *time.Time `json:"responded_at,omitempty"`
}

type ListIncomingInvitesResponse struct {
	Items []ProjectFlowIncomingInviteResponse `json:"items"`
}

type ProjectFlowOutgoingApplicationResponse struct {
	ProjectID     string     `json:"project_id"`
	ProjectTitle  string     `json:"project_title"`
	ProjectStatus string     `json:"project_status"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	RespondedAt   *time.Time `json:"responded_at,omitempty"`
}

type ListOutgoingApplicationsResponse struct {
	Items []ProjectFlowOutgoingApplicationResponse `json:"items"`
}

type ProjectFlowAccessCatalogItemResponse struct {
	Code            string   `json:"code"`
	DisplayCode     string   `json:"display_code,omitempty"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	PermissionCodes []string `json:"permission_codes"`
	Custom          bool     `json:"custom"`
}

type ListAccessCatalogResponse struct {
	Items []ProjectFlowAccessCatalogItemResponse `json:"items"`
}

type ProjectFlowPermissionCatalogItemResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type ListProjectAccessPermissionsResponse struct {
	Items []ProjectFlowPermissionCatalogItemResponse `json:"items"`
}

type ProjectFlowMemberAccessResponse struct {
	UserID                   string   `json:"user_id"`
	RoleCodes                []string `json:"role_codes"`
	ManagedRoleCodes         []string `json:"managed_role_codes"`
	EffectivePermissionCodes []string `json:"effective_permission_codes"`
}

type MyPermissionsResponse struct {
	Permissions []string `json:"permissions"`
}

type ProjectFlowReadinessResponse struct {
	ProjectID       string `json:"project_id"`
	Status          string `json:"status"`
	RequiredMembers int    `json:"required_members"`
	ActiveMembers   int    `json:"active_members"`
	HasProfessor    bool   `json:"has_professor"`
	ProfessorStatus string `json:"professor_status"`
	CriteriaCount   int    `json:"criteria_count"`
	CanActivate     bool   `json:"can_activate"`
}

type ApproveProjectResponse struct {
	Project   ProjectResponse              `json:"project"`
	Readiness ProjectFlowReadinessResponse `json:"readiness"`
}

type ProjectReadinessConflictResponse struct {
	Error     string                       `json:"error"`
	Readiness ProjectFlowReadinessResponse `json:"readiness"`
}

type ProjectEnvelopeResponse struct {
	Project ProjectResponse `json:"project"`
}

func CriterionGradesFromRequest(items []GradingItemRequest) []projectflow.CriterionGrade {
	if items == nil {
		return nil
	}
	out := make([]projectflow.CriterionGrade, 0, len(items))
	for _, item := range items {
		out = append(out, projectflow.CriterionGrade{
			CriterionID: strings.TrimSpace(item.CriterionID),
			IsMet:       item.IsMet,
			Comment:     strings.TrimSpace(item.Comment),
		})
	}
	return out
}

func ProjectStackResponsesFromService(items []projectflow.Stack) []ProjectStackResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectStackResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectStackResponse{Code: item.Code})
	}
	return out
}

func ProjectFlowCriterionResponseFromService(item projectflow.Criterion) ProjectFlowCriterionResponse {
	return ProjectFlowCriterionResponse{
		ID:          item.ID,
		ProjectID:   item.ProjectID,
		Title:       item.Title,
		Description: item.Description,
		Weight:      item.Weight,
		CreatedBy:   item.CreatedBy,
		CreatedAt:   item.CreatedAt,
	}
}

func ProjectFlowCriterionResponsesFromService(items []projectflow.Criterion) []ProjectFlowCriterionResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowCriterionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowCriterionResponseFromService(item))
	}
	return out
}

func ProjectFlowCriterionGradeResponseFromService(item projectflow.CriterionGrade) ProjectFlowCriterionGradeResponse {
	return ProjectFlowCriterionGradeResponse{
		CriterionID: item.CriterionID,
		IsMet:       item.IsMet,
		Comment:     item.Comment,
		UpdatedAt:   item.UpdatedAt,
	}
}

func ProjectFlowCriterionGradeResponsesFromService(items []projectflow.CriterionGrade) []ProjectFlowCriterionGradeResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowCriterionGradeResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowCriterionGradeResponseFromService(item))
	}
	return out
}

func ProjectFlowPositionResponseFromService(item projectflow.Position) ProjectFlowPositionResponse {
	return ProjectFlowPositionResponse{
		ID:        item.ID,
		ProjectID: item.ProjectID,
		Code:      item.Code,
		Name:      item.Name,
		Capacity:  item.Capacity,
		CreatedAt: item.CreatedAt,
	}
}

func ProjectFlowPositionResponsesFromService(items []projectflow.Position) []ProjectFlowPositionResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowPositionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowPositionResponseFromService(item))
	}
	return out
}

func ProjectFlowMemberResponseFromService(item projectflow.Member) ProjectFlowMemberResponse {
	return ProjectFlowMemberResponse{
		ID:             item.ID,
		ProjectID:      item.ProjectID,
		UserID:         item.UserID,
		FullName:       item.FullName,
		Email:          item.Email,
		PositionID:     item.PositionID,
		PositionCode:   item.PositionCode,
		PositionName:   item.PositionName,
		AccessRoleName: item.AccessRoleName,
		AccessRoleCode: item.AccessRoleCode,
		Status:         item.Status,
		InviteComment:  item.InviteComment,
		InvitedBy:      item.InvitedBy,
		RespondedAt:    item.RespondedAt,
		JoinedAt:       item.JoinedAt,
		CreatedAt:      item.CreatedAt,
	}
}

func ProjectFlowMemberResponsesFromService(items []projectflow.Member) []ProjectFlowMemberResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowMemberResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowMemberResponseFromService(item))
	}
	return out
}

func ProjectFlowTaskResponseFromService(item projectflow.Task) ProjectFlowTaskResponse {
	return ProjectFlowTaskResponse{
		ID:             item.ID,
		ProjectID:      item.ProjectID,
		Title:          item.Title,
		Description:    item.Description,
		PositionID:     item.PositionID,
		PositionCode:   item.PositionCode,
		PositionName:   item.PositionName,
		AssigneeUserID: item.AssigneeUserID,
		Status:         item.Status,
		CreatedBy:      item.CreatedBy,
		DueAt:          item.DueAt,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func ProjectFlowTaskResponsesFromService(items []projectflow.Task) []ProjectFlowTaskResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowTaskResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowTaskResponseFromService(item))
	}
	return out
}

func ProjectFlowTaskActivityResponsesFromService(items []projectflow.TaskActivity) []ProjectFlowTaskActivityResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowTaskActivityResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowTaskActivityResponse{
			ID:          item.ID,
			ProjectID:   item.ProjectID,
			TaskID:      item.TaskID,
			ActorUserID: item.ActorUserID,
			ActorName:   item.ActorName,
			ActorEmail:  item.ActorEmail,
			EventType:   item.EventType,
			FromStatus:  item.FromStatus,
			ToStatus:    item.ToStatus,
			Title:       item.Title,
			Comment:     item.Comment,
			Attachments: item.Attachments,
			CreatedAt:   item.CreatedAt,
		})
	}
	return out
}

func ProjectFlowStudentCandidateResponsesFromService(items []projectflow.StudentCandidate) []ProjectFlowStudentCandidateResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowStudentCandidateResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowStudentCandidateResponse{
			UserID:         item.UserID,
			FullName:       item.FullName,
			Email:          item.Email,
			DepartmentCode: item.DepartmentCode,
		})
	}
	return out
}

func ProjectFlowProfessorCandidateResponseFromService(item projectflow.ProfessorCandidate) ProjectFlowProfessorCandidateResponse {
	return ProjectFlowProfessorCandidateResponse{
		UserID:         item.UserID,
		FullName:       item.FullName,
		Email:          item.Email,
		DepartmentCode: item.DepartmentCode,
	}
}

func ProjectFlowProfessorCandidateResponsesFromService(items []projectflow.ProfessorCandidate) []ProjectFlowProfessorCandidateResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowProfessorCandidateResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowProfessorCandidateResponseFromService(item))
	}
	return out
}

func AssignedProfessorResponseFromService(item *projectflow.ProfessorCandidate) AssignedProfessorResponse {
	if item == nil {
		return AssignedProfessorResponse{}
	}
	resp := ProjectFlowProfessorCandidateResponseFromService(*item)
	return AssignedProfessorResponse{Professor: &resp}
}

func ProjectFlowIncomingInviteResponsesFromService(items []projectflow.IncomingInvite) []ProjectFlowIncomingInviteResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowIncomingInviteResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowIncomingInviteResponse{
			ProjectID:     item.ProjectID,
			ProjectTitle:  item.ProjectTitle,
			ProjectStatus: item.ProjectStatus,
			Status:        item.Status,
			UserID:        item.UserID,
			InviteComment: item.InviteComment,
			InvitedBy:     item.InvitedBy,
			InviterName:   item.InviterName,
			InviterEmail:  item.InviterEmail,
			CreatedAt:     item.CreatedAt,
			RespondedAt:   item.RespondedAt,
		})
	}
	return out
}

func ProjectFlowOutgoingApplicationResponsesFromService(items []projectflow.OutgoingApplication) []ProjectFlowOutgoingApplicationResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowOutgoingApplicationResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowOutgoingApplicationResponse{
			ProjectID:     item.ProjectID,
			ProjectTitle:  item.ProjectTitle,
			ProjectStatus: item.ProjectStatus,
			Status:        item.Status,
			CreatedAt:     item.CreatedAt,
			RespondedAt:   item.RespondedAt,
		})
	}
	return out
}

func ProjectFlowAccessCatalogItemResponsesFromService(items []projectflow.AccessCatalogItem) []ProjectFlowAccessCatalogItemResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowAccessCatalogItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowAccessCatalogItemResponse{
			Code:            item.Code,
			DisplayCode:     item.DisplayCode,
			Name:            item.Name,
			Description:     item.Description,
			PermissionCodes: item.PermissionCodes,
			Custom:          item.Custom,
		})
	}
	return out
}

func ProjectFlowPermissionCatalogItemResponsesFromService(items []projectflow.ProjectPermissionItem) []ProjectFlowPermissionCatalogItemResponse {
	if items == nil {
		return nil
	}
	out := make([]ProjectFlowPermissionCatalogItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ProjectFlowPermissionCatalogItemResponse{
			Code:        item.Code,
			Description: item.Description,
		})
	}
	return out
}

func ProjectFlowMemberAccessResponseFromService(item *projectflow.MemberAccess) *ProjectFlowMemberAccessResponse {
	if item == nil {
		return nil
	}
	return &ProjectFlowMemberAccessResponse{
		UserID:                   item.UserID,
		RoleCodes:                item.RoleCodes,
		ManagedRoleCodes:         item.ManagedRoleCodes,
		EffectivePermissionCodes: item.EffectivePermissionCodes,
	}
}

func ProjectFlowReadinessResponseFromService(item projectflow.Readiness) ProjectFlowReadinessResponse {
	return ProjectFlowReadinessResponse{
		ProjectID:       item.ProjectID,
		Status:          item.Status,
		RequiredMembers: item.RequiredMembers,
		ActiveMembers:   item.ActiveMembers,
		HasProfessor:    item.HasProfessor,
		ProfessorStatus: item.ProfessorStatus,
		CriteriaCount:   item.CriteriaCount,
		CanActivate:     item.CanActivate,
	}
}
