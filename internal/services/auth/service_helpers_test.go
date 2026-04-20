package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"idsai-core-up/internal/security/passwords"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type authRepoRecorder struct {
	fakeRepo

	facultiesOut   []Faculty
	departmentsOut []Department
	groupsOut      []StudentGroup
	departmentID   uuid.UUID
	facultyID      uuid.UUID

	findTenantCode string

	findDepartmentTenantID uuid.UUID
	findDepartmentCode     string

	listGroupsTenantID  uuid.UUID
	listGroupsCode      string
	listDepartmentsType string

	listOwnRequestsOut []GroupChangeRequest
	listOwnLimit       int

	listGroupRequestsOut []GroupChangeRequest
	listGroupStatus      string
	listGroupSearch      string
	listGroupLimit       int

	reviewTenantID uuid.UUID
	reviewRequest  uuid.UUID
	reviewReviewer uuid.UUID
	reviewDecision string
	reviewComment  string
	reviewAt       time.Time
	reviewOut      GroupChangeRequest

	treeOut            []DepartmentGroupsTree
	treeTenantID       uuid.UUID
	treeDepartmentCode string
	treeSearch         string
	treeEducationType  string

	updateProfileTenantID uuid.UUID
	updateProfileUserID   uuid.UUID
	updateProfileInput    ProfileUpdate
	updateProfileAt       time.Time

	emailInUse          bool
	setPendingEmail     string
	setPendingRequested time.Time
	setPendingTenantID  uuid.UUID
	setPendingUserID    uuid.UUID
	consumeTokenID      uuid.UUID
	consumeTokenAt      time.Time
	revokeTenantID      uuid.UUID
	revokeUserID        uuid.UUID
	revokeRefreshCalls  int
	activateTenantID    uuid.UUID
	activateUserID      uuid.UUID
	activatePendingAt   time.Time
}

func (r *authRepoRecorder) FindTenantByCode(ctx context.Context, tenantCode string) (uuid.UUID, error) {
	r.findTenantCode = tenantCode
	if r.tenantID == uuid.Nil {
		r.tenantID = uuid.New()
	}
	return r.tenantID, nil
}

func (r *authRepoRecorder) UpdateUserProfile(ctx context.Context, tenantID, userID uuid.UUID, in ProfileUpdate, updatedAt time.Time) error {
	r.updateProfileTenantID = tenantID
	r.updateProfileUserID = userID
	r.updateProfileInput = in
	r.updateProfileAt = updatedAt
	return nil
}

func (r *authRepoRecorder) ListFaculties(ctx context.Context, tenantID uuid.UUID) ([]Faculty, error) {
	return r.facultiesOut, nil
}

func (r *authRepoRecorder) ListDepartments(ctx context.Context, tenantID uuid.UUID, educationType string) ([]Department, error) {
	r.listDepartmentsType = educationType
	return r.departmentsOut, nil
}

func (r *authRepoRecorder) FindDepartment(ctx context.Context, tenantID uuid.UUID, departmentCode string) (uuid.UUID, uuid.UUID, error) {
	r.findDepartmentTenantID = tenantID
	r.findDepartmentCode = departmentCode
	if r.departmentID == uuid.Nil {
		r.departmentID = uuid.New()
	}
	if r.facultyID == uuid.Nil {
		r.facultyID = uuid.New()
	}
	return r.departmentID, r.facultyID, nil
}

func (r *authRepoRecorder) FindDepartmentInFaculty(ctx context.Context, tenantID, facultyID uuid.UUID, departmentCode string) (uuid.UUID, error) {
	r.findDepartmentTenantID = tenantID
	r.findDepartmentCode = departmentCode
	r.facultyID = facultyID
	if r.departmentID == uuid.Nil {
		r.departmentID = uuid.New()
	}
	return r.departmentID, nil
}

func (r *authRepoRecorder) FindSchoolRegistrationScope(ctx context.Context, tenantID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	if r.facultyID == uuid.Nil {
		r.facultyID = uuid.New()
	}
	if r.departmentID == uuid.Nil {
		r.departmentID = uuid.New()
	}
	return r.facultyID, r.departmentID, nil
}

func (r *authRepoRecorder) ListGroupsByDepartmentCode(ctx context.Context, tenantID uuid.UUID, departmentCode string) ([]StudentGroup, error) {
	r.listGroupsTenantID = tenantID
	r.listGroupsCode = departmentCode
	return r.groupsOut, nil
}

func (r *authRepoRecorder) ListOwnGroupChangeRequests(ctx context.Context, tenantID, studentID uuid.UUID, limit int) ([]GroupChangeRequest, error) {
	r.listOwnLimit = limit
	return r.listOwnRequestsOut, nil
}

