package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"idsai-core-up/internal/infra/images"
	"idsai-core-up/internal/security/passwords"
	notifsvc "idsai-core-up/internal/services/notifications"
	"math/big"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const dummyPasswordHash = "$2a$10$6zP4Lb6Tx0Jj8N4A7JwK3eVj3ljmd725LLJoPLD114F8CbnMD4Hzy"

var groupCodePattern = regexp.MustCompile(`^[A-Z]{2,8}-[0-9]{1,4}$`)
var groupNumberPattern = regexp.MustCompile(`^[0-9]{1,4}$`)

const (
	maxProfileNameLen         = 120
	maxProfileHeadlineLen     = 96
	maxProfileAboutLen        = 1200
	maxProfileRoleLen         = 80
	maxProfileSemesterLen     = 40
	maxProfileAvailabilityLen = 80
	maxProfileGoalsLen        = 600
	maxProfileLinkLen         = 240
	maxProfileStacks          = 12
	maxProfileInterests       = 16
	maxProfileTagLen          = 40
)

type User struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	Email            string
	PendingEmail     string
	PasswordHash     string
	Status           string
	FacultyID        uuid.UUID
	DepartmentID     uuid.UUID
	DepartmentCode   string
	GroupID          *uuid.UUID
	GroupCode        string
	GroupNumber      *int
	FullName         string
	Headline         string
	About            string
	PreferredRole    string
	Semester         string
	Availability     string
	Goals            string
	GithubURL        string
	Telegram         string
	PortfolioURL     string
	Stacks           []string
	Interests        []string
	ProfileUpdatedAt time.Time
	IsAdmin          bool
	IsProfessor      bool
	EmailVerifiedAt  *time.Time
	AvatarKey        string
	PendingEmailAt   *time.Time
	AvatarUpdatedAt  *time.Time
	AvatarURL        string
}

type ProfileUpdate struct {
	FullName      string
	Headline      string
	About         string
	PreferredRole string
	Semester      string
	Availability  string
	Goals         string
	GithubURL     string
	Telegram      string
	PortfolioURL  string
	Stacks        []string
	Interests     []string
}

type CreateUserParams struct {
	TenantID          uuid.UUID
	Email             string
	PasswordHash      string
	Status            string
	EmailVerifiedAt   *time.Time
	PasswordChangedAt time.Time
}

type AuthTokenRecord struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	UserID     uuid.UUID
	Purpose    string
	TokenHash  string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
}

type Session struct {
	User   User
	Tokens Tokens
}

type Department struct {
	ID        uuid.UUID
	FacultyID uuid.UUID
	Code      string
	Name      string
	ShortCode string
	CreatedAt time.Time
}

