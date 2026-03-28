package middleware

import (
	"net/http"
	"strings"

	"idsai-core-up/internal/requestctx"
	authsvc "idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	TenantID     string `json:"tenant_id"`
	FacultyID    string `json:"faculty_id"`
	DepartmentID string `json:"department_id"`
	IsAdmin      bool   `json:"is_admin"`
	IsProfessor  bool   `json:"is_professor"`
	jwt.RegisteredClaims
}

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	secret := []byte(jwtSecret)

	return func(c *gin.Context) {
		raw := ""
		if cookie, err := c.Cookie(authsvc.AccessCookieName); err == nil {
			raw = strings.TrimSpace(cookie)
		}
		if raw == "" {
			h := c.GetHeader("Authorization")
			if h != "" && strings.HasPrefix(h, "Bearer ") {
				raw = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			}
		}
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
			return
		}

		tok, err := jwt.ParseWithClaims(
			raw,
			&AccessClaims{},
			func(token *jwt.Token) (any, error) {
				return secret, nil
			},
			jwt.WithIssuer(authsvc.TokenIssuer),
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := tok.Claims.(*AccessClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		uid, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid sub"})
			return
		}
		tid, err := uuid.Parse(claims.TenantID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid tenant_id"})
			return
		}
		fid, err := uuid.Parse(claims.FacultyID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid faculty_id"})
			return
		}
		did, err := uuid.Parse(claims.DepartmentID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid department_id"})
			return
		}

		c.Set("userID", uid)
		c.Set("tenantID", tid)
		c.Set("facultyID", fid)
		c.Set("departmentID", did)
		c.Set("isAdmin", claims.IsAdmin)
		c.Set("isProfessor", claims.IsProfessor)
		c.Request = c.Request.WithContext(requestctx.WithIdentity(c.Request.Context(), uid, tid, fid, did))
		c.Next()
	}
}

func UserIDFromCtx(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func FacultyIDFromCtx(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("facultyID")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func TenantIDFromCtx(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("tenantID")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func DepartmentIDFromCtx(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("departmentID")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func IsAdminFromCtx(c *gin.Context) (bool, bool) {
	v, ok := c.Get("isAdmin")
	if !ok {
		return false, false
	}
	isAdmin, ok := v.(bool)
	return isAdmin, ok
}

func IsProfessorFromCtx(c *gin.Context) (bool, bool) {
	v, ok := c.Get("isProfessor")
	if !ok {
		return false, false
	}
	isProfessor, ok := v.(bool)
	return isProfessor, ok
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, ok := IsAdminFromCtx(c)
		if !ok || !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
