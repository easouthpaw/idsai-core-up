package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/http/middleware"
	"idsai-core-up/internal/requestctx"
	authsvc "idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func signedAccessToken(t *testing.T, secret string, claims middleware.AccessClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func baseAccessClaims(now time.Time) middleware.AccessClaims {
	return middleware.AccessClaims{
		TenantID:          uuid.NewString(),
		FacultyID:         uuid.NewString(),
		DepartmentID:      uuid.NewString(),
		IsAdmin:           true,
		IsProfessor:       true,
		PasswordChangedAt: now.UnixMilli(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			Issuer:    authsvc.TokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
}

func TestAuthRequired_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.AuthRequired("01234567890123456789012345678901", nil))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_InvalidHeaderToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.AuthRequired("01234567890123456789012345678901", nil))
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_RejectsInvalidClaimUUIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "01234567890123456789012345678901"
	now := time.Now().UTC()
	cases := []struct {
		name   string
		mutate func(claims *middleware.AccessClaims)
	}{
		{
			name: "invalid sub",
			mutate: func(claims *middleware.AccessClaims) {
				claims.Subject = "bad-uuid"
			},
		},
		{
			name: "invalid tenant",
			mutate: func(claims *middleware.AccessClaims) {
				claims.TenantID = "bad-uuid"
			},
		},
		{
			name: "invalid faculty",
			mutate: func(claims *middleware.AccessClaims) {
				claims.FacultyID = "bad-uuid"
			},
		},
		{
			name: "invalid department",
			mutate: func(claims *middleware.AccessClaims) {
				claims.DepartmentID = "bad-uuid"
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims := baseAccessClaims(now)
			tc.mutate(&claims)

			r := gin.New()
			r.Use(middleware.AuthRequired(secret, nil))
			r.GET("/protected", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+signedAccessToken(t, secret, claims))
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthRequired_RejectsInvalidStateReaderResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "01234567890123456789012345678901"
	now := time.Now().UTC()
	claims := baseAccessClaims(now)
	claims.PasswordChangedAt = now.UnixMilli()
	token := signedAccessToken(t, secret, claims)

	readerCases := []struct {
		name   string
		reader stubAuthStateReader
	}{
		{
			name:   "reader error",
			reader: stubAuthStateReader{err: assertiveErr("state reader failed")},
		},
		{
			name: "inactive user",
			reader: stubAuthStateReader{state: authsvc.UserAuthState{
				Status:            "BLOCKED",
				PasswordChangedAt: now,
				EmailVerifiedAt:   &now,
			}},
		},
		{
			name: "email not verified",
			reader: stubAuthStateReader{state: authsvc.UserAuthState{
				Status:            authsvc.StatusActive,
				PasswordChangedAt: now,
				EmailVerifiedAt:   nil,
			}},
		},
	}

	for _, tc := range readerCases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(middleware.AuthRequired(secret, tc.reader))
			r.GET("/protected", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthRequired_SetsContextAndRequestIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "01234567890123456789012345678901"
	now := time.Now().UTC()
	userID := uuid.New()
	tenantID := uuid.New()
	facultyID := uuid.New()
	departmentID := uuid.New()
	claims := middleware.AccessClaims{
		TenantID:          tenantID.String(),
		FacultyID:         facultyID.String(),
		DepartmentID:      departmentID.String(),
		IsAdmin:           true,
		IsProfessor:       true,
		PasswordChangedAt: now.UnixMilli(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    authsvc.TokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}

	r := gin.New()
	r.Use(middleware.AuthRequired(secret, stubAuthStateReader{
		state: authsvc.UserAuthState{
			Status:            authsvc.StatusActive,
			PasswordChangedAt: now,
			EmailVerifiedAt:   &now,
		},
	}))
	r.GET("/protected", func(c *gin.Context) {
		gotUserID, ok := middleware.UserIDFromCtx(c)
		require.True(t, ok)
		require.Equal(t, userID, gotUserID)

		gotTenantID, ok := middleware.TenantIDFromCtx(c)
		require.True(t, ok)
		require.Equal(t, tenantID, gotTenantID)

		gotFacultyID, ok := middleware.FacultyIDFromCtx(c)
		require.True(t, ok)
		require.Equal(t, facultyID, gotFacultyID)

		gotDepartmentID, ok := middleware.DepartmentIDFromCtx(c)
		require.True(t, ok)
		require.Equal(t, departmentID, gotDepartmentID)

		isAdmin, ok := middleware.IsAdminFromCtx(c)
		require.True(t, ok)
		require.True(t, isAdmin)

		isProfessor, ok := middleware.IsProfessorFromCtx(c)
		require.True(t, ok)
		require.True(t, isProfessor)

		ctxUserID, ok := requestctx.UserID(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, userID, ctxUserID)

		ctxTenantID, ok := requestctx.TenantID(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, tenantID, ctxTenantID)

		ctxFacultyID, ok := requestctx.FacultyID(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, facultyID, ctxFacultyID)

		ctxDepartmentID, ok := requestctx.DepartmentID(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, departmentID, ctxDepartmentID)

		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedAccessToken(t, secret, claims))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestContextReaders_ReturnFalseWhenValueMissingOrWrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	_, ok := middleware.UserIDFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.TenantIDFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.FacultyIDFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.DepartmentIDFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.IsAdminFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.IsProfessorFromCtx(c)
	require.False(t, ok)

	c.Set("userID", "not-uuid")
	c.Set("tenantID", "not-uuid")
	c.Set("facultyID", "not-uuid")
	c.Set("departmentID", "not-uuid")
	c.Set("isAdmin", "yes")
	c.Set("isProfessor", "yes")

	_, ok = middleware.UserIDFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.TenantIDFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.FacultyIDFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.DepartmentIDFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.IsAdminFromCtx(c)
	require.False(t, ok)
	_, ok = middleware.IsProfessorFromCtx(c)
	require.False(t, ok)
}

func TestAdminRequired_RejectsAndAllows(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("forbidden when missing admin flag", func(t *testing.T) {
		r := gin.New()
		r.GET("/admin", middleware.AdminRequired(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("allows admin", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("isAdmin", true)
			c.Next()
		})
		r.GET("/admin", middleware.AdminRequired(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})
}

type assertiveErr string

func (e assertiveErr) Error() string { return string(e) }
