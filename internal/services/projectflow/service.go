package projectflow

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrRecruitmentOpen  = errors.New("recruitment is not open")
	ErrProjectNotReady  = errors.New("project is not ready for activation")
	ErrProjectNotActive = errors.New("project is not active")
	ErrPositionFull     = errors.New("position capacity reached")
)

type RoleGrantor interface {
	GrantRoleByCode(ctx context.Context, userID uuid.UUID, roleCode string, scope rbac.Scope, expiresAt *time.Time) error
}

type Service struct {
	db      *pgxpool.Pool
	authz   *rbac.Service
	grantor RoleGrantor
	now     func() time.Time
}

func NewService(db *pgxpool.Pool, authz *rbac.Service, grantor RoleGrantor) *Service {
	return &Service{db: db, authz: authz, grantor: grantor, now: time.Now}
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

type Position struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Capacity  int       `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	UserID       string     `json:"user_id"`
	PositionID   *string    `json:"position_id,omitempty"`
	PositionCode *string    `json:"position_code,omitempty"`
	PositionName *string    `json:"position_name,omitempty"`
	Status       string     `json:"status"`
	JoinedAt     *time.Time `json:"joined_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
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

type Readiness struct {
	ProjectID       string `json:"project_id"`
	Status          string `json:"status"`
	RequiredMembers int    `json:"required_members"`
	ActiveMembers   int    `json:"active_members"`
	HasProfessor    bool   `json:"has_professor"`
	CriteriaCount   int    `json:"criteria_count"`
	CanActivate     bool   `json:"can_activate"`
}

func (s *Service) projectByID(ctx context.Context, projectID uuid.UUID) (domain.Project, error) {
	const q = `
SELECT id, title, description, status, is_public, created_by, professor_id,
       faculty_id, visibility, group_id,
       created_at, updated_at
FROM projects
WHERE id = $1;
`
	var p domain.Project
	var professorID *uuid.UUID
	var groupID *uuid.UUID

	err := s.db.QueryRow(ctx, q, projectID).Scan(
		&p.ID,
		&p.Title,
		&p.Description,
		&p.Status,
		&p.IsPublic,
		&p.CreatedBy,
		&professorID,
		&p.FacultyID,
		&p.Visibility,
		&groupID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return domain.Project{}, err
	}
	p.ProfessorID = professorID
	p.GroupID = groupID
	return p, nil
}

func (s *Service) requireProjectPermission(ctx context.Context, userID uuid.UUID, permission string, projectID uuid.UUID) error {
	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}
	ok, err := s.authz.Can(ctx, userID, permission, scope)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) requireFacultyPermission(ctx context.Context, userID uuid.UUID, permission string, facultyID uuid.UUID) error {
	scope := rbac.Scope{Type: rbac.ScopeFaculty, ID: &facultyID}
	ok, err := s.authz.Can(ctx, userID, permission, scope)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) ensureProjectRole(ctx context.Context, userID uuid.UUID, roleCode string, projectID uuid.UUID) error {
	const q = `
SELECT EXISTS (
	SELECT 1
	FROM role_assignments ra
	JOIN roles r ON r.id = ra.role_id
	WHERE ra.user_id = $1
	  AND ra.scope_type = 'PROJECT'
	  AND ra.scope_id = $2
	  AND r.code = $3
) AS ok;
`
	var ok bool
	if err := s.db.QueryRow(ctx, q, userID, projectID, roleCode).Scan(&ok); err != nil {
		return err
	}
	if ok {
		return nil
	}
	return s.grantor.GrantRoleByCode(ctx, userID, roleCode, rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}, nil)
}

