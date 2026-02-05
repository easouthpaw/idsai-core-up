package rbac_test

import (
	"context"
	"testing"
	"time"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	wantUser  uuid.UUID
	wantPerm  string
	wantScope rbac.Scope
	retBool   bool
	retErr    error
	called    bool
}

func (f *fakeRepo) HasPermission(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope, now time.Time) (bool, error) {
	f.called = true
	f.wantUser = userID
	f.wantPerm = permissionCode
	f.wantScope = scope
	return f.retBool, f.retErr
}

func TestService_Can_InvalidScope_SystemWithID(t *testing.T) {
	repo := &fakeRepo{}
	svc := rbac.NewService(repo)

	id := uuid.New()
	ok, err := svc.Can(context.Background(), uuid.New(), "project.create", rbac.Scope{
		Type: rbac.ScopeSystem,
		ID:   &id, // invalid
	})

	require.ErrorIs(t, err, rbac.ErrInvalidScope)
	require.False(t, ok)
	require.False(t, repo.called)
}

func TestService_Can_InvalidScope_ProjectWithoutID(t *testing.T) {
	repo := &fakeRepo{}
	svc := rbac.NewService(repo)

	ok, err := svc.Can(context.Background(), uuid.New(), "task.view", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   nil, // invalid
	})

	require.ErrorIs(t, err, rbac.ErrInvalidScope)
	require.False(t, ok)
	require.False(t, repo.called)
}

func TestService_Can_DelegatesToRepo(t *testing.T) {
	repo := &fakeRepo{retBool: true, retErr: nil}
	svc := rbac.NewService(repo)

	userID := uuid.New()
	projectID := uuid.New()

	ok, err := svc.Can(context.Background(), userID, "task.view", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	})

	require.NoError(t, err)
	require.True(t, ok)
	require.True(t, repo.called)
	require.Equal(t, userID, repo.wantUser)
	require.Equal(t, "task.view", repo.wantPerm)
	require.Equal(t, rbac.ScopeProject, repo.wantScope.Type)
	require.NotNil(t, repo.wantScope.ID)
	require.Equal(t, projectID, *repo.wantScope.ID)
}
