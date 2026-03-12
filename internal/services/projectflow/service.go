package projectflow

import (
	"context"
	"errors"
	"time"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrRecruitmentOpen  = errors.New("recruitment is not open")
	ErrProjectNotReady  = errors.New("project is not ready for activation")
	ErrProjectNotActive = errors.New("project is not active")
	ErrPositionFull     = errors.New("position capacity reached")
	ErrInviteNotFound   = errors.New("invite not found")
	ErrNotFound         = errors.New("not found")
	ErrSchemaMissing    = errors.New("schema relation is missing")
)

type RoleGrantor interface {
	GrantRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scope rbac.Scope, expiresAt *time.Time) error
}

type Service struct {
	authz          *rbac.Service
	grantor        RoleGrantor
	projectsRepo   ProjectsRepository
	stacksRepo     StacksRepository
	positionsRepo  PositionsRepository
	membersRepo    MembersRepository
	professorsRepo ProfessorsRepository
	criteriaRepo   CriteriaRepository
	lifecycleRepo  LifecycleRepository
	tasksRepo      TasksRepository
	now            func() time.Time
}

func NewService(
	authz *rbac.Service,
	grantor RoleGrantor,
	projectsRepo ProjectsRepository,
	stacksRepo StacksRepository,
	positionsRepo PositionsRepository,
	membersRepo MembersRepository,
	professorsRepo ProfessorsRepository,
	criteriaRepo CriteriaRepository,
	lifecycleRepo LifecycleRepository,
	tasksRepo TasksRepository,
) *Service {
	if projectsRepo == nil {
		panic("projectflow.NewService: projectsRepo is nil")
	}
	if stacksRepo == nil {
		panic("projectflow.NewService: stacksRepo is nil")
	}
	if positionsRepo == nil {
		panic("projectflow.NewService: positionsRepo is nil")
	}
	if membersRepo == nil {
		panic("projectflow.NewService: membersRepo is nil")
	}
	if professorsRepo == nil {
		panic("projectflow.NewService: professorsRepo is nil")
	}
	if criteriaRepo == nil {
		panic("projectflow.NewService: criteriaRepo is nil")
	}
	if lifecycleRepo == nil {
		panic("projectflow.NewService: lifecycleRepo is nil")
	}
	if tasksRepo == nil {
		panic("projectflow.NewService: tasksRepo is nil")
	}

	return &Service{
		authz:          authz,
		grantor:        grantor,
		projectsRepo:   projectsRepo,
		stacksRepo:     stacksRepo,
		positionsRepo:  positionsRepo,
		membersRepo:    membersRepo,
		professorsRepo: professorsRepo,
		criteriaRepo:   criteriaRepo,
		lifecycleRepo:  lifecycleRepo,
		tasksRepo:      tasksRepo,
		now:            time.Now,
	}
}

type Stack struct {
	Code string `json:"code"`
}

type Criterion struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Weight      int       `json:"weight"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type CriterionGrade struct {
	CriterionID string     `json:"criterion_id"`
	IsMet       *bool      `json:"is_met,omitempty"`
	Comment     string     `json:"comment,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type CriterionGradeUpsert struct {
	CriterionID uuid.UUID
	IsMet       *bool
	Comment     string
}

type Position struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Capacity  int       `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	UserID        string     `json:"user_id"`
	FullName      string     `json:"full_name,omitempty"`
	Email         string     `json:"email,omitempty"`
	PositionID    *string    `json:"position_id,omitempty"`
	PositionCode  *string    `json:"position_code,omitempty"`
	PositionName  *string    `json:"position_name,omitempty"`
	Status        string     `json:"status"`
	InviteComment string     `json:"invite_comment,omitempty"`
	InvitedBy     *string    `json:"invited_by,omitempty"`
	RespondedAt   *time.Time `json:"responded_at,omitempty"`
	JoinedAt      *time.Time `json:"joined_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Task struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"project_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	PositionID     string     `json:"position_id"`
	PositionCode   string     `json:"position_code"`
	PositionName   string     `json:"position_name"`
	AssigneeUserID *string    `json:"assignee_user_id,omitempty"`
	Status         string     `json:"status"`
	CreatedBy      string     `json:"created_by"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type TaskActivity struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	TaskID      string    `json:"task_id"`
	ActorUserID *string   `json:"actor_user_id,omitempty"`
	ActorName   string    `json:"actor_name,omitempty"`
	ActorEmail  string    `json:"actor_email,omitempty"`
	EventType   string    `json:"event_type"`
	FromStatus  string    `json:"from_status,omitempty"`
	ToStatus    string    `json:"to_status,omitempty"`
	Title       string    `json:"title,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	Attachments []string  `json:"attachments,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type StudentCandidate struct {
	UserID         string `json:"user_id"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	DepartmentCode string `json:"department_code"`
}

type ProfessorCandidate struct {
	UserID         string `json:"user_id"`
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	DepartmentCode string `json:"department_code"`
}

type IncomingInvite struct {
	ProjectID     string     `json:"project_id"`
	ProjectTitle  string     `json:"project_title"`
	ProjectStatus string     `json:"project_status"`
	Status        string     `json:"status"`
	UserID        string     `json:"user_id,omitempty"`
	InviteComment string     `json:"invite_comment,omitempty"`
	InvitedBy     *string    `json:"invited_by,omitempty"`
	InviterName   string     `json:"inviter_name,omitempty"`
	InviterEmail  string     `json:"inviter_email,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	RespondedAt   *time.Time `json:"responded_at,omitempty"`
}

type OutgoingApplication struct {
	ProjectID     string     `json:"project_id"`
	ProjectTitle  string     `json:"project_title"`
	ProjectStatus string     `json:"project_status"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	RespondedAt   *time.Time `json:"responded_at,omitempty"`
}

type Readiness struct {
	ProjectID       string `json:"project_id"`
	Status          string `json:"status"`
	RequiredMembers int    `json:"required_members"`
	ActiveMembers   int    `json:"active_members"`
	HasProfessor    bool   `json:"has_professor"`
	ProfessorStatus string `json:"professor_status"`
	CriteriaCount   int    `json:"criteria_count"`
	CanActivate     bool   `json:"can_activate"`
}