func (r *authRepoRecorder) ListGroupChangeRequests(ctx context.Context, tenantID uuid.UUID, status, search string, limit int) ([]GroupChangeRequest, error) {
	r.listGroupStatus = status
	r.listGroupSearch = search
	r.listGroupLimit = limit
	return r.listGroupRequestsOut, nil
}

func (r *authRepoRecorder) ReviewGroupChangeRequest(ctx context.Context, tenantID, requestID, reviewerID uuid.UUID, decision, comment string, reviewedAt time.Time) (GroupChangeRequest, error) {
	r.reviewTenantID = tenantID
	r.reviewRequest = requestID
	r.reviewReviewer = reviewerID
	r.reviewDecision = decision
	r.reviewComment = comment
	r.reviewAt = reviewedAt
	return r.reviewOut, nil
}

func (r *authRepoRecorder) ListDepartmentGroupsTree(ctx context.Context, tenantID uuid.UUID, departmentCode, search, educationType string) ([]DepartmentGroupsTree, error) {
	r.treeTenantID = tenantID
	r.treeDepartmentCode = departmentCode
	r.treeSearch = search
	r.treeEducationType = educationType
	return r.treeOut, nil
}

func (r *authRepoRecorder) IsEmailInUse(ctx context.Context, tenantID, excludeUserID uuid.UUID, email string) (bool, error) {
	return r.emailInUse, nil
}

func (r *authRepoRecorder) SetPendingEmail(ctx context.Context, tenantID, userID uuid.UUID, pendingEmail string, requestedAt time.Time) error {
	r.setPendingTenantID = tenantID
	r.setPendingUserID = userID
	r.setPendingEmail = pendingEmail
	r.setPendingRequested = requestedAt
	return nil
}

func (r *authRepoRecorder) ActivatePendingEmail(ctx context.Context, tenantID, userID uuid.UUID, activatedAt time.Time) (string, error) {
	r.activateTenantID = tenantID
	r.activateUserID = userID
	r.activatePendingAt = activatedAt
	r.user.Email = r.setPendingEmail
	return r.user.Email, nil
}

func (r *authRepoRecorder) ConsumeAuthToken(ctx context.Context, tokenID uuid.UUID, consumedAt time.Time) error {
	r.consumeTokenID = tokenID
	r.consumeTokenAt = consumedAt
	return nil
}

func (r *authRepoRecorder) RevokeUserRefreshTokens(ctx context.Context, tenantID, userID uuid.UUID) error {
	r.revokeTenantID = tenantID
	r.revokeUserID = userID
	r.revokeRefreshCalls++
	return nil
}

type authTestStorage struct {
	available bool
}

func (s *authTestStorage) PutObject(ctx context.Context, key, contentType string, body []byte) error {
	return nil
}

func (s *authTestStorage) DeleteObject(ctx context.Context, key string) error {
	return nil
}

func (s *authTestStorage) PublicURL(key string) string {
	return "https://cdn.example.com/" + strings.TrimSpace(key)
}

func (s *authTestStorage) Available() bool {
	return s.available
}

func TestServiceConfigurationAndHelperFunctions(t *testing.T) {
	svc := NewService(&fakeRepo{}, Config{
		JWTSecret:             "01234567890123456789012345678901",
		AccessTTL:             2 * time.Minute,
		RefreshTTL:            3 * time.Hour,
		PasswordResetTTL:      4 * time.Hour,
		AutoVerifyRegistrants: true,
	})

	storage := &authTestStorage{available: true}
	svc.SetStorage(storage)

	require.Equal(t, 2*time.Minute, svc.AccessTTL())
	require.Equal(t, 3*time.Hour, svc.RefreshTTL())
	require.Equal(t, 4*time.Hour, svc.PasswordResetTTL())
	require.False(t, svc.RegistrationRequiresVerification())
	require.Equal(t, storage, svc.storage)

	payload, err := normalizeProfileUpdate(ProfileUpdate{
		FullName:      "  Student Example  ",
		Headline:      "  Backend  ",
		Stacks:        []string{" Go ", "go", "", strings.Repeat("x", 41)},
		Interests:     []string{"AI", "ai", "Testing"},
		PortfolioURL:  " https://example.com ",
		Availability:  " evenings ",
		PreferredRole: " backend ",
	})
	require.NoError(t, err)
	require.Equal(t, "Student Example", payload.FullName)
	require.Equal(t, "Backend", payload.Headline)
	require.Equal(t, []string{"Go"}, payload.Stacks)
	require.Equal(t, []string{"AI", "Testing"}, payload.Interests)
	require.Equal(t, "https://example.com", payload.PortfolioURL)
	require.Equal(t, "evenings", payload.Availability)

	_, err = normalizeProfileUpdate(ProfileUpdate{FullName: "A"})
	require.ErrorIs(t, err, ErrInvalidInput)
	require.Equal(t, []string{}, normalizeProfileTags(nil, 4))
	require.Equal(t, "CORE", normalizeTenantCode(""))
	require.Equal(t, "cpi@example.edu", normalizeEmail(" CPI@EXAMPLE.EDU "))
	require.Equal(t, "CPI-2201", normalizeDepartmentGroupCode("cpi", "2201"))

	groupNumber, err := groupNumberFromCode("CPI-2201")
	require.NoError(t, err)
	require.Equal(t, 2201, groupNumber)

	rawToken, err := randomToken(4)
	require.NoError(t, err)
	require.Len(t, rawToken, 8)

	code, err := randomNumericCode(6)
	require.NoError(t, err)
	require.Len(t, code, 6)
	require.True(t, isVerificationCode(code))
	_, err = randomNumericCode(19)
	require.ErrorIs(t, err, ErrInvalidInput)

	require.True(t, isVerificationCode("123456"))
	require.False(t, isVerificationCode("12345a"))
	require.Equal(t, int64(0), passwordChangedUnixMilli(time.Time{}))
	require.NotZero(t, passwordChangedUnixMilli(time.Now()))
}

