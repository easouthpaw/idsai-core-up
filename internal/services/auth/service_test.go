package auth

import (
	"bytes"
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"idsai-core-up/internal/security/passwords"
	notifsvc "idsai-core-up/internal/services/notifications"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type fakeRepo struct {
	tenantID              uuid.UUID
	user                  User
	faculties             []Faculty
	findUserErr           error
	findGroupErr          error
	createdUser           CreateUserParams
	createUserCount       int
	refreshInsertCount    int
	updatedPasswordHash   string
	updatedPasswordUserID uuid.UUID
	resetToken            AuthTokenRecord
	insertAuthTokenCount  int
	invalidateTokenCount  int
	refreshTenantID       uuid.UUID
	refreshUserID         uuid.UUID
	refreshExpiresAt      time.Time
	refreshRevokedAt      *time.Time
	revokeRefreshCount    int
	createdGroupID        uuid.UUID
	createdGroupCode      string
	createdGroupNumber    int
	profileGroupID        uuid.UUID
	profileInstitution    InstitutionSelection
}

type fakeNotifier struct {
	err    error
	calls  int
	lastIn notifsvc.CreateInput
}

func (f *fakeNotifier) Notify(ctx context.Context, in notifsvc.CreateInput) (notifsvc.Notification, error) {
	f.calls++
	f.lastIn = in
	if f.err != nil {
		return notifsvc.Notification{}, f.err
	}
	return notifsvc.Notification{ID: uuid.NewString()}, nil
}

func captureAuthLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	prevWriter := log.Writer()
	prevFlags := log.Flags()
	prevPrefix := log.Prefix()
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(prevWriter)
		log.SetFlags(prevFlags)
		log.SetPrefix(prevPrefix)
	})
	return &buf
}

func (f *fakeRepo) FindTenantByCode(ctx context.Context, tenantCode string) (uuid.UUID, error) {
	return f.tenantID, nil
}

func (f *fakeRepo) CreateUser(ctx context.Context, in CreateUserParams) (uuid.UUID, error) {
	f.createdUser = in
	f.createUserCount++
	return uuid.New(), nil
}

func (f *fakeRepo) CreateProfile(ctx context.Context, tenantID, userID uuid.UUID, fullName string, facultyID, departmentID uuid.UUID, groupID *uuid.UUID, institution InstitutionSelection) error {
	if groupID != nil {
		f.profileGroupID = *groupID
	}
	f.profileInstitution = institution
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
	f.user.FullName = in.FullName
	f.user.Headline = in.Headline
	f.user.Stacks = append([]string(nil), in.Stacks...)
	f.user.Interests = append([]string(nil), in.Interests...)
	f.user.ProfileUpdatedAt = updatedAt
	return nil
}

func (f *fakeRepo) UpdateUserPasswordHash(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string, changedAt time.Time) error {
	f.updatedPasswordHash = passwordHash
	f.updatedPasswordUserID = userID
	return nil
}

func (f *fakeRepo) MarkUserEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verifiedAt time.Time) error {
	f.user.EmailVerifiedAt = &verifiedAt
	f.user.Status = StatusActive
	return nil
}

func (f *fakeRepo) IsEmailInUse(ctx context.Context, tenantID, excludeUserID uuid.UUID, email string) (bool, error) {
	return false, nil
}

func (f *fakeRepo) SetPendingEmail(ctx context.Context, tenantID, userID uuid.UUID, pendingEmail string, requestedAt time.Time) error {
	f.user.PendingEmail = pendingEmail
	f.user.PendingEmailAt = &requestedAt
	return nil
}

func (f *fakeRepo) ActivatePendingEmail(ctx context.Context, tenantID, userID uuid.UUID, activatedAt time.Time) (string, error) {
	if f.user.PendingEmail != "" {
		f.user.Email = f.user.PendingEmail
		f.user.PendingEmail = ""
		f.user.PendingEmailAt = nil
	}
	return f.user.Email, nil
}

func (f *fakeRepo) UpdateUserAvatarKey(ctx context.Context, tenantID, userID uuid.UUID, avatarKey *string, updatedAt time.Time) error {
	if avatarKey == nil {
		f.user.AvatarKey = ""
	} else {
		f.user.AvatarKey = *avatarKey
	}
	f.user.AvatarUpdatedAt = &updatedAt
	return nil
}

