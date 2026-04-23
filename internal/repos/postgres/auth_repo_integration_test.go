//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/repos/postgres"
	authsvc "idsai-core-up/internal/services/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestAuthRepo_Integration_UserProfileTokensAndGroupFlow(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewAuthRepo(pool)
	scope := seedAuthScope(t, ctx, pool)

	foundTenantID, err := repo.FindTenantByCode(ctx, scope.tenantCode)
	require.NoError(t, err)
	require.Equal(t, scope.tenantID, foundTenantID)
	_, err = repo.FindTenantByCode(ctx, "missing-"+scope.tenantCode)
	require.ErrorIs(t, err, authsvc.ErrNotFound)

	faculties, err := repo.ListFaculties(ctx, scope.tenantID)
	require.NoError(t, err)
	require.Len(t, faculties, 1)
	require.Equal(t, scope.facultyID, faculties[0].ID)

	departments, err := repo.ListDepartments(ctx, scope.tenantID, authsvc.EducationTypeUniversity)
	require.NoError(t, err)
	require.Len(t, departments, 1)
	require.Equal(t, scope.departmentID, departments[0].ID)

	schoolDepartments, err := repo.ListDepartments(ctx, scope.tenantID, authsvc.EducationTypeSchool)
	require.NoError(t, err)
	require.Len(t, schoolDepartments, 1)
	require.Equal(t, "CLASS", schoolDepartments[0].Code)

	departmentID, facultyID, err := repo.FindDepartment(ctx, scope.tenantID, scope.departmentCode)
	require.NoError(t, err)
	require.Equal(t, scope.departmentID, departmentID)
	require.Equal(t, scope.facultyID, facultyID)

	departmentID, err = repo.FindDepartmentInFaculty(ctx, scope.tenantID, scope.facultyID, scope.departmentCode)
	require.NoError(t, err)
	require.Equal(t, scope.departmentID, departmentID)
	_, _, err = repo.FindDepartment(ctx, scope.tenantID, "NOPE")
	require.ErrorIs(t, err, authsvc.ErrDepartmentNotFound)

	schoolFacultyID, schoolDepartmentID, err := repo.FindSchoolRegistrationScope(ctx, scope.tenantID)
	require.NoError(t, err)
	require.Equal(t, scope.schoolFacultyID, schoolFacultyID)
	require.Equal(t, scope.schoolDepartmentID, schoolDepartmentID)

	groupID1, err := repo.CreateGroupInDepartment(ctx, scope.tenantID, scope.facultyID, scope.departmentID, scope.groupCode1, 101)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, groupID1)
	groupID2, err := repo.CreateGroupInDepartment(ctx, scope.tenantID, scope.facultyID, scope.departmentID, scope.groupCode2, 102)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, groupID2)

	foundGroupID, err := repo.FindGroupByCodeInDepartment(ctx, scope.tenantID, scope.departmentID, scope.groupCode1)
	require.NoError(t, err)
	require.Equal(t, groupID1, foundGroupID)
	_, err = repo.FindGroupByCodeInDepartment(ctx, scope.tenantID, scope.departmentID, "NOPE-1")
	require.ErrorIs(t, err, authsvc.ErrGroupNotFound)

	groups, err := repo.ListGroupsByDepartmentCode(ctx, scope.tenantID, scope.departmentCode)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(groups), 2)

	createdAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	studentEmail := fmt.Sprintf("auth-student-%s@example.local", uuid.NewString())
	studentID, err := repo.CreateUser(ctx, authsvc.CreateUserParams{
		TenantID:          scope.tenantID,
		Email:             studentEmail,
		PasswordHash:      "hash-one",
		Status:            authsvc.StatusPending,
		PasswordChangedAt: createdAt,
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, studentID)
	_, err = repo.CreateUser(ctx, authsvc.CreateUserParams{
		TenantID:          scope.tenantID,
		Email:             studentEmail,
		PasswordHash:      "hash-two",
		Status:            authsvc.StatusActive,
		PasswordChangedAt: createdAt,
	})
	require.ErrorIs(t, err, authsvc.ErrUserExists)

	require.NoError(t, repo.CreateProfile(ctx, scope.tenantID, studentID, "Auth Student", scope.facultyID, scope.departmentID, &groupID1, authsvc.InstitutionSelection{
		Provider:   "local",
		ExternalID: "institution-1",
		Name:       "Integration University",
		Address:    "Astana",
	}))
	require.NoError(t, repo.GrantStudentFacultyRole(ctx, scope.tenantID, studentID, scope.facultyID))

	user, err := repo.FindUserByEmail(ctx, scope.tenantID, studentEmail)
	require.NoError(t, err)
	require.Equal(t, studentID, user.ID)
	require.Equal(t, "Auth Student", user.FullName)
	require.Equal(t, scope.groupCode1, user.GroupCode)
	require.Equal(t, "Integration University", user.Institution.Name)
	require.False(t, user.IsAdmin)

	user, err = repo.FindUserByID(ctx, scope.tenantID, studentID)
	require.NoError(t, err)
	require.Equal(t, studentEmail, user.Email)
	_, err = repo.FindUserByID(ctx, scope.tenantID, uuid.New())
	require.ErrorIs(t, err, authsvc.ErrNotFound)

	updatedAt := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, repo.UpdateUserProfile(ctx, scope.tenantID, studentID, authsvc.ProfileUpdate{
		FullName:      "Auth Student Updated",
		Headline:      "Backend learner",
		About:         "Loves integration tests",
		PreferredRole: "Backend",
		Semester:      "6",
		Availability:  "Evenings",
		Goals:         "Ship diploma",
		GithubURL:     "https://github.com/example",
		Telegram:      "@auth_student",
		PortfolioURL:  "https://example.local",
		Stacks:        []string{"go", "postgres"},
		Interests:     []string{"security", "testing"},
	}, updatedAt))
	require.NoError(t, repo.UpdateUserPasswordHash(ctx, scope.tenantID, studentID, "hash-new", updatedAt))
	require.NoError(t, repo.MarkUserEmailVerified(ctx, scope.tenantID, studentID, updatedAt))

	state, err := repo.GetUserAuthState(ctx, scope.tenantID, studentID)
	require.NoError(t, err)
	require.Equal(t, authsvc.StatusActive, state.Status)
	require.NotNil(t, state.EmailVerifiedAt)

	user, err = repo.FindUserByID(ctx, scope.tenantID, studentID)
	require.NoError(t, err)
	require.Equal(t, "Auth Student Updated", user.FullName)
	require.Equal(t, "hash-new", user.PasswordHash)
	require.Equal(t, []string{"go", "postgres"}, user.Stacks)
	require.Equal(t, []string{"security", "testing"}, user.Interests)

	inUse, err := repo.IsEmailInUse(ctx, scope.tenantID, studentID, studentEmail)
	require.NoError(t, err)
	require.False(t, inUse)

	newEmail := fmt.Sprintf("auth-new-%s@example.local", uuid.NewString())
	require.NoError(t, repo.SetPendingEmail(ctx, scope.tenantID, studentID, newEmail, updatedAt))
	inUse, err = repo.IsEmailInUse(ctx, scope.tenantID, uuid.New(), newEmail)
	require.NoError(t, err)
	require.True(t, inUse)
	activatedEmail, err := repo.ActivatePendingEmail(ctx, scope.tenantID, studentID, updatedAt)
	require.NoError(t, err)
	require.Equal(t, newEmail, activatedEmail)
	_, err = repo.ActivatePendingEmail(ctx, scope.tenantID, studentID, updatedAt)
	require.ErrorIs(t, err, authsvc.ErrNoPendingEmail)

	avatarKey := "avatars/student.jpg"
	require.NoError(t, repo.UpdateUserAvatarKey(ctx, scope.tenantID, studentID, &avatarKey, updatedAt))
	user, err = repo.FindUserByID(ctx, scope.tenantID, studentID)
	require.NoError(t, err)
	require.Equal(t, avatarKey, user.AvatarKey)
	require.NoError(t, repo.UpdateUserAvatarKey(ctx, scope.tenantID, studentID, nil, updatedAt))

	refreshHash := "refresh-" + uuid.NewString()
	refreshExpiresAt := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	require.NoError(t, repo.InsertRefreshToken(ctx, scope.tenantID, studentID, refreshHash, refreshExpiresAt))
	tokenTenantID, tokenUserID, expiresAt, revokedAt, err := repo.FindRefreshToken(ctx, refreshHash)
	require.NoError(t, err)
	require.Equal(t, scope.tenantID, tokenTenantID)
	require.Equal(t, studentID, tokenUserID)
	require.WithinDuration(t, refreshExpiresAt, expiresAt, time.Second)
	require.Nil(t, revokedAt)
	require.NoError(t, repo.RevokeRefreshToken(ctx, refreshHash))
	_, _, _, revokedAt, err = repo.FindRefreshToken(ctx, refreshHash)
	require.NoError(t, err)
	require.NotNil(t, revokedAt)
	_, _, _, err = repo.RevokeAndReturnRefreshToken(ctx, refreshHash)
	require.ErrorIs(t, err, authsvc.ErrNotFound)

	secondRefreshHash := "refresh-" + uuid.NewString()
	require.NoError(t, repo.InsertRefreshToken(ctx, scope.tenantID, studentID, secondRefreshHash, refreshExpiresAt))
	tokenTenantID, tokenUserID, expiresAt, err = repo.RevokeAndReturnRefreshToken(ctx, secondRefreshHash)
	require.NoError(t, err)
	require.Equal(t, scope.tenantID, tokenTenantID)
	require.Equal(t, studentID, tokenUserID)
	require.WithinDuration(t, refreshExpiresAt, expiresAt, time.Second)
	require.NoError(t, repo.RevokeUserRefreshTokens(ctx, scope.tenantID, studentID))

	authTokenHash := "auth-" + uuid.NewString()
	authTokenExpiresAt := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	require.NoError(t, repo.InsertAuthToken(ctx, scope.tenantID, studentID, authsvc.TokenPurposeEmailVerification, authTokenHash, authTokenExpiresAt))
	authToken, err := repo.FindAuthToken(ctx, authsvc.TokenPurposeEmailVerification, authTokenHash)
	require.NoError(t, err)
	require.Equal(t, scope.tenantID, authToken.TenantID)
	require.Equal(t, studentID, authToken.UserID)
	require.Nil(t, authToken.ConsumedAt)
	require.NoError(t, repo.ConsumeAuthToken(ctx, authToken.ID, updatedAt))
	err = repo.ConsumeAuthToken(ctx, authToken.ID, updatedAt)
	require.ErrorIs(t, err, authsvc.ErrNotFound)

	resetTokenHash := "reset-" + uuid.NewString()
	require.NoError(t, repo.InsertAuthToken(ctx, scope.tenantID, studentID, authsvc.TokenPurposePasswordReset, resetTokenHash, authTokenExpiresAt))
	require.NoError(t, repo.InvalidateAuthTokens(ctx, scope.tenantID, studentID, authsvc.TokenPurposePasswordReset))
	authToken, err = repo.FindAuthToken(ctx, authsvc.TokenPurposePasswordReset, resetTokenHash)
	require.NoError(t, err)
	require.NotNil(t, authToken.ConsumedAt)

	reviewerID := seedAuthUserProfile(t, ctx, pool, scope, "reviewer", groupID1)
	request, err := repo.InsertGroupChangeRequest(ctx, scope.tenantID, studentID, groupID1, groupID2, createdAt)
	require.NoError(t, err)
	require.Equal(t, studentID, request.StudentID)
	require.Equal(t, "PENDING", request.Status)
	_, err = repo.InsertGroupChangeRequest(ctx, scope.tenantID, studentID, groupID1, groupID2, createdAt)
	require.ErrorIs(t, err, authsvc.ErrPendingGroupRequestExists)

	ownRequests, err := repo.ListOwnGroupChangeRequests(ctx, scope.tenantID, studentID, 10)
	require.NoError(t, err)
	require.Len(t, ownRequests, 1)
	allRequests, err := repo.ListGroupChangeRequests(ctx, scope.tenantID, "PENDING", "Student", 10)
	require.NoError(t, err)
	require.Len(t, allRequests, 1)

	reviewed, err := repo.ReviewGroupChangeRequest(ctx, scope.tenantID, request.ID, reviewerID, "APPROVED", "ok", updatedAt)
	require.NoError(t, err)
	require.Equal(t, "APPROVED", reviewed.Status)
	require.Equal(t, "ok", reviewed.AdminComment)
	_, err = repo.ReviewGroupChangeRequest(ctx, scope.tenantID, request.ID, reviewerID, "APPROVED", "again", updatedAt)
	require.ErrorIs(t, err, authsvc.ErrGroupRequestReviewed)
	_, err = repo.ReviewGroupChangeRequest(ctx, scope.tenantID, uuid.New(), reviewerID, "APPROVED", "missing", updatedAt)
	require.ErrorIs(t, err, authsvc.ErrGroupRequestNotFound)

	user, err = repo.FindUserByID(ctx, scope.tenantID, studentID)
	require.NoError(t, err)
	require.NotNil(t, user.GroupID)
	require.Equal(t, groupID2, *user.GroupID)
	require.Equal(t, scope.groupCode2, user.GroupCode)

	tree, err := repo.ListDepartmentGroupsTree(ctx, scope.tenantID, scope.departmentCode, "Updated", authsvc.EducationTypeUniversity)
	require.NoError(t, err)
	require.Len(t, tree, 1)
	require.NotEmpty(t, tree[0].Groups)
	require.Equal(t, scope.departmentCode, tree[0].Code)
}

