package auth

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"idsai-core-up/internal/security/passwords"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_SessionTokenAndVerificationFlows(t *testing.T) {
	now := time.Now().UTC()
	repo := authServiceRepo(t, now)
	repo.refreshTenantID = repo.user.TenantID
	repo.refreshUserID = repo.user.ID
	repo.refreshExpiresAt = now.Add(time.Hour)
	repo.resetToken = AuthTokenRecord{
		ID:        uuid.New(),
		TenantID:  repo.user.TenantID,
		UserID:    repo.user.ID,
		Purpose:   TokenPurposePasswordReset,
		TokenHash: hashToken("token"),
		ExpiresAt: now.Add(time.Hour),
	}
	svc := NewService(repo, Config{JWTSecret: "01234567890123456789012345678901"})

	session, err := svc.Refresh(context.Background(), "refresh-token")
	require.NoError(t, err)
	require.Equal(t, repo.user.ID, session.User.ID)
	require.NotEmpty(t, session.Tokens.AccessToken)
	require.NotEmpty(t, session.Tokens.RefreshToken)

	require.NoError(t, svc.Logout(context.Background(), "refresh-token"))
	require.Equal(t, 1, repo.revokeRefreshCount)

	ttl, err := svc.ValidatePasswordResetToken(context.Background(), "token")
	require.NoError(t, err)
	require.Positive(t, ttl)

	require.NoError(t, svc.ResetPassword(context.Background(), "token", "new-valid1"))
	require.Equal(t, repo.user.ID, repo.updatedPasswordUserID)
	require.NotEmpty(t, repo.updatedPasswordHash)

	repo.user.PendingEmail = "changed@example.edu"
	repo.resetToken.Purpose = TokenPurposeEmailChange
	changed, err := svc.ConfirmEmailChangeCode(context.Background(), repo.user.TenantID, repo.user.ID, "123456")
	require.NoError(t, err)
	require.Equal(t, "changed@example.edu", changed.Email)

	repo.user.Status = StatusPending
	repo.user.EmailVerifiedAt = nil
	repo.resetToken.Purpose = TokenPurposeEmailVerification
	require.NoError(t, svc.VerifyEmail(context.Background(), "token"))
	require.Equal(t, StatusActive, repo.user.Status)
	require.NotNil(t, repo.user.EmailVerifiedAt)
}

func TestService_ProfileGroupAndAvatarFlows(t *testing.T) {
	now := time.Now().UTC()
	groupID := uuid.New()
	repo := authServiceRepo(t, now)
	repo.findGroupErr = ErrGroupNotFound
	repo.user.GroupID = &groupID
	repo.user.GroupCode = "CS-101"
	svc := NewService(repo, Config{
		JWTSecret:     "01234567890123456789012345678901",
		PublicBaseURL: "https://idsai.example",
	})

	updated, err := svc.UpdateProfile(context.Background(), repo.user.TenantID, repo.user.ID, ProfileUpdate{
		FullName:  "Updated User",
		Headline:  "Backend",
		Stacks:    []string{"go", "go", "postgres"},
		Interests: []string{"security"},
	})
	require.NoError(t, err)
	require.Equal(t, "Updated User", updated.FullName)
	require.Equal(t, []string{"go", "postgres"}, updated.Stacks)

	require.NoError(t, svc.StartEmailChange(context.Background(), "actor", repo.user.TenantID, repo.user.ID, "next@example.edu"))
	require.Equal(t, "next@example.edu", repo.user.PendingEmail)
	require.NoError(t, svc.ResendEmailChange(context.Background(), "actor", repo.user.TenantID, repo.user.ID))

	require.NoError(t, svc.ChangePassword(context.Background(), repo.user.TenantID, repo.user.ID, "valid-password1", "new-valid1"))
	require.NotEmpty(t, repo.updatedPasswordHash)

	request, err := svc.SubmitGroupChangeRequest(context.Background(), repo.user.TenantID, repo.user.ID, "CS", "202")
	require.NoError(t, err)
	require.Equal(t, GroupChangeRequest{}, request)
	require.Equal(t, "CS-202", repo.createdGroupCode)
	require.Equal(t, 202, repo.createdGroupNumber)

	repo.user.AvatarKey = "old-avatar.jpg"
	deleted, err := svc.DeleteAvatar(context.Background(), repo.user.TenantID, repo.user.ID)
	require.NoError(t, err)
	require.Empty(t, deleted.AvatarKey)

	storage := &fakeAuthStorage{}
	svc.SetStorage(storage)
	withAvatar, err := svc.UpdateAvatar(context.Background(), repo.user.TenantID, repo.user.ID, validAvatarPNG(t))
	require.NoError(t, err)
	require.NotEmpty(t, withAvatar.AvatarKey)
	require.Equal(t, 1, storage.puts)
}

func TestEducationHelpers(t *testing.T) {
	require.Equal(t, EducationTypeSchool, EducationTypeFromFacultyCode("CORE_SCHOOL"))
	require.Equal(t, EducationTypeUniversity, EducationTypeFromFacultyCode("IDSAI_ENU"))
	require.Equal(t, "10A", SchoolClassFromGroupCode("CLASS-10A"))
	require.Empty(t, SchoolClassFromGroupCode("CS-101"))
}

func authServiceRepo(t *testing.T, now time.Time) *fakeRepo {
	t.Helper()
	hash, err := passwords.Hash("valid-password1")
	require.NoError(t, err)
	return &fakeRepo{
		tenantID: uuid.New(),
		user: User{
			ID:                uuid.New(),
			TenantID:          uuid.New(),
			Email:             "student@example.edu",
			PasswordHash:      hash,
			PasswordChangedAt: now,
			Status:            StatusActive,
			FacultyID:         uuid.New(),
			FacultyCode:       "IDSAI_ENU",
			DepartmentID:      uuid.New(),
			DepartmentCode:    "CS",
			FullName:          "Student User",
			ProfileUpdatedAt:  now,
			EmailVerifiedAt:   &now,
		},
	}
}

type fakeAuthStorage struct {
	puts int
}

func (f *fakeAuthStorage) PutObject(ctx context.Context, key, contentType string, body []byte) error {
	f.puts++
	return nil
}

func (f *fakeAuthStorage) DeleteObject(ctx context.Context, key string) error {
	return nil
}

func (f *fakeAuthStorage) PublicURL(key string) string {
	return "https://cdn.example/" + key
}

func (f *fakeAuthStorage) Available() bool {
	return true
}

func validAvatarPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 420, 420))
	for y := 0; y < 420; y++ {
		for x := 0; x < 420; x++ {
			img.Set(x, y, color.RGBA{R: 20, G: 120, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}
