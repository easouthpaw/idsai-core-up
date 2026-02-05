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
		// 1) auth (temporary): read user from header
		userHeader := c.GetHeader("X-User-ID")
		if userHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-User-ID"})
			return
		}

		userID, err := uuid.Parse(userHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid X-User-ID"})
			return
		}

		// 2) resolve scope
		scope, ok := resolveScope(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
			return
		}

		// 3) rbac
		allowed, err := authz.Can(c.Request.Context(), userID, permission, scope)
		if err != nil {
			// invalid scope or repo error
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

func SystemScope() ScopeResolver {
	return func(c *gin.Context) (rbac.Scope, bool) {
		return rbac.Scope{Type: rbac.ScopeSystem, ID: nil}, true
	}
}