type authScope struct {
	tenantID           uuid.UUID
	tenantCode         string
	facultyID          uuid.UUID
	departmentID       uuid.UUID
	departmentCode     string
	schoolFacultyID    uuid.UUID
	schoolDepartmentID uuid.UUID
	groupCode1         string
	groupCode2         string
}

func seedAuthScope(t *testing.T, ctx context.Context, pool *pgxpool.Pool) authScope {
	t.Helper()

	tenantID := uuid.New()
	facultyID := uuid.New()
	departmentID := uuid.New()
	schoolFacultyID := uuid.New()
	schoolDepartmentID := uuid.New()
	short := tenantID.String()[:8]
	tenantCode := "AUTH_T_" + short
	departmentCode := "AUTHD" + short[:3]

	_, err := pool.Exec(ctx, `
INSERT INTO tenants(id, code, name, status)
VALUES ($1, $2, $3, 'ACTIVE');
`, tenantID, tenantCode, "Auth Tenant")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO faculties(id, tenant_id, code, name)
VALUES
  ($1, $2, $3, 'Auth Faculty'),
  ($4, $2, $5, 'Auth School');
`, facultyID, tenantID, "AUTH_F_"+short, schoolFacultyID, tenantCode+"_SCHOOL")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO departments(id, tenant_id, faculty_id, code, name)
VALUES
  ($1, $2, $3, $4, 'Auth Department'),
  ($5, $2, $6, 'CLASS', 'School Classes');
`, departmentID, tenantID, facultyID, departmentCode, schoolDepartmentID, schoolFacultyID)
	require.NoError(t, err)

	return authScope{
		tenantID:           tenantID,
		tenantCode:         tenantCode,
		facultyID:          facultyID,
		departmentID:       departmentID,
		departmentCode:     departmentCode,
		schoolFacultyID:    schoolFacultyID,
		schoolDepartmentID: schoolDepartmentID,
		groupCode1:         departmentCode + "-101",
		groupCode2:         departmentCode + "-102",
	}
}

func seedAuthUserProfile(t *testing.T, ctx context.Context, pool *pgxpool.Pool, scope authScope, label string, groupID uuid.UUID) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	_, err := pool.Exec(ctx, `
INSERT INTO users(id, tenant_id, email, password_hash, status, password_changed_at)
VALUES ($1, $2, $3, 'integration-hash', 'ACTIVE', now());
`, userID, scope.tenantID, fmt.Sprintf("auth-%s-%s@example.local", label, userID.String()))
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO user_profiles(tenant_id, user_id, full_name, faculty_id, department_id, group_id)
VALUES ($1, $2, $3, $4, $5, $6);
`, scope.tenantID, userID, "Auth "+label, scope.facultyID, scope.departmentID, groupID)
	require.NoError(t, err)

	return userID
}