type StudentGroup struct {
	ID           uuid.UUID
	DepartmentID uuid.UUID
	GroupCode    string
	GroupNumber  int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GroupChangeRequest struct {
	ID               uuid.UUID
	StudentID        uuid.UUID
	StudentName      string
	StudentEmail     string
	CurrentGroupID   uuid.UUID
	CurrentGroupCode string
	RequestedGroupID uuid.UUID
	RequestedCode    string
	Status           string
	AdminComment     string
	CreatedAt        time.Time
	ReviewedAt       *time.Time
	ReviewedBy       *uuid.UUID
	ReviewedByName   string
}

type GroupStudent struct {
	UserID    uuid.UUID
	FullName  string
	Email     string
	AvatarURL string
	Status    string
	Role      string
}

type GroupNode struct {
	ID            uuid.UUID
	GroupCode     string
	GroupNumber   int
	TotalStudents int
	Students      []GroupStudent
}

type DepartmentGroupsTree struct {
	ID        uuid.UUID
	Code      string
	Name      string
	ShortCode string
	Groups    []GroupNode
}

type Repository interface {
	FindTenantByCode(ctx context.Context, tenantCode string) (uuid.UUID, error)
	CreateUser(ctx context.Context, in CreateUserParams) (uuid.UUID, error)
	CreateProfile(ctx context.Context, tenantID, userID uuid.UUID, fullName string, facultyID, departmentID uuid.UUID, groupID *uuid.UUID) error
	GrantStudentFacultyRole(ctx context.Context, tenantID, userID, facultyID uuid.UUID) error
	FindUserByEmail(ctx context.Context, tenantID uuid.UUID, email string) (User, error)
	FindUserByID(ctx context.Context, tenantID, userID uuid.UUID) (User, error)
	UpdateUserProfile(ctx context.Context, tenantID, userID uuid.UUID, in ProfileUpdate, updatedAt time.Time) error
	UpdateUserPasswordHash(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string, changedAt time.Time) error
	MarkUserEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verifiedAt time.Time) error
	IsEmailInUse(ctx context.Context, tenantID, excludeUserID uuid.UUID, email string) (bool, error)
	SetPendingEmail(ctx context.Context, tenantID, userID uuid.UUID, pendingEmail string, requestedAt time.Time) error
	ActivatePendingEmail(ctx context.Context, tenantID, userID uuid.UUID, activatedAt time.Time) (string, error)
	UpdateUserAvatarKey(ctx context.Context, tenantID, userID uuid.UUID, avatarKey *string, updatedAt time.Time) error
	InsertRefreshToken(ctx context.Context, tenantID, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	FindRefreshToken(ctx context.Context, tokenHash string) (tenantID uuid.UUID, userID uuid.UUID, expiresAt time.Time, revokedAt *time.Time, err error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAndReturnRefreshToken(ctx context.Context, tokenHash string) (tenantID uuid.UUID, userID uuid.UUID, expiresAt time.Time, err error)
	RevokeUserRefreshTokens(ctx context.Context, tenantID, userID uuid.UUID) error
	FindDepartment(ctx context.Context, tenantID uuid.UUID, departmentCode string) (departmentID uuid.UUID, facultyID uuid.UUID, err error)
	FindGroupByCodeInDepartment(ctx context.Context, tenantID, departmentID uuid.UUID, groupCode string) (groupID uuid.UUID, err error)
	CreateGroupInDepartment(ctx context.Context, tenantID, facultyID, departmentID uuid.UUID, groupCode string, groupNumber int) (groupID uuid.UUID, err error)
	ListDepartments(ctx context.Context, tenantID uuid.UUID) ([]Department, error)
	ListGroupsByDepartmentCode(ctx context.Context, tenantID uuid.UUID, departmentCode string) ([]StudentGroup, error)
	InsertGroupChangeRequest(ctx context.Context, tenantID, studentID, currentGroupID, requestedGroupID uuid.UUID, createdAt time.Time) (GroupChangeRequest, error)
	ListOwnGroupChangeRequests(ctx context.Context, tenantID, studentID uuid.UUID, limit int) ([]GroupChangeRequest, error)
	ListGroupChangeRequests(ctx context.Context, tenantID uuid.UUID, status, search string, limit int) ([]GroupChangeRequest, error)
	ReviewGroupChangeRequest(ctx context.Context, tenantID, requestID, reviewerID uuid.UUID, decision, comment string, reviewedAt time.Time) (GroupChangeRequest, error)
	ListDepartmentGroupsTree(ctx context.Context, tenantID uuid.UUID, departmentCode, search string) ([]DepartmentGroupsTree, error)
	InsertAuthToken(ctx context.Context, tenantID, userID uuid.UUID, purpose, tokenHash string, expiresAt time.Time) error
	FindAuthToken(ctx context.Context, purpose, tokenHash string) (AuthTokenRecord, error)
	ConsumeAuthToken(ctx context.Context, tokenID uuid.UUID, consumedAt time.Time) error
	InvalidateAuthTokens(ctx context.Context, tenantID, userID uuid.UUID, purpose string) error
}

type ObjectStorage interface {
	PutObject(ctx context.Context, key, contentType string, body []byte) error
	DeleteObject(ctx context.Context, key string) error
	PublicURL(key string) string
	Available() bool
}

type Config struct {
	JWTSecret              string
	PublicBaseURL          string
	AccessTTL              time.Duration
	RefreshTTL             time.Duration
	VerificationTTL        time.Duration
	EmailChangeTTL         time.Duration
	PasswordResetTTL       time.Duration
	AutoVerifyRegistrants  bool
	MaxFailedLoginAttempts int
	LoginAttemptWindow     time.Duration
}

type Service struct {
	repo             Repository
	jwtSecret        []byte
	publicBaseURL    string
	accessTTL        time.Duration
	refreshTTL       time.Duration
	verificationTTL  time.Duration
	emailChangeTTL   time.Duration
	passwordResetTTL time.Duration
	autoVerifyRegs   bool
	loginLimiter     *attemptLimiter
	recoveryLimiter  *attemptLimiter
	notifier         NotificationPublisher
	storage          ObjectStorage
}

type NotificationPublisher interface {
	Notify(ctx context.Context, in notifsvc.CreateInput) (notifsvc.Notification, error)
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

func NewService(repo Repository, cfg Config) *Service {
	publicBaseURL := strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	if publicBaseURL == "" {
		publicBaseURL = "http://localhost:8080"
	}
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 15 * time.Minute
	}
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = 7 * 24 * time.Hour
	}
	if cfg.VerificationTTL <= 0 {
		cfg.VerificationTTL = 24 * time.Hour
	}
	if cfg.EmailChangeTTL <= 0 {
		cfg.EmailChangeTTL = 24 * time.Hour
	}
	if cfg.PasswordResetTTL <= 0 {
		cfg.PasswordResetTTL = 30 * time.Minute
	}

	return &Service{
		repo:             repo,
		jwtSecret:        []byte(cfg.JWTSecret),
		publicBaseURL:    publicBaseURL,
		accessTTL:        cfg.AccessTTL,
		refreshTTL:       cfg.RefreshTTL,
		verificationTTL:  cfg.VerificationTTL,
		emailChangeTTL:   cfg.EmailChangeTTL,
		passwordResetTTL: cfg.PasswordResetTTL,
		autoVerifyRegs:   cfg.AutoVerifyRegistrants,
		loginLimiter:     newAttemptLimiter(cfg.LoginAttemptWindow, cfg.MaxFailedLoginAttempts),
		recoveryLimiter:  newAttemptLimiter(15*time.Minute, 5),
	}
}

func (s *Service) SetNotifier(notifier NotificationPublisher) {
	s.notifier = notifier
}

func (s *Service) SetStorage(storage ObjectStorage) {
	s.storage = storage
}

func (s *Service) AccessTTL() time.Duration {
	return s.accessTTL
}

func (s *Service) RefreshTTL() time.Duration {
	return s.refreshTTL
}

func (s *Service) PasswordResetTTL() time.Duration {
	return s.passwordResetTTL
}

func (s *Service) RegistrationRequiresVerification() bool {
	return !s.autoVerifyRegs
}

func normalizeDepartmentGroupCode(departmentCode, rawGroup string) string {
	departmentCode = strings.ToUpper(strings.TrimSpace(departmentCode))
	groupCode := strings.ToUpper(strings.TrimSpace(rawGroup))
	if departmentCode == "" || groupCode == "" {
		return groupCode
	}
	if groupNumberPattern.MatchString(groupCode) {
		return departmentCode + "-" + groupCode
	}
	return groupCode
}

func groupNumberFromCode(groupCode string) (int, error) {
	_, rawNumber, ok := strings.Cut(strings.ToUpper(strings.TrimSpace(groupCode)), "-")
	if !ok || !groupNumberPattern.MatchString(rawNumber) {
		return 0, ErrGroupMismatch
	}
	number, err := strconv.Atoi(rawNumber)
	if err != nil || number <= 0 {
		return 0, ErrGroupMismatch
	}
	return number, nil
}

func (s *Service) resolveOrCreateGroupByCode(
	ctx context.Context,
	tenantID, facultyID, departmentID uuid.UUID,
	groupCode string,
) (uuid.UUID, error) {
	groupID, err := s.repo.FindGroupByCodeInDepartment(ctx, tenantID, departmentID, groupCode)
	if err == nil {
		return groupID, nil
	}
	if !errors.Is(err, ErrGroupNotFound) {
		return uuid.Nil, err
	}

	groupNumber, err := groupNumberFromCode(groupCode)
	if err != nil {
		return uuid.Nil, err
	}
	return s.repo.CreateGroupInDepartment(ctx, tenantID, facultyID, departmentID, groupCode, groupNumber)
}

func (s *Service) RegisterStudent(ctx context.Context, tenantCode, email, password, fullName, departmentCode, groupCode string) error {
	tenantCode = normalizeTenantCode(tenantCode)
	tenantID, err := s.repo.FindTenantByCode(ctx, tenantCode)
	if err != nil {
		return err
	}

	email = normalizeEmail(email)
	fullName = strings.TrimSpace(fullName)
	departmentCode = strings.ToUpper(strings.TrimSpace(departmentCode))
	groupCode = normalizeDepartmentGroupCode(departmentCode, groupCode)
	if email == "" || departmentCode == "" || groupCode == "" {
		return ErrInvalidInput
	}
	if !groupCodePattern.MatchString(groupCode) || !strings.HasPrefix(groupCode, departmentCode+"-") {
		return ErrGroupMismatch
	}
	if err := passwords.Validate(password); err != nil {
		return err
	}

	deptID, facultyID, err := s.repo.FindDepartment(ctx, tenantID, departmentCode)
	if err != nil {
		return err
	}
	groupID, err := s.resolveOrCreateGroupByCode(ctx, tenantID, facultyID, deptID, groupCode)
	if err != nil {
		return err
	}

	hash, err := passwords.Hash(password)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	status := StatusPending
	var verifiedAt *time.Time
	if s.autoVerifyRegs {
		status = StatusActive
		verifiedAt = &now
	}
	userID, err := s.repo.CreateUser(ctx, CreateUserParams{
		TenantID:          tenantID,
		Email:             email,
		PasswordHash:      hash,
		Status:            status,
		EmailVerifiedAt:   verifiedAt,
		PasswordChangedAt: now,
	})
	if err != nil {
		return err
	}
	if err := s.repo.CreateProfile(ctx, tenantID, userID, fullName, facultyID, deptID, &groupID); err != nil {
		return err
	}
	if err := s.repo.GrantStudentFacultyRole(ctx, tenantID, userID, facultyID); err != nil {
		return err
	}
	if s.autoVerifyRegs {
		return nil
	}

	return s.issueVerification(ctx, tenantID, userID, email)
}

func (s *Service) Login(ctx context.Context, actorKey, tenantCode, email, password string) (Session, error) {
	email = normalizeEmail(email)
	tenantCode = normalizeTenantCode(tenantCode)
	limitKey := loginAttemptKey(actorKey, tenantCode, email)
	if !s.loginLimiter.Allow(limitKey) {
		return Session{}, ErrTooManyAttempts
	}

	tenantID, err := s.repo.FindTenantByCode(ctx, tenantCode)
	if err != nil {
		s.loginLimiter.Fail(limitKey)
		return Session{}, ErrInvalidCredentials
	}

	u, err := s.repo.FindUserByEmail(ctx, tenantID, email)
	if err != nil {
		_, _ = passwords.Verify(dummyPasswordHash, password)
		s.loginLimiter.Fail(limitKey)
		if errors.Is(err, ErrNotFound) {
			return Session{}, ErrInvalidCredentials
		}
		return Session{}, err
	}

	verifyResult, err := passwords.Verify(u.PasswordHash, password)
	if err != nil {
		s.loginLimiter.Fail(limitKey)
		return Session{}, err
	}
	if !verifyResult.Valid {
		s.loginLimiter.Fail(limitKey)
		return Session{}, ErrInvalidCredentials
	}

	if verifyResult.NeedsRehash {
		newHash, hashErr := passwords.Hash(password)
		if hashErr == nil {
			_ = s.repo.UpdateUserPasswordHash(ctx, u.TenantID, u.ID, newHash, time.Now().UTC())
		}
	}

	if u.EmailVerifiedAt == nil || u.Status == StatusPending {
		s.loginLimiter.Reset(limitKey)
		return Session{}, ErrEmailVerificationRequired
	}
	if u.Status != StatusActive {
		s.loginLimiter.Fail(limitKey)
		return Session{}, ErrInvalidCredentials
	}
	u.AvatarURL = s.resolveAvatarURL(u.AvatarKey)

	tokens, err := s.issueTokens(ctx, u.TenantID, u.ID, u.FacultyID, u.DepartmentID, u.IsAdmin, u.IsProfessor)
	if err != nil {
		return Session{}, err
	}

	s.loginLimiter.Reset(limitKey)
	return Session{
		User:   u,
		Tokens: tokens,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (Session, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return Session{}, ErrInvalidInput
	}

	h := hashToken(refreshToken)

	// Atomic revoke-and-return: prevents TOCTOU race condition
	// where two concurrent requests with the same token could both succeed.
	tenantID, userID, exp, err := s.repo.RevokeAndReturnRefreshToken(ctx, h)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Session{}, ErrSessionInvalid
		}
		return Session{}, err
	}
	if time.Now().UTC().After(exp) {
		return Session{}, ErrSessionExpired
	}

	u, err := s.repo.FindUserByID(ctx, tenantID, userID)
	if err != nil {
		return Session{}, err
	}
	if u.Status != StatusActive || u.EmailVerifiedAt == nil {
		return Session{}, ErrSessionInvalid
	}
	u.AvatarURL = s.resolveAvatarURL(u.AvatarKey)

	tokens, err := s.issueTokens(ctx, u.TenantID, u.ID, u.FacultyID, u.DepartmentID, u.IsAdmin, u.IsProfessor)
	if err != nil {
		return Session{}, err
	}

	return Session{
		User:   u,
		Tokens: tokens,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil
	}
	if err := s.repo.RevokeRefreshToken(ctx, hashToken(refreshToken)); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return nil
}

