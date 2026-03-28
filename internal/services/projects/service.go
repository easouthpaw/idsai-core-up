package projects

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/infra/images"
	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrNotFound      = errors.New("project not found")
	ErrGroupNotFound = errors.New("group not found")
	ErrStorage       = errors.New("storage unavailable")
)

type ObjectStorage interface {
	PutObject(ctx context.Context, key, contentType string, body []byte) error
	DeleteObject(ctx context.Context, key string) error
	PublicURL(key string) string
	Available() bool
}

type Service struct {
	repo    Repository
	grantor RoleGrantor
	storage ObjectStorage
}

func NewService(repo Repository, grantor RoleGrantor) *Service {
	return &Service{repo: repo, grantor: grantor}
}

func (s *Service) SetStorage(storage ObjectStorage) {
	s.storage = storage
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
	p, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	return s.decorateProjectMedia(p), nil
}

func (s *Service) GetProjectForViewer(ctx context.Context, projectID, viewerID, viewerFacultyID uuid.UUID) (domain.Project, error) {
	view, err := s.GetProjectViewForViewer(ctx, projectID, viewerID, viewerFacultyID)
	if err != nil {
		return domain.Project{}, err
	}
	return s.decorateProjectMedia(view.Project), nil
}

func (s *Service) GetProjectViewForViewer(ctx context.Context, projectID, viewerID, viewerFacultyID uuid.UUID) (ProjectView, error) {
	p, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return ProjectView{}, err
	}
	p = s.decorateProjectMedia(p)

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
		hasProjectView, err = s.repo.HasResolvedProjectPermission(ctx, viewerID, projectID, "project.view")
		if err != nil {
			return ProjectView{}, err
		}
	}
	access.CanViewProjectDetails = hasProjectView

	sameFacultyRecruitment := p.Status == domain.ProjectRecruitment && p.FacultyID == viewerFacultyID
	access.CanApply = sameFacultyRecruitment && !hasProjectView && p.CreatedBy != viewerID

	if !(p.IsPublic || access.CanViewProjectDetails || sameFacultyRecruitment) {
		return ProjectView{}, domain.ErrForbidden
	}

	var summary *ReviewSummary
	access.CanViewFinalGrade = (p.Status == domain.ProjectCompleted || p.Status == domain.ProjectArchive) && (p.IsPublic || access.CanViewProjectDetails)
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
	items, err := s.repo.ListByCreator(ctx, createdBy)
	if err != nil {
		return nil, err
	}
	return s.decorateProjectsMedia(items), nil
}

func (s *Service) ListProjectsByFaculty(ctx context.Context, facultyID uuid.UUID) ([]domain.Project, error) {
	items, err := s.repo.ListByFaculty(ctx, facultyID)
	if err != nil {
		return nil, err
	}
	return s.decorateProjectsMedia(items), nil
}

func (s *Service) ListPublicProjects(ctx context.Context) ([]domain.Project, error) {
	items, err := s.repo.ListPublic(ctx)
	if err != nil {
		return nil, err
	}
	return s.decorateProjectsMedia(items), nil
}

func (s *Service) ResolveGroupByCode(ctx context.Context, facultyID uuid.UUID, groupCode string) (uuid.UUID, error) {
	return s.repo.FindGroupIDByCode(ctx, facultyID, groupCode)
}

func (s *Service) ListGroupsByFaculty(ctx context.Context, facultyID uuid.UUID) ([]Group, error) {
	return s.repo.ListGroupsByFaculty(ctx, facultyID)
}

func (s *Service) UpdateProjectImage(ctx context.Context, userID, projectID uuid.UUID, raw []byte) (domain.Project, error) {
	if userID == uuid.Nil || projectID == uuid.Nil || len(raw) == 0 {
		return domain.Project{}, ErrInvalidInput
	}
	if s.storage == nil || !s.storage.Available() {
		return domain.Project{}, ErrStorage
	}

	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if err := s.ensureCanEditProject(ctx, userID, project); err != nil {
		return domain.Project{}, err
	}

	processed, contentType, err := images.ProcessProjectCover(raw)
	if err != nil {
		return domain.Project{}, err
	}

	now := time.Now().UTC()
	key := fmt.Sprintf("projects/covers/%s/%d-%s.jpg", projectID.String(), now.Unix(), uuid.NewString())
	if err := s.storage.PutObject(ctx, key, contentType, processed); err != nil {
		return domain.Project{}, ErrStorage
	}
	if err := s.repo.SetProjectImage(ctx, projectID, key, now); err != nil {
		_ = s.storage.DeleteObject(ctx, key)
		return domain.Project{}, err
	}
	if strings.TrimSpace(project.ImageKey) != "" && project.ImageKey != key {
		_ = s.storage.DeleteObject(ctx, project.ImageKey)
	}

	updated, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	return s.decorateProjectMedia(updated), nil
}

func (s *Service) DeleteProjectImage(ctx context.Context, userID, projectID uuid.UUID) (domain.Project, error) {
	if userID == uuid.Nil || projectID == uuid.Nil {
		return domain.Project{}, ErrInvalidInput
	}
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if err := s.ensureCanEditProject(ctx, userID, project); err != nil {
		return domain.Project{}, err
	}

	oldKey := strings.TrimSpace(project.ImageKey)
	if err := s.repo.ClearProjectImage(ctx, projectID); err != nil {
		return domain.Project{}, err
	}
	if oldKey != "" && s.storage != nil && s.storage.Available() {
		_ = s.storage.DeleteObject(ctx, oldKey)
	}

	updated, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	return s.decorateProjectMedia(updated), nil
}

func (s *Service) ResolveProjectImageURL(imageKey string) string {
	imageKey = strings.TrimSpace(imageKey)
	if imageKey == "" || s.storage == nil {
		return ""
	}
	return strings.TrimSpace(s.storage.PublicURL(imageKey))
}

func (s *Service) decorateProjectsMedia(items []domain.Project) []domain.Project {
	if len(items) == 0 {
		return items
	}
	out := make([]domain.Project, 0, len(items))
	for _, item := range items {
		out = append(out, s.decorateProjectMedia(item))
	}
	return out
}

func (s *Service) decorateProjectMedia(item domain.Project) domain.Project {
	if item.DefaultCoverVariant <= 0 {
		item.DefaultCoverVariant = 1
	}
	item.ImageURL = s.ResolveProjectImageURL(item.ImageKey)
	return item
}

func (s *Service) ensureCanEditProject(ctx context.Context, userID uuid.UUID, project domain.Project) error {
	if project.CreatedBy == userID {
		return nil
	}
	allowed, err := s.repo.HasProjectPermission(ctx, userID, project.ID, "project.edit")
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}
