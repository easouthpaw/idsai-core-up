package admin

import (
	"context"
	"errors"
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
	ErrInvalidProjectStatus = errors.New("project status must be DRAFT, REVIEW, RECRUITMENT, ACTIVE, GRADING or ARCHIVE")
	ErrInvalidInput         = errors.New("invalid input")
	ErrUserNotFound         = errors.New("user not found")
	ErrProjectNotFound      = errors.New("project not found")
	ErrDepartmentNotFound   = errors.New("department not found")
	ErrUserExists           = errors.New("user already exists")
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

type ProjectPosition struct {
	ID       uuid.UUID `json:"id"`
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	Capacity int       `json:"capacity"`
}

type ProjectMember struct {
	UserID       uuid.UUID  `json:"user_id"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	RoleCode     string     `json:"role_code"`
	Status       string     `json:"status"`
	PositionCode string     `json:"position_code"`
	PositionName string     `json:"position_name"`
	JoinedAt     *time.Time `json:"joined_at,omitempty"`
	RespondedAt  *time.Time `json:"responded_at,omitempty"`
}

type ProjectTask struct {
	ID             uuid.UUID  `json:"id"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	PositionCode   string     `json:"position_code"`
	AssigneeUserID *uuid.UUID `json:"assignee_user_id,omitempty"`
	AssigneeName   string     `json:"assignee_name"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ProjectCriterion struct {
	ID        uuid.UUID  `json:"id"`
	Title     string     `json:"title"`
	Weight    int        `json:"weight"`
	CreatedBy uuid.UUID  `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	IsMet     *bool      `json:"is_met,omitempty"`
	Comment   string     `json:"comment,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type ProjectObservationSummary struct {
	MembersTotal   int `json:"members_total"`
	MembersActive  int `json:"members_active"`
	MembersApplied int `json:"members_applied"`
	MembersInvited int `json:"members_invited"`
	TasksTotal     int `json:"tasks_total"`
	TasksDone      int `json:"tasks_done"`
	CriteriaTotal  int `json:"criteria_total"`
}

type ProjectObservation struct {
	Project   Project                   `json:"project"`
	Positions []ProjectPosition         `json:"positions"`
	Members   []ProjectMember           `json:"members"`
	Tasks     []ProjectTask             `json:"tasks"`
	Criteria  []ProjectCriterion        `json:"criteria"`
	Summary   ProjectObservationSummary `json:"summary"`
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