func TestServiceMeAndUpdateProfile(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	repo := &authRepoRecorder{
		fakeRepo: fakeRepo{
			user: User{
				ID:               userID,
				TenantID:         tenantID,
				FacultyID:        uuid.New(),
				DepartmentID:     uuid.New(),
				DepartmentCode:   "CPI",
				Email:            "student@example.edu",
				FullName:         "Student Example",
				AvatarKey:        "avatars/user.jpg",
				ProfileUpdatedAt: time.Now().UTC(),
			},
		},
	}
	svc := NewService(repo, Config{JWTSecret: "01234567890123456789012345678901"})
	svc.SetStorage(&authTestStorage{available: true})

	me, err := svc.Me(context.Background(), tenantID, userID)
	require.NoError(t, err)
	require.Equal(t, "https://cdn.example.com/avatars/user.jpg", me.AvatarURL)

	updated, err := svc.UpdateProfile(context.Background(), tenantID, userID, ProfileUpdate{
		FullName:  "  Updated Name  ",
		Stacks:    []string{"Go", "go"},
		Interests: []string{"AI"},
	})
	require.NoError(t, err)
	require.Equal(t, tenantID, repo.updateProfileTenantID)
	require.Equal(t, userID, repo.updateProfileUserID)
	require.Equal(t, "Updated Name", repo.updateProfileInput.FullName)
	require.Equal(t, []string{"Go"}, repo.updateProfileInput.Stacks)
	require.False(t, repo.updateProfileAt.IsZero())
	require.Equal(t, "https://cdn.example.com/avatars/user.jpg", updated.AvatarURL)
}