func (s *Service) Me(ctx context.Context, tenantID, userID uuid.UUID) (User, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return User{}, ErrInvalidInput
	}
	u, err := s.repo.FindUserByID(ctx, tenantID, userID)
	if err != nil {
		return User{}, err
	}
	u.AvatarURL = s.resolveAvatarURL(u.AvatarKey)
	return u, nil
}

func (s *Service) UpdateProfile(ctx context.Context, tenantID, userID uuid.UUID, in ProfileUpdate) (User, error) {
	payload, err := normalizeProfileUpdate(in)
	if err != nil {
		return User{}, err
	}
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return User{}, ErrInvalidInput
	}
	if err := s.repo.UpdateUserProfile(ctx, tenantID, userID, payload, time.Now().UTC()); err != nil {
		return User{}, err
	}
	return s.Me(ctx, tenantID, userID)
}

func normalizeProfileUpdate(in ProfileUpdate) (ProfileUpdate, error) {
	out := ProfileUpdate{
		FullName:      strings.TrimSpace(in.FullName),
		Headline:      strings.TrimSpace(in.Headline),
		About:         strings.TrimSpace(in.About),
		PreferredRole: strings.TrimSpace(in.PreferredRole),
		Semester:      strings.TrimSpace(in.Semester),
		Availability:  strings.TrimSpace(in.Availability),
		Goals:         strings.TrimSpace(in.Goals),
		GithubURL:     strings.TrimSpace(in.GithubURL),
		Telegram:      strings.TrimSpace(in.Telegram),
		PortfolioURL:  strings.TrimSpace(in.PortfolioURL),
		Stacks:        normalizeProfileTags(in.Stacks, maxProfileStacks),
		Interests:     normalizeProfileTags(in.Interests, maxProfileInterests),
	}

	switch {
	case len(out.FullName) < 2 || len(out.FullName) > maxProfileNameLen:
		return ProfileUpdate{}, ErrInvalidInput
	case len(out.Headline) > maxProfileHeadlineLen:
		return ProfileUpdate{}, ErrInvalidInput
	case len(out.About) > maxProfileAboutLen:
		return ProfileUpdate{}, ErrInvalidInput
	case len(out.PreferredRole) > maxProfileRoleLen:
		return ProfileUpdate{}, ErrInvalidInput
	case len(out.Semester) > maxProfileSemesterLen:
		return ProfileUpdate{}, ErrInvalidInput
	case len(out.Availability) > maxProfileAvailabilityLen:
		return ProfileUpdate{}, ErrInvalidInput
	case len(out.Goals) > maxProfileGoalsLen:
		return ProfileUpdate{}, ErrInvalidInput
	case len(out.GithubURL) > maxProfileLinkLen:
		return ProfileUpdate{}, ErrInvalidInput
	case len(out.Telegram) > maxProfileLinkLen:
		return ProfileUpdate{}, ErrInvalidInput
	case len(out.PortfolioURL) > maxProfileLinkLen:
		return ProfileUpdate{}, ErrInvalidInput
	}

	return out, nil
}

