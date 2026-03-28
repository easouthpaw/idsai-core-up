package middleware

import (
	"log"
	"net/http"
	"sync/atomic"

	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScopeResolver func(c *gin.Context) (rbac.Scope, bool)

var (
	rbacUnauthorizedCount atomic.Uint64
	rbacBadScopeCount     atomic.Uint64
	rbacDeniedCount       atomic.Uint64
)

func RequirePermissionIf(enabled bool, authz rbac.Authorizer, permission string, resolveScope ScopeResolver) gin.HandlerFunc {
	if !enabled {
		return func(c *gin.Context) { c.Next() }
	}
	return RequirePermission(authz, permission, resolveScope)
}

func RequirePermission(authz rbac.Authorizer, permission string, resolveScope ScopeResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := UserIDFromCtx(c)
		if !ok {
			rbacUnauthorizedCount.Add(1)
			logRBACDecision(c, http.StatusUnauthorized, uuid.Nil, permission, rbac.Scope{}, "missing_user")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		scope, ok := resolveScope(c)
		if !ok {
			rbacBadScopeCount.Add(1)
			logRBACDecision(c, http.StatusBadRequest, userID, permission, rbac.Scope{}, "invalid_scope")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
			return
		}

		allowed, err := authz.Can(c.Request.Context(), userID, permission, scope)
		if err != nil {
			rbacDeniedCount.Add(1)
			logRBACDecision(c, http.StatusForbidden, userID, permission, scope, "authorizer_error:"+err.Error())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if !allowed {
			rbacDeniedCount.Add(1)
			logRBACDecision(c, http.StatusForbidden, userID, permission, scope, "denied")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}

// Helpers
func ProjectScopeFromParam(param string) ScopeResolver {
	return func(c *gin.Context) (rbac.Scope, bool) {
		raw := c.Param(param)
		id, err := uuid.Parse(raw)
		if err != nil {
			return rbac.Scope{}, false
		}
		return rbac.Scope{Type: rbac.ScopeProject, ID: &id}, true
	}
}

func FacultyScopeFromHeader(header string) ScopeResolver {
	return func(c *gin.Context) (rbac.Scope, bool) {
		raw := c.GetHeader(header)
		id, err := uuid.Parse(raw)
		if err != nil {
			return rbac.Scope{}, false
		}
		return rbac.Scope{Type: rbac.ScopeFaculty, ID: &id}, true
	}
}

func FacultyScopeFromCtx() ScopeResolver {
	return func(c *gin.Context) (rbac.Scope, bool) {
		id, ok := FacultyIDFromCtx(c)
		if !ok {
			return rbac.Scope{}, false
		}
		return rbac.Scope{Type: rbac.ScopeFaculty, ID: &id}, true
	}
}

func TenantScopeFromCtx() ScopeResolver {
	return func(c *gin.Context) (rbac.Scope, bool) {
		id, ok := TenantIDFromCtx(c)
		if !ok {
			return rbac.Scope{}, false
		}
		return rbac.Scope{Type: rbac.ScopeTenant, ID: &id}, true
	}
}

func DepartmentScopeFromCtx() ScopeResolver {
	return func(c *gin.Context) (rbac.Scope, bool) {
		id, ok := DepartmentIDFromCtx(c)
		if !ok {
			return rbac.Scope{}, false
		}
		return rbac.Scope{Type: rbac.ScopeDepartment, ID: &id}, true
	}
}

func SystemScope() ScopeResolver {
	return func(c *gin.Context) (rbac.Scope, bool) {
		return rbac.Scope{Type: rbac.ScopeSystem, ID: nil}, true
	}
}

func logRBACDecision(c *gin.Context, status int, userID uuid.UUID, permission string, scope rbac.Scope, reason string) {
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	tenantID, _ := TenantIDFromCtx(c)
	scopeID := ""
	if scope.ID != nil {
		scopeID = scope.ID.String()
	}

	log.Printf(
		"rbac_deny status=%d method=%s path=%s user_id=%s tenant_id=%s permission=%s scope_type=%s scope_id=%s reason=%s",
		status,
		c.Request.Method,
		path,
		userID.String(),
		tenantID.String(),
		permission,
		scope.Type,
		scopeID,
		reason,
	)
}