func normalizeStackCodes(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, raw := range input {
		v := strings.ToUpper(strings.TrimSpace(raw))
		if v == "" {
			continue
		}
		if len(v) > 40 {
			v = v[:40]
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func stacksFromCodes(codes []string) []Stack {
	out := make([]Stack, 0, len(codes))
	for _, code := range codes {
		out = append(out, Stack{Code: code})
	}
	return out
}

func isUndefinedRelation(err error, relation string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "42P01" {
		return false
	}
	if strings.EqualFold(pgErr.TableName, relation) {
		return true
	}
	return strings.Contains(strings.ToLower(pgErr.Message), strings.ToLower(relation))
}

func normalizePositionCode(code, name string) string {
	v := strings.ToUpper(strings.TrimSpace(code))
	if v == "" {
		v = strings.ToUpper(strings.TrimSpace(name))
	}
	v = strings.ReplaceAll(v, " ", "_")
	v = strings.ReplaceAll(v, "-", "_")
	if len(v) > 40 {
		v = v[:40]
	}
	return v
}

func (s *Service) ensurePositionExists(ctx context.Context, projectID, positionID uuid.UUID) (capacity int, err error) {
	err = s.db.QueryRow(ctx, `
SELECT capacity
FROM project_positions
WHERE project_id = $1 AND id = $2
`, projectID, positionID).Scan(&capacity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%w: unknown position_id", ErrInvalidInput)
		}
		return 0, err
	}
	return capacity, nil
}

func (s *Service) ensurePositionCapacity(ctx context.Context, projectID, positionID uuid.UUID, excludeUserID *uuid.UUID) error {
	capacity, err := s.ensurePositionExists(ctx, projectID, positionID)
	if err != nil {
		return err
	}

	var occupied int
	if excludeUserID != nil {
		err = s.db.QueryRow(ctx, `
SELECT COUNT(*)
FROM project_members
WHERE project_id = $1
  AND position_id = $2
  AND status = 'ACTIVE'
  AND user_id <> $3
`, projectID, positionID, *excludeUserID).Scan(&occupied)
	} else {
		err = s.db.QueryRow(ctx, `
SELECT COUNT(*)
FROM project_members
WHERE project_id = $1
  AND position_id = $2
  AND status = 'ACTIVE'
`, projectID, positionID).Scan(&occupied)
	}
	if err != nil {
		return err
	}

	if occupied >= capacity {
		return ErrPositionFull
	}
	return nil
}

func (s *Service) ensureAssigneeMatchesPosition(ctx context.Context, projectID, userID, positionID uuid.UUID) error {
	const q = `
SELECT status, position_id
FROM project_members
WHERE project_id = $1 AND user_id = $2;
`
	var status string
	var assignedPositionID *uuid.UUID
	if err := s.db.QueryRow(ctx, q, projectID, userID).Scan(&status, &assignedPositionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: assignee is not a project member", ErrInvalidInput)
		}
		return err
	}
	if status != "ACTIVE" {
		return fmt.Errorf("%w: assignee must be ACTIVE member", ErrInvalidInput)
	}
	if assignedPositionID == nil || *assignedPositionID != positionID {
		return fmt.Errorf("%w: assignee role does not match task position", ErrInvalidInput)
	}
	return nil
}

func (s *Service) taskByID(ctx context.Context, projectID, taskID uuid.UUID) (Task, error) {
	const q = `
SELECT t.id, t.project_id, t.title, t.description, t.position_id,
       t.assignee_user_id, t.status, t.created_by, t.due_at, t.created_at, t.updated_at,
       p.code, p.name
FROM tasks t
JOIN project_positions p ON p.id = t.position_id
WHERE t.project_id = $1
  AND t.id = $2;
`
	var t Task
	var id uuid.UUID
	var pid uuid.UUID
	var positionID uuid.UUID
	var assignee *uuid.UUID
	var createdBy uuid.UUID

	if err := s.db.QueryRow(ctx, q, projectID, taskID).Scan(
		&id,
		&pid,
		&t.Title,
		&t.Description,
		&positionID,
		&assignee,
		&t.Status,
		&createdBy,
		&t.DueAt,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.PositionCode,
		&t.PositionName,
	); err != nil {
		return Task{}, err
	}

	t.ID = id.String()
	t.ProjectID = pid.String()
	t.PositionID = positionID.String()
	t.CreatedBy = createdBy.String()
	if assignee != nil {
		s := assignee.String()
		t.AssigneeUserID = &s
	}
	return t, nil
}

func (s *Service) UpdateProject(ctx context.Context, userID, projectID uuid.UUID, title, description *string) (domain.Project, error) {
	if title == nil && description == nil {
		return domain.Project{}, ErrInvalidInput
	}
	if err := s.requireProjectPermission(ctx, userID, "project.edit", projectID); err != nil {
		return domain.Project{}, err
	}

	titleVal := ""
	titleSet := false
	if title != nil {
		titleSet = true
		titleVal = strings.TrimSpace(*title)
		if titleVal == "" {
			return domain.Project{}, ErrInvalidInput
		}
	}
	descVal := ""
	descSet := false
	if description != nil {
		descSet = true
		descVal = strings.TrimSpace(*description)
	}

	const q = `
UPDATE projects
SET title = CASE WHEN $2::boolean THEN $3 ELSE title END,
    description = CASE WHEN $4::boolean THEN $5 ELSE description END,
    updated_at = now()
WHERE id = $1;
`
	ct, err := s.db.Exec(ctx, q, projectID, titleSet, titleVal, descSet, descVal)
	if err != nil {
		return domain.Project{}, err
	}
	if ct.RowsAffected() == 0 {
		return domain.Project{}, pgx.ErrNoRows
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) SetStacks(ctx context.Context, userID, projectID uuid.UUID, stacks []string) ([]Stack, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.edit", projectID); err != nil {
		return nil, err
	}
	norm := normalizeStackCodes(stacks)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM project_stacks WHERE project_id = $1`, projectID); err != nil {
		if isUndefinedRelation(err, "project_stacks") {
			return stacksFromCodes(norm), nil
		}
		return nil, err
	}
	for _, code := range norm {
		if _, err := tx.Exec(ctx, `INSERT INTO project_stacks(project_id, stack_code) VALUES ($1, $2)`, projectID, code); err != nil {
			if isUndefinedRelation(err, "project_stacks") {
				return stacksFromCodes(norm), nil
			}
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.ListStacks(ctx, projectID)
}

func (s *Service) ListStacks(ctx context.Context, projectID uuid.UUID) ([]Stack, error) {
	rows, err := s.db.Query(ctx, `SELECT stack_code FROM project_stacks WHERE project_id = $1 ORDER BY stack_code ASC`, projectID)
	if err != nil {
		if isUndefinedRelation(err, "project_stacks") {
			return []Stack{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]Stack, 0, 8)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, Stack{Code: code})
	}
	return out, rows.Err()
}

func (s *Service) OpenRecruitment(ctx context.Context, userID, projectID uuid.UUID) (domain.Project, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.edit", projectID); err != nil {
		return domain.Project{}, err
	}
	const q = `
UPDATE projects
SET status = 'RECRUITMENT', updated_at = now()
WHERE id = $1
  AND status IN ('DRAFT','REVIEW','RECRUITMENT');
`
	ct, err := s.db.Exec(ctx, q, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if ct.RowsAffected() == 0 {
		return domain.Project{}, ErrInvalidInput
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) CreatePosition(ctx context.Context, userID, projectID uuid.UUID, code, name string, capacity int) (Position, error) {
	if err := s.requireProjectPermission(ctx, userID, "position.create", projectID); err != nil {
		return Position{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Position{}, ErrInvalidInput
	}
	if capacity <= 0 {
		capacity = 1
	}
	code = normalizePositionCode(code, name)
	if code == "" {
		return Position{}, ErrInvalidInput
	}

	const q = `
INSERT INTO project_positions(project_id, code, name, capacity)
VALUES ($1, $2, $3, $4)
RETURNING id, project_id, code, name, capacity, created_at;
`
	var p Position
	err := s.db.QueryRow(ctx, q, projectID, code, name, capacity).Scan(
		&p.ID,
		&p.ProjectID,
		&p.Code,
		&p.Name,
		&p.Capacity,
		&p.CreatedAt,
	)
	return p, err
}

func (s *Service) ListPositions(ctx context.Context, projectID uuid.UUID) ([]Position, error) {
	const q = `
SELECT id, project_id, code, name, capacity, created_at
FROM project_positions
WHERE project_id = $1
ORDER BY created_at ASC;
`
	rows, err := s.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Position, 0, 8)
	for rows.Next() {
		var p Position
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Code, &p.Name, &p.Capacity, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Service) ApplyMember(ctx context.Context, userID, projectID uuid.UUID) (Member, error) {
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Member{}, err
	}
	if err := s.requireFacultyPermission(ctx, userID, "member.apply", p.FacultyID); err != nil {
		return Member{}, err
	}
	if p.Status != domain.ProjectRecruitment {
		return Member{}, ErrRecruitmentOpen
	}

	const q = `
INSERT INTO project_members(project_id, user_id, status, position_id, joined_at)
VALUES ($1, $2, 'APPLIED', NULL, NULL)
ON CONFLICT (project_id, user_id)
DO UPDATE SET status='APPLIED', position_id=NULL, joined_at=NULL
RETURNING id, project_id, user_id, position_id, status, joined_at, created_at;
`
	var m Member
	var positionID *uuid.UUID
	err = s.db.QueryRow(ctx, q, projectID, userID).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&positionID,
		&m.Status,
		&m.JoinedAt,
		&m.CreatedAt,
	)
	if err != nil {
		return Member{}, err
	}
	if positionID != nil {
		s := positionID.String()
		m.PositionID = &s
	}
	return m, nil
}

func (s *Service) ListMembers(ctx context.Context, projectID uuid.UUID) ([]Member, error) {
	const q = `
SELECT m.id, m.project_id, m.user_id, m.position_id, m.status, m.joined_at, m.created_at,
       p.code, p.name
FROM project_members m
LEFT JOIN project_positions p ON p.id = m.position_id
WHERE m.project_id = $1
ORDER BY m.created_at ASC;
`
	rows, err := s.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Member, 0, 8)
	for rows.Next() {
		var m Member
		var positionID *uuid.UUID
		var posCode *string
		var posName *string
		if err := rows.Scan(
			&m.ID,
			&m.ProjectID,
			&m.UserID,
			&positionID,
			&m.Status,
			&m.JoinedAt,
			&m.CreatedAt,
			&posCode,
			&posName,
		); err != nil {
			return nil, err
		}
		if positionID != nil {
			s := positionID.String()
			m.PositionID = &s
		}
		m.PositionCode = posCode
		m.PositionName = posName
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Service) ApproveMember(ctx context.Context, userID, projectID, memberUserID, positionID uuid.UUID) (Member, error) {
	if err := s.requireProjectPermission(ctx, userID, "member.approve", projectID); err != nil {
		return Member{}, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Member{}, err
	}
	if p.Status != domain.ProjectRecruitment {
		return Member{}, ErrRecruitmentOpen
	}
	if err := s.ensurePositionCapacity(ctx, projectID, positionID, &memberUserID); err != nil {
		return Member{}, err
	}

	const q = `
UPDATE project_members
SET status='ACTIVE', position_id=$3, joined_at=now()
WHERE project_id=$1 AND user_id=$2 AND status IN ('APPLIED', 'ACTIVE')
RETURNING id, project_id, user_id, position_id, status, joined_at, created_at;
`
	var m Member
	var posID *uuid.UUID
	err = s.db.QueryRow(ctx, q, projectID, memberUserID, positionID).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&posID,
		&m.Status,
		&m.JoinedAt,
		&m.CreatedAt,
	)
	if err != nil {
		return Member{}, err
	}
	if posID != nil {
		s := posID.String()
		m.PositionID = &s
	}
	if err := s.ensureProjectRole(ctx, memberUserID, "MEMBER", projectID); err != nil {
		return Member{}, err
	}
	return m, nil
}

func (s *Service) SetMemberPosition(ctx context.Context, userID, projectID, memberUserID, positionID uuid.UUID) (Member, error) {
	if err := s.requireProjectPermission(ctx, userID, "member.approve", projectID); err != nil {
		return Member{}, err
	}
	if err := s.ensurePositionCapacity(ctx, projectID, positionID, &memberUserID); err != nil {
		return Member{}, err
	}

	const q = `
UPDATE project_members
SET position_id=$3
WHERE project_id=$1 AND user_id=$2 AND status='ACTIVE'
RETURNING id, project_id, user_id, position_id, status, joined_at, created_at;
`
	var m Member
	var posID *uuid.UUID
	err := s.db.QueryRow(ctx, q, projectID, memberUserID, positionID).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&posID,
		&m.Status,
		&m.JoinedAt,
		&m.CreatedAt,
	)
	if err != nil {
		return Member{}, err
	}
	if posID != nil {
		s := posID.String()
		m.PositionID = &s
	}
	return m, nil
}

func (s *Service) AssignProfessor(ctx context.Context, userID, projectID, professorID uuid.UUID) (domain.Project, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.invite_professor", projectID); err != nil {
		return domain.Project{}, err
	}
	const q = `
UPDATE projects
SET professor_id = $2,
    status = CASE WHEN status IN ('DRAFT','RECRUITMENT') THEN 'REVIEW' ELSE status END,
    updated_at = now()
WHERE id = $1;
`
	ct, err := s.db.Exec(ctx, q, projectID, professorID)
	if err != nil {
		return domain.Project{}, err
	}
	if ct.RowsAffected() == 0 {
		return domain.Project{}, pgx.ErrNoRows
	}
	if err := s.ensureProjectRole(ctx, professorID, "PROJECT_PROFESSOR", projectID); err != nil {
		return domain.Project{}, err
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) CreateCriterion(ctx context.Context, userID, projectID uuid.UUID, title, description string, weight int) (Criterion, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.set_criteria", projectID); err != nil {
		return Criterion{}, err
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return Criterion{}, ErrInvalidInput
	}
	if weight <= 0 {
		weight = 1
	}
	if weight > 100 {
		weight = 100
	}

	const q = `
INSERT INTO project_criteria(project_id, title, description, weight, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, title, description, weight, created_by, created_at;
`
	var c Criterion
	err := s.db.QueryRow(ctx, q, projectID, title, description, weight, userID).Scan(
		&c.ID,
		&c.ProjectID,
		&c.Title,
		&c.Description,
		&c.Weight,
		&c.CreatedBy,
		&c.CreatedAt,
	)
	return c, err
}

func (s *Service) ListCriteria(ctx context.Context, projectID uuid.UUID) ([]Criterion, error) {
	const q = `
SELECT id, project_id, title, description, weight, created_by, created_at
FROM project_criteria
WHERE project_id=$1
ORDER BY created_at ASC;
`
	rows, err := s.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Criterion, 0, 8)
	for rows.Next() {
		var c Criterion
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Title, &c.Description, &c.Weight, &c.CreatedBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Service) Readiness(ctx context.Context, projectID uuid.UUID) (Readiness, error) {
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return Readiness{}, err
	}

	var requiredMembers int
	if err := s.db.QueryRow(ctx, `SELECT COALESCE(SUM(capacity), 0) FROM project_positions WHERE project_id=$1`, projectID).Scan(&requiredMembers); err != nil {
		return Readiness{}, err
	}

	var activeMembers int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM project_members WHERE project_id=$1 AND status='ACTIVE'`, projectID).Scan(&activeMembers); err != nil {
		return Readiness{}, err
	}

	var criteriaCount int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM project_criteria WHERE project_id=$1`, projectID).Scan(&criteriaCount); err != nil {
		return Readiness{}, err
	}

	hasProfessor := p.ProfessorID != nil
	canActivate := requiredMembers > 0 && activeMembers >= requiredMembers && hasProfessor && criteriaCount > 0

	return Readiness{
		ProjectID:       projectID.String(),
		Status:          string(p.Status),
		RequiredMembers: requiredMembers,
		ActiveMembers:   activeMembers,
		HasProfessor:    hasProfessor,
		CriteriaCount:   criteriaCount,
		CanActivate:     canActivate,
	}, nil
}

func (s *Service) ApproveProject(ctx context.Context, userID, projectID uuid.UUID) (domain.Project, Readiness, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.approve", projectID); err != nil {
		return domain.Project{}, Readiness{}, err
	}
	ready, err := s.Readiness(ctx, projectID)
	if err != nil {
		return domain.Project{}, Readiness{}, err
	}
	if !ready.CanActivate {
		return domain.Project{}, ready, fmt.Errorf("%w: members=%d/%d professor=%v criteria=%d", ErrProjectNotReady, ready.ActiveMembers, ready.RequiredMembers, ready.HasProfessor, ready.CriteriaCount)
	}

	const q = `
UPDATE projects
SET status='ACTIVE', updated_at=now()
WHERE id=$1;
`
	ct, err := s.db.Exec(ctx, q, projectID)
	if err != nil {
		return domain.Project{}, Readiness{}, err
	}
	if ct.RowsAffected() == 0 {
		return domain.Project{}, Readiness{}, pgx.ErrNoRows
	}
	p, err := s.projectByID(ctx, projectID)
	return p, ready, err
}

func (s *Service) ensureActiveProject(ctx context.Context, projectID uuid.UUID) error {
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return err
	}
	if p.Status != domain.ProjectActive {
		return ErrProjectNotActive
	}
	return nil
}

func normalizeTaskStatus(status string) (string, bool) {
	s := strings.ToUpper(strings.TrimSpace(status))
	if s == "OPEN" || s == "IN_PROGRESS" || s == "DONE" {
		return s, true
	}
	return "", false
}

func (s *Service) CreateTask(ctx context.Context, userID, projectID uuid.UUID, title, description string, positionID uuid.UUID, assigneeUserID *uuid.UUID, dueAt *time.Time) (Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.create", projectID); err != nil {
		return Task{}, err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return Task{}, err
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return Task{}, ErrInvalidInput
	}
	if _, err := s.ensurePositionExists(ctx, projectID, positionID); err != nil {
		return Task{}, err
	}
	if assigneeUserID != nil {
		if err := s.ensureAssigneeMatchesPosition(ctx, projectID, *assigneeUserID, positionID); err != nil {
			return Task{}, err
		}
	}

	status := "OPEN"
	if assigneeUserID != nil {
		status = "IN_PROGRESS"
	}

	const q = `
INSERT INTO tasks(project_id, title, description, position_id, assignee_user_id, status, created_by, due_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;
`
	var taskID uuid.UUID
	err := s.db.QueryRow(ctx, q, projectID, title, description, positionID, assigneeUserID, status, userID, dueAt).Scan(&taskID)
	if err != nil {
		return Task{}, err
	}
	return s.taskByID(ctx, projectID, taskID)
}

func (s *Service) ListTasks(ctx context.Context, userID, projectID uuid.UUID) ([]Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.view", projectID); err != nil {
		return nil, err
	}
	const q = `
SELECT t.id, t.project_id, t.title, t.description, t.position_id,
       t.assignee_user_id, t.status, t.created_by, t.due_at, t.created_at, t.updated_at,
       p.code, p.name
FROM tasks t
JOIN project_positions p ON p.id = t.position_id
WHERE t.project_id = $1
ORDER BY t.created_at ASC;
`
	rows, err := s.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Task, 0, 16)
	for rows.Next() {
		var t Task
		var id uuid.UUID
		var pid uuid.UUID
		var positionID uuid.UUID
		var assignee *uuid.UUID
		var createdBy uuid.UUID
		if err := rows.Scan(
			&id,
			&pid,
			&t.Title,
			&t.Description,
			&positionID,
			&assignee,
			&t.Status,
			&createdBy,
			&t.DueAt,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.PositionCode,
			&t.PositionName,
		); err != nil {
			return nil, err
		}
		if assignee != nil {
			s := assignee.String()
			t.AssigneeUserID = &s
		}
		t.ID = id.String()
		t.ProjectID = pid.String()
		t.PositionID = positionID.String()
		t.CreatedBy = createdBy.String()
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Service) UpdateTaskStatus(ctx context.Context, userID, projectID, taskID uuid.UUID, status string) (Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.update", projectID); err != nil {
		return Task{}, err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return Task{}, err
	}
	status, ok := normalizeTaskStatus(status)
	if !ok {
		return Task{}, ErrInvalidInput
	}

	const q = `
UPDATE tasks
SET status=$3, updated_at=now()
WHERE project_id=$1 AND id=$2
RETURNING id;
`
	var outTaskID uuid.UUID
	err := s.db.QueryRow(ctx, q, projectID, taskID, status).Scan(&outTaskID)
	if err != nil {
		return Task{}, err
	}
	return s.taskByID(ctx, projectID, outTaskID)
}

func (s *Service) AssignTask(ctx context.Context, userID, projectID, taskID, assigneeUserID uuid.UUID) (Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.assign", projectID); err != nil {
		return Task{}, err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return Task{}, err
	}

	var positionID uuid.UUID
	if err := s.db.QueryRow(ctx, `
SELECT position_id
FROM tasks
WHERE project_id = $1 AND id = $2
`, projectID, taskID).Scan(&positionID); err != nil {
		return Task{}, err
	}
	if err := s.ensureAssigneeMatchesPosition(ctx, projectID, assigneeUserID, positionID); err != nil {
		return Task{}, err
	}

	const q = `
UPDATE tasks
SET assignee_user_id = $3,
    status = CASE WHEN status='OPEN' THEN 'IN_PROGRESS' ELSE status END,
    updated_at = now()
WHERE project_id=$1 AND id=$2
RETURNING id;
`
	var outTaskID uuid.UUID
	err := s.db.QueryRow(ctx, q, projectID, taskID, assigneeUserID).Scan(&outTaskID)
	if err != nil {
		return Task{}, err
	}
	return s.taskByID(ctx, projectID, outTaskID)
}

func (s *Service) ClaimTask(ctx context.Context, userID, projectID, taskID uuid.UUID) error {
	if err := s.requireProjectPermission(ctx, userID, "task.claim", projectID); err != nil {
		return err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return err
	}
	const q = `
UPDATE tasks t
SET assignee_user_id = $3, status='IN_PROGRESS', updated_at=now()
WHERE t.id=$2
  AND t.project_id=$1
  AND t.status='OPEN'
  AND t.assignee_user_id IS NULL
  AND EXISTS (
    SELECT 1 FROM project_members m
    WHERE m.project_id=$1
      AND m.user_id=$3
      AND m.status='ACTIVE'
      AND m.position_id = t.position_id
  );
`
	ct, err := s.db.Exec(ctx, q, projectID, taskID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrForbidden
	}
	return nil
}
