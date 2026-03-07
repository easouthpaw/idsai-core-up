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
}

func (f *fakeAdminRepo) ListUsers(ctx context.Context, roleCode, search string) ([]admin.User, error) {
	return nil, nil
}

func (f *fakeAdminRepo) ListProjects(ctx context.Context, status, search string) ([]admin.Project, error) {
	return nil, nil
}

func (f *fakeAdminRepo) CreateUser(ctx context.Context, in admin.CreateUserParams) (admin.User, error) {
	return admin.User{}, nil
}

func (f *fakeAdminRepo) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return nil
}

func (f *fakeAdminRepo) UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status string) error {
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
	return admin.Project{}, nil
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
