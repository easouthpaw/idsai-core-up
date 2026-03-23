package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/http/middleware"
	authsvc "idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthRequiredPrefersValidCookieOverInvalidHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "01234567890123456789012345678901"
	now := time.Now()
	claims := middleware.AccessClaims{
		TenantID:     uuid.NewString(),
		FacultyID:    uuid.NewString(),
		DepartmentID: uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			Issuer:    authsvc.TokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	r := gin.New()
	r.Use(middleware.AuthRequired(secret))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer definitely-invalid")
	req.AddCookie(&http.Cookie{Name: authsvc.AccessCookieName, Value: signed, Path: "/"})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
