package rbacmodule

import (
	"testing"

	"idsai-core-up/internal/infra/cache"
	"idsai-core-up/internal/services/rbac"
)

func TestNew_WiresRBACModuleWithoutRedis(t *testing.T) {
	out := New(nil, nil)

	if out.Repo == nil || out.Service == nil || out.Authorizer == nil {
		t.Fatalf("expected rbac module output to be fully wired")
	}
	if _, ok := out.Authorizer.(*rbac.Service); !ok {
		t.Fatalf("expected plain service authorizer when redis is disabled")
	}
	if !out.Service.Conditions().Evaluate("task.edit", map[string]interface{}{
		"user_id":        "u1",
		"task_author_id": "u1",
		"project_status": "ACTIVE",
	}) {
		t.Fatalf("expected built-in task.edit ABAC condition to allow matching author in active project")
	}
	if out.Service.Conditions().Evaluate("task.edit", map[string]interface{}{
		"user_id":        "u1",
		"task_author_id": "u2",
		"project_status": "ACTIVE",
	}) {
		t.Fatalf("expected built-in task.edit ABAC condition to deny foreign author")
	}
}

func TestNew_WiresRBACModuleWithRedis(t *testing.T) {
	redisClient := cache.NewRedisClient("127.0.0.1:0", "", 0)
	defer redisClient.Close()

	out := New(nil, redisClient)

	if _, ok := out.Authorizer.(*rbac.CachedAuthorizer); !ok {
		t.Fatalf("expected cached authorizer when redis client is configured")
	}
}
