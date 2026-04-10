package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type recordingAuthz struct {
	allow           bool
	err             error
	lastPermissions []string
	lastAttrs       map[string]interface{}
}

func (a *recordingAuthz) Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope) (bool, error) {
	return a.allow, a.err
}

func (a *recordingAuthz) CanAll(ctx context.Context, userID uuid.UUID, permissions []string, scope rbac.Scope) (bool, error) {
	a.lastPermissions = append([]string(nil), permissions...)
	return a.allow, a.err
}

func (a *recordingAuthz) CanWithAttributes(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope, attrs map[string]interface{}) (bool, error) {
	a.lastAttrs = attrs
	return a.allow, a.err
}

func (a *recordingAuthz) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope rbac.Scope) ([]string, error) {
	return nil, a.err
}

func TestRequireAllPermissions_Allows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authz := &recordingAuthz{allow: true}
	user := uuid.New()
	r.Use(withUser(user))
	r.GET("/x", middleware.RequireAllPermissions(authz, []string{"task.view", "task.update"}, middleware.SystemScope()), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, []string{"task.view", "task.update"}, authz.lastPermissions)
}

func TestRequireAllPermissions_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authz := &recordingAuthz{allow: false}
	user := uuid.New()
	r.Use(withUser(user))
	r.GET("/x", middleware.RequireAllPermissions(authz, []string{"task.view", "task.update"}, middleware.SystemScope()), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAllPermissions_AuthorizerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authz := &recordingAuthz{err: errors.New("rbac backend failed")}
	user := uuid.New()
	r.Use(withUser(user))
	r.GET("/x", middleware.RequireAllPermissions(authz, []string{"task.view"}, middleware.SystemScope()), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermissionWithAttrs_PassesResolvedAttributes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authz := &recordingAuthz{allow: true}
	user := uuid.New()
	r.Use(withUser(user))
	r.GET("/x", middleware.RequirePermissionWithAttrs(
		authz,
		"task.update",
		middleware.SystemScope(),
		func(c *gin.Context) map[string]interface{} {
			return map[string]interface{}{
				"status":     c.Query("status"),
				"actor_role": "TEAM_LEAD",
			}
		},
	), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x?status=IN_PROGRESS", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, map[string]interface{}{
		"status":     "IN_PROGRESS",
		"actor_role": "TEAM_LEAD",
	}, authz.lastAttrs)
}

func TestRequirePermissionWithAttrs_DeniesWhenABACRejects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authz := &recordingAuthz{allow: false}
	user := uuid.New()
	r.Use(withUser(user))
	r.GET("/x", middleware.RequirePermissionWithAttrs(
		authz,
		"task.update",
		middleware.SystemScope(),
		func(c *gin.Context) map[string]interface{} {
			return map[string]interface{}{
				"status": "DONE",
			}
		},
	), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermissionWithAttrs_AuthorizerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authz := &recordingAuthz{err: errors.New("abac backend failed")}
	user := uuid.New()
	r.Use(withUser(user))
	r.GET("/x", middleware.RequirePermissionWithAttrs(
		authz,
		"task.update",
		middleware.SystemScope(),
		func(c *gin.Context) map[string]interface{} {
			return map[string]interface{}{"status": "IN_PROGRESS"}
		},
	), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermissionIf_DisabledBypassesAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/x", middleware.RequirePermissionIf(false, &recordingAuthz{}, "task.view", middleware.SystemScope()), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestScopeResolvers_HeaderAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	facultyID := uuid.New()
	tenantID := uuid.New()
	departmentID := uuid.New()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	c.Request.Header.Set("X-Faculty-ID", facultyID.String())
	c.Set("facultyID", facultyID)
	c.Set("tenantID", tenantID)
	c.Set("departmentID", departmentID)

	scope, ok := middleware.FacultyScopeFromHeader("X-Faculty-ID")(c)
	require.True(t, ok)
	require.Equal(t, rbac.ScopeFaculty, scope.Type)
	require.Equal(t, facultyID, *scope.ID)

	scope, ok = middleware.FacultyScopeFromCtx()(c)
	require.True(t, ok)
	require.Equal(t, rbac.ScopeFaculty, scope.Type)
	require.Equal(t, facultyID, *scope.ID)

	scope, ok = middleware.TenantScopeFromCtx()(c)
	require.True(t, ok)
	require.Equal(t, rbac.ScopeTenant, scope.Type)
	require.Equal(t, tenantID, *scope.ID)

	scope, ok = middleware.DepartmentScopeFromCtx()(c)
	require.True(t, ok)
	require.Equal(t, rbac.ScopeDepartment, scope.Type)
	require.Equal(t, departmentID, *scope.ID)
}
