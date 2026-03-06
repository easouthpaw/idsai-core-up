package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	FacultyID    string `json:"faculty_id"`
	DepartmentID string `json:"department_id"`
	IsAdmin      bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	secret := []byte(jwtSecret)

	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")

		tok, err := jwt.ParseWithClaims(raw, &AccessClaims{}, func(token *jwt.Token) (any, error) {
			return secret, nil
		})
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
		c.Set("facultyID", fid)
		c.Set("departmentID", did)
		c.Set("isAdmin", claims.IsAdmin)
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

func IsAdminFromCtx(c *gin.Context) (bool, bool) {
	v, ok := c.Get("isAdmin")
	if !ok {
		return false, false
	}
	isAdmin, ok := v.(bool)
	return isAdmin, ok
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
