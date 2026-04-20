package projectflow

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

// AccessCatalogItem describes an assignable delegated project role.
type AccessCatalogItem struct {
	Code            string
	DisplayCode     string
	Name            string
	Description     string
	PermissionCodes []string
	Custom          bool
}

type ProjectPermissionItem struct {
	Code        string
	Description string
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
		DisplayCode:     "CO_LEAD",
		Name:            "Co-Lead",
		Description:     "Помощник тимлида: может редактировать проект, управлять набором, создавать задачи.",
		PermissionCodes: []string{"project.edit", "project.invite_professor", "project.submit_for_review", "position.create", "member.approve", "task.create", "task.assign"},
	},
	{
		Code:            "RECRUITER",
		DisplayCode:     "RECRUITER",
		Name:            "Recruiter",
		Description:     "Может одобрять и приглашать участников в проект.",
		PermissionCodes: []string{"member.approve"},
	},
	{
		Code:            "TASK_MANAGER",
		DisplayCode:     "TASK_MANAGER",
		Name:            "Task Manager",
		Description:     "Может создавать, назначать и удалять задачи в проекте.",
		PermissionCodes: []string{"task.create", "task.assign", "task.delete"},
	},
}

var ProjectAssignablePermissions = []ProjectPermissionItem{
	{Code: "project.view", Description: "Просматривать проект и базовую информацию."},
	{Code: "project.edit", Description: "Редактировать карточку проекта."},
	{Code: "project.invite_professor", Description: "Приглашать преподавателя в проект."},
	{Code: "project.submit_for_review", Description: "Отправлять проект на оценивание."},
	{Code: "position.create", Description: "Создавать рабочие роли и места в команде."},
	{Code: "member.approve", Description: "Принимать, приглашать и удалять участников."},
	{Code: "member.access.manage", Description: "Создавать роли доступа и назначать их участникам."},
	{Code: "task.view", Description: "Просматривать задачи проекта."},
	{Code: "task.create", Description: "Создавать задачи."},
	{Code: "task.assign", Description: "Назначать задачи участникам."},
	{Code: "task.update", Description: "Менять статус и содержание задач."},
	{Code: "task.delete", Description: "Удалять задачи."},
	{Code: "task.claim", Description: "Брать задачи в работу на себя."},
	{Code: "grading.view", Description: "Просматривать критерии и оценки проекта."},
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

func projectPermissionCodeSet() map[string]bool {
	out := make(map[string]bool, len(ProjectAssignablePermissions))
	for _, item := range ProjectAssignablePermissions {
		out[item.Code] = true
	}
	return out
}

func normalizeProjectAccessRoleCode(code string) (string, error) {
	raw := strings.TrimSpace(code)
	if raw == "" {
		return "", ErrInvalidInput
	}

	var b strings.Builder
	prevUnderscore := false
	for _, r := range raw {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToUpper(r))
			prevUnderscore = false
		case r == '_' || r == '-' || unicode.IsSpace(r):
			if !prevUnderscore && b.Len() > 0 {
				b.WriteRune('_')
				prevUnderscore = true
			}
		}
	}

	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "", ErrInvalidInput
	}
	if assignableCodeSet[out] || out == "TEAM_LEAD" || out == "MEMBER" || out == "INVITED_MEMBER" || out == "PROJECT_PROFESSOR" {
		return "", fmt.Errorf("%w: reserved role code", ErrInvalidInput)
	}
	return out, nil
}

func projectAccessRoleCode(projectID uuid.UUID, displayCode string) string {
	return "PROJECT_" + strings.ReplaceAll(projectID.String(), "-", "") + "_" + displayCode
}

func accessCatalogCodes(items []AccessCatalogItem) []string {
	codes := assignableRoleCodes()
	for _, item := range items {
		if strings.TrimSpace(item.Code) != "" {
			codes = append(codes, item.Code)
		}
	}
	sort.Strings(codes)
	return codes
}