func normalizeProfileTags(items []string, limit int) []string {
	if len(items) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, min(limit, len(items)))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" || len(value) > maxProfileTagLen {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func (s *Service) ListDepartments(ctx context.Context, tenantCode string) ([]Department, error) {
	tenantID, err := s.repo.FindTenantByCode(ctx, normalizeTenantCode(tenantCode))
	if err != nil {
		return nil, err
	}
	return s.repo.ListDepartments(ctx, tenantID)
}

func (s *Service) ListGroupsByDepartmentCode(ctx context.Context, tenantCode, departmentCode string) ([]StudentGroup, error) {
	departmentCode = strings.ToUpper(strings.TrimSpace(departmentCode))
	if departmentCode == "" {
		return nil, ErrInvalidInput
	}
	tenantID, err := s.repo.FindTenantByCode(ctx, normalizeTenantCode(tenantCode))
	if err != nil {
		return nil, err
	}
	if _, _, err := s.repo.FindDepartment(ctx, tenantID, departmentCode); err != nil {
		return nil, err
	}
	return s.repo.ListGroupsByDepartmentCode(ctx, tenantID, departmentCode)
}

func (s *Service) SubmitGroupChangeRequest(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	departmentCode, requestedGroupCode string,
) (GroupChangeRequest, error) {
	departmentCode = strings.ToUpper(strings.TrimSpace(departmentCode))
	requestedGroupCode = normalizeDepartmentGroupCode(departmentCode, requestedGroupCode)
	if tenantID == uuid.Nil || userID == uuid.Nil || departmentCode == "" || requestedGroupCode == "" {
		return GroupChangeRequest{}, ErrInvalidInput
	}
	if !groupCodePattern.MatchString(requestedGroupCode) || !strings.HasPrefix(requestedGroupCode, departmentCode+"-") {
		return GroupChangeRequest{}, ErrGroupMismatch
	}

	u, err := s.repo.FindUserByID(ctx, tenantID, userID)
	if err != nil {
		return GroupChangeRequest{}, err
	}
	if u.GroupID == nil || *u.GroupID == uuid.Nil {
		return GroupChangeRequest{}, ErrInvalidInput
	}

	targetDeptID, targetFacultyID, err := s.repo.FindDepartment(ctx, tenantID, departmentCode)
	if err != nil {
		return GroupChangeRequest{}, err
	}
	targetGroupID, err := s.resolveOrCreateGroupByCode(ctx, tenantID, targetFacultyID, targetDeptID, requestedGroupCode)
	if err != nil {
		return GroupChangeRequest{}, err
	}
	if targetGroupID == *u.GroupID {
		return GroupChangeRequest{}, ErrGroupUnchanged
	}

	return s.repo.InsertGroupChangeRequest(
		ctx,
		tenantID,
		userID,
		*u.GroupID,
		targetGroupID,
		time.Now().UTC(),
	)
}

