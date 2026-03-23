package projects

import (
	"context"
	"errors"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("project not found")
	ErrGroupNotFound = errors.New("group not found")
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

func (s *Service) GetProjectForViewer(ctx context.Context, projectID, viewerID, viewerFacultyID uuid.UUID) (domain.Project, error) {
	view, err := s.GetProjectViewForViewer(ctx, projectID, viewerID, viewerFacultyID)
	if err != nil {
		return domain.Project{}, err
	}
	return view.Project, nil
}

func (s *Service) GetProjectViewForViewer(ctx context.Context, projectID, viewerID, viewerFacultyID uuid.UUID) (ProjectView, error) {
	p, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return ProjectView{}, err
	}

	access := ViewerAccess{}
	if p.CreatedBy == viewerID {
		access.CanViewWorkspace = true
	} else {
		access.CanViewWorkspace, err = s.repo.HasProjectPermission(ctx, viewerID, projectID, "grading.view")
		if err != nil {
			return ProjectView{}, err
		}
	}

	hasProjectView := access.CanViewWorkspace || p.CreatedBy == viewerID
	if !hasProjectView {
		hasProjectView, err = s.repo.HasProjectPermission(ctx, viewerID, projectID, "project.view")
		if err != nil {
			return ProjectView{}, err
		}
	}

	sameFacultyRecruitment := p.Status == domain.ProjectRecruitment && p.FacultyID == viewerFacultyID
	access.CanApply = sameFacultyRecruitment && !hasProjectView && p.CreatedBy != viewerID

	if !(p.IsPublic || hasProjectView || sameFacultyRecruitment) {
		return ProjectView{}, domain.ErrForbidden
	}

	var summary *ReviewSummary
	access.CanViewFinalGrade = p.Status == domain.ProjectArchive && (p.IsPublic || access.CanViewWorkspace)
	if access.CanViewFinalGrade {
		summary, err = s.repo.GetProjectReviewSummary(ctx, projectID)
		if err != nil {
			return ProjectView{}, err
		}
	}

	return ProjectView{
		Project:       p,
		Access:        access,
		ReviewSummary: summary,
	}, nil
}

func (s *Service) ListProjectsByCreator(ctx context.Context, createdBy uuid.UUID) ([]domain.Project, error) {
	return s.repo.ListByCreator(ctx, createdBy)
}

func (s *Service) ListPublicProjects(ctx context.Context) ([]domain.Project, error) {
	return s.repo.ListPublic(ctx)
}

func (s *Service) ResolveGroupByCode(ctx context.Context, facultyID uuid.UUID, groupCode string) (uuid.UUID, error) {
	return s.repo.FindGroupIDByCode(ctx, facultyID, groupCode)
}

func (s *Service) ListGroupsByFaculty(ctx context.Context, facultyID uuid.UUID) ([]Group, error) {
	return s.repo.ListGroupsByFaculty(ctx, facultyID)
}