// GetAccessCatalog returns the list of assignable delegated roles.
func (s *Service) GetAccessCatalog(ctx context.Context, callerID, projectID uuid.UUID) ([]AccessCatalogItem, error) {
	if err := s.requireProjectPermission(ctx, callerID, "member.access.manage", projectID); err != nil {
		return nil, err
	}
	custom, err := s.accessRepo.ListProjectAccessRoles(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]AccessCatalogItem, 0, len(AssignableRoles)+len(custom))
	out = append(out, AssignableRoles...)
	out = append(out, custom...)
	return out, nil
}

func (s *Service) ListProjectAccessPermissions(ctx context.Context, callerID, projectID uuid.UUID) ([]ProjectPermissionItem, error) {
	if err := s.requireProjectPermission(ctx, callerID, "member.access.manage", projectID); err != nil {
		return nil, err
	}
	out := append([]ProjectPermissionItem(nil), ProjectAssignablePermissions...)
	return out, nil
}

func (s *Service) CreateProjectAccessRole(ctx context.Context, callerID, projectID uuid.UUID, displayCode, name, description string, permissionCodes []string) (AccessCatalogItem, error) {
	if err := s.requireProjectPermission(ctx, callerID, "member.access.manage", projectID); err != nil {
		return AccessCatalogItem{}, err
	}

	code, err := normalizeProjectAccessRoleCode(displayCode)
	if err != nil {
		return AccessCatalogItem{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return AccessCatalogItem{}, ErrInvalidInput
	}
	description = strings.TrimSpace(description)

	existing, err := s.accessRepo.ListProjectAccessRoles(ctx, projectID)
	if err != nil {
		return AccessCatalogItem{}, err
	}
	for _, item := range existing {
		if strings.EqualFold(item.DisplayCode, code) {
			return AccessCatalogItem{}, fmt.Errorf("%w: role code already exists", ErrInvalidInput)
		}
	}

	allowed := projectPermissionCodeSet()
	seen := make(map[string]bool, len(permissionCodes))
	validated := make([]string, 0, len(permissionCodes))
	for _, raw := range permissionCodes {
		perm := strings.TrimSpace(raw)
		if perm == "" {
			continue
		}
		if !allowed[perm] {
			return AccessCatalogItem{}, fmt.Errorf("%w: permission is not assignable", ErrInvalidInput)
		}
		if !seen[perm] {
			seen[perm] = true
			validated = append(validated, perm)
		}
	}
	sort.Strings(validated)

	return s.accessRepo.CreateProjectAccessRole(ctx, projectID, callerID, projectAccessRoleCode(projectID, code), code, name, description, validated)
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

	customRoles, err := s.accessRepo.ListProjectAccessRoles(ctx, projectID)
	if err != nil {
		return nil, err
	}
	allowedCodes := make(map[string]bool, len(AssignableRoles)+len(customRoles))
	for _, code := range assignableRoleCodes() {
		allowedCodes[code] = true
	}
	for _, role := range customRoles {
		allowedCodes[role.Code] = true
	}

	// Deduplicate and validate role codes. Only one delegated/custom role is allowed.
	seen := make(map[string]bool, len(managedRoleCodes))
	validated := make([]string, 0, len(managedRoleCodes))
	for _, code := range managedRoleCodes {
		if !allowedCodes[code] {
			return nil, ErrUnknownRoleCode
		}
		if !seen[code] {
			seen[code] = true
			validated = append(validated, code)
		}
	}
	if len(validated) > 1 {
		return nil, fmt.Errorf("%w: only one managed role is allowed", ErrInvalidInput)
	}
	sort.Strings(validated)

	if err := s.accessRepo.ReplaceAssignableRoles(ctx, targetUserID, projectID, accessCatalogCodes(customRoles), validated); err != nil {
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
	customRoles, err := s.accessRepo.ListProjectAccessRoles(ctx, projectID)
	if err != nil {
		return nil, err
	}
	managedSet := make(map[string]bool, len(AssignableRoles)+len(customRoles))
	for _, code := range assignableRoleCodes() {
		managedSet[code] = true
	}
	for _, role := range customRoles {
		managedSet[role.Code] = true
	}
	for _, code := range allRoles {
		if managedSet[code] {
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
