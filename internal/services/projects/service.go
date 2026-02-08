package projects

import (
	"context"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

type Service struct {
	repo    Repository
	grantor RoleGrantor
}

func NewService(repo Repository, grantor RoleGrantor) *Service {
	return &Service{repo: repo, grantor: grantor}
}

func (s *Service) CreateProject(ctx context.Context, title, description string, facultyID uuid.UUID, visibility string, groupID *uuid.UUID, createdBy uuid.UUID) (uuid.UUID, error) {
	projectID, err := s.repo.Create(ctx, title, description, facultyID, visibility, groupID, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}
	if err := s.grantor.GrantRoleByCode(ctx, createdBy, "TEAM_LEAD", scope, nil); err != nil {
		return uuid.Nil, err
	}

	return projectID, nil
}

func (s *Service) GetProject(ctx context.Context, projectID uuid.UUID) (domain.Project, error) {
	return s.repo.GetByID(ctx, projectID)
}