func (s *Service) ListOwnGroupChangeRequests(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]GroupChangeRequest, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListOwnGroupChangeRequests(ctx, tenantID, userID, limit)
}

func (s *Service) ListGroupChangeRequests(ctx context.Context, tenantID uuid.UUID, status, search string, limit int) ([]GroupChangeRequest, error) {
	if tenantID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	switch status {
	case "", "PENDING", "APPROVED", "REJECTED":
	default:
		return nil, ErrInvalidInput
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.ListGroupChangeRequests(ctx, tenantID, status, strings.TrimSpace(search), limit)
}

func (s *Service) ReviewGroupChangeRequest(
	ctx context.Context,
	tenantID, reviewerID, requestID uuid.UUID,
	decision, comment string,
) (GroupChangeRequest, error) {
	if tenantID == uuid.Nil || reviewerID == uuid.Nil || requestID == uuid.Nil {
		return GroupChangeRequest{}, ErrInvalidInput
	}
	decision = strings.ToUpper(strings.TrimSpace(decision))
	switch decision {
	case "APPROVE":
		decision = "APPROVED"
	case "REJECT":
		decision = "REJECTED"
	case "APPROVED", "REJECTED":
	default:
		return GroupChangeRequest{}, ErrInvalidInput
	}
	return s.repo.ReviewGroupChangeRequest(ctx, tenantID, requestID, reviewerID, decision, strings.TrimSpace(comment), time.Now().UTC())
}

func (s *Service) ListDepartmentGroupsTree(
	ctx context.Context,
	tenantID uuid.UUID,
	departmentCode, search string,
) ([]DepartmentGroupsTree, error) {
	if tenantID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	tree, err := s.repo.ListDepartmentGroupsTree(ctx, tenantID, strings.ToUpper(strings.TrimSpace(departmentCode)), strings.TrimSpace(search))
	if err != nil {
		return nil, err
	}
	for depIdx := range tree {
		for groupIdx := range tree[depIdx].Groups {
			for studentIdx := range tree[depIdx].Groups[groupIdx].Students {
				avatarURL := s.resolveAvatarURL(tree[depIdx].Groups[groupIdx].Students[studentIdx].AvatarURL)
				tree[depIdx].Groups[groupIdx].Students[studentIdx].AvatarURL = avatarURL
			}
		}
	}
	return tree, nil
}

func (s *Service) StartEmailChange(ctx context.Context, actorKey string, tenantID, userID uuid.UUID, nextEmail string) error {
	nextEmail = normalizeEmail(nextEmail)
	if tenantID == uuid.Nil || userID == uuid.Nil || nextEmail == "" {
		return ErrInvalidInput
	}

	user, err := s.repo.FindUserByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(user.Email), nextEmail) {
		return ErrInvalidInput
	}

	limitKey := recoveryAttemptKey("email-change", actorKey, tenantID.String(), nextEmail)
	if !s.recoveryLimiter.Allow(limitKey) {
		return ErrTooManyAttempts
	}
	defer s.recoveryLimiter.Fail(limitKey)

	inUse, err := s.repo.IsEmailInUse(ctx, tenantID, userID, nextEmail)
	if err != nil {
		return err
	}
	if inUse {
		return ErrEmailInUse
	}

	now := time.Now().UTC()
	if err := s.repo.SetPendingEmail(ctx, tenantID, userID, nextEmail, now); err != nil {
		return err
	}
	return s.issueEmailChangeToken(ctx, tenantID, userID, nextEmail)
}

