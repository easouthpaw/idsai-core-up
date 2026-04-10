package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type benchmarkAuthz struct {
	allow bool
}

func (a benchmarkAuthz) Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope) (bool, error) {
	return a.allow, nil
}

func (a benchmarkAuthz) CanAll(ctx context.Context, userID uuid.UUID, permissions []string, scope rbac.Scope) (bool, error) {
	return a.allow, nil
}

func (a benchmarkAuthz) CanWithAttributes(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope, attrs map[string]interface{}) (bool, error) {
	return a.allow, nil
}

func (a benchmarkAuthz) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope rbac.Scope) ([]string, error) {
	return nil, nil
}

func benchmarkRoute(b *testing.B, router *gin.Engine, path string, wantStatus int) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(w, req)
		if w.Code != wantStatus {
			b.Fatalf("unexpected status: got=%d want=%d", w.Code, wantStatus)
		}
	}
}

func BenchmarkRequirePermission_Allowed(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(withUser(uuid.New()))
	router.GET("/x", middleware.RequirePermission(benchmarkAuthz{allow: true}, "task.view", middleware.SystemScope()), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	benchmarkRoute(b, router, "/x", http.StatusOK)
}

func BenchmarkRequireAllPermissions_Allowed(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(withUser(uuid.New()))
	router.GET("/x", middleware.RequireAllPermissions(benchmarkAuthz{allow: true}, []string{
		"project.view",
		"task.view",
		"task.update",
	}, middleware.SystemScope()), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	benchmarkRoute(b, router, "/x", http.StatusOK)
}

func BenchmarkRequirePermissionWithAttrs_Allowed(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(withUser(uuid.New()))
	router.GET("/x", middleware.RequirePermissionWithAttrs(
		benchmarkAuthz{allow: true},
		"task.update",
		middleware.SystemScope(),
		func(c *gin.Context) map[string]interface{} {
			return map[string]interface{}{
				"status":     "IN_PROGRESS",
				"actor_role": "TEAM_LEAD",
			}
		},
	), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	benchmarkRoute(b, router, "/x", http.StatusOK)
}
