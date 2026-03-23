package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"idsai-core-up/internal/security/passwords"
	notifsvc "idsai-core-up/internal/services/notifications"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const dummyPasswordHash = "$2a$10$6zP4Lb6Tx0Jj8N4A7JwK3eVj3ljmd725LLJoPLD114F8CbnMD4Hzy"

type User struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	Email           string
	PasswordHash    string
	Status          string
	FacultyID       uuid.UUID
	DepartmentID    uuid.UUID
	FullName        string
	IsAdmin         bool
	IsProfessor     bool
	EmailVerifiedAt *time.Time
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

type Repository interface {
	FindTenantByCode(ctx context.Context, tenantCode string) (uuid.UUID, error)
	CreateUser(ctx context.Context, in CreateUserParams) (uuid.UUID, error)
	CreateProfile(ctx context.Context, tenantID, userID uuid.UUID, fullName string, facultyID, departmentID uuid.UUID) error
	GrantStudentFacultyRole(ctx context.Context, tenantID, userID, facultyID uuid.UUID) error
	FindUserByEmail(ctx context.Context, tenantID uuid.UUID, email string) (User, error)
	FindUserByID(ctx context.Context, tenantID, userID uuid.UUID) (User, error)
	UpdateUserPasswordHash(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string, changedAt time.Time) error
	MarkUserEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verifiedAt time.Time) error
	InsertRefreshToken(ctx context.Context, tenantID, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	FindRefreshToken(ctx context.Context, tokenHash string) (tenantID uuid.UUID, userID uuid.UUID, expiresAt time.Time, revokedAt *time.Time, err error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeUserRefreshTokens(ctx context.Context, tenantID, userID uuid.UUID) error
	FindDepartment(ctx context.Context, tenantID uuid.UUID, departmentCode string) (departmentID uuid.UUID, facultyID uuid.UUID, err error)
	InsertAuthToken(ctx context.Context, tenantID, userID uuid.UUID, purpose, tokenHash string, expiresAt time.Time) error
	FindAuthToken(ctx context.Context, purpose, tokenHash string) (AuthTokenRecord, error)
	ConsumeAuthToken(ctx context.Context, tokenID uuid.UUID, consumedAt time.Time) error
	InvalidateAuthTokens(ctx context.Context, tenantID, userID uuid.UUID, purpose string) error
}

type Config struct {
	JWTSecret              string
	PublicBaseURL          string
	AccessTTL              time.Duration
	RefreshTTL             time.Duration
	VerificationTTL        time.Duration
	PasswordResetTTL       time.Duration
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
	passwordResetTTL time.Duration
	loginLimiter     *attemptLimiter
	recoveryLimiter  *attemptLimiter
	notifier         NotificationPublisher
}

type NotificationPublisher interface {
	Notify(ctx context.Context, in notifsvc.CreateInput) (notifsvc.Notification, error)
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AccessClaims struct {
	TenantID     string `json:"tenant_id"`
	FacultyID    string `json:"faculty_id"`
	DepartmentID string `json:"department_id"`
	IsAdmin      bool   `json:"is_admin"`
	IsProfessor  bool   `json:"is_professor"`
	jwt.RegisteredClaims
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
		passwordResetTTL: cfg.PasswordResetTTL,
		loginLimiter:     newAttemptLimiter(cfg.LoginAttemptWindow, cfg.MaxFailedLoginAttempts),
		recoveryLimiter:  newAttemptLimiter(15*time.Minute, 5),
	}
}

func (s *Service) SetNotifier(notifier NotificationPublisher) {
	s.notifier = notifier
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

func (s *Service) RegisterStudent(ctx context.Context, tenantCode, email, password, fullName, departmentCode string) error {
	tenantCode = normalizeTenantCode(tenantCode)
	tenantID, err := s.repo.FindTenantByCode(ctx, tenantCode)
	if err != nil {
		return err
	}

	email = normalizeEmail(email)
	fullName = strings.TrimSpace(fullName)
	departmentCode = strings.ToUpper(strings.TrimSpace(departmentCode))
	if email == "" || departmentCode == "" {
		return ErrInvalidInput
	}
	if err := passwords.Validate(password); err != nil {
		return err
	}

	deptID, facultyID, err := s.repo.FindDepartment(ctx, tenantID, departmentCode)
	if err != nil {
		return err
	}

	hash, err := passwords.Hash(password)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	userID, err := s.repo.CreateUser(ctx, CreateUserParams{
		TenantID:          tenantID,
		Email:             email,
		PasswordHash:      hash,
		Status:            StatusPending,
		EmailVerifiedAt:   nil,
		PasswordChangedAt: now,
	})
	if err != nil {
		return err
	}
	if err := s.repo.CreateProfile(ctx, tenantID, userID, fullName, facultyID, deptID); err != nil {
		return err
	}
	if err := s.repo.GrantStudentFacultyRole(ctx, tenantID, userID, facultyID); err != nil {
		return err
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
	tenantID, userID, exp, revokedAt, err := s.repo.FindRefreshToken(ctx, h)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Session{}, ErrSessionInvalid
		}
		return Session{}, err
	}
	if revokedAt != nil {
		return Session{}, ErrSessionInvalid
	}
	if time.Now().UTC().After(exp) {
		_ = s.repo.RevokeRefreshToken(ctx, h)
		return Session{}, ErrSessionExpired
	}

	u, err := s.repo.FindUserByID(ctx, tenantID, userID)
	if err != nil {
		return Session{}, err
	}
	if u.Status != StatusActive || u.EmailVerifiedAt == nil {
		_ = s.repo.RevokeRefreshToken(ctx, h)
		return Session{}, ErrSessionInvalid
	}

	if err := s.repo.RevokeRefreshToken(ctx, h); err != nil {
		return Session{}, err
	}

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
	return s.repo.FindUserByID(ctx, tenantID, userID)
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
		return nil
	}
	u, err := s.repo.FindUserByEmail(ctx, tenantID, email)
	if err != nil {
		return nil
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
		return nil
	}

	u, err := s.repo.FindUserByEmail(ctx, tenantID, email)
	if err != nil {
		return nil
	}
	if u.Status != StatusActive || u.EmailVerifiedAt == nil {
		return nil
	}

	if err := s.repo.InvalidateAuthTokens(ctx, u.TenantID, u.ID, TokenPurposePasswordReset); err != nil {
		return err
	}

	raw, err := randomToken(32)
	if err != nil {
		return err
	}

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
			Body:      "Мы получили запрос на сброс пароля. Перейдите по ссылке из письма, чтобы задать новый пароль.",
			Payload:   map[string]any{"action": "password_reset"},
			WithEmail: true,
			EmailTo:   u.Email,
			EmailSubj: "IDSAI: сброс пароля",
			EmailBody: fmt.Sprintf("Чтобы задать новый пароль, перейдите по ссылке: %s\nСсылка действует %d минут.", link, int(s.passwordResetTTL.Minutes())),
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

	claims := AccessClaims{
		TenantID:     tenantID.String(),
		FacultyID:    facultyID.String(),
		DepartmentID: deptID.String(),
		IsAdmin:      isAdmin,
		IsProfessor:  isProfessor,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			Issuer:    TokenIssuer,
			ID:        uuid.NewString(),
		},
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

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
