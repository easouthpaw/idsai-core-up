package projects_test

import (
	"context"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projects"
	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeProjectsRepo struct {
	id                    uuid.UUID
	err                   error
	project               domain.Project
	getErr                error
	hasProjectPermission  bool
	hasProjectPermissionE error
	list                  []domain.Project
	listErr               error
	listPublic            []domain.Project
	listPublicErr         error
	groupID               uuid.UUID
	groupErr              error
	groups                []projects.Group
	groupsErr             error
}

func (f fakeProjectsRepo) Create(ctx context.Context, title, description string, facultyID uuid.UUID, visibility string, groupID *uuid.UUID, createdBy uuid.UUID) (uuid.UUID, error) {
	return f.id, f.err
}

func (f fakeProjectsRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	return f.project, f.getErr
}

func (f fakeProjectsRepo) HasProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permissionCode string) (bool, error) {
	return f.hasProjectPermission, f.hasProjectPermissionE
}

func (f fakeProjectsRepo) ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]domain.Project, error) {
	return f.list, f.listErr
}

func (f fakeProjectsRepo) ListPublic(ctx context.Context) ([]domain.Project, error) {
	if f.listPublic != nil || f.listPublicErr != nil {
		return f.listPublic, f.listPublicErr
	}
	return f.list, f.listErr
}

func (f fakeProjectsRepo) FindGroupIDByCode(ctx context.Context, facultyID uuid.UUID, code string) (uuid.UUID, error) {
	return f.groupID, f.groupErr
}

func (f fakeProjectsRepo) ListGroupsByFaculty(ctx context.Context, facultyID uuid.UUID) ([]projects.Group, error) {
	return f.groups, f.groupsErr
}

type fakeGrantor struct {
	called   bool
	userID   uuid.UUID
	roleCode string
	scope    rbac.Scope
	err      error
}

func TestService_GetProject_ReturnsProject(t *testing.T) {
	pid := uuid.New()
	fid := uuid.New()
	uid := uuid.New()

	want := domain.Project{
		ID:          pid,
		Title:       "T",
		Description: "D",
		Status:      domain.ProjectDraft,
		FacultyID:   fid,
		Visibility:  "FACULTY",
		CreatedBy:   uid,
	}

	repo := fakeProjectsRepo{project: want}
	grantor := &fakeGrantor{}
	svc := projects.NewService(repo, grantor)

	got, err := svc.GetProject(context.Background(), pid)
	require.NoError(t, err)
	require.Equal(t, want.ID, got.ID)
	require.Equal(t, want.Title, got.Title)
	require.Equal(t, want.FacultyID, got.FacultyID)
	require.Equal(t, want.Visibility, got.Visibility)
}

func TestService_GetProjectForViewer_AllowsPublicProject(t *testing.T) {
	pid := uuid.New()
	viewerID := uuid.New()
	ownerID := uuid.New()
	viewerFacultyID := uuid.New()

	repo := fakeProjectsRepo{
		project: domain.Project{
			ID:        pid,
			Title:     "Public Project",
			IsPublic:  true,
			CreatedBy: ownerID,
		},
	}
	svc := projects.NewService(repo, &fakeGrantor{})

	got, err := svc.GetProjectForViewer(context.Background(), pid, viewerID, viewerFacultyID)
	require.NoError(t, err)
	require.Equal(t, pid, got.ID)
	require.True(t, got.IsPublic)
}

func TestService_GetProjectForViewer_DeniesPrivateWithoutPermission(t *testing.T) {
	pid := uuid.New()
	viewerID := uuid.New()
	ownerID := uuid.New()
	viewerFacultyID := uuid.New()

	repo := fakeProjectsRepo{
		project: domain.Project{
			ID:        pid,
			Title:     "Private Project",
			IsPublic:  false,
			CreatedBy: ownerID,
		},
		hasProjectPermission: false,
	}
	svc := projects.NewService(repo, &fakeGrantor{})

	_, err := svc.GetProjectForViewer(context.Background(), pid, viewerID, viewerFacultyID)
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestService_GetProjectForViewer_AllowsPrivateWithPermission(t *testing.T) {
	pid := uuid.New()
	viewerID := uuid.New()
	ownerID := uuid.New()
	viewerFacultyID := uuid.New()

	repo := fakeProjectsRepo{
		project: domain.Project{
			ID:        pid,
			Title:     "Private Project",
			IsPublic:  false,
			CreatedBy: ownerID,
		},
		hasProjectPermission: true,
	}
	svc := projects.NewService(repo, &fakeGrantor{})

	got, err := svc.GetProjectForViewer(context.Background(), pid, viewerID, viewerFacultyID)
	require.NoError(t, err)
	require.Equal(t, pid, got.ID)
}

func TestService_GetProjectForViewer_AllowsRecruitmentForSameFaculty(t *testing.T) {
	pid := uuid.New()
	viewerID := uuid.New()
	ownerID := uuid.New()
	facultyID := uuid.New()

	repo := fakeProjectsRepo{
		project: domain.Project{
			ID:        pid,
			Title:     "Recruitment Project",
			Status:    domain.ProjectRecruitment,
			IsPublic:  false,
			CreatedBy: ownerID,
			FacultyID: facultyID,
		},
		hasProjectPermission: false,
	}
	svc := projects.NewService(repo, &fakeGrantor{})

	got, err := svc.GetProjectForViewer(context.Background(), pid, viewerID, facultyID)
	require.NoError(t, err)
	require.Equal(t, pid, got.ID)
}

func (g *fakeGrantor) GrantRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scope rbac.Scope, expiresAt *time.Time) error {
	g.called = true
	g.userID = userID
	g.roleCode = roleCode
	g.scope = scope
	return g.err
}

func TestService_CreateProject_GrantsTeamLead(t *testing.T) {
	projectID := uuid.New()
	createdBy := uuid.New()

	repo := fakeProjectsRepo{id: projectID}
	grantor := &fakeGrantor{}

	svc := projects.NewService(repo, grantor)

	facultyID := uuid.New()
	gotID, err := svc.CreateProject(context.Background(), "X", "Y", facultyID, "FACULTY", nil, createdBy)
	require.NoError(t, err)
	require.Equal(t, projectID, gotID)

	require.True(t, grantor.called)
	require.Equal(t, createdBy, grantor.userID)
	require.Equal(t, "TEAM_LEAD", grantor.roleCode)
	require.Equal(t, rbac.ScopeProject, grantor.scope.Type)
	require.NotNil(t, grantor.scope.ID)
	require.Equal(t, projectID, *grantor.scope.ID)
}

func TestService_ListPublicProjects_ReturnsItems(t *testing.T) {
	creator := uuid.New()
	items := []domain.Project{
		{
			ID:        uuid.New(),
			Title:     "Public One",
			IsPublic:  true,
			CreatedBy: creator,
		},
	}

	repo := fakeProjectsRepo{listPublic: items}
	svc := projects.NewService(repo, &fakeGrantor{})

	got, err := svc.ListPublicProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "Public One", got[0].Title)
	require.True(t, got[0].IsPublic)
}
