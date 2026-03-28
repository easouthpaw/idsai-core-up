package admin_test

import (
	"context"
	"errors"
	"testing"

	"idsai-core-up/internal/services/admin"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeAdminRepo struct {
	deleteUserID     uuid.UUID
	deleteUserErr    error
	deleteProjectID  uuid.UUID
	deleteProjectErr error
	projectByID      admin.Project
	projectByIDErr   error
	projectStatusID  uuid.UUID
	projectStatus    string
	roleUserID       uuid.UUID
	roleCode         string
	passwordUserID   uuid.UUID
	passwordHash     string
	revokedUserID    uuid.UUID
	userByID         admin.User
	userByIDErr      error
}

func (f *fakeAdminRepo) ListUsers(ctx context.Context, roleCode, search string) ([]admin.User, error) {
	return nil, nil
}

func (f *fakeAdminRepo) ListProjects(ctx context.Context, status, search string) ([]admin.Project, error) {
	return nil, nil
}

func (f *fakeAdminRepo) GetProjectObservation(ctx context.Context, projectID uuid.UUID) (admin.ProjectObservation, error) {
	return admin.ProjectObservation{}, nil
}

func (f *fakeAdminRepo) CreateUser(ctx context.Context, in admin.CreateUserParams) (admin.User, error) {
	return admin.User{}, nil
}

func (f *fakeAdminRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (admin.User, error) {
	return f.userByID, f.userByIDErr
}

func (f *fakeAdminRepo) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return nil
}

func (f *fakeAdminRepo) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode string) error {
	f.roleUserID = userID
	f.roleCode = roleCode
	return nil
}

func (f *fakeAdminRepo) UpdateUserPasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	f.passwordUserID = userID
	f.passwordHash = passwordHash
	return nil
}

func (f *fakeAdminRepo) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	f.revokedUserID = userID
	return nil
}

func (f *fakeAdminRepo) UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status string) error {
	f.projectStatusID = projectID
	f.projectStatus = status
	return nil
}

func (f *fakeAdminRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	f.deleteUserID = userID
	return f.deleteUserErr
}

func (f *fakeAdminRepo) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	f.deleteProjectID = projectID
	return f.deleteProjectErr
}

func (f *fakeAdminRepo) GetProjectByID(ctx context.Context, projectID uuid.UUID) (admin.Project, error) {
	return f.projectByID, f.projectByIDErr
}

func TestService_DeleteUser_CallsRepo(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := admin.NewService(repo)

	userID := uuid.New()
	err := svc.DeleteUser(context.Background(), userID)
	require.NoError(t, err)
	require.Equal(t, userID, repo.deleteUserID)
}

func TestService_DeleteUser_PropagatesError(t *testing.T) {
	repo := &fakeAdminRepo{deleteUserErr: errors.New("boom")}
	svc := admin.NewService(repo)

	err := svc.DeleteUser(context.Background(), uuid.New())
	require.Error(t, err)
	require.Equal(t, "boom", err.Error())
}

func TestService_DeleteProject_CallsRepo(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := admin.NewService(repo)

	projectID := uuid.New()
	err := svc.DeleteProject(context.Background(), projectID)
	require.NoError(t, err)
	require.Equal(t, projectID, repo.deleteProjectID)
}

func TestService_SetProjectStatus_ArchivesCompletedProject(t *testing.T) {
	projectID := uuid.New()
	repo := &fakeAdminRepo{
		projectByID: admin.Project{ID: projectID, Status: "COMPLETED"},
	}
	svc := admin.NewService(repo)

	_, err := svc.SetProjectStatus(context.Background(), projectID, "archive")
	require.NoError(t, err)
	require.Equal(t, projectID, repo.projectStatusID)
	require.Equal(t, "ARCHIVE", repo.projectStatus)
}

func TestService_SetProjectStatus_RejectsArchiveForActiveProject(t *testing.T) {
	projectID := uuid.New()
	repo := &fakeAdminRepo{
		projectByID: admin.Project{ID: projectID, Status: "ACTIVE"},
	}
	svc := admin.NewService(repo)

	_, err := svc.SetProjectStatus(context.Background(), projectID, "archive")
	require.ErrorIs(t, err, admin.ErrInvalidInput)
	require.Equal(t, uuid.Nil, repo.projectStatusID)
}

func TestService_SetUserRole_UpdatesAndReturnsUser(t *testing.T) {
	userID := uuid.New()
	repo := &fakeAdminRepo{
		userByID: admin.User{ID: userID, RoleCode: admin.RoleProfessor},
	}
	svc := admin.NewService(repo)

	got, err := svc.SetUserRole(context.Background(), userID, "professor")
	require.NoError(t, err)
	require.Equal(t, userID, repo.roleUserID)
	require.Equal(t, admin.RoleProfessor, repo.roleCode)
	require.Equal(t, admin.RoleProfessor, got.RoleCode)
}

func TestService_ResetUserPassword_HashesValue(t *testing.T) {
	userID := uuid.New()
	repo := &fakeAdminRepo{}
	svc := admin.NewService(repo)

	err := svc.ResetUserPassword(context.Background(), userID, "new-password-123")
	require.NoError(t, err)
	require.Equal(t, userID, repo.passwordUserID)
	require.NotEmpty(t, repo.passwordHash)
	require.NotEqual(t, "new-password-123", repo.passwordHash)
	require.Equal(t, userID, repo.revokedUserID)
}
