package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Status       string
	FacultyID    uuid.UUID
	DepartmentID uuid.UUID
	FullName     string
}

type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash, status string) (uuid.UUID, error)
	CreateProfile(ctx context.Context, userID uuid.UUID, fullName string, facultyID, departmentID uuid.UUID) error
	GrantStudentFacultyRole(ctx context.Context, userID, facultyID uuid.UUID) error
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, userID uuid.UUID) (User, error)
	InsertRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	FindRefreshToken(ctx context.Context, tokenHash string) (userID uuid.UUID, expiresAt time.Time, revokedAt *time.Time, err error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	FindDepartment(ctx context.Context, departmentCode string) (departmentID uuid.UUID, facultyID uuid.UUID, err error)
}

type Service struct {
	repo       Repository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(repo Repository, jwtSecret string) *Service {
	return &Service{
		repo:       repo,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  15 * time.Minute,
		refreshTTL: 14 * 24 * time.Hour,
	}
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type accessClaims struct {
	FacultyID    string `json:"faculty_id"`
	DepartmentID string `json:"department_id"`
	jwt.RegisteredClaims
}

func (s *Service) RegisterStudent(ctx context.Context, email, password, fullName, departmentCode string) (Tokens, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return Tokens{}, errors.New("email/password required")
	}

	deptID, facultyID, err := s.repo.FindDepartment(ctx, departmentCode)
	if err != nil {
		return Tokens{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Tokens{}, err
	}

	userID, err := s.repo.CreateUser(ctx, email, string(hash), "ACTIVE")
	if err != nil {
		return Tokens{}, err
	}
	if err := s.repo.CreateProfile(ctx, userID, fullName, facultyID, deptID); err != nil {
		return Tokens{}, err
	}
	if err := s.repo.GrantStudentFacultyRole(ctx, userID, facultyID); err != nil {
		return Tokens{}, err
	}

	// tokens
	return s.issueTokens(ctx, userID, facultyID, deptID)
}

func (s *Service) Login(ctx context.Context, email, password string) (Tokens, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return Tokens{}, err
	}
	if u.Status != "ACTIVE" {
		return Tokens{}, errors.New("user is not active")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return Tokens{}, errors.New("invalid credentials")
	}
	return s.issueTokens(ctx, u.ID, u.FacultyID, u.DepartmentID)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, error) {
	h := hashToken(refreshToken)
	userID, exp, revokedAt, err := s.repo.FindRefreshToken(ctx, h)
	if err != nil {
		return "", err
	}
	if revokedAt != nil {
		return "", errors.New("refresh revoked")
	}
	if time.Now().After(exp) {
		return "", errors.New("refresh expired")
	}

	u, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if u.Status != "ACTIVE" {
		return "", errors.New("user is not active")
	}

	now := time.Now()
	claims := accessClaims{
		FacultyID:    u.FacultyID.String(),
		DepartmentID: u.DepartmentID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.jwtSecret)
}

func (s *Service) issueTokens(ctx context.Context, userID, facultyID, deptID uuid.UUID) (Tokens, error) {
	now := time.Now()

	claims := accessClaims{
		FacultyID:    facultyID.String(),
		DepartmentID: deptID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
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
	if err := s.repo.InsertRefreshToken(ctx, userID, refreshHash, now.Add(s.refreshTTL)); err != nil {
		return Tokens{}, err
	}

	return Tokens{AccessToken: access, RefreshToken: refreshRaw}, nil
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
