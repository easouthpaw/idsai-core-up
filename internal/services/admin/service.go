package admin

import (
	"context"
	"errors"
	"fmt"
	"idsai-core-up/internal/security/passwords"
	"strings"
	"time"

	"github.com/google/uuid"
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
	ErrInvalidProjectStatus = errors.New("project status must be DRAFT, REVIEW, RECRUITMENT, ACTIVE, GRADING, COMPLETED or ARCHIVE")
	ErrInvalidInput         = errors.New("invalid input")
	ErrUserNotFound         = errors.New("user not found")
	ErrProjectNotFound      = errors.New("project not found")
	ErrDepartmentNotFound   = errors.New("department not found")
	ErrUserExists           = errors.New("user already exists")
)

type User struct {
	ID             uuid.UUID
	FullName       string
	Email          string
	RoleCode       string
	Status         string
	FacultyCode    string
	DepartmentCode string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Project struct {
	ID             uuid.UUID
	Title          string
	Description    string
	Status         string
	Visibility     string
	IsPublic       bool
	CreatedBy      uuid.UUID
	AuthorName     string
	AuthorEmail    string
	FacultyCode    string
	DepartmentCode string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ProjectPosition struct {
	ID       uuid.UUID
	Code     string
	Name     string
	Capacity int
}

type ProjectMember struct {
	UserID       uuid.UUID
	FullName     string
	Email        string
	RoleCode     string
	Status       string
	PositionCode string
	PositionName string
	JoinedAt     *time.Time
	RespondedAt  *time.Time
}

type ProjectTask struct {
	ID             uuid.UUID
	Title          string
	Status         string
	PositionCode   string
	AssigneeUserID *uuid.UUID
	AssigneeName   string
	DueAt          *time.Time
	UpdatedAt      time.Time
}

type ProjectCriterion struct {
	ID        uuid.UUID
	Title     string
	Weight    int
	CreatedBy uuid.UUID
	CreatedAt time.Time
	IsMet     *bool
	Comment   string
	UpdatedAt *time.Time
}

type ProjectObservationSummary struct {
	MembersTotal   int
	MembersActive  int
	MembersApplied int
	MembersInvited int
	TasksTotal     int
	TasksDone      int
	CriteriaTotal  int
}

type ProjectObservation struct {
	Project   Project
	Positions []ProjectPosition
	Members   []ProjectMember
	Tasks     []ProjectTask
	Criteria  []ProjectCriterion
	Summary   ProjectObservationSummary
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
	GetProjectObservation(ctx context.Context, projectID uuid.UUID) (ProjectObservation, error)
	CreateUser(ctx context.Context, in CreateUserParams) (User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (User, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
	UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode string) error
	UpdateUserPasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error
	RevokeUserSessions(ctx context.Context, userID uuid.UUID) error
	UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status string) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	DeleteProject(ctx context.Context, projectID uuid.UUID) error
	GetProjectByID(ctx context.Context, projectID uuid.UUID) (Project, error)
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
	hash, err := passwords.Hash(password)
	if err != nil {
		return User{}, ErrInvalidInput
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
	if err := s.repo.UpdateUserStatus(ctx, userID, normalized); err != nil {
		return err
	}
	if normalized == StatusDisabled {
		return s.repo.RevokeUserSessions(ctx, userID)
	}
	return nil
}

func (s *Service) SetUserRole(ctx context.Context, userID uuid.UUID, roleCode string) (User, error) {
	role, err := normalizeRole(roleCode)
	if err != nil {
		return User{}, err
	}
	if err := s.repo.UpdateUserRole(ctx, userID, role); err != nil {
		return User{}, err
	}
	return s.repo.GetUserByID(ctx, userID)
}

func (s *Service) ResetUserPassword(ctx context.Context, userID uuid.UUID, password string) error {
	password = strings.TrimSpace(password)
	hash, err := passwords.Hash(password)
	if err != nil {
		return ErrInvalidInput
	}
	if err := s.repo.UpdateUserPasswordHash(ctx, userID, hash); err != nil {
		return err
	}
	return s.repo.RevokeUserSessions(ctx, userID)
}

func (s *Service) SetProjectStatus(ctx context.Context, projectID uuid.UUID, status string) (Project, error) {
	normalized, err := normalizeProjectStatus(status)
	if err != nil {
		return Project{}, err
	}
	current, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return Project{}, err
	}
	if err := validateAdminProjectStatusChange(current.Status, normalized); err != nil {
		return Project{}, err
	}
	if err := s.repo.UpdateProjectStatus(ctx, projectID, normalized); err != nil {
		return Project{}, err
	}
	return s.repo.GetProjectByID(ctx, projectID)
}

func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteUser(ctx, userID)
}

func (s *Service) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	return s.repo.DeleteProject(ctx, projectID)
}

func (s *Service) ObserveProject(ctx context.Context, projectID uuid.UUID) (ProjectObservation, error) {
	return s.repo.GetProjectObservation(ctx, projectID)
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
	case "COMPLETED":
		return "COMPLETED", nil
	case "ARCHIVE":
		return "ARCHIVE", nil
	default:
		return "", ErrInvalidProjectStatus
	}
}

func validateAdminProjectStatusChange(current, next string) error {
	current = strings.ToUpper(strings.TrimSpace(current))
	next = strings.ToUpper(strings.TrimSpace(next))

	switch next {
	case "ACTIVE":
		if current == "ACTIVE" || current == "REVIEW" || current == "RECRUITMENT" {
			return nil
		}
		return fmt.Errorf("%w: project can be activated only from REVIEW or RECRUITMENT", ErrInvalidInput)
	case "ARCHIVE":
		if current == "ARCHIVE" || current == "COMPLETED" {
			return nil
		}
		return fmt.Errorf("%w: project can be archived only from COMPLETED", ErrInvalidInput)
	default:
		return fmt.Errorf("%w: admin can change project status only to ACTIVE or ARCHIVE", ErrInvalidInput)
	}
}

func normalizeProjectStatusFilter(status string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return "", nil
	}
	return normalizeProjectStatus(status)
}