func (s *Service) ResendEmailChange(ctx context.Context, actorKey string, tenantID, userID uuid.UUID) error {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return ErrInvalidInput
	}
	user, err := s.repo.FindUserByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	pendingEmail := normalizeEmail(user.PendingEmail)
	if pendingEmail == "" {
		return ErrNoPendingEmail
	}

	limitKey := recoveryAttemptKey("email-change-resend", actorKey, tenantID.String(), pendingEmail)
	if !s.recoveryLimiter.Allow(limitKey) {
		return ErrTooManyAttempts
	}
	defer s.recoveryLimiter.Fail(limitKey)

	return s.issueEmailChangeToken(ctx, tenantID, userID, pendingEmail)
}

func (s *Service) ConfirmEmailChange(ctx context.Context, rawToken string) (User, error) {
	record, err := s.findUsableAuthToken(ctx, TokenPurposeEmailChange, rawToken)
	if err != nil {
		return User{}, err
	}

	now := time.Now().UTC()
	if _, err := s.repo.ActivatePendingEmail(ctx, record.TenantID, record.UserID, now); err != nil {
		return User{}, err
	}
	if err := s.repo.ConsumeAuthToken(ctx, record.ID, now); err != nil {
		return User{}, err
	}
	if err := s.repo.RevokeUserRefreshTokens(ctx, record.TenantID, record.UserID); err != nil {
		return User{}, err
	}
	return s.Me(ctx, record.TenantID, record.UserID)
}

func (s *Service) ConfirmEmailChangeCode(ctx context.Context, tenantID, userID uuid.UUID, code string) (User, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil || !isVerificationCode(code) {
		return User{}, ErrInvalidInput
	}
	return s.ConfirmEmailChange(ctx, scopedCodeToken(userID, code))
}

func (s *Service) ChangePassword(ctx context.Context, tenantID, userID uuid.UUID, currentPassword, newPassword string) error {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return ErrInvalidInput
	}
	if err := passwords.Validate(newPassword); err != nil {
		return err
	}

	u, err := s.repo.FindUserByID(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	verifyResult, err := passwords.Verify(u.PasswordHash, currentPassword)
	if err != nil {
		return err
	}
	if !verifyResult.Valid {
		return ErrInvalidCurrentPassword
	}

	hash, err := passwords.Hash(newPassword)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := s.repo.UpdateUserPasswordHash(ctx, tenantID, userID, hash, now); err != nil {
		return err
	}
	if err := s.repo.RevokeUserRefreshTokens(ctx, tenantID, userID); err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateAvatar(ctx context.Context, tenantID, userID uuid.UUID, raw []byte) (User, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil || len(raw) == 0 {
		return User{}, ErrInvalidInput
	}
	if s.storage == nil || !s.storage.Available() {
		return User{}, ErrStorageUnavailable
	}

	processed, contentType, err := images.ProcessAvatar(raw)
	if err != nil {
		return User{}, err
	}

	current, err := s.repo.FindUserByID(ctx, tenantID, userID)
	if err != nil {
		return User{}, err
	}

	now := time.Now().UTC()
	key := fmt.Sprintf("users/avatars/%s/%d-%s.jpg", userID.String(), now.Unix(), uuid.NewString())
	if err := s.storage.PutObject(ctx, key, contentType, processed); err != nil {
		return User{}, ErrStorageUnavailable
	}
	if err := s.repo.UpdateUserAvatarKey(ctx, tenantID, userID, &key, now); err != nil {
		_ = s.storage.DeleteObject(ctx, key)
		return User{}, err
	}

	if oldKey := strings.TrimSpace(current.AvatarKey); oldKey != "" && oldKey != key {
		_ = s.storage.DeleteObject(ctx, oldKey)
	}
	return s.Me(ctx, tenantID, userID)
}

func (s *Service) DeleteAvatar(ctx context.Context, tenantID, userID uuid.UUID) (User, error) {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return User{}, ErrInvalidInput
	}
	current, err := s.repo.FindUserByID(ctx, tenantID, userID)
	if err != nil {
		return User{}, err
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateUserAvatarKey(ctx, tenantID, userID, nil, now); err != nil {
		return User{}, err
	}
	if s.storage != nil && s.storage.Available() {
		if oldKey := strings.TrimSpace(current.AvatarKey); oldKey != "" {
			_ = s.storage.DeleteObject(ctx, oldKey)
		}
	}
	return s.Me(ctx, tenantID, userID)
}

func (s *Service) ResendVerification(ctx context.Context, actorKey, tenantCode, email string) error {
	email = normalizeEmail(email)
	tenantCode = normalizeTenantCode(tenantCode)
	limitKey := recoveryAttemptKey("verify", actorKey, tenantCode, email)
	if !s.recoveryLimiter.Allow(limitKey) {
		return ErrTooManyAttempts
	}

	defer s.recoveryLimiter.Fail(limitKey)

	tenantID, err := s.repo.FindTenantByCode(ctx, tenantCode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	u, err := s.repo.FindUserByEmail(ctx, tenantID, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	if u.EmailVerifiedAt != nil || u.Status == StatusDisabled {
		return nil
	}

	return s.issueVerification(ctx, tenantID, u.ID, u.Email)
}

func (s *Service) RequestPasswordReset(ctx context.Context, actorKey, tenantCode, email string) error {
	email = normalizeEmail(email)
	tenantCode = normalizeTenantCode(tenantCode)
	limitKey := recoveryAttemptKey("reset", actorKey, tenantCode, email)
	if !s.recoveryLimiter.Allow(limitKey) {
		return ErrTooManyAttempts
	}

	defer s.recoveryLimiter.Fail(limitKey)

	tenantID, err := s.repo.FindTenantByCode(ctx, tenantCode)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrPasswordResetUnavailable
		}
		return err
	}

	u, err := s.repo.FindUserByEmail(ctx, tenantID, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrPasswordResetUnavailable
		}
		return err
	}
	if u.Status != StatusActive || u.EmailVerifiedAt == nil {
		return ErrPasswordResetUnavailable
	}

	if err := s.repo.InvalidateAuthTokens(ctx, u.TenantID, u.ID, TokenPurposePasswordReset); err != nil {
		return err
	}

	code, err := randomNumericCode(6)
	if err != nil {
		return err
	}
	raw := scopedCodeToken(u.ID, code)

	expiresAt := time.Now().UTC().Add(s.passwordResetTTL)
	if err := s.repo.InsertAuthToken(ctx, u.TenantID, u.ID, TokenPurposePasswordReset, hashToken(raw), expiresAt); err != nil {
		return err
	}

	if s.notifier != nil {
		link := s.publicBaseURL + "/v2/auth/password-reset?token=" + url.QueryEscape(raw)
		_, _ = s.notifier.Notify(ctx, notifsvc.CreateInput{
			TenantID:  u.TenantID,
			UserID:    u.ID,
			Type:      "account.password_reset",
			Title:     "Сброс пароля",
			Body:      "Мы получили запрос на сброс пароля. Используйте код подтверждения из письма, чтобы задать новый пароль.",
			Payload:   map[string]any{"action": "password_reset"},
			WithEmail: true,
			EmailTo:   u.Email,
			EmailSubj: "IDSAI: сброс пароля",
			EmailBody: fmt.Sprintf(
				"Код подтверждения для сброса пароля: %s\nКод действует %d минут.\n\nЕсли удобнее, можно открыть ссылку: %s",
				code,
				int(s.passwordResetTTL.Minutes()),
				link,
			),
		})
	}

	return nil
}

