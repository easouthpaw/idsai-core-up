package auth

import (
	"context"
	"testing"
	"time"

	"idsai-core-up/internal/security/passwords"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepo struct {
	tenantID              uuid.UUID
	user                  User
	findUserErr           error
	createdUser           CreateUserParams
	createUserCount       int
	refreshInsertCount    int
	updatedPasswordHash   string
	updatedPasswordUserID uuid.UUID
	resetToken            AuthTokenRecord
	insertAuthTokenCount  int
	invalidateTokenCount  int
}

func (f *fakeRepo) FindTenantByCode(ctx context.Context, tenantCode string) (uuid.UUID, error) {
	return f.tenantID, nil
}

func (f *fakeRepo) CreateUser(ctx context.Context, in CreateUserParams) (uuid.UUID, error) {
	f.createdUser = in
	f.createUserCount++
	return uuid.New(), nil
}

func (f *fakeRepo) CreateProfile(ctx context.Context, tenantID, userID uuid.UUID, fullName string, facultyID, departmentID uuid.UUID, groupID *uuid.UUID) error {
	return nil
}

func (f *fakeRepo) GrantStudentFacultyRole(ctx context.Context, tenantID, userID, facultyID uuid.UUID) error {
	return nil
}

func (f *fakeRepo) FindUserByEmail(ctx context.Context, tenantID uuid.UUID, email string) (User, error) {
	if f.findUserErr != nil {
		return User{}, f.findUserErr
	}
	return f.user, nil
}

func (f *fakeRepo) FindUserByID(ctx context.Context, tenantID, userID uuid.UUID) (User, error) {
	return f.user, nil
}

func (f *fakeRepo) UpdateUserProfile(ctx context.Context, tenantID, userID uuid.UUID, in ProfileUpdate, updatedAt time.Time) error {
	return nil
}

func (f *fakeRepo) UpdateUserPasswordHash(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string, changedAt time.Time) error {
	f.updatedPasswordHash = passwordHash
	f.updatedPasswordUserID = userID
	return nil
}

func (f *fakeRepo) MarkUserEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verifiedAt time.Time) error {
	return nil
}

func (f *fakeRepo) IsEmailInUse(ctx context.Context, tenantID, excludeUserID uuid.UUID, email string) (bool, error) {
	return false, nil
}

func (f *fakeRepo) SetPendingEmail(ctx context.Context, tenantID, userID uuid.UUID, pendingEmail string, requestedAt time.Time) error {
	return nil
}

func (f *fakeRepo) ActivatePendingEmail(ctx context.Context, tenantID, userID uuid.UUID, activatedAt time.Time) (string, error) {
	return f.user.Email, nil
}

func (f *fakeRepo) UpdateUserAvatarKey(ctx context.Context, tenantID, userID uuid.UUID, avatarKey *string, updatedAt time.Time) error {
	return nil
}

func (f *fakeRepo) InsertRefreshToken(ctx context.Context, tenantID, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	f.refreshInsertCount++
	return nil
}

func (f *fakeRepo) FindRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, uuid.UUID, time.Time, *time.Time, error) {
	return uuid.Nil, uuid.Nil, time.Time{}, nil, ErrNotFound
}

func (f *fakeRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return nil
}

func (f *fakeRepo) RevokeAndReturnRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, uuid.UUID, time.Time, error) {
	return uuid.Nil, uuid.Nil, time.Time{}, ErrNotFound
}

func (f *fakeRepo) RevokeUserRefreshTokens(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}

func (f *fakeRepo) FindDepartment(ctx context.Context, tenantID uuid.UUID, departmentCode string) (uuid.UUID, uuid.UUID, error) {
	return uuid.New(), uuid.New(), nil
}

func (f *fakeRepo) FindGroupByCodeInDepartment(ctx context.Context, tenantID, departmentID uuid.UUID, groupCode string) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (f *fakeRepo) ListDepartments(ctx context.Context, tenantID uuid.UUID) ([]Department, error) {
	return []Department{}, nil
}

