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
	"github.com/stretchr/testify/require"
)

type fakeAuthz struct {
	allow bool
	err   error
}

func (f fakeAuthz) Can(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope) (bool, error) {
	return f.allow, f.err
}

func TestRequirePermission_MissingUserHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/x", middleware.RequirePermission(fakeAuthz{allow: true}, "task.view", middleware.SystemScope()), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequirePermission_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	user := uuid.New().String()
	r.GET("/x", middleware.RequirePermission(fakeAuthz{allow: false}, "task.view", middleware.SystemScope()), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-User-ID", user)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermission_Allows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	user := uuid.New().String()
	r.GET("/x", middleware.RequirePermission(fakeAuthz{allow: true}, "task.view", middleware.SystemScope()), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-User-ID", user)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
