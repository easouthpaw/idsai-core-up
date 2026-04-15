package dto

import (
	"time"

	"idsai-core-up/internal/services/auth"

	"github.com/google/uuid"
)

type ListDepartmentsResponse struct {
	Departments []DepartmentResponse `json:"departments"`
}

type ListFacultiesResponse struct {
	Faculties []FacultyResponse `json:"faculties"`
}

type FacultyResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type DepartmentResponse struct {
	ID        uuid.UUID `json:"id"`
	FacultyID uuid.UUID `json:"faculty_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	ShortCode string    `json:"short_code"`
	CreatedAt time.Time `json:"created_at"`
}

type ListDepartmentGroupsResponse struct {
	Groups []StudentGroupResponse `json:"groups"`
}

type StudentGroupResponse struct {
	ID           uuid.UUID `json:"id"`
	DepartmentID uuid.UUID `json:"department_id"`
	GroupCode    string    `json:"group_code"`
	GroupNumber  int       `json:"group_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SubmitGroupChangeRequest struct {
	DepartmentCode string `json:"department_code" binding:"required"`
	GroupCode      string `json:"group_code" binding:"required"`
}

type ListGroupChangeRequestsResponse struct {
	Requests []GroupChangeRequestResponse `json:"requests"`
}

type GroupChangeRequestResponse struct {
	ID               uuid.UUID  `json:"id"`
	StudentID        uuid.UUID  `json:"student_id"`
	StudentName      string     `json:"student_name"`
	StudentEmail     string     `json:"student_email"`
	CurrentGroupID   uuid.UUID  `json:"current_group_id"`
	CurrentGroupCode string     `json:"current_group_code"`
	RequestedGroupID uuid.UUID  `json:"requested_group_id"`
	RequestedCode    string     `json:"requested_group_code"`
	Status           string     `json:"status"`
	AdminComment     string     `json:"admin_comment,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy       *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedByName   string     `json:"reviewed_by_name,omitempty"`
}

type ReviewGroupChangeRequest struct {
	Action  string `json:"action" binding:"required"`
	Comment string `json:"comment"`
}

type ListDepartmentGroupsTreeResponse struct {
	Departments []DepartmentGroupsTreeResponse `json:"departments"`
}

type DepartmentGroupsTreeResponse struct {
	ID        uuid.UUID           `json:"id"`
	Code      string              `json:"code"`
	Name      string              `json:"name"`
	ShortCode string              `json:"short_code"`
	Groups    []GroupNodeResponse `json:"groups"`
}

type GroupNodeResponse struct {
	ID            uuid.UUID              `json:"id"`
	GroupCode     string                 `json:"group_code"`
	GroupNumber   int                    `json:"group_number"`
	TotalStudents int                    `json:"total_students"`
	Students      []GroupStudentResponse `json:"students"`
}

type GroupStudentResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Status    string    `json:"status"`
	Role      string    `json:"role"`
}

func DepartmentResponsesFromService(items []auth.Department) []DepartmentResponse {
	if items == nil {
		return nil
	}
	out := make([]DepartmentResponse, 0, len(items))
	for _, item := range items {
		out = append(out, DepartmentResponse{
			ID:        item.ID,
			FacultyID: item.FacultyID,
			Code:      item.Code,
			Name:      item.Name,
			ShortCode: item.ShortCode,
			CreatedAt: item.CreatedAt,
		})
	}
	return out
}

func FacultyResponsesFromService(items []auth.Faculty) []FacultyResponse {
	if items == nil {
		return nil
	}
	out := make([]FacultyResponse, 0, len(items))
	for _, item := range items {
		out = append(out, FacultyResponse{
			ID:        item.ID,
			Code:      item.Code,
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
		})
	}
	return out
}

func StudentGroupResponsesFromService(items []auth.StudentGroup) []StudentGroupResponse {
	if items == nil {
		return nil
	}
	out := make([]StudentGroupResponse, 0, len(items))
	for _, item := range items {
		out = append(out, StudentGroupResponse{
			ID:           item.ID,
			DepartmentID: item.DepartmentID,
			GroupCode:    item.GroupCode,
			GroupNumber:  item.GroupNumber,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		})
	}
	return out
}

func GroupChangeRequestResponseFromService(item auth.GroupChangeRequest) GroupChangeRequestResponse {
	return GroupChangeRequestResponse{
		ID:               item.ID,
		StudentID:        item.StudentID,
		StudentName:      item.StudentName,
		StudentEmail:     item.StudentEmail,
		CurrentGroupID:   item.CurrentGroupID,
		CurrentGroupCode: item.CurrentGroupCode,
		RequestedGroupID: item.RequestedGroupID,
		RequestedCode:    item.RequestedCode,
		Status:           item.Status,
		AdminComment:     item.AdminComment,
		CreatedAt:        item.CreatedAt,
		ReviewedAt:       item.ReviewedAt,
		ReviewedBy:       item.ReviewedBy,
		ReviewedByName:   item.ReviewedByName,
	}
}

func GroupChangeRequestResponsesFromService(items []auth.GroupChangeRequest) []GroupChangeRequestResponse {
	if items == nil {
		return nil
	}
	out := make([]GroupChangeRequestResponse, 0, len(items))
	for _, item := range items {
		out = append(out, GroupChangeRequestResponseFromService(item))
	}
	return out
}

func DepartmentGroupsTreeResponsesFromService(items []auth.DepartmentGroupsTree) []DepartmentGroupsTreeResponse {
	if items == nil {
		return nil
	}
	out := make([]DepartmentGroupsTreeResponse, 0, len(items))
	for _, item := range items {
		out = append(out, DepartmentGroupsTreeResponse{
			ID:        item.ID,
			Code:      item.Code,
			Name:      item.Name,
			ShortCode: item.ShortCode,
			Groups:    GroupNodeResponsesFromService(item.Groups),
		})
	}
	return out
}

func GroupNodeResponsesFromService(items []auth.GroupNode) []GroupNodeResponse {
	if items == nil {
		return nil
	}
	out := make([]GroupNodeResponse, 0, len(items))
	for _, item := range items {
		out = append(out, GroupNodeResponse{
			ID:            item.ID,
			GroupCode:     item.GroupCode,
			GroupNumber:   item.GroupNumber,
			TotalStudents: item.TotalStudents,
			Students:      GroupStudentResponsesFromService(item.Students),
		})
	}
	return out
}

func GroupStudentResponsesFromService(items []auth.GroupStudent) []GroupStudentResponse {
	if items == nil {
		return nil
	}
	out := make([]GroupStudentResponse, 0, len(items))
	for _, item := range items {
		out = append(out, GroupStudentResponse{
			UserID:    item.UserID,
			FullName:  item.FullName,
			Email:     item.Email,
			AvatarURL: item.AvatarURL,
			Status:    item.Status,
			Role:      item.Role,
		})
	}
	return out
}