func (f *fakeRepo) ListGroupsByDepartmentCode(ctx context.Context, tenantID uuid.UUID, departmentCode string) ([]StudentGroup, error) {
	return []StudentGroup{}, nil
}

func (f *fakeRepo) InsertGroupChangeRequest(ctx context.Context, tenantID, studentID, currentGroupID, requestedGroupID uuid.UUID, createdAt time.Time) (GroupChangeRequest, error) {
	return GroupChangeRequest{}, nil
}

func (f *fakeRepo) ListOwnGroupChangeRequests(ctx context.Context, tenantID, studentID uuid.UUID, limit int) ([]GroupChangeRequest, error) {
	return []GroupChangeRequest{}, nil
}

func (f *fakeRepo) ListGroupChangeRequests(ctx context.Context, tenantID uuid.UUID, status, search string, limit int) ([]GroupChangeRequest, error) {
	return []GroupChangeRequest{}, nil
}

func (f *fakeRepo) ReviewGroupChangeRequest(ctx context.Context, tenantID, requestID, reviewerID uuid.UUID, decision, comment string, reviewedAt time.Time) (GroupChangeRequest, error) {
	return GroupChangeRequest{}, nil
}

func (f *fakeRepo) ListDepartmentGroupsTree(ctx context.Context, tenantID uuid.UUID, departmentCode, search string) ([]DepartmentGroupsTree, error) {
	return []DepartmentGroupsTree{}, nil
}

func (f *fakeRepo) InsertAuthToken(ctx context.Context, tenantID, userID uuid.UUID, purpose, tokenHash string, expiresAt time.Time) error {
	f.insertAuthTokenCount++
	return nil
}

func (f *fakeRepo) FindAuthToken(ctx context.Context, purpose, tokenHash string) (AuthTokenRecord, error) {
	if f.resetToken.ID == uuid.Nil {
		return AuthTokenRecord{}, ErrNotFound
	}
	return f.resetToken, nil
}

func (f *fakeRepo) ConsumeAuthToken(ctx context.Context, tokenID uuid.UUID, consumedAt time.Time) error {
	return nil
}

func (f *fakeRepo) InvalidateAuthTokens(ctx context.Context, tenantID, userID uuid.UUID, purpose string) error {
	f.invalidateTokenCount++
	return nil
}

func TestLoginRateLimitedAfterFailures(t *testing.T) {
	hash, err := passwords.Hash("valid-password1")
	require.NoError(t, err)

	now := time.Now().UTC()
	repo := &fakeRepo{
		tenantID: uuid.New(),
		user: User{
			ID:              uuid.New(),
			TenantID:        uuid.New(),
			Email:           "student@example.edu",
			PasswordHash:    hash,
			Status:          StatusActive,
			FacultyID:       uuid.New(),
			DepartmentID:    uuid.New(),
			EmailVerifiedAt: &now,
		},
	}
	svc := NewService(repo, Config{
		JWTSecret:              "01234567890123456789012345678901",
		MaxFailedLoginAttempts: 5,
		LoginAttemptWindow:     time.Minute,
	})

	for range 5 {
		_, err = svc.Login(context.Background(), "127.0.0.1", "CORE", repo.user.Email, "wrong-password1")
		require.ErrorIs(t, err, ErrInvalidCredentials)
	}

	_, err = svc.Login(context.Background(), "127.0.0.1", "CORE", repo.user.Email, "wrong-password1")
	require.ErrorIs(t, err, ErrTooManyAttempts)
}