func (f *fakeRepo) InsertRefreshToken(ctx context.Context, tenantID, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	f.refreshInsertCount++
	return nil
}

func (f *fakeRepo) FindRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, uuid.UUID, time.Time, *time.Time, error) {
	if f.refreshTenantID != uuid.Nil {
		return f.refreshTenantID, f.refreshUserID, f.refreshExpiresAt, f.refreshRevokedAt, nil
	}
	return uuid.Nil, uuid.Nil, time.Time{}, nil, ErrNotFound
}

func (f *fakeRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	f.revokeRefreshCount++
	return nil
}

func (f *fakeRepo) RevokeAndReturnRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, uuid.UUID, time.Time, error) {
	if f.refreshTenantID != uuid.Nil && f.refreshRevokedAt == nil {
		now := time.Now().UTC()
		f.refreshRevokedAt = &now
		return f.refreshTenantID, f.refreshUserID, f.refreshExpiresAt, nil
	}
	return uuid.Nil, uuid.Nil, time.Time{}, ErrNotFound
}

func (f *fakeRepo) RevokeUserRefreshTokens(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}

func (f *fakeRepo) ListFaculties(ctx context.Context, tenantID uuid.UUID) ([]Faculty, error) {
	return f.faculties, nil
}

func (f *fakeRepo) FindDepartment(ctx context.Context, tenantID uuid.UUID, departmentCode string) (uuid.UUID, uuid.UUID, error) {
	return uuid.New(), uuid.New(), nil
}

func (f *fakeRepo) FindDepartmentInFaculty(ctx context.Context, tenantID, facultyID uuid.UUID, departmentCode string) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (f *fakeRepo) FindSchoolRegistrationScope(ctx context.Context, tenantID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	return uuid.New(), uuid.New(), nil
}

func (f *fakeRepo) FindGroupByCodeInDepartment(ctx context.Context, tenantID, departmentID uuid.UUID, groupCode string) (uuid.UUID, error) {
	if f.findGroupErr != nil {
		return uuid.Nil, f.findGroupErr
	}
	return uuid.New(), nil
}

func (f *fakeRepo) CreateGroupInDepartment(ctx context.Context, tenantID, facultyID, departmentID uuid.UUID, groupCode string, groupNumber int) (uuid.UUID, error) {
	if f.createdGroupID == uuid.Nil {
		f.createdGroupID = uuid.New()
	}
	f.createdGroupCode = groupCode
	f.createdGroupNumber = groupNumber
	return f.createdGroupID, nil
}

func (f *fakeRepo) ListDepartments(ctx context.Context, tenantID uuid.UUID, educationType string) ([]Department, error) {
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

func (f *fakeRepo) ListDepartmentGroupsTree(ctx context.Context, tenantID uuid.UUID, departmentCode, search, educationType string) ([]DepartmentGroupsTree, error) {
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

func TestRegisterStudentCreatesMissingGroupFromManualNumber(t *testing.T) {
	repo := &fakeRepo{
		tenantID:     uuid.New(),
		findGroupErr: ErrGroupNotFound,
	}
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
		"101",
	)
	require.NoError(t, err)
	require.Equal(t, "CS-101", repo.createdGroupCode)
	require.Equal(t, 101, repo.createdGroupNumber)
	require.Equal(t, repo.createdGroupID, repo.profileGroupID)
}

func TestRegisterSchoolCreatesMissingClassGroup(t *testing.T) {
	repo := &fakeRepo{
		tenantID:     uuid.New(),
		findGroupErr: ErrGroupNotFound,
	}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})

	err := svc.Register(
		context.Background(),
		"CORE",
		RegistrationInput{
			Email:         "school.student@example.edu",
			Password:      "DemoPass123!",
			FullName:      "School Student",
			EducationType: EducationTypeSchool,
			SchoolClass:   "10A",
			Institution: InstitutionSelection{
				Provider:   InstitutionProviderPhoton,
				ExternalID: "school-17",
				Name:       "Школа-лицей №17",
				Address:    "Астана, ул. Абая, 1",
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, "CLASS-10A", repo.createdGroupCode)
	require.Equal(t, 1065, repo.createdGroupNumber)
	require.Equal(t, repo.createdGroupID, repo.profileGroupID)
	require.Equal(t, InstitutionProviderPhoton, repo.profileInstitution.Provider)
	require.Equal(t, "school-17", repo.profileInstitution.ExternalID)
	require.Equal(t, "Школа-лицей №17", repo.profileInstitution.Name)
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
	require.NoError(t, err)
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
	require.NoError(t, err)
	require.Zero(t, repo.insertAuthTokenCount)
}

func TestRequestPasswordResetIssuesSeparateCodeAndLinkTokens(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{
		tenantID: uuid.New(),
		user: User{
			ID:                uuid.New(),
			TenantID:          uuid.New(),
			Email:             "student@example.edu",
			Status:            StatusActive,
			EmailVerifiedAt:   &now,
			PasswordChangedAt: now,
		},
	}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})

	err := svc.RequestPasswordReset(context.Background(), "127.0.0.1", "CORE", repo.user.Email)
	require.NoError(t, err)
	require.Equal(t, 1, repo.invalidateTokenCount)
	require.Equal(t, 2, repo.insertAuthTokenCount)
}

