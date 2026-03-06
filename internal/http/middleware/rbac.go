package middleware

import (
	"net/http"

	"idsai-core-up/internal/services/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScopeResolver func(c *gin.Context) (rbac.Scope, bool)

func RequirePermission(authz rbac.Authorizer, permission string, resolveScope ScopeResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := UserIDFromCtx(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		scope, ok := resolveScope(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
			return
		}

		allowed, err := authz.Can(c.Request.Context(), userID, permission, scope)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if !allowed {
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

func SystemScope() ScopeResolver {
	return func(c *gin.Context) (rbac.Scope, bool) {
		return rbac.Scope{Type: rbac.ScopeSystem, ID: nil}, true
	}
}