func TestLoginRehashesLegacyBcryptPassword(t *testing.T) {
	legacyHash, err := bcrypt.GenerateFromPassword([]byte("legacy-password1"), bcrypt.DefaultCost)
	require.NoError(t, err)

	now := time.Now().UTC()
	tenantID := uuid.New()
	userID := uuid.New()
	repo := &fakeRepo{
		tenantID: tenantID,
		user: User{
			ID:              userID,
			TenantID:        tenantID,
			Email:           "student@example.edu",
			PasswordHash:    string(legacyHash),
			Status:          StatusActive,
			FacultyID:       uuid.New(),
			DepartmentID:    uuid.New(),
			EmailVerifiedAt: &now,
		},
	}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})

	session, err := svc.Login(context.Background(), "127.0.0.1", "CORE", repo.user.Email, "legacy-password1")
	require.NoError(t, err)
	require.NotEmpty(t, session.Tokens.AccessToken)
	require.NotEmpty(t, session.Tokens.RefreshToken)
	require.Equal(t, 1, repo.refreshInsertCount)
	require.Equal(t, userID, repo.updatedPasswordUserID)
	require.NotEmpty(t, repo.updatedPasswordHash)
	require.NotEqual(t, string(legacyHash), repo.updatedPasswordHash)
	require.Contains(t, repo.updatedPasswordHash, "$argon2id$")
}

func TestResetPasswordRejectsExpiredToken(t *testing.T) {
	repo := &fakeRepo{
		resetToken: AuthTokenRecord{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			UserID:    uuid.New(),
			Purpose:   TokenPurposePasswordReset,
			ExpiresAt: time.Now().UTC().Add(-time.Minute),
		},
	}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})

	err := svc.ResetPassword(context.Background(), "raw-token", "new-password-123")
	require.ErrorIs(t, err, ErrTokenExpired)
}

func TestRegisterStudentRequiresVerificationByDefault(t *testing.T) {
	repo := &fakeRepo{tenantID: uuid.New()}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})

	err := svc.RegisterStudent(
		context.Background(),
		"CORE",
		"student@example.edu",
		"DemoPass123!",
		"Student User",
		"CS",
		"CS-101",
	)
	require.NoError(t, err)
	require.Equal(t, 1, repo.createUserCount)
	require.Equal(t, StatusPending, repo.createdUser.Status)
	require.Nil(t, repo.createdUser.EmailVerifiedAt)
	require.Equal(t, 1, repo.invalidateTokenCount)
	require.Equal(t, 1, repo.insertAuthTokenCount)
}

func TestRegisterStudentAutoVerifiesWhenEnabled(t *testing.T) {
	repo := &fakeRepo{tenantID: uuid.New()}
	svc := NewService(repo, Config{
		JWTSecret:             "01234567890123456789012345678901",
		AutoVerifyRegistrants: true,
	})

	err := svc.RegisterStudent(
		context.Background(),
		"CORE",
		"student@example.edu",
		"DemoPass123!",
		"Student User",
		"CS",
		"CS-101",
	)
	require.NoError(t, err)
	require.Equal(t, 1, repo.createUserCount)
	require.Equal(t, StatusActive, repo.createdUser.Status)
	require.NotNil(t, repo.createdUser.EmailVerifiedAt)
	require.Zero(t, repo.invalidateTokenCount)
	require.Zero(t, repo.insertAuthTokenCount)
}

func TestRequestPasswordResetRejectsUnknownAccount(t *testing.T) {
	repo := &fakeRepo{
		tenantID:    uuid.New(),
		findUserErr: ErrNotFound,
	}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})

	err := svc.RequestPasswordReset(context.Background(), "127.0.0.1", "CORE", "missing@example.edu")
	require.ErrorIs(t, err, ErrPasswordResetUnavailable)
	require.Zero(t, repo.insertAuthTokenCount)
}

func TestRequestPasswordResetRejectsInactiveOrUnverifiedAccount(t *testing.T) {
	repo := &fakeRepo{
		tenantID: uuid.New(),
		user: User{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Email:    "pending@example.edu",
			Status:   StatusPending,
		},
	}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})

	err := svc.RequestPasswordReset(context.Background(), "127.0.0.1", "CORE", "pending@example.edu")
	require.ErrorIs(t, err, ErrPasswordResetUnavailable)
	require.Zero(t, repo.insertAuthTokenCount)
}
