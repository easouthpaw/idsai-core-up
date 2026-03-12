package projectflow

import (
	"context"
	"errors"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

func (s *Service) projectByID(ctx context.Context, projectID uuid.UUID) (domain.Project, error) {
	return s.projectsRepo.GetProjectByID(ctx, projectID)
}

func (s *Service) requireProjectPermission(ctx context.Context, userID uuid.UUID, permission string, projectID uuid.UUID) error {
	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}
	ok, err := s.authz.Can(ctx, userID, permission, scope)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) requireFacultyPermission(ctx context.Context, userID uuid.UUID, permission string, facultyID uuid.UUID) error {
	scope := rbac.Scope{Type: rbac.ScopeFaculty, ID: &facultyID}
	ok, err := s.authz.Can(ctx, userID, permission, scope)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) isActiveProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	return s.projectsRepo.IsActiveProjectMember(ctx, userID, projectID)
}

func (s *Service) requireProjectEditAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	if err := s.requireProjectPermission(ctx, userID, "project.edit", projectID); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrForbidden) {
		return err
	}

	ok, err := s.isActiveProjectMember(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) requireProjectSubmitAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	if err := s.requireProjectPermission(ctx, userID, "project.submit_for_review", projectID); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrForbidden) {
		return err
	}

	ok, err := s.isActiveProjectMember(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) ensureProjectRole(ctx context.Context, userID uuid.UUID, roleCode string, projectID uuid.UUID) error {
	ok, err := s.projectsRepo.HasProjectRole(ctx, userID, projectID, roleCode)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return s.grantor.GrantRoleByCode(ctx, userID, roleCode, rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}, nil)
}

func (s *Service) revokeProjectRole(ctx context.Context, userID uuid.UUID, roleCode string, projectID uuid.UUID) error {
	return s.projectsRepo.RevokeProjectRole(ctx, userID, projectID, roleCode)
}
