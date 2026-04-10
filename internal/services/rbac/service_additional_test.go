package rbac_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type trackingRepo struct {
	results map[string]bool
	errFor  string
	err     error
	calls   []string
	lastNow time.Time
}

func (r *trackingRepo) HasPermission(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope, now time.Time) (bool, error) {
	r.calls = append(r.calls, permissionCode)
	r.lastNow = now
	if r.errFor == permissionCode && r.err != nil {
		return false, r.err
	}
	return r.results[permissionCode], nil
}

func (r *trackingRepo) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope rbac.Scope, now time.Time) ([]string, error) {
	r.lastNow = now
	return nil, nil
}

func TestConditionRegistry_RegisterEvaluateAndHas(t *testing.T) {
	registry := rbac.NewConditionRegistry()

	require.False(t, registry.Has("task.update"))
	require.True(t, registry.Evaluate("task.update", nil))

	registry.Register("task.update", func(attrs map[string]interface{}) bool {
		status, _ := attrs["status"].(string)
		return status == "IN_PROGRESS"
	})
	registry.Register("task.update", func(attrs map[string]interface{}) bool {
		role, _ := attrs["actor_role"].(string)
		return role == "TEAM_LEAD"
	})

	require.True(t, registry.Has("task.update"))
	require.True(t, registry.Evaluate("task.update", map[string]interface{}{
		"status":     "IN_PROGRESS",
		"actor_role": "TEAM_LEAD",
	}))
	require.False(t, registry.Evaluate("task.update", map[string]interface{}{
		"status":     "DONE",
		"actor_role": "TEAM_LEAD",
	}))
}

func TestService_CanAll_AllowsWhenAllPermissionsPresent(t *testing.T) {
	repo := &trackingRepo{
		results: map[string]bool{
			"project.view": true,
			"task.view":    true,
			"task.update":  true,
		},
	}
	svc := rbac.NewService(repo, nil)

	projectID := uuid.New()
	ok, err := svc.CanAll(context.Background(), uuid.New(), []string{
		"project.view",
		"task.view",
		"task.update",
	}, rbac.Scope{Type: rbac.ScopeProject, ID: &projectID})

	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []string{"project.view", "task.view", "task.update"}, repo.calls)
}

func TestService_CanAll_StopsOnFirstDeniedPermission(t *testing.T) {
	repo := &trackingRepo{
		results: map[string]bool{
			"project.view": true,
			"task.view":    false,
			"task.update":  true,
		},
	}
	svc := rbac.NewService(repo, nil)

	projectID := uuid.New()
	ok, err := svc.CanAll(context.Background(), uuid.New(), []string{
		"project.view",
		"task.view",
		"task.update",
	}, rbac.Scope{Type: rbac.ScopeProject, ID: &projectID})

	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, []string{"project.view", "task.view"}, repo.calls)
}

func TestService_CanAll_PropagatesRepoError(t *testing.T) {
	repo := &trackingRepo{
		results: map[string]bool{
			"project.view": true,
		},
		errFor: "task.view",
		err:    errors.New("repo unavailable"),
	}
	svc := rbac.NewService(repo, nil)

	projectID := uuid.New()
	ok, err := svc.CanAll(context.Background(), uuid.New(), []string{
		"project.view",
		"task.view",
	}, rbac.Scope{Type: rbac.ScopeProject, ID: &projectID})

	require.ErrorContains(t, err, "repo unavailable")
	require.False(t, ok)
	require.Equal(t, []string{"project.view", "task.view"}, repo.calls)
}

func TestService_CanWithAttributes_DeniesWhenConditionFails(t *testing.T) {
	repo := &trackingRepo{
		results: map[string]bool{
			"task.update": true,
		},
	}
	registry := rbac.NewConditionRegistry()
	registry.Register("task.update", func(attrs map[string]interface{}) bool {
		role, _ := attrs["actor_role"].(string)
		return role == "TEAM_LEAD"
	})

	svc := rbac.NewService(repo, registry)
	projectID := uuid.New()
	ok, err := svc.CanWithAttributes(context.Background(), uuid.New(), "task.update", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	}, map[string]interface{}{
		"actor_role": "MEMBER",
	})

	require.NoError(t, err)
	require.False(t, ok)
	require.Equal(t, []string{"task.update"}, repo.calls)
}

func TestService_CanWithAttributes_AllowsWhenConditionsPass(t *testing.T) {
	repo := &trackingRepo{
		results: map[string]bool{
			"task.update": true,
		},
	}
	registry := rbac.NewConditionRegistry()
	registry.Register("task.update", func(attrs map[string]interface{}) bool {
		status, _ := attrs["status"].(string)
		return status == "IN_PROGRESS"
	})
	registry.Register("task.update", func(attrs map[string]interface{}) bool {
		role, _ := attrs["actor_role"].(string)
		return role == "TEAM_LEAD"
	})

	svc := rbac.NewService(repo, registry)
	projectID := uuid.New()
	ok, err := svc.CanWithAttributes(context.Background(), uuid.New(), "task.update", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	}, map[string]interface{}{
		"status":     "IN_PROGRESS",
		"actor_role": "TEAM_LEAD",
	})

	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []string{"task.update"}, repo.calls)
}

func TestService_CanWithAttributes_SkipsConditionsWhenBaseRBACDenies(t *testing.T) {
	repo := &trackingRepo{
		results: map[string]bool{
			"task.update": false,
		},
	}
	registry := rbac.NewConditionRegistry()

	conditionCalled := false
	registry.Register("task.update", func(attrs map[string]interface{}) bool {
		conditionCalled = true
		return true
	})

	svc := rbac.NewService(repo, registry)
	projectID := uuid.New()
	ok, err := svc.CanWithAttributes(context.Background(), uuid.New(), "task.update", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	}, map[string]interface{}{
		"status": "IN_PROGRESS",
	})

	require.NoError(t, err)
	require.False(t, ok)
	require.False(t, conditionCalled)
}

func TestService_Conditions_ReturnsRegistry(t *testing.T) {
	registry := rbac.NewConditionRegistry()
	svc := rbac.NewService(&trackingRepo{}, registry)

	require.Same(t, registry, svc.Conditions())
}

func TestService_SetNow_OverridesClock(t *testing.T) {
	repo := &trackingRepo{
		results: map[string]bool{
			"task.view": true,
		},
	}
	svc := rbac.NewService(repo, nil)

	fixedNow := time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC)
	svc.SetNow(func() time.Time { return fixedNow })

	projectID := uuid.New()
	ok, err := svc.Can(context.Background(), uuid.New(), "task.view", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	})

	require.NoError(t, err)
	require.True(t, ok)
	require.True(t, repo.lastNow.Equal(fixedNow))
}