func TestRequestPasswordResetLogsQueuedEmail(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{
		tenantID: uuid.New(),
		user: User{
			ID:                uuid.New(),
			TenantID:          uuid.New(),
			Email:             "student@example.edu",
			Status:            StatusActive,
			EmailVerifiedAt:   &now,
			PasswordChangedAt: now,
		},
	}
	notifier := &fakeNotifier{}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})
	svc.SetNotifier(notifier)
	logs := captureAuthLogs(t)

	err := svc.RequestPasswordReset(context.Background(), "127.0.0.1", "CORE", repo.user.Email)
	require.NoError(t, err)
	require.Equal(t, 1, notifier.calls)
	require.Contains(t, logs.String(), "auth password reset email queued")
	require.Contains(t, logs.String(), repo.user.Email)
}

func TestRequestPasswordResetLogsQueueFailure(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{
		tenantID: uuid.New(),
		user: User{
			ID:                uuid.New(),
			TenantID:          uuid.New(),
			Email:             "student@example.edu",
			Status:            StatusActive,
			EmailVerifiedAt:   &now,
			PasswordChangedAt: now,
		},
	}
	notifier := &fakeNotifier{err: errors.New("outbox insert failed")}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})
	svc.SetNotifier(notifier)
	logs := captureAuthLogs(t)

	err := svc.RequestPasswordReset(context.Background(), "127.0.0.1", "CORE", repo.user.Email)
	require.NoError(t, err)
	require.Equal(t, 1, notifier.calls)
	require.Contains(t, logs.String(), "auth password reset email enqueue failed")
	require.Contains(t, logs.String(), "outbox insert failed")
}

func TestResetPasswordInvalidatesAllPasswordResetTokens(t *testing.T) {
	repo := &fakeRepo{
		resetToken: AuthTokenRecord{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			UserID:    uuid.New(),
			Purpose:   TokenPurposePasswordReset,
			ExpiresAt: time.Now().UTC().Add(time.Minute),
		},
	}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})

	err := svc.ResetPassword(context.Background(), "raw-token", "new-password-123")
	require.NoError(t, err)
	require.Equal(t, 1, repo.invalidateTokenCount)
	require.NotEmpty(t, repo.updatedPasswordHash)
}

func TestResetPasswordByCodeRateLimitedAfterFailures(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepo{
		tenantID: uuid.New(),
		user: User{
			ID:                uuid.New(),
			TenantID:          uuid.New(),
			Email:             "student@example.edu",
			Status:            StatusActive,
			EmailVerifiedAt:   &now,
			PasswordChangedAt: now,
		},
	}
	svc := NewService(repo, Config{
		JWTSecret: "01234567890123456789012345678901",
	})

	for range 5 {
		err := svc.ResetPasswordByCode(context.Background(), "127.0.0.1", "CORE", repo.user.Email, "123456", "new-password-123")
		require.ErrorIs(t, err, ErrTokenInvalid)
	}

	err := svc.ResetPasswordByCode(context.Background(), "127.0.0.1", "CORE", repo.user.Email, "123456", "new-password-123")
	require.ErrorIs(t, err, ErrTooManyAttempts)
}