func (s *Service) ValidatePasswordResetToken(ctx context.Context, rawToken string) (time.Duration, error) {
	record, err := s.findUsableAuthToken(ctx, TokenPurposePasswordReset, rawToken)
	if err != nil {
		return 0, err
	}
	return time.Until(record.ExpiresAt), nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if err := passwords.Validate(newPassword); err != nil {
		return err
	}

	record, err := s.findUsableAuthToken(ctx, TokenPurposePasswordReset, rawToken)
	if err != nil {
		return err
	}

	hash, err := passwords.Hash(newPassword)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateUserPasswordHash(ctx, record.TenantID, record.UserID, hash, now); err != nil {
		return err
	}
	if err := s.repo.RevokeUserRefreshTokens(ctx, record.TenantID, record.UserID); err != nil {
		return err
	}
	if err := s.repo.ConsumeAuthToken(ctx, record.ID, now); err != nil {
		return err
	}
	return nil
}

func (s *Service) ResetPasswordByCode(ctx context.Context, tenantCode, email, code, newPassword string) error {
	code = strings.TrimSpace(code)
	if !isVerificationCode(code) {
		return ErrInvalidInput
	}

	tenantID, err := s.repo.FindTenantByCode(ctx, normalizeTenantCode(tenantCode))
	if err != nil {
		return ErrTokenInvalid
	}
	u, err := s.repo.FindUserByEmail(ctx, tenantID, normalizeEmail(email))
	if err != nil {
		return ErrTokenInvalid
	}

	return s.ResetPassword(ctx, scopedCodeToken(u.ID, code), newPassword)
}

func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	record, err := s.findUsableAuthToken(ctx, TokenPurposeEmailVerification, rawToken)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if err := s.repo.MarkUserEmailVerified(ctx, record.TenantID, record.UserID, now); err != nil {
		return err
	}
	if err := s.repo.ConsumeAuthToken(ctx, record.ID, now); err != nil {
		return err
	}
	return nil
}

func (s *Service) issueVerification(ctx context.Context, tenantID, userID uuid.UUID, email string) error {
	if err := s.repo.InvalidateAuthTokens(ctx, tenantID, userID, TokenPurposeEmailVerification); err != nil {
		return err
	}

	raw, err := randomToken(32)
	if err != nil {
		return err
	}

	expiresAt := time.Now().UTC().Add(s.verificationTTL)
	if err := s.repo.InsertAuthToken(ctx, tenantID, userID, TokenPurposeEmailVerification, hashToken(raw), expiresAt); err != nil {
		return err
	}

	if s.notifier != nil {
		link := s.publicBaseURL + "/v2/auth/verify-email?token=" + url.QueryEscape(raw)
		_, _ = s.notifier.Notify(ctx, notifsvc.CreateInput{
			TenantID:  tenantID,
			UserID:    userID,
			Type:      "account.verify_email",
			Title:     "Подтвердите email",
			Body:      "Подтвердите email, чтобы активировать аккаунт и войти в систему.",
			Payload:   map[string]any{"action": "verify_email"},
			WithEmail: true,
			EmailTo:   email,
			EmailSubj: "IDSAI: подтверждение email",
			EmailBody: fmt.Sprintf("Подтвердите email по ссылке: %s\nСсылка действует %d часов.", link, int(s.verificationTTL.Hours())),
		})
	}

	return nil
}

