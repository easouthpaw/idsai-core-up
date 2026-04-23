package handlers

import (
	"bytes"
	"mime/multipart"
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

func TestAuthHandlerSessionProfileAndCapabilitiesRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, user := newAuthHandlerWithUser(t, "")
	handler.SetAuthorizer(&projectFlowTestDeps{permissions: []string{"project.view", "task.create"}})

	router := gin.New()
	router.Use(withFlowContext(user.ID, user.TenantID, user.FacultyID))
	router.POST("/auth/logout", handler.Logout)
	router.GET("/auth/me", handler.Me)
	router.GET("/auth/capabilities", handler.Capabilities)
	router.GET("/settings", handler.SettingsGet)
	router.GET("/profiles/:user_id", handler.GetProfile)
	router.PUT("/settings/profile", handler.SettingsUpdateProfile)
	router.POST("/settings/email/start", handler.SettingsStartEmailChange)
	router.POST("/settings/password", handler.SettingsChangePassword)
	router.POST("/settings/avatar", handler.SettingsUploadAvatar)
	router.DELETE("/settings/avatar", handler.SettingsDeleteAvatar)

	requireStatus(t, router, http.MethodPost, "/auth/logout", `{}`, http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/auth/me", "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/auth/capabilities?scope_type=FACULTY&scope_id="+user.FacultyID.String(), "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/settings", "", http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/profiles/"+user.ID.String(), "", http.StatusOK)
	requireStatus(t, router, http.MethodPut, "/settings/profile", `{"full_name":"Updated User","headline":"Backend","about":"About","preferred_role":"Backend","semester":"6","availability":"Evenings","goals":"Ship","github_url":"https://github.com/example","telegram":"@example","portfolio_url":"https://example.local","stacks":["go"],"interests":["security"]}`, http.StatusOK)
	requireStatus(t, router, http.MethodPost, "/settings/email/start", `{"email":"new.student@example.edu"}`, http.StatusAccepted)
	requireStatus(t, router, http.MethodPost, "/settings/password", `{"current_password":"valid-password1","new_password":"new-valid1","confirm_password":"new-valid1"}`, http.StatusOK)
	requireStatus(t, router, http.MethodPost, "/settings/password", `{"current_password":"valid-password1","new_password":"new-valid1","confirm_password":"different1"}`, http.StatusBadRequest)
	requireStatus(t, router, http.MethodDelete, "/settings/avatar", "", http.StatusOK)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("avatar", "avatar.txt")
	require.NoError(t, err)
	_, err = file.Write([]byte("not image"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	req := httptest.NewRequest(http.MethodPost, "/settings/avatar", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestAuthHandlerResendAndRedirectErrorRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, user := newAuthHandlerWithUser(t, "pending@example.edu")

	router := gin.New()
	router.Use(withFlowContext(user.ID, user.TenantID, user.FacultyID))
	router.POST("/auth/verification/resend", handler.ResendVerification)
	router.GET("/auth/password-reset", handler.PasswordResetLanding)
	router.POST("/auth/password-reset/confirm", handler.PasswordResetConfirm)
	router.GET("/auth/verify-email", handler.VerifyEmail)
	router.POST("/settings/email/resend", handler.SettingsResendEmailChange)
	router.GET("/settings/email/verify", handler.SettingsVerifyEmailChange)
	router.POST("/settings/email/confirm", handler.SettingsConfirmEmailChange)

	requireStatus(t, router, http.MethodPost, "/auth/verification/resend", `{"email":"student@example.edu"}`, http.StatusAccepted)
	requireStatus(t, router, http.MethodPost, "/settings/email/resend", `{}`, http.StatusAccepted)
	requireStatus(t, router, http.MethodPost, "/settings/email/confirm", `{}`, http.StatusBadRequest)
	requireStatus(t, router, http.MethodPost, "/auth/password-reset/confirm", `{"email":"student@example.edu","code":"bad","password":"new-valid1"}`, http.StatusBadRequest)

	req := httptest.NewRequest(http.MethodGet, "/auth/password-reset?token=bad", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Contains(t, rec.Header().Get("Location"), "reset=expired")

	req = httptest.NewRequest(http.MethodGet, "/auth/verify-email?token=bad", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Contains(t, rec.Header().Get("Location"), "verified=0")

	req = httptest.NewRequest(http.MethodGet, "/settings/email/verify?token=bad", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	require.Contains(t, rec.Header().Get("Location"), "email_change=expired")
}

func TestAuthHandlerCapabilitiesUnavailableAndInvalidScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler, user := newAuthHandlerWithUser(t, "")
	router := gin.New()
	router.Use(withFlowContext(user.ID, user.TenantID, user.FacultyID))
	router.GET("/auth/capabilities", handler.Capabilities)
	requireStatus(t, router, http.MethodGet, "/auth/capabilities", "", http.StatusServiceUnavailable)

	handler.SetAuthorizer(&projectFlowTestDeps{})
	requireStatus(t, router, http.MethodGet, "/auth/capabilities?scope_type=PROJECT&scope_id=bad", "", http.StatusBadRequest)
}

func newAuthHandlerWithUser(t *testing.T, pendingEmail string) (*AuthHandler, auth.User) {
	t.Helper()

	hash, err := passwords.Hash("valid-password1")
	require.NoError(t, err)
	now := time.Now().UTC()
	user := auth.User{
		ID:                uuid.New(),
		TenantID:          uuid.New(),
		Email:             "student@example.edu",
		PendingEmail:      pendingEmail,
		PasswordHash:      hash,
		PasswordChangedAt: now,
		Status:            auth.StatusActive,
		FacultyID:         uuid.New(),
		FacultyCode:       "IDSAI_ENU",
		DepartmentID:      uuid.New(),
		DepartmentCode:    "CS",
		FullName:          "Student User",
		ProfileUpdatedAt:  now,
		EmailVerifiedAt:   &now,
		AvatarKey:         "old-avatar.jpg",
	}
	if pendingEmail != "" {
		user.PendingEmailAt = &now
	}
	repo := &authHandlerRepo{
		tenantID: user.TenantID,
		user:     user,
	}
	svc := auth.NewService(repo, auth.Config{JWTSecret: "01234567890123456789012345678901"})
	return NewAuthHandler(svc), user
}
