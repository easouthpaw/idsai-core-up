package dto

import (
	"strings"
	"time"

	"idsai-core-up/internal/services/auth"
)

type RegisterRequest struct {
	Email                 string `json:"email" binding:"required,email"`
	Password              string `json:"password" binding:"required"`
	FullName              string `json:"full_name"`
	EducationType         string `json:"education_type"`
	FacultyID             string `json:"faculty_id"`
	DepartmentCode        string `json:"department_code"`
	GroupCode             string `json:"group_code"`
	SchoolClass           string `json:"school_class"`
	InstitutionProvider   string `json:"institution_provider"`
	InstitutionExternalID string `json:"institution_external_id"`
	InstitutionName       string `json:"institution_name"`
	InstitutionAddress    string `json:"institution_address"`
}

type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe *bool  `json:"remember_me"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetConfirmRequest struct {
	Token    string `json:"token"`
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password" binding:"required"`
}

type AuthStatusResponse struct {
	Status string `json:"status"`
}

type AccessResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Status       string `json:"status,omitempty"`
}

type MeResponse struct {
	UserID                string   `json:"user_id"`
	TenantID              string   `json:"tenant_id"`
	FacultyID             string   `json:"faculty_id"`
	FacultyCode           string   `json:"faculty_code"`
	DepartmentID          string   `json:"department_id"`
	DepartmentCode        string   `json:"department_code"`
	GroupID               string   `json:"group_id,omitempty"`
	GroupCode             string   `json:"group_code,omitempty"`
	GroupNumber           *int     `json:"group_number,omitempty"`
	EducationType         string   `json:"education_type"`
	SchoolClass           string   `json:"school_class,omitempty"`
	InstitutionProvider   string   `json:"institution_provider,omitempty"`
	InstitutionExternalID string   `json:"institution_external_id,omitempty"`
	InstitutionName       string   `json:"institution_name,omitempty"`
	InstitutionAddress    string   `json:"institution_address,omitempty"`
	Email                 string   `json:"email"`
	PendingEmail          string   `json:"pending_email,omitempty"`
	PendingStatus         string   `json:"pending_email_status,omitempty"`
	FullName              string   `json:"full_name"`
	AvatarURL             string   `json:"avatar_url,omitempty"`
	Headline              string   `json:"headline,omitempty"`
	About                 string   `json:"about,omitempty"`
	PreferredRole         string   `json:"preferred_role,omitempty"`
	Semester              string   `json:"semester,omitempty"`
	Availability          string   `json:"availability,omitempty"`
	Goals                 string   `json:"goals,omitempty"`
	GithubURL             string   `json:"github_url,omitempty"`
	Telegram              string   `json:"telegram,omitempty"`
	PortfolioURL          string   `json:"portfolio_url,omitempty"`
	Stacks                []string `json:"stacks,omitempty"`
	Interests             []string `json:"interests,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
	IsAdmin               bool     `json:"is_admin"`
	IsProfessor           bool     `json:"is_professor"`
	EmailVerified         bool     `json:"email_verified"`
}

type CapabilitiesResponse struct {
	ScopeType   string   `json:"scope_type"`
	ScopeID     string   `json:"scope_id,omitempty"`
	Permissions []string `json:"permissions"`
}

type UpdateProfileRequest struct {
	FullName      string   `json:"full_name"`
	Headline      string   `json:"headline"`
	About         string   `json:"about"`
	PreferredRole string   `json:"preferred_role"`
	Semester      string   `json:"semester"`
	Availability  string   `json:"availability"`
	Goals         string   `json:"goals"`
	GithubURL     string   `json:"github_url"`
	Telegram      string   `json:"telegram"`
	PortfolioURL  string   `json:"portfolio_url"`
	Stacks        []string `json:"stacks"`
	Interests     []string `json:"interests"`
}

type StartEmailChangeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ConfirmEmailChangeRequest struct {
	Token string `json:"token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func MeResponseFromUser(u auth.User) MeResponse {
	return profileResponseFromUser(u, true)
}

func PublicProfileResponseFromUser(u auth.User) MeResponse {
	return profileResponseFromUser(u, false)
}

func profileResponseFromUser(u auth.User, includePrivate bool) MeResponse {
	groupID := ""
	if u.GroupID != nil {
		groupID = u.GroupID.String()
	}
	resp := MeResponse{
		UserID:                u.ID.String(),
		TenantID:              u.TenantID.String(),
		FacultyID:             u.FacultyID.String(),
		FacultyCode:           strings.TrimSpace(u.FacultyCode),
		DepartmentID:          u.DepartmentID.String(),
		DepartmentCode:        strings.TrimSpace(u.DepartmentCode),
		GroupID:               groupID,
		GroupCode:             strings.TrimSpace(u.GroupCode),
		GroupNumber:           u.GroupNumber,
		EducationType:         auth.EducationTypeFromFacultyCode(u.FacultyCode),
		SchoolClass:           auth.SchoolClassFromGroupCode(u.GroupCode),
		InstitutionProvider:   strings.TrimSpace(u.Institution.Provider),
		InstitutionExternalID: strings.TrimSpace(u.Institution.ExternalID),
		InstitutionName:       strings.TrimSpace(u.Institution.Name),
		InstitutionAddress:    strings.TrimSpace(u.Institution.Address),
		Email:                 u.Email,
		FullName:              u.FullName,
		AvatarURL:             strings.TrimSpace(u.AvatarURL),
		Headline:              strings.TrimSpace(u.Headline),
		About:                 strings.TrimSpace(u.About),
		PreferredRole:         strings.TrimSpace(u.PreferredRole),
		Semester:              strings.TrimSpace(u.Semester),
		Availability:          strings.TrimSpace(u.Availability),
		Goals:                 strings.TrimSpace(u.Goals),
		GithubURL:             strings.TrimSpace(u.GithubURL),
		Telegram:              strings.TrimSpace(u.Telegram),
		PortfolioURL:          strings.TrimSpace(u.PortfolioURL),
		Stacks:                append([]string(nil), u.Stacks...),
		Interests:             append([]string(nil), u.Interests...),
		UpdatedAt:             u.ProfileUpdatedAt.UTC().Format(time.RFC3339),
		IsAdmin:               u.IsAdmin,
		IsProfessor:           u.IsProfessor,
		EmailVerified:         u.EmailVerifiedAt != nil,
	}
	if u.IsProfessor || u.IsAdmin {
		resp.GroupID = ""
		resp.GroupCode = ""
		resp.GroupNumber = nil
		resp.SchoolClass = ""
	}
	if includePrivate {
		resp.PendingEmail = strings.TrimSpace(u.PendingEmail)
		resp.PendingStatus = pendingEmailStatus(u)
	}
	return resp
}

func pendingEmailStatus(u auth.User) string {
	pending := strings.TrimSpace(u.PendingEmail)
	if pending == "" {
		return ""
	}
	if u.PendingEmailAt == nil {
		return "pending_verification"
	}
	return "verification_sent"
}
