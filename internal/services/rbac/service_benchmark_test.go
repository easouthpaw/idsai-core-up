package rbac_test

import (
	"context"
	"testing"
	"time"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

type benchmarkRepo struct{}

func (benchmarkRepo) HasPermission(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope, now time.Time) (bool, error) {
	return true, nil
}

func (benchmarkRepo) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope rbac.Scope, now time.Time) ([]string, error) {
	return []string{"project.view", "task.view", "task.update"}, nil
}

func BenchmarkServiceCan_ProjectScope(b *testing.B) {
	svc := rbac.NewService(benchmarkRepo{}, nil)
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()
	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ok, err := svc.Can(ctx, userID, "task.view", scope)
		if err != nil || !ok {
			b.Fatalf("unexpected result: ok=%v err=%v", ok, err)
		}
	}
}

func BenchmarkServiceCanAll_FivePermissions(b *testing.B) {
	svc := rbac.NewService(benchmarkRepo{}, nil)
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()
	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}
	permissions := []string{
		"project.view",
		"task.view",
		"task.update",
		"member.apply",
		"grading.view",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ok, err := svc.CanAll(ctx, userID, permissions, scope)
		if err != nil || !ok {
			b.Fatalf("unexpected result: ok=%v err=%v", ok, err)
		}
	}
}

func BenchmarkServiceCanWithAttributes_ThreeConditions(b *testing.B) {
	registry := rbac.NewConditionRegistry()
	registry.Register("task.update", func(attrs map[string]interface{}) bool {
		status, _ := attrs["status"].(string)
		return status == "IN_PROGRESS"
	})
	registry.Register("task.update", func(attrs map[string]interface{}) bool {
		role, _ := attrs["actor_role"].(string)
		return role == "TEAM_LEAD"
	})
	registry.Register("task.update", func(attrs map[string]interface{}) bool {
		authorID, _ := attrs["author_id"].(uuid.UUID)
		return authorID != uuid.Nil
	})

	svc := rbac.NewService(benchmarkRepo{}, registry)
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()
	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}
	attrs := map[string]interface{}{
		"status":     "IN_PROGRESS",
		"actor_role": "TEAM_LEAD",
		"author_id":  userID,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ok, err := svc.CanWithAttributes(ctx, userID, "task.update", scope, attrs)
		if err != nil || !ok {
			b.Fatalf("unexpected result: ok=%v err=%v", ok, err)
		}
	}
}
