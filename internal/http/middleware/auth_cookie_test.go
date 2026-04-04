package middleware_test

import (
	"context"
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

type stubAuthStateReader struct {
	state authsvc.UserAuthState
	err   error
}

func (s stubAuthStateReader) GetUserAuthState(ctx context.Context, tenantID, userID uuid.UUID) (authsvc.UserAuthState, error) {
	return s.state, s.err
}

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
	r.Use(middleware.AuthRequired(secret, nil))
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

func TestAuthRequiredRejectsTokenIssuedBeforePasswordChange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "01234567890123456789012345678901"
	now := time.Now().UTC()
	userID := uuid.New()
	tenantID := uuid.New()
	claims := middleware.AccessClaims{
		TenantID:          tenantID.String(),
		FacultyID:         uuid.NewString(),
		DepartmentID:      uuid.NewString(),
		PasswordChangedAt: now.Add(-time.Minute).UnixMilli(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    authsvc.TokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	r := gin.New()
	r.Use(middleware.AuthRequired(secret, stubAuthStateReader{
		state: authsvc.UserAuthState{
			PasswordChangedAt: now,
			Status:            authsvc.StatusActive,
			EmailVerifiedAt:   &now,
		},
	}))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: authsvc.AccessCookieName, Value: signed, Path: "/"})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}
