package admin_test

import (
	"context"
	"testing"

	"idsai-core-up/internal/services/admin"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type recordingAdminRepo struct {
	usersOut        []admin.User
	usersErr        error
	listUsersRole   string
	listUsersSearch string

	projectsOut        []admin.Project
	projectsErr        error
	listProjectsStatus string
	listProjectsSearch string

	createUserIn  admin.CreateUserParams
	createUserOut admin.User
	createUserErr error

	statusUserID uuid.UUID
	statusValue  string
	statusErr    error
	revokeErr    error

	observation    admin.ProjectObservation
	observationID  uuid.UUID
	observationErr error
}

func (r *recordingAdminRepo) ListUsers(ctx context.Context, roleCode, search string) ([]admin.User, error) {
	r.listUsersRole = roleCode
	r.listUsersSearch = search
	return r.usersOut, r.usersErr
}

func (r *recordingAdminRepo) ListProjects(ctx context.Context, status, search string) ([]admin.Project, error) {
	r.listProjectsStatus = status
	r.listProjectsSearch = search
	return r.projectsOut, r.projectsErr
}

func (r *recordingAdminRepo) GetProjectObservation(ctx context.Context, projectID uuid.UUID) (admin.ProjectObservation, error) {
	r.observationID = projectID
	return r.observation, r.observationErr
}

func (r *recordingAdminRepo) CreateUser(ctx context.Context, in admin.CreateUserParams) (admin.User, error) {
	r.createUserIn = in
	return r.createUserOut, r.createUserErr
}

func (r *recordingAdminRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (admin.User, error) {
	return admin.User{}, nil
}

func (r *recordingAdminRepo) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	r.statusUserID = userID
	r.statusValue = status
	return r.statusErr
}

func (r *recordingAdminRepo) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode string) error {
	return nil
}

func (r *recordingAdminRepo) UpdateUserPasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return nil
}

func (r *recordingAdminRepo) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	return r.revokeErr
}

func (r *recordingAdminRepo) UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status string) error {
	return nil
}

func (r *recordingAdminRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (r *recordingAdminRepo) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	return nil
}

func (r *recordingAdminRepo) GetProjectByID(ctx context.Context, projectID uuid.UUID) (admin.Project, error) {
	return admin.Project{}, nil
}

func TestServiceListUsersAndProjects_NormalizesFilters(t *testing.T) {
	repo := &recordingAdminRepo{
		usersOut:    []admin.User{{ID: uuid.New(), RoleCode: admin.RoleProfessor}},
		projectsOut: []admin.Project{{ID: uuid.New(), Status: "REVIEW"}},
	}
	svc := admin.NewService(repo)

	users, err := svc.ListUsers(context.Background(), " professor ", "  Alice  ")
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, admin.RoleProfessor, repo.listUsersRole)
	require.Equal(t, "Alice", repo.listUsersSearch)

	projects, err := svc.ListProjects(context.Background(), " review ", "  AI  ")
	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, "REVIEW", repo.listProjectsStatus)
	require.Equal(t, "AI", repo.listProjectsSearch)
}

func TestServiceCreateUser_NormalizesPayloadAndHashesPassword(t *testing.T) {
	repo := &recordingAdminRepo{
		createUserOut: admin.User{ID: uuid.New(), RoleCode: admin.RoleStudent},
	}
	svc := admin.NewService(repo)

	user, err := svc.CreateUser(context.Background(), admin.CreateUserInput{
		Email:          " STUDENT@EXAMPLE.EDU ",
		Password:       "strong-password-123",
		FullName:       " Student Example ",
		DepartmentCode: " cpi ",
		RoleCode:       " student ",
	})

	require.NoError(t, err)
	require.Equal(t, admin.RoleStudent, user.RoleCode)
	require.Equal(t, "student@example.edu", repo.createUserIn.Email)
	require.Equal(t, "Student Example", repo.createUserIn.FullName)
	require.Equal(t, "CPI", repo.createUserIn.DepartmentCode)
	require.Equal(t, admin.RoleStudent, repo.createUserIn.RoleCode)
	require.NotEmpty(t, repo.createUserIn.PasswordHash)
	require.NotEqual(t, "strong-password-123", repo.createUserIn.PasswordHash)
}

func TestServiceSetUserStatus_DisabledRevokesSessions(t *testing.T) {
	repo := &recordingAdminRepo{}
	svc := admin.NewService(repo)
	userID := uuid.New()

	err := svc.SetUserStatus(context.Background(), userID, " disabled ")
	require.NoError(t, err)
	require.Equal(t, userID, repo.statusUserID)
	require.Equal(t, admin.StatusDisabled, repo.statusValue)
}

func TestServiceObserveProject_ReturnsObservation(t *testing.T) {
	projectID := uuid.New()
	repo := &recordingAdminRepo{
		observation: admin.ProjectObservation{
			Project: admin.Project{ID: projectID, Title: "AI Platform"},
			Summary: admin.ProjectObservationSummary{TasksTotal: 3},
		},
	}
	svc := admin.NewService(repo)

	got, err := svc.ObserveProject(context.Background(), projectID)
	require.NoError(t, err)
	require.Equal(t, projectID, repo.observationID)
	require.Equal(t, "AI Platform", got.Project.Title)
	require.Equal(t, 3, got.Summary.TasksTotal)
}
