package projectflow

import (
	"context"
	"errors"
	"time"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrRecruitmentOpen     = errors.New("recruitment is not open")
	ErrProjectNotReady     = errors.New("project is not ready for activation")
	ErrProjectNotActive    = errors.New("project is not active")
	ErrPositionFull        = errors.New("position capacity reached")
	ErrInviteNotFound      = errors.New("invite not found")
	ErrNotFound            = errors.New("not found")
	ErrSchemaMissing       = errors.New("schema relation is missing")
	ErrSystemManagedAccess = errors.New("cannot modify system-managed access")
	ErrUnknownRoleCode     = errors.New("unknown managed role code")
)

type RoleGrantor interface {
	GrantRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scope rbac.Scope, expiresAt *time.Time) error
}

type Service struct {
	authz          rbac.Authorizer
	grantor        RoleGrantor
	projectsRepo   ProjectsRepository
	stacksRepo     StacksRepository
	positionsRepo  PositionsRepository
	membersRepo    MembersRepository
	professorsRepo ProfessorsRepository
	criteriaRepo   CriteriaRepository
	lifecycleRepo  LifecycleRepository
	tasksRepo      TasksRepository
	accessRepo     AccessRepository
	now            func() time.Time
}

func NewService(
	authz rbac.Authorizer,
	grantor RoleGrantor,
	projectsRepo ProjectsRepository,
	stacksRepo StacksRepository,
	positionsRepo PositionsRepository,
	membersRepo MembersRepository,
	professorsRepo ProfessorsRepository,
	criteriaRepo CriteriaRepository,
	lifecycleRepo LifecycleRepository,
	tasksRepo TasksRepository,
	accessRepo AccessRepository,
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
	if accessRepo == nil {
		panic("projectflow.NewService: accessRepo is nil")
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
		accessRepo:     accessRepo,
		now:            time.Now,
	}
}

type Stack struct {
	Code string
}

type Criterion struct {
	ID          string
	ProjectID   string
	Title       string
	Description string
	Weight      int
	CreatedBy   string
	CreatedAt   time.Time
}

type CriterionGrade struct {
	CriterionID string
	IsMet       *bool
	Comment     string
	UpdatedAt   *time.Time
}

type CriterionGradeUpsert struct {
	CriterionID uuid.UUID
	IsMet       *bool
	Comment     string
}

type Position struct {
	ID        string
	ProjectID string
	Code      string
	Name      string
	Capacity  int
	CreatedAt time.Time
}

type Member struct {
	ID            string
	ProjectID     string
	UserID        string
	FullName      string
	Email         string
	PositionID    *string
	PositionCode  *string
	PositionName  *string
	Status        string
	InviteComment string
	InvitedBy     *string
	RespondedAt   *time.Time
	JoinedAt      *time.Time
	CreatedAt     time.Time
}

type Task struct {
	ID             string
	ProjectID      string
	Title          string
	Description    string
	PositionID     string
	PositionCode   string
	PositionName   string
	AssigneeUserID *string
	Status         string
	CreatedBy      string
	DueAt          *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TaskActivity struct {
	ID          string
	ProjectID   string
	TaskID      string
	ActorUserID *string
	ActorName   string
	ActorEmail  string
	EventType   string
	FromStatus  string
	ToStatus    string
	Title       string
	Comment     string
	Attachments []string
	CreatedAt   time.Time
}

type StudentCandidate struct {
	UserID         string
	FullName       string
	Email          string
	DepartmentCode string
}

type ProfessorCandidate struct {
	UserID         string
	FullName       string
	Email          string
	DepartmentCode string
}

type IncomingInvite struct {
	ProjectID     string
	ProjectTitle  string
	ProjectStatus string
	Status        string
	UserID        string
	InviteComment string
	InvitedBy     *string
	InviterName   string
	InviterEmail  string
	CreatedAt     time.Time
	RespondedAt   *time.Time
}

type OutgoingApplication struct {
	ProjectID     string
	ProjectTitle  string
	ProjectStatus string
	Status        string
	CreatedAt     time.Time
	RespondedAt   *time.Time
}

type Readiness struct {
	ProjectID       string
	Status          string
	RequiredMembers int
	ActiveMembers   int
	HasProfessor    bool
	ProfessorStatus string
	CriteriaCount   int
	CanActivate     bool
}