func TestServiceGroupListAndReviewFlows(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	requestID := uuid.New()
	reviewerID := uuid.New()
	departmentID := uuid.New()
	repo := &authRepoRecorder{
		fakeRepo: fakeRepo{tenantID: tenantID},
		facultiesOut: []Faculty{{
			ID:   uuid.New(),
			Code: "IDSAI_ENU",
			Name: "IDSAI ENU",
		}},
		departmentsOut: []Department{{ID: departmentID, Code: "CPI", Name: "Computer Science"}},
		groupsOut:      []StudentGroup{{ID: uuid.New(), GroupCode: "CPI-2201", GroupNumber: 2201}},
		listOwnRequestsOut: []GroupChangeRequest{{
			ID: requestID,
		}},
		listGroupRequestsOut: []GroupChangeRequest{{
			ID: requestID,
		}},
		reviewOut: GroupChangeRequest{ID: requestID, Status: "APPROVED"},
		treeOut: []DepartmentGroupsTree{{
			ID:   departmentID,
			Code: "CPI",
			Groups: []GroupNode{{
				GroupCode: "CPI-2201",
				Students: []GroupStudent{{
					UserID:    userID,
					AvatarURL: "avatars/student.jpg",
				}},
			}},
		}},
	}
	svc := NewService(repo, Config{JWTSecret: "01234567890123456789012345678901"})
	svc.SetStorage(&authTestStorage{available: true})

	faculties, err := svc.ListFaculties(context.Background(), " core ")
	require.NoError(t, err)
	require.Equal(t, "CORE", repo.findTenantCode)
	require.Len(t, faculties, 1)

	departments, err := svc.ListDepartments(context.Background(), " core ")
	require.NoError(t, err)
	require.Equal(t, "CORE", repo.findTenantCode)
	require.Equal(t, EducationTypeUniversity, repo.listDepartmentsType)
	require.Len(t, departments, 1)

	departments, err = svc.ListDepartments(context.Background(), "core", " school ")
	require.NoError(t, err)
	require.Equal(t, EducationTypeSchool, repo.listDepartmentsType)
	require.Len(t, departments, 1)

	groups, err := svc.ListGroupsByDepartmentCode(context.Background(), "core", " cpi ")
	require.NoError(t, err)
	require.Equal(t, tenantID, repo.findDepartmentTenantID)
	require.Equal(t, "CPI", repo.findDepartmentCode)
	require.Equal(t, "CPI", repo.listGroupsCode)
	require.Len(t, groups, 1)

	ownRequests, err := svc.ListOwnGroupChangeRequests(context.Background(), tenantID, userID, 0)
	require.NoError(t, err)
	require.Len(t, ownRequests, 1)
	require.Equal(t, 50, repo.listOwnLimit)

	requests, err := svc.ListGroupChangeRequests(context.Background(), tenantID, " approved ", "  Alice  ", 999)
	require.NoError(t, err)
	require.Len(t, requests, 1)
	require.Equal(t, "APPROVED", repo.listGroupStatus)
	require.Equal(t, "Alice", repo.listGroupSearch)
	require.Equal(t, 100, repo.listGroupLimit)

	reviewed, err := svc.ReviewGroupChangeRequest(context.Background(), tenantID, reviewerID, requestID, " approve ", " looks good ")
	require.NoError(t, err)
	require.Equal(t, requestID, reviewed.ID)
	require.Equal(t, "APPROVED", repo.reviewDecision)
	require.Equal(t, "looks good", repo.reviewComment)
	require.False(t, repo.reviewAt.IsZero())

	tree, err := svc.ListDepartmentGroupsTree(context.Background(), tenantID, " cpi ", " student ", " school ")
	require.NoError(t, err)
	require.Equal(t, "CPI", repo.treeDepartmentCode)
	require.Equal(t, "student", repo.treeSearch)
	require.Equal(t, EducationTypeSchool, repo.treeEducationType)
	require.Equal(t, "https://cdn.example.com/avatars/student.jpg", tree[0].Groups[0].Students[0].AvatarURL)
}

func TestServiceStartEmailChangeConfirmCodeAndChangePassword(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()
	passwordHash, err := passwords.Hash("current-password-123")
	require.NoError(t, err)

	repo := &authRepoRecorder{
		fakeRepo: fakeRepo{
			user: User{
				ID:               userID,
				TenantID:         tenantID,
				FacultyID:        uuid.New(),
				DepartmentID:     uuid.New(),
				DepartmentCode:   "CPI",
				Email:            "student@example.edu",
				PasswordHash:     passwordHash,
				Status:           StatusActive,
				ProfileUpdatedAt: now,
			},
			resetToken: AuthTokenRecord{
				ID:        uuid.New(),
				TenantID:  tenantID,
				UserID:    userID,
				ExpiresAt: now.Add(time.Hour),
			},
		},
	}
	svc := NewService(repo, Config{
		JWTSecret:     "01234567890123456789012345678901",
		PublicBaseURL: "https://idsai.example",
	})

	err = svc.StartEmailChange(context.Background(), "127.0.0.1", tenantID, userID, " next@example.edu ")
	require.NoError(t, err)
	require.Equal(t, "next@example.edu", repo.setPendingEmail)
	require.Equal(t, tenantID, repo.setPendingTenantID)
	require.Equal(t, userID, repo.setPendingUserID)
	require.Equal(t, 1, repo.insertAuthTokenCount)
	require.Equal(t, 1, repo.invalidateTokenCount)

	confirmed, err := svc.ConfirmEmailChangeCode(context.Background(), tenantID, userID, "123456")
	require.NoError(t, err)
	require.Equal(t, repo.resetToken.ID, repo.consumeTokenID)
	require.Equal(t, 1, repo.revokeRefreshCalls)
	require.Equal(t, "next@example.edu", confirmed.Email)

	err = svc.ChangePassword(context.Background(), tenantID, userID, "current-password-123", "new-password-123")
	require.NoError(t, err)
	require.Equal(t, userID, repo.updatedPasswordUserID)
	require.NotEmpty(t, repo.updatedPasswordHash)
	require.NotEqual(t, passwordHash, repo.updatedPasswordHash)
	require.Equal(t, 2, repo.revokeRefreshCalls)
}
