package projectflow

import (
	"context"
	"sort"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

// AccessCatalogItem describes an assignable delegated project role.
type AccessCatalogItem struct {
	Code            string
	Name            string
	Description     string
	PermissionCodes []string
}

// MemberAccess describes a member's access state in a project.
type MemberAccess struct {
	UserID                   string
	RoleCodes                []string
	ManagedRoleCodes         []string
	EffectivePermissionCodes []string
}

// AssignableRoles is the static catalog of delegated roles that can be managed via the access API.
var AssignableRoles = []AccessCatalogItem{
	{
		Code:            "CO_LEAD",
		Name:            "Co-Lead",
		Description:     "Помощник тимлида: может редактировать проект, управлять набором, создавать задачи.",
		PermissionCodes: []string{"project.edit", "project.invite_professor", "project.submit_for_review", "position.create", "member.approve", "task.create", "task.assign"},
	},
	{
		Code:            "RECRUITER",
		Name:            "Recruiter",
		Description:     "Может одобрять и приглашать участников в проект.",
		PermissionCodes: []string{"member.approve"},
	},
	{
		Code:            "TASK_MANAGER",
		Name:            "Task Manager",
		Description:     "Может создавать, назначать и удалять задачи в проекте.",
		PermissionCodes: []string{"task.create", "task.assign", "task.delete"},
	},
}

var assignableCodeSet map[string]bool

func init() {
	assignableCodeSet = make(map[string]bool, len(AssignableRoles))
	for _, r := range AssignableRoles {
		assignableCodeSet[r.Code] = true
	}
}

func assignableRoleCodes() []string {
	codes := make([]string, 0, len(AssignableRoles))
	for _, r := range AssignableRoles {
		codes = append(codes, r.Code)
	}
	return codes
}

// GetAccessCatalog returns the list of assignable delegated roles.
func (s *Service) GetAccessCatalog(ctx context.Context, callerID, projectID uuid.UUID) ([]AccessCatalogItem, error) {
	if err := s.requireProjectPermission(ctx, callerID, "member.access.manage", projectID); err != nil {
		return nil, err
	}
	return AssignableRoles, nil
}

// GetMemberAccess returns the access state for a specific member.
func (s *Service) GetMemberAccess(ctx context.Context, callerID, projectID, targetUserID uuid.UUID) (*MemberAccess, error) {
	if err := s.requireProjectPermission(ctx, callerID, "member.access.manage", projectID); err != nil {
		return nil, err
	}

	status, _, err := s.accessRepo.GetMemberStatusAndCreator(ctx, targetUserID, projectID)
	if err != nil {
		return nil, err
	}
	if status != "ACTIVE" {
		return nil, ErrInvalidInput
	}

	return s.buildMemberAccess(ctx, targetUserID, projectID)
}

// ReplaceMemberAccess atomically replaces assignable roles for a member.
func (s *Service) ReplaceMemberAccess(ctx context.Context, callerID, projectID, targetUserID uuid.UUID, managedRoleCodes []string) (*MemberAccess, error) {
	if err := s.requireProjectPermission(ctx, callerID, "member.access.manage", projectID); err != nil {
		return nil, err
	}

	status, creatorID, err := s.accessRepo.GetMemberStatusAndCreator(ctx, targetUserID, projectID)
	if err != nil {
		return nil, err
	}
	if status != "ACTIVE" {
		return nil, ErrInvalidInput
	}
	if targetUserID == creatorID {
		return nil, ErrSystemManagedAccess
	}

	// Deduplicate and validate role codes.
	seen := make(map[string]bool, len(managedRoleCodes))
	validated := make([]string, 0, len(managedRoleCodes))
	for _, code := range managedRoleCodes {
		if !assignableCodeSet[code] {
			return nil, ErrUnknownRoleCode
		}
		if !seen[code] {
			seen[code] = true
			validated = append(validated, code)
		}
	}
	sort.Strings(validated)

	if err := s.accessRepo.ReplaceAssignableRoles(ctx, targetUserID, projectID, assignableRoleCodes(), validated); err != nil {
		return nil, err
	}

	return s.buildMemberAccess(ctx, targetUserID, projectID)
}

// MyPermissions returns the current user's effective permission codes in the project scope.
func (s *Service) MyPermissions(ctx context.Context, userID, projectID uuid.UUID) ([]string, error) {
	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}
	return s.authz.ListPermissionCodes(ctx, userID, scope)
}

func (s *Service) buildMemberAccess(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (*MemberAccess, error) {
	allRoles, err := s.accessRepo.ListProjectRoleCodes(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	managed := make([]string, 0, len(AssignableRoles))
	for _, code := range allRoles {
		if assignableCodeSet[code] {
			managed = append(managed, code)
		}
	}

	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}
	effective, err := s.authz.ListPermissionCodes(ctx, userID, scope)
	if err != nil {
		return nil, err
	}

	return &MemberAccess{
		UserID:                   userID.String(),
		RoleCodes:                allRoles,
		ManagedRoleCodes:         managed,
		EffectivePermissionCodes: effective,
	}, nil
}
