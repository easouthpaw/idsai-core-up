package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleStudent   = "STUDENT"
	RoleProfessor = "PROFESSOR"

	StatusActive   = "ACTIVE"
	StatusPending  = "PENDING"
	StatusDisabled = "DISABLED"
)

var (
	ErrInvalidRole          = errors.New("role must be STUDENT or PROFESSOR")
	ErrInvalidStatus        = errors.New("status must be ACTIVE, PENDING or DISABLED")
	ErrInvalidProjectStatus = errors.New("project status must be DRAFT, REVIEW, RECRUITMENT, ACTIVE, GRADING or ARCHIVE")
	ErrInvalidInput         = errors.New("invalid input")
)

type User struct {
	ID             uuid.UUID `json:"id"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	RoleCode       string    `json:"role_code"`
	Status         string    `json:"status"`
	FacultyCode    string    `json:"faculty_code"`
	DepartmentCode string    `json:"department_code"`
}

type Project struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	Visibility     string    `json:"visibility"`
	IsPublic       bool      `json:"is_public"`
	CreatedBy      uuid.UUID `json:"created_by"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	FacultyCode    string    `json:"faculty_code"`
	DepartmentCode string    `json:"department_code"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateUserInput struct {
	Email          string
	Password       string
	FullName       string
	DepartmentCode string
	RoleCode       string
}

type CreateUserParams struct {
	Email          string
	PasswordHash   string
	FullName       string
	DepartmentCode string
	RoleCode       string
}

type Repository interface {
	ListUsers(ctx context.Context, roleCode, search string) ([]User, error)
	ListProjects(ctx context.Context, status, search string) ([]Project, error)
	CreateUser(ctx context.Context, in CreateUserParams) (User, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
	UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUsers(ctx context.Context, roleCode, search string) ([]User, error) {
	role, err := normalizeRoleFilter(roleCode)
	if err != nil {
		return nil, err
	}
	return s.repo.ListUsers(ctx, role, strings.TrimSpace(search))
}

func (s *Service) ListProjects(ctx context.Context, status, search string) ([]Project, error) {
	projectStatus, err := normalizeProjectStatusFilter(status)
	if err != nil {
		return nil, err
	}
	return s.repo.ListProjects(ctx, projectStatus, strings.TrimSpace(search))
}

func (s *Service) CreateUser(ctx context.Context, in CreateUserInput) (User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	password := strings.TrimSpace(in.Password)
	fullName := strings.TrimSpace(in.FullName)
	departmentCode := strings.ToUpper(strings.TrimSpace(in.DepartmentCode))
	role, err := normalizeRole(in.RoleCode)
	if err != nil {
		return User{}, err
	}
	if email == "" || password == "" || departmentCode == "" || fullName == "" {
		return User{}, ErrInvalidInput
	}
	if len(password) < 6 {
		return User{}, ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	return s.repo.CreateUser(ctx, CreateUserParams{
		Email:          email,
		PasswordHash:   string(hash),
		FullName:       fullName,
		DepartmentCode: departmentCode,
		RoleCode:       role,
	})
}

func (s *Service) SetUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	normalized, err := normalizeStatus(status)
	if err != nil {
		return err
	}
	return s.repo.UpdateUserStatus(ctx, userID, normalized)
}

func (s *Service) SetProjectStatus(ctx context.Context, projectID uuid.UUID, status string) error {
	normalized, err := normalizeProjectStatus(status)
	if err != nil {
		return err
	}
	return s.repo.UpdateProjectStatus(ctx, projectID, normalized)
}

func normalizeRole(role string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case RoleStudent:
		return RoleStudent, nil
	case RoleProfessor:
		return RoleProfessor, nil
	default:
		return "", ErrInvalidRole
	}
}

func normalizeRoleFilter(role string) (string, error) {
	role = strings.ToUpper(strings.TrimSpace(role))
	if role == "" {
		return "", nil
	}
	if role == "SUPER_ADMIN" || role == RoleStudent || role == RoleProfessor {
		return role, nil
	}
	return "", ErrInvalidRole
}

func normalizeStatus(status string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case StatusActive:
		return StatusActive, nil
	case StatusPending:
		return StatusPending, nil
	case StatusDisabled:
		return StatusDisabled, nil
	default:
		return "", ErrInvalidStatus
	}
}

func normalizeProjectStatus(status string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "DRAFT":
		return "DRAFT", nil
	case "REVIEW":
		return "REVIEW", nil
	case "RECRUITMENT":
		return "RECRUITMENT", nil
	case "ACTIVE":
		return "ACTIVE", nil
	case "GRADING":
		return "GRADING", nil
	case "ARCHIVE":
		return "ARCHIVE", nil
	default:
		return "", ErrInvalidProjectStatus
	}
}

func normalizeProjectStatusFilter(status string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return "", nil
	}
	return normalizeProjectStatus(status)
}
