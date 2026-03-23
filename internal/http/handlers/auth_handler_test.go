package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/security/passwords"
	"idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type authHandlerRepo struct {
	tenantID uuid.UUID
	user     auth.User
}

func (f *authHandlerRepo) FindTenantByCode(ctx context.Context, tenantCode string) (uuid.UUID, error) {
	return f.tenantID, nil
}

func (f *authHandlerRepo) CreateUser(ctx context.Context, in auth.CreateUserParams) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (f *authHandlerRepo) CreateProfile(ctx context.Context, tenantID, userID uuid.UUID, fullName string, facultyID, departmentID uuid.UUID) error {
	return nil
}

func (f *authHandlerRepo) GrantStudentFacultyRole(ctx context.Context, tenantID, userID, facultyID uuid.UUID) error {
	return nil
}

func (f *authHandlerRepo) FindUserByEmail(ctx context.Context, tenantID uuid.UUID, email string) (auth.User, error) {
	return f.user, nil
}

func (f *authHandlerRepo) FindUserByID(ctx context.Context, tenantID, userID uuid.UUID) (auth.User, error) {
	return f.user, nil
}

func (f *authHandlerRepo) UpdateUserPasswordHash(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string, changedAt time.Time) error {
	return nil
}

func (f *authHandlerRepo) MarkUserEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verifiedAt time.Time) error {
	return nil
}

func (f *authHandlerRepo) InsertRefreshToken(ctx context.Context, tenantID, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return nil
}

func (f *authHandlerRepo) FindRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, uuid.UUID, time.Time, *time.Time, error) {
	return uuid.Nil, uuid.Nil, time.Time{}, nil, auth.ErrNotFound
}

func (f *authHandlerRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return nil
}

func (f *authHandlerRepo) RevokeUserRefreshTokens(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}

func (f *authHandlerRepo) FindDepartment(ctx context.Context, tenantID uuid.UUID, departmentCode string) (uuid.UUID, uuid.UUID, error) {
	return uuid.New(), uuid.New(), nil
}

func (f *authHandlerRepo) InsertAuthToken(ctx context.Context, tenantID, userID uuid.UUID, purpose, tokenHash string, expiresAt time.Time) error {
	return nil
}

func (f *authHandlerRepo) FindAuthToken(ctx context.Context, purpose, tokenHash string) (auth.AuthTokenRecord, error) {
	return auth.AuthTokenRecord{}, auth.ErrNotFound
}

func (f *authHandlerRepo) ConsumeAuthToken(ctx context.Context, tokenID uuid.UUID, consumedAt time.Time) error {
	return nil
}

func (f *authHandlerRepo) InvalidateAuthTokens(ctx context.Context, tenantID, userID uuid.UUID, purpose string) error {
	return nil
}

func newAuthHandlerForTest(t *testing.T) *AuthHandler {
	t.Helper()

	hash, err := passwords.Hash("valid-password")
	require.NoError(t, err)

	now := time.Now().UTC()
	repo := &authHandlerRepo{
		tenantID: uuid.New(),
		user: auth.User{
			ID:              uuid.New(),
			TenantID:        uuid.New(),
			Email:           "student@example.edu",
			FullName:        "Student User",
			PasswordHash:    hash,
			Status:          auth.StatusActive,
			FacultyID:       uuid.New(),
			DepartmentID:    uuid.New(),
			EmailVerifiedAt: &now,
		},
	}

	svc := auth.NewService(repo, auth.Config{
		JWTSecret: "01234567890123456789012345678901",
	})
	return NewAuthHandler(svc)
}

func TestAuthHandlerLoginSetsHttpOnlyCookiesByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newAuthHandlerForTest(t)
	r := gin.New()
	r.POST("/v2/auth/login", h.Login)

	body, err := json.Marshal(map[string]string{
		"email":    "student@example.edu",
		"password": "valid-password",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v2/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotContains(t, w.Body.String(), "access_token")
	require.NotContains(t, w.Body.String(), "refresh_token")

	cookies := w.Result().Cookies()
	var cookieNames []string
	for _, cookie := range cookies {
		cookieNames = append(cookieNames, cookie.Name)
		require.True(t, cookie.HttpOnly)
	}
	require.Contains(t, cookieNames, auth.AccessCookieName)
	require.Contains(t, cookieNames, auth.RefreshCookieName)
}

func TestAuthHandlerLoginTokenModeReturnsTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newAuthHandlerForTest(t)
	r := gin.New()
	r.POST("/v2/auth/login", h.Login)

	body, err := json.Marshal(map[string]string{
		"email":    "student@example.edu",
		"password": "valid-password",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v2/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Mode", "token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "access_token")
	require.Contains(t, w.Body.String(), "refresh_token")
}

func TestAuthHandlerRefreshWithoutSessionReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newAuthHandlerForTest(t)
	r := gin.New()
	r.POST("/v2/auth/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/v2/auth/refresh", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}