func (s *Service) issueEmailChangeToken(ctx context.Context, tenantID, userID uuid.UUID, pendingEmail string) error {
	if err := s.repo.InvalidateAuthTokens(ctx, tenantID, userID, TokenPurposeEmailChange); err != nil {
		return err
	}

	code, err := randomNumericCode(6)
	if err != nil {
		return err
	}
	raw := scopedCodeToken(userID, code)

	expiresAt := time.Now().UTC().Add(s.emailChangeTTL)
	if err := s.repo.InsertAuthToken(ctx, tenantID, userID, TokenPurposeEmailChange, hashToken(raw), expiresAt); err != nil {
		return err
	}

	if s.notifier != nil {
		link := s.publicBaseURL + "/v2/auth/settings/email/verify?token=" + url.QueryEscape(raw)
		_, _ = s.notifier.Notify(ctx, notifsvc.CreateInput{
			TenantID:  tenantID,
			UserID:    userID,
			Type:      "account.email_change",
			Title:     "Подтверждение нового email",
			Body:      "Подтвердите новый email кодом из письма, чтобы завершить изменение адреса для входа.",
			Payload:   map[string]any{"action": "email_change"},
			WithEmail: true,
			EmailTo:   pendingEmail,
			EmailSubj: "IDSAI: подтверждение нового email",
			EmailBody: fmt.Sprintf(
				"Код подтверждения нового email: %s\nКод действует %d часов.\n\nЕсли удобнее, можно открыть ссылку: %s",
				code,
				int(s.emailChangeTTL.Hours()),
				link,
			),
		})
	}

	return nil
}

func (s *Service) findUsableAuthToken(ctx context.Context, purpose, rawToken string) (AuthTokenRecord, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return AuthTokenRecord{}, ErrInvalidInput
	}

	record, err := s.repo.FindAuthToken(ctx, purpose, hashToken(rawToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return AuthTokenRecord{}, ErrTokenInvalid
		}
		return AuthTokenRecord{}, err
	}
	if record.ConsumedAt != nil {
		return AuthTokenRecord{}, ErrTokenInvalid
	}
	if time.Now().UTC().After(record.ExpiresAt) {
		return AuthTokenRecord{}, ErrTokenExpired
	}
	return record, nil
}

func (s *Service) issueTokens(ctx context.Context, tenantID, userID, facultyID, deptID uuid.UUID, isAdmin bool, isProfessor bool) (Tokens, error) {
	now := time.Now().UTC()

	claims := jwt.MapClaims{
		"tenant_id":     tenantID.String(),
		"faculty_id":    facultyID.String(),
		"department_id": deptID.String(),
		"is_admin":      isAdmin,
		"is_professor":  isProfessor,
		"sub":           userID.String(),
		"iat":           now.Unix(),
		"exp":           now.Add(s.accessTTL).Unix(),
		"iss":           TokenIssuer,
		"jti":           uuid.NewString(),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	access, err := tok.SignedString(s.jwtSecret)
	if err != nil {
		return Tokens{}, err
	}

	refreshRaw, err := randomToken(32)
	if err != nil {
		return Tokens{}, err
	}
	refreshHash := hashToken(refreshRaw)
	if err := s.repo.InsertRefreshToken(ctx, tenantID, userID, refreshHash, now.Add(s.refreshTTL)); err != nil {
		return Tokens{}, err
	}

	return Tokens{AccessToken: access, RefreshToken: refreshRaw}, nil
}

func normalizeTenantCode(tenantCode string) string {
	tenantCode = strings.ToUpper(strings.TrimSpace(tenantCode))
	if tenantCode == "" {
		return "CORE"
	}
	return tenantCode
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func loginAttemptKey(actorKey, tenantCode, email string) string {
	return strings.Join([]string{"login", strings.TrimSpace(actorKey), tenantCode, email}, "|")
}

func recoveryAttemptKey(kind, actorKey, tenantCode, email string) string {
	return strings.Join([]string{kind, strings.TrimSpace(actorKey), tenantCode, email}, "|")
}

func (s *Service) resolveAvatarURL(key string) string {
	key = strings.TrimSpace(key)
	if key == "" || s.storage == nil {
		return ""
	}
	return strings.TrimSpace(s.storage.PublicURL(key))
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randomNumericCode(digits int) (string, error) {
	if digits <= 0 || digits > 18 {
		return "", ErrInvalidInput
	}
	max := big.NewInt(1)
	for i := 0; i < digits; i++ {
		max.Mul(max, big.NewInt(10))
	}
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", digits, n.Int64()), nil
}

func scopedCodeToken(userID uuid.UUID, code string) string {
	return userID.String() + "." + strings.TrimSpace(code)
}

func isVerificationCode(code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
