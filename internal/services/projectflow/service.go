package projectflow

import (
	"context"
	"encoding/json"
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
	ErrInviteNotFound   = errors.New("invite not found")
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

type CriterionGrade struct {
	CriterionID string     `json:"criterion_id"`
	IsMet       *bool      `json:"is_met,omitempty"`
	Comment     string     `json:"comment,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
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

func (s *Service) projectByID(ctx context.Context, projectID uuid.UUID) (domain.Project, error) {
	const q = `
SELECT id, title, description, status, is_public, created_by, professor_id,
       professor_review_status,
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
		&p.ProfessorReviewStatus,
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

func (s *Service) isActiveProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	const q = `
SELECT EXISTS (
  SELECT 1
  FROM projects p
  WHERE p.id = $1
    AND (
      p.created_by = $2
      OR EXISTS (
        SELECT 1
        FROM project_members pm
        WHERE pm.project_id = p.id
          AND pm.user_id = $2
          AND pm.status = 'ACTIVE'
      )
    )
) AS ok;
`
	var ok bool
	if err := s.db.QueryRow(ctx, q, projectID, userID).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (s *Service) requireProjectEditAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	if err := s.requireProjectPermission(ctx, userID, "project.edit", projectID); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrForbidden) {
		return err
	}

	ok, err := s.isActiveProjectMember(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) requireProjectSubmitAccess(ctx context.Context, userID, projectID uuid.UUID) error {
	if err := s.requireProjectPermission(ctx, userID, "project.submit_for_review", projectID); err == nil {
		return nil
	} else if !errors.Is(err, domain.ErrForbidden) {
		return err
	}

	ok, err := s.isActiveProjectMember(ctx, userID, projectID)
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

func (s *Service) revokeProjectRole(ctx context.Context, userID uuid.UUID, roleCode string, projectID uuid.UUID) error {
	const q = `
DELETE FROM role_assignments ra
USING roles r
WHERE ra.role_id = r.id
  AND ra.user_id = $1
  AND ra.scope_type = 'PROJECT'
  AND ra.scope_id = $2
  AND r.code = $3;
`
	_, err := s.db.Exec(ctx, q, userID, projectID, roleCode)
	return err
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

func normalizeSearchQuery(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func clampLimit(limit, max int) int {
	if limit <= 0 {
		limit = 20
	}
	if max <= 0 {
		max = 20
	}
	if limit > max {
		return max
	}
	return limit
}

func normalizeTaskAttachments(items []string) []string {
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, raw := range items {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if len(v) > 1000 {
			v = v[:1000]
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
		if len(out) >= 10 {
			break
		}
	}
	return out
}

func encodeStringSliceJSON(items []string) []byte {
	b, err := json.Marshal(items)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func decodeStringSliceJSON(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return []string{}
	}
	return normalizeTaskAttachments(out)
}

func (s *Service) appendTaskActivity(
	ctx context.Context,
	projectID, taskID uuid.UUID,
	actorUserID *uuid.UUID,
	eventType, fromStatus, toStatus, title, comment string,
	attachments []string,
) error {
	eventType = strings.ToUpper(strings.TrimSpace(eventType))
	if eventType == "" {
		eventType = "STATUS_CHANGED"
	}
	fromStatus = strings.ToUpper(strings.TrimSpace(fromStatus))
	toStatus = strings.ToUpper(strings.TrimSpace(toStatus))
	title = strings.TrimSpace(title)
	comment = strings.TrimSpace(comment)
	if len(comment) > 3000 {
		comment = comment[:3000]
	}
	attachments = normalizeTaskAttachments(attachments)

	const q = `
INSERT INTO task_activity_logs(
  tenant_id, project_id, task_id, actor_user_id, event_type,
  from_status, to_status, title, comment, attachments
)
VALUES (
  (SELECT tenant_id FROM projects WHERE id = $1),
  $1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8, $9::jsonb
);
`
	_, err := s.db.Exec(ctx, q, projectID, taskID, actorUserID, eventType, fromStatus, toStatus, title, comment, encodeStringSliceJSON(attachments))
	if err != nil && isUndefinedRelation(err, "task_activity_logs") {
		return nil
	}
	return err
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
	if err := s.requireProjectEditAccess(ctx, userID, projectID); err != nil {
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
	if err := s.requireProjectEditAccess(ctx, userID, projectID); err != nil {
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
		if _, err := tx.Exec(ctx, `INSERT INTO project_stacks(tenant_id, project_id, stack_code) VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2)`, projectID, code); err != nil {
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
INSERT INTO project_positions(tenant_id, project_id, code, name, capacity)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, $3, $4)
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

func (s *Service) ListStudentCandidates(ctx context.Context, userID, projectID uuid.UUID, query string, limit int) ([]StudentCandidate, error) {
	if err := s.requireProjectPermission(ctx, userID, "member.approve", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	term := normalizeSearchQuery(query)
	limit = clampLimit(limit, 100)

	const q = `
SELECT u.id,
       COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(u.email, '@', 1)) AS full_name,
       u.email,
       COALESCE(d.code, '') AS department_code
FROM users u
JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN departments d ON d.id = up.department_id
WHERE u.status = 'ACTIVE'
  AND up.faculty_id = $1
  AND u.id <> $3
  AND u.id <> $4
  AND EXISTS (
    SELECT 1
    FROM role_assignments ra
    JOIN roles r ON r.id = ra.role_id
    WHERE ra.user_id = u.id
      AND ra.tenant_id = u.tenant_id
      AND r.code = 'STUDENT'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM project_members pm
    WHERE pm.project_id = $2
      AND pm.user_id = u.id
      AND pm.status IN ('ACTIVE', 'INVITED', 'APPLIED')
  )
  AND ($5 = '' OR lower(up.full_name) LIKE '%' || $5 || '%' OR lower(u.email) LIKE '%' || $5 || '%')
ORDER BY up.full_name ASC, u.email ASC
LIMIT $6;
`
	rows, err := s.db.Query(ctx, q, p.FacultyID, projectID, userID, p.CreatedBy, term, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]StudentCandidate, 0, limit)
	for rows.Next() {
		var item StudentCandidate
		if err := rows.Scan(&item.UserID, &item.FullName, &item.Email, &item.DepartmentCode); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Service) InviteStudent(ctx context.Context, userID, projectID, studentID uuid.UUID, comment string) (Member, error) {
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
	if studentID == userID || studentID == p.CreatedBy {
		return Member{}, ErrInvalidInput
	}

	const verifyStudentQ = `
SELECT EXISTS (
  SELECT 1
  FROM users u
  JOIN user_profiles up ON up.user_id = u.id
  WHERE u.id = $1
    AND u.status = 'ACTIVE'
    AND up.tenant_id = u.tenant_id
    AND up.faculty_id = $2
    AND EXISTS (
      SELECT 1
      FROM role_assignments ra
      JOIN roles r ON r.id = ra.role_id
      WHERE ra.user_id = u.id
        AND ra.tenant_id = u.tenant_id
        AND r.code = 'STUDENT'
    )
) AS ok;
`
	var ok bool
	if err := s.db.QueryRow(ctx, verifyStudentQ, studentID, p.FacultyID).Scan(&ok); err != nil {
		return Member{}, err
	}
	if !ok {
		return Member{}, ErrInvalidInput
	}

	comment = strings.TrimSpace(comment)
	if len(comment) > 500 {
		comment = comment[:500]
	}

	const q = `
INSERT INTO project_members(tenant_id, project_id, user_id, status, position_id, joined_at, invite_comment, invited_by, responded_at)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, 'INVITED', NULL, NULL, $3, $4, NULL)
ON CONFLICT (project_id, user_id)
DO UPDATE SET status='INVITED', position_id=NULL, joined_at=NULL, invite_comment=EXCLUDED.invite_comment, invited_by=EXCLUDED.invited_by, responded_at=NULL
WHERE project_members.status IN ('INVITED', 'APPLIED', 'REJECTED', 'REMOVED')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	var m Member
	var posID *uuid.UUID
	var invitedBy *uuid.UUID
	if err := s.db.QueryRow(ctx, q, projectID, studentID, comment, userID).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&posID,
		&m.Status,
		&m.InviteComment,
		&invitedBy,
		&m.RespondedAt,
		&m.JoinedAt,
		&m.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Member{}, ErrInvalidInput
		}
		return Member{}, err
	}
	if posID != nil {
		s := posID.String()
		m.PositionID = &s
	}
	if invitedBy != nil {
		s := invitedBy.String()
		m.InvitedBy = &s
	}
	if err := s.ensureProjectRole(ctx, studentID, "INVITED_MEMBER", projectID); err != nil {
		return Member{}, err
	}
	return m, nil
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
INSERT INTO project_members(tenant_id, project_id, user_id, status, position_id, joined_at, invite_comment, invited_by, responded_at)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, 'APPLIED', NULL, NULL, '', NULL, now())
ON CONFLICT (project_id, user_id)
DO UPDATE SET status='APPLIED', position_id=NULL, joined_at=NULL, invite_comment='', invited_by=NULL, responded_at=now()
WHERE project_members.status IN ('APPLIED', 'INVITED', 'REJECTED', 'REMOVED')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	var m Member
	var positionID *uuid.UUID
	var invitedBy *uuid.UUID
	err = s.db.QueryRow(ctx, q, projectID, userID).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&positionID,
		&m.Status,
		&m.InviteComment,
		&invitedBy,
		&m.RespondedAt,
		&m.JoinedAt,
		&m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Member{}, ErrInvalidInput
		}
		return Member{}, err
	}
	if positionID != nil {
		s := positionID.String()
		m.PositionID = &s
	}
	if invitedBy != nil {
		s := invitedBy.String()
		m.InvitedBy = &s
	}
	return m, nil
}

func (s *Service) ListMembers(ctx context.Context, projectID uuid.UUID) ([]Member, error) {
	const q = `
SELECT m.id, m.project_id, m.user_id, m.position_id, m.status, m.invite_comment, m.invited_by, m.responded_at, m.joined_at, m.created_at,
       p.code, p.name,
       COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), '') AS full_name,
       COALESCE(u.email, '') AS email
FROM project_members m
LEFT JOIN project_positions p ON p.id = m.position_id
LEFT JOIN users u ON u.id = m.user_id
LEFT JOIN user_profiles up ON up.user_id = m.user_id
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
		var invitedBy *uuid.UUID
		var posCode *string
		var posName *string
		if err := rows.Scan(
			&m.ID,
			&m.ProjectID,
			&m.UserID,
			&positionID,
			&m.Status,
			&m.InviteComment,
			&invitedBy,
			&m.RespondedAt,
			&m.JoinedAt,
			&m.CreatedAt,
			&posCode,
			&posName,
			&m.FullName,
			&m.Email,
		); err != nil {
			return nil, err
		}
		if positionID != nil {
			s := positionID.String()
			m.PositionID = &s
		}
		if invitedBy != nil {
			s := invitedBy.String()
			m.InvitedBy = &s
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
SET status='ACTIVE', position_id=$3, joined_at=now(), responded_at=now()
	WHERE project_id=$1 AND user_id=$2 AND status IN ('APPLIED', 'ACTIVE')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	var m Member
	var posID *uuid.UUID
	var invitedBy *uuid.UUID
	err = s.db.QueryRow(ctx, q, projectID, memberUserID, positionID).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&posID,
		&m.Status,
		&m.InviteComment,
		&invitedBy,
		&m.RespondedAt,
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
	if invitedBy != nil {
		s := invitedBy.String()
		m.InvitedBy = &s
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

func (s *Service) RespondMemberInvite(ctx context.Context, userID, projectID uuid.UUID, accept bool) (Member, error) {
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

	if accept {
		var invitePositionID *uuid.UUID
		err := s.db.QueryRow(ctx, `
SELECT position_id
FROM project_members
WHERE project_id = $1
  AND user_id = $2
  AND status = 'INVITED'
`, projectID, userID).Scan(&invitePositionID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return Member{}, ErrInviteNotFound
			}
			return Member{}, err
		}
		if invitePositionID != nil {
			if err := s.ensurePositionCapacity(ctx, projectID, *invitePositionID, &userID); err != nil {
				return Member{}, err
			}
		}
	}

	q := `
UPDATE project_members
SET status = CASE WHEN $3::boolean THEN 'ACTIVE' ELSE 'REJECTED' END,
    joined_at = CASE WHEN $3::boolean THEN now() ELSE NULL END,
    responded_at = now()
WHERE project_id = $1
  AND user_id = $2
  AND status = 'INVITED'
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`

	var m Member
	var posID *uuid.UUID
	var invitedBy *uuid.UUID
	err = s.db.QueryRow(ctx, q, projectID, userID, accept).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&posID,
		&m.Status,
		&m.InviteComment,
		&invitedBy,
		&m.RespondedAt,
		&m.JoinedAt,
		&m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Member{}, ErrInviteNotFound
		}
		return Member{}, err
	}
	if posID != nil {
		s := posID.String()
		m.PositionID = &s
	}
	if invitedBy != nil {
		s := invitedBy.String()
		m.InvitedBy = &s
	}
	if accept {
		if err := s.ensureProjectRole(ctx, userID, "MEMBER", projectID); err != nil {
			return Member{}, err
		}
		if err := s.revokeProjectRole(ctx, userID, "INVITED_MEMBER", projectID); err != nil {
			return Member{}, err
		}
	} else {
		if err := s.revokeProjectRole(ctx, userID, "INVITED_MEMBER", projectID); err != nil {
			return Member{}, err
		}
	}
	return m, nil
}

func (s *Service) SearchProfessors(ctx context.Context, userID, projectID uuid.UUID, query string, limit int) ([]ProfessorCandidate, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.invite_professor", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	term := normalizeSearchQuery(query)
	limit = clampLimit(limit, 50)

	const q = `
SELECT u.id,
       COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(u.email, '@', 1)) AS full_name,
       u.email,
       COALESCE(d.code, '') AS department_code
FROM users u
JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN departments d ON d.id = up.department_id
WHERE u.status = 'ACTIVE'
  AND up.faculty_id = $1
  AND u.id <> $4
  AND u.id <> $5
  AND EXISTS (
    SELECT 1
    FROM role_assignments ra
    JOIN roles r ON r.id = ra.role_id
    WHERE ra.user_id = u.id
      AND ra.tenant_id = u.tenant_id
      AND r.code = 'PROFESSOR'
  )
  AND ($2 = '' OR lower(up.full_name) LIKE '%' || $2 || '%' OR lower(u.email) LIKE '%' || $2 || '%')
ORDER BY up.full_name ASC, u.email ASC
LIMIT $3;
`
	rows, err := s.db.Query(ctx, q, p.FacultyID, term, limit, userID, p.CreatedBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]ProfessorCandidate, 0, limit)
	for rows.Next() {
		var item ProfessorCandidate
		if err := rows.Scan(&item.UserID, &item.FullName, &item.Email, &item.DepartmentCode); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Service) AssignProfessor(ctx context.Context, userID, projectID, professorID uuid.UUID) (domain.Project, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.invite_professor", projectID); err != nil {
		return domain.Project{}, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if professorID == userID || professorID == p.CreatedBy {
		return domain.Project{}, ErrInvalidInput
	}

	const verifyProfessorQ = `
SELECT EXISTS (
  SELECT 1
  FROM users u
  JOIN user_profiles up ON up.user_id = u.id
  WHERE u.id = $1
    AND u.status = 'ACTIVE'
    AND up.tenant_id = u.tenant_id
    AND up.faculty_id = $2
    AND EXISTS (
      SELECT 1
      FROM role_assignments ra
      JOIN roles r ON r.id = ra.role_id
      WHERE ra.user_id = u.id
        AND ra.tenant_id = u.tenant_id
        AND r.code = 'PROFESSOR'
    )
) AS ok;
`
	var professorOK bool
	if err := s.db.QueryRow(ctx, verifyProfessorQ, professorID, p.FacultyID).Scan(&professorOK); err != nil {
		return domain.Project{}, err
	}
	if !professorOK {
		return domain.Project{}, ErrInvalidInput
	}

	const q = `
UPDATE projects
SET professor_id = $2,
    professor_review_status = 'PENDING',
    professor_invited_at = now(),
    professor_responded_at = NULL,
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
	return s.projectByID(ctx, projectID)
}

func (s *Service) GetAssignedProfessor(ctx context.Context, userID, projectID uuid.UUID) (*ProfessorCandidate, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.view", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p.ProfessorID == nil {
		return nil, nil
	}

	const q = `
SELECT u.id,
       COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(u.email, '@', 1)) AS full_name,
       u.email,
       COALESCE(d.code, '') AS department_code
FROM users u
JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN departments d ON d.id = up.department_id
WHERE u.id = $1
  AND up.faculty_id = $2
LIMIT 1;
`
	var item ProfessorCandidate
	if err := s.db.QueryRow(ctx, q, *p.ProfessorID, p.FacultyID).Scan(
		&item.UserID,
		&item.FullName,
		&item.Email,
		&item.DepartmentCode,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (s *Service) RespondProfessorInvite(ctx context.Context, professorID, projectID uuid.UUID, accept bool) (domain.Project, error) {
	nextStatus := "REJECTED"
	if accept {
		nextStatus = "ACCEPTED"
	}

	const q = `
UPDATE projects
SET professor_review_status = $3,
    professor_responded_at = now(),
    updated_at = now()
WHERE id = $1
  AND professor_id = $2
  AND professor_review_status = 'PENDING'
RETURNING id, title, description, status, is_public, created_by, professor_id,
          professor_review_status, faculty_id, visibility, group_id, created_at, updated_at;
`
	var p domain.Project
	var profID *uuid.UUID
	var groupID *uuid.UUID
	err := s.db.QueryRow(ctx, q, projectID, professorID, nextStatus).Scan(
		&p.ID,
		&p.Title,
		&p.Description,
		&p.Status,
		&p.IsPublic,
		&p.CreatedBy,
		&profID,
		&p.ProfessorReviewStatus,
		&p.FacultyID,
		&p.Visibility,
		&groupID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, ErrInviteNotFound
		}
		return domain.Project{}, err
	}
	p.ProfessorID = profID
	p.GroupID = groupID

	if accept {
		if err := s.ensureProjectRole(ctx, professorID, "PROJECT_PROFESSOR", projectID); err != nil {
			return domain.Project{}, err
		}
	}
	return p, nil
}

func (s *Service) ListProfessorReviewInvites(ctx context.Context, professorID uuid.UUID, query string, limit int) ([]domain.Project, error) {
	term := normalizeSearchQuery(query)
	limit = clampLimit(limit, 100)

	const q = `
SELECT id, title, description, status, is_public, created_by, professor_id,
       professor_review_status, faculty_id, visibility, group_id, created_at, updated_at
FROM projects
WHERE professor_id = $1
  AND professor_review_status = 'PENDING'
  AND ($2 = '' OR lower(title) LIKE '%' || $2 || '%' OR lower(description) LIKE '%' || $2 || '%')
ORDER BY updated_at DESC
LIMIT $3;
`
	rows, err := s.db.Query(ctx, q, professorID, term, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Project, 0, limit)
	for rows.Next() {
		var p domain.Project
		var profID *uuid.UUID
		var groupID *uuid.UUID
		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Status,
			&p.IsPublic,
			&p.CreatedBy,
			&profID,
			&p.ProfessorReviewStatus,
			&p.FacultyID,
			&p.Visibility,
			&groupID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.ProfessorID = profID
		p.GroupID = groupID
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Service) ListIncomingInvites(ctx context.Context, userID uuid.UUID, limit int) ([]IncomingInvite, error) {
	limit = clampLimit(limit, 100)

	const q = `
SELECT
  p.id,
  p.title,
  p.status,
  pm.invite_comment,
  pm.invited_by,
  COALESCE(NULLIF(TRIM(inv.full_name), ''), split_part(inv_u.email, '@', 1), '') AS inviter_name,
  COALESCE(inv_u.email, '') AS inviter_email,
  pm.created_at,
  pm.responded_at
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
LEFT JOIN users inv_u ON inv_u.id = pm.invited_by
LEFT JOIN user_profiles inv ON inv.user_id = pm.invited_by
WHERE pm.user_id = $1
  AND pm.status = 'INVITED'
ORDER BY pm.created_at DESC
LIMIT $2;
`
	rows, err := s.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]IncomingInvite, 0, limit)
	for rows.Next() {
		var item IncomingInvite
		var projectID uuid.UUID
		var invitedBy *uuid.UUID
		if err := rows.Scan(
			&projectID,
			&item.ProjectTitle,
			&item.ProjectStatus,
			&item.InviteComment,
			&invitedBy,
			&item.InviterName,
			&item.InviterEmail,
			&item.CreatedAt,
			&item.RespondedAt,
		); err != nil {
			return nil, err
		}
		item.ProjectID = projectID.String()
		if invitedBy != nil {
			s := invitedBy.String()
			item.InvitedBy = &s
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Service) ListOutgoingApplications(ctx context.Context, userID uuid.UUID, limit int) ([]OutgoingApplication, error) {
	limit = clampLimit(limit, 100)

	const q = `
SELECT
  p.id,
  p.title,
  p.status,
  pm.status,
  pm.created_at,
  pm.responded_at
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
WHERE pm.user_id = $1
  AND pm.invited_by IS NULL
  AND pm.status IN ('APPLIED', 'REJECTED', 'ACTIVE', 'REMOVED')
ORDER BY COALESCE(pm.responded_at, pm.created_at) DESC
LIMIT $2;
`
	rows, err := s.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]OutgoingApplication, 0, limit)
	for rows.Next() {
		var item OutgoingApplication
		var projectID uuid.UUID
		if err := rows.Scan(
			&projectID,
			&item.ProjectTitle,
			&item.ProjectStatus,
			&item.Status,
			&item.CreatedAt,
			&item.RespondedAt,
		); err != nil {
			return nil, err
		}
		item.ProjectID = projectID.String()
		out = append(out, item)
	}
	return out, rows.Err()
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
	var currentWeight int
	if err := s.db.QueryRow(ctx, `
SELECT COALESCE(SUM(weight), 0)
FROM project_criteria
WHERE project_id = $1
`, projectID).Scan(&currentWeight); err != nil {
		return Criterion{}, err
	}
	if currentWeight+weight > 100 {
		return Criterion{}, fmt.Errorf("%w: total criteria weight exceeds 100", ErrInvalidInput)
	}

	const q = `
INSERT INTO project_criteria(tenant_id, project_id, title, description, weight, created_by)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, $3, $4, $5)
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

func (s *Service) GetGrading(ctx context.Context, userID, projectID uuid.UUID) ([]CriterionGrade, error) {
	if err := s.requireProjectPermission(ctx, userID, "project.view", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	professorID := userID
	if p.ProfessorID != nil {
		professorID = *p.ProfessorID
	}
	const q = `
SELECT criterion_id, is_met, comment, updated_at
FROM project_criterion_reviews
WHERE project_id = $1
  AND professor_id = $2
ORDER BY updated_at DESC;
`
	rows, err := s.db.Query(ctx, q, projectID, professorID)
	if err != nil {
		if isUndefinedRelation(err, "project_criterion_reviews") {
			return []CriterionGrade{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]CriterionGrade, 0, 16)
	for rows.Next() {
		var item CriterionGrade
		if err := rows.Scan(&item.CriterionID, &item.IsMet, &item.Comment, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Service) UpsertGrading(ctx context.Context, userID, projectID uuid.UUID, items []CriterionGrade) ([]CriterionGrade, error) {
	if err := s.requireProjectPermission(ctx, userID, "grading.mark_criteria", projectID); err != nil {
		return nil, err
	}
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.ProjectReview && p.Status != domain.ProjectGrading {
		return nil, ErrInvalidInput
	}

	uniq := make(map[uuid.UUID]CriterionGrade, len(items))
	for _, item := range items {
		cid, err := uuid.Parse(strings.TrimSpace(item.CriterionID))
		if err != nil {
			return nil, ErrInvalidInput
		}
		comment := strings.TrimSpace(item.Comment)
		if len(comment) > 3000 {
			comment = comment[:3000]
		}
		uniq[cid] = CriterionGrade{
			CriterionID: cid.String(),
			IsMet:       item.IsMet,
			Comment:     comment,
		}
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const qExists = `
SELECT EXISTS (
  SELECT 1
  FROM project_criteria
  WHERE id = $1
    AND project_id = $2
) AS ok;
`
	const qUpsert = `
INSERT INTO project_criterion_reviews(tenant_id, project_id, criterion_id, professor_id, is_met, comment)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, $3, $4, $5)
ON CONFLICT (project_id, criterion_id, professor_id)
DO UPDATE SET is_met = EXCLUDED.is_met, comment = EXCLUDED.comment, updated_at = now();
`

	for cid, item := range uniq {
		var ok bool
		if err := tx.QueryRow(ctx, qExists, cid, projectID).Scan(&ok); err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrInvalidInput
		}
		if _, err := tx.Exec(ctx, qUpsert, projectID, cid, userID, item.IsMet, item.Comment); err != nil {
			if isUndefinedRelation(err, "project_criterion_reviews") {
				return nil, ErrInvalidInput
			}
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetGrading(ctx, userID, projectID)
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
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM project_members WHERE project_id=$1 AND status='ACTIVE' AND position_id IS NOT NULL`, projectID).Scan(&activeMembers); err != nil {
		return Readiness{}, err
	}

	var criteriaCount int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM project_criteria WHERE project_id=$1`, projectID).Scan(&criteriaCount); err != nil {
		return Readiness{}, err
	}

	hasProfessor := p.ProfessorID != nil
	professorStatus := strings.ToUpper(strings.TrimSpace(p.ProfessorReviewStatus))
	if professorStatus == "" {
		professorStatus = "NONE"
	}
	professorAccepted := professorStatus == "ACCEPTED"
	canActivate := requiredMembers > 0 && activeMembers >= requiredMembers && professorAccepted && criteriaCount > 0

	return Readiness{
		ProjectID:       projectID.String(),
		Status:          string(p.Status),
		RequiredMembers: requiredMembers,
		ActiveMembers:   activeMembers,
		HasProfessor:    hasProfessor,
		ProfessorStatus: professorStatus,
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
		return domain.Project{}, ready, fmt.Errorf("%w: members=%d/%d professor=%s criteria=%d", ErrProjectNotReady, ready.ActiveMembers, ready.RequiredMembers, ready.ProfessorStatus, ready.CriteriaCount)
	}

	const q = `
UPDATE projects
SET status='ACTIVE', updated_at=now()
WHERE id=$1
  AND status IN ('REVIEW', 'RECRUITMENT');
`
	ct, err := s.db.Exec(ctx, q, projectID)
	if err != nil {
		return domain.Project{}, Readiness{}, err
	}
	if ct.RowsAffected() == 0 {
		return domain.Project{}, Readiness{}, ErrInvalidInput
	}
	p, err := s.projectByID(ctx, projectID)
	return p, ready, err
}

func (s *Service) SubmitProjectForGrading(ctx context.Context, userID, projectID uuid.UUID) (domain.Project, error) {
	if err := s.requireProjectSubmitAccess(ctx, userID, projectID); err != nil {
		return domain.Project{}, err
	}

	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if p.Status != domain.ProjectActive {
		return domain.Project{}, fmt.Errorf("%w: project status must be ACTIVE", ErrInvalidInput)
	}
	if p.ProfessorID == nil {
		return domain.Project{}, fmt.Errorf("%w: professor is not assigned", ErrInvalidInput)
	}
	if strings.ToUpper(strings.TrimSpace(p.ProfessorReviewStatus)) != "ACCEPTED" {
		return domain.Project{}, fmt.Errorf("%w: professor invitation is not accepted", ErrInvalidInput)
	}

	var tasksTotal int
	var tasksDone int
	if err := s.db.QueryRow(ctx, `
SELECT
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE status = 'DONE') AS done
FROM tasks
WHERE project_id = $1;
`, projectID).Scan(&tasksTotal, &tasksDone); err != nil {
		return domain.Project{}, err
	}
	if tasksTotal == 0 {
		return domain.Project{}, fmt.Errorf("%w: at least one task is required", ErrInvalidInput)
	}
	if tasksDone < tasksTotal {
		return domain.Project{}, fmt.Errorf("%w: all tasks must be DONE before grading", ErrInvalidInput)
	}

	ct, err := s.db.Exec(ctx, `
UPDATE projects
SET status = 'GRADING', updated_at = now()
WHERE id = $1
  AND status = 'ACTIVE';
`, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if ct.RowsAffected() == 0 {
		return domain.Project{}, ErrInvalidInput
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) PublishGrading(ctx context.Context, userID, projectID uuid.UUID) (domain.Project, error) {
	if err := s.requireProjectPermission(ctx, userID, "grading.publish", projectID); err != nil {
		return domain.Project{}, err
	}

	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if p.Status != domain.ProjectGrading {
		return domain.Project{}, fmt.Errorf("%w: project status must be GRADING", ErrInvalidInput)
	}
	if p.ProfessorID == nil || *p.ProfessorID != userID {
		return domain.Project{}, domain.ErrForbidden
	}

	var criteriaTotal int
	if err := s.db.QueryRow(ctx, `
SELECT COUNT(*)
FROM project_criteria
WHERE project_id = $1;
`, projectID).Scan(&criteriaTotal); err != nil {
		return domain.Project{}, err
	}
	if criteriaTotal == 0 {
		return domain.Project{}, fmt.Errorf("%w: criteria are not configured", ErrInvalidInput)
	}

	var gradedTotal int
	rowsQ := `
SELECT COUNT(*)
FROM project_criterion_reviews
WHERE project_id = $1
  AND professor_id = $2
  AND is_met IS NOT NULL;
`
	if err := s.db.QueryRow(ctx, rowsQ, projectID, userID).Scan(&gradedTotal); err != nil {
		if isUndefinedRelation(err, "project_criterion_reviews") {
			return domain.Project{}, fmt.Errorf("%w: grading table is missing", ErrInvalidInput)
		}
		return domain.Project{}, err
	}
	if gradedTotal < criteriaTotal {
		return domain.Project{}, fmt.Errorf("%w: grading is incomplete (%d/%d)", ErrInvalidInput, gradedTotal, criteriaTotal)
	}

	ct, err := s.db.Exec(ctx, `
UPDATE projects
SET status = 'ARCHIVE', updated_at = now()
WHERE id = $1
  AND status = 'GRADING';
`, projectID)
	if err != nil {
		return domain.Project{}, err
	}
	if ct.RowsAffected() == 0 {
		return domain.Project{}, ErrInvalidInput
	}
	return s.projectByID(ctx, projectID)
}

func (s *Service) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var ownerID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT created_by FROM projects WHERE id = $1`, projectID).Scan(&ownerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return err
	}
	if ownerID != userID {
		return domain.ErrForbidden
	}

	if _, err := tx.Exec(ctx, `
DELETE FROM role_assignments
WHERE scope_type = 'PROJECT'
  AND scope_id = $1;
`, projectID); err != nil {
		return err
	}

	ct, err := tx.Exec(ctx, `DELETE FROM projects WHERE id = $1`, projectID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return tx.Commit(ctx)
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

	const q = `
INSERT INTO tasks(tenant_id, project_id, title, description, position_id, assignee_user_id, status, created_by, due_at)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;
`
	var taskID uuid.UUID
	err := s.db.QueryRow(ctx, q, projectID, title, description, positionID, assigneeUserID, status, userID, dueAt).Scan(&taskID)
	if err != nil {
		return Task{}, err
	}
	if err := s.appendTaskActivity(ctx, projectID, taskID, &userID, "CREATED", "", status, title, description, nil); err != nil {
		return Task{}, err
	}
	if assigneeUserID != nil {
		if err := s.appendTaskActivity(
			ctx,
			projectID,
			taskID,
			&userID,
			"ASSIGNED",
			status,
			status,
			title,
			fmt.Sprintf("Назначен исполнитель: %s", assigneeUserID.String()),
			nil,
		); err != nil {
			return Task{}, err
		}
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
	var prevStatus string
	var taskTitle string
	if err := s.db.QueryRow(ctx, `
SELECT status, title
FROM tasks
WHERE project_id=$1 AND id=$2
`, projectID, taskID).Scan(&prevStatus, &taskTitle); err != nil {
		return Task{}, err
	}
	prevStatus = strings.ToUpper(strings.TrimSpace(prevStatus))

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
	if prevStatus != status {
		if err := s.appendTaskActivity(ctx, projectID, taskID, &userID, "STATUS_CHANGED", prevStatus, status, taskTitle, "", nil); err != nil {
			return Task{}, err
		}
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
	var prevStatus string
	var taskTitle string
	var prevAssignee *uuid.UUID
	if err := s.db.QueryRow(ctx, `
SELECT position_id, status, title, assignee_user_id
FROM tasks
WHERE project_id = $1 AND id = $2
`, projectID, taskID).Scan(&positionID, &prevStatus, &taskTitle, &prevAssignee); err != nil {
		return Task{}, err
	}
	prevStatus = strings.ToUpper(strings.TrimSpace(prevStatus))
	if err := s.ensureAssigneeMatchesPosition(ctx, projectID, assigneeUserID, positionID); err != nil {
		return Task{}, err
	}

	const q = `
UPDATE tasks
SET assignee_user_id = $3,
    updated_at = now()
WHERE project_id=$1 AND id=$2
RETURNING id;
`
	var outTaskID uuid.UUID
	err := s.db.QueryRow(ctx, q, projectID, taskID, assigneeUserID).Scan(&outTaskID)
	if err != nil {
		return Task{}, err
	}
	if prevAssignee == nil || *prevAssignee != assigneeUserID {
		if err := s.appendTaskActivity(ctx, projectID, taskID, &userID, "ASSIGNED", prevStatus, prevStatus, taskTitle, fmt.Sprintf("Назначен исполнитель: %s", assigneeUserID.String()), nil); err != nil {
			return Task{}, err
		}
	}
	return s.taskByID(ctx, projectID, outTaskID)
}

func (s *Service) ListTaskActivities(ctx context.Context, userID, projectID uuid.UUID, taskID *uuid.UUID) ([]TaskActivity, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.view", projectID); err != nil {
		return nil, err
	}

	baseQ := `
SELECT
  a.id, a.project_id, a.task_id, a.actor_user_id,
  COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), '') AS actor_name,
  COALESCE(u.email, '') AS actor_email,
  a.event_type,
  COALESCE(a.from_status, '') AS from_status,
  COALESCE(a.to_status, '') AS to_status,
  COALESCE(a.title, '') AS title,
  COALESCE(a.comment, '') AS comment,
  a.attachments,
  a.created_at
FROM task_activity_logs a
LEFT JOIN users u ON u.id = a.actor_user_id
LEFT JOIN user_profiles up ON up.user_id = a.actor_user_id
WHERE a.project_id = $1
`
	args := []any{projectID}
	if taskID != nil {
		baseQ += ` AND a.task_id = $2`
		args = append(args, *taskID)
	}
	baseQ += ` ORDER BY a.created_at ASC;`

	rows, err := s.db.Query(ctx, baseQ, args...)
	if err != nil {
		if isUndefinedRelation(err, "task_activity_logs") {
			return []TaskActivity{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]TaskActivity, 0, 32)
	for rows.Next() {
		var item TaskActivity
		var id uuid.UUID
		var pid uuid.UUID
		var tid uuid.UUID
		var actorID *uuid.UUID
		var rawAttachments []byte
		if err := rows.Scan(
			&id,
			&pid,
			&tid,
			&actorID,
			&item.ActorName,
			&item.ActorEmail,
			&item.EventType,
			&item.FromStatus,
			&item.ToStatus,
			&item.Title,
			&item.Comment,
			&rawAttachments,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.ID = id.String()
		item.ProjectID = pid.String()
		item.TaskID = tid.String()
		item.EventType = strings.ToUpper(strings.TrimSpace(item.EventType))
		item.FromStatus = strings.ToUpper(strings.TrimSpace(item.FromStatus))
		item.ToStatus = strings.ToUpper(strings.TrimSpace(item.ToStatus))
		item.Attachments = decodeStringSliceJSON(rawAttachments)
		if actorID != nil {
			s := actorID.String()
			item.ActorUserID = &s
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Service) CompleteTask(ctx context.Context, userID, projectID, taskID uuid.UUID, comment string, attachments []string) (Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.update", projectID); err != nil {
		return Task{}, err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return Task{}, err
	}
	comment = strings.TrimSpace(comment)
	if len(comment) > 3000 {
		comment = comment[:3000]
	}
	attachments = normalizeTaskAttachments(attachments)

	var assigneeID *uuid.UUID
	var currentStatus string
	var taskTitle string
	if err := s.db.QueryRow(ctx, `
SELECT assignee_user_id, status, title
FROM tasks
WHERE project_id=$1 AND id=$2
`, projectID, taskID).Scan(&assigneeID, &currentStatus, &taskTitle); err != nil {
		return Task{}, err
	}
	if assigneeID == nil || *assigneeID != userID {
		return Task{}, domain.ErrForbidden
	}
	currentStatus = strings.ToUpper(strings.TrimSpace(currentStatus))
	if currentStatus != "IN_PROGRESS" {
		return Task{}, ErrInvalidInput
	}

	const upsertSubmissionQ = `
INSERT INTO task_submissions(tenant_id, project_id, task_id, user_id, comment, attachments)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, $3, $4, $5::jsonb)
ON CONFLICT (task_id)
DO UPDATE SET comment = EXCLUDED.comment, attachments = EXCLUDED.attachments, updated_at = now(), submitted_at = now();
`
	if _, err := s.db.Exec(ctx, upsertSubmissionQ, projectID, taskID, userID, comment, encodeStringSliceJSON(attachments)); err != nil {
		if !isUndefinedRelation(err, "task_submissions") {
			return Task{}, err
		}
	}

	const q = `
UPDATE tasks
SET status='DONE', updated_at=now()
WHERE project_id=$1 AND id=$2
RETURNING id;
`
	var outTaskID uuid.UUID
	if err := s.db.QueryRow(ctx, q, projectID, taskID).Scan(&outTaskID); err != nil {
		return Task{}, err
	}

	if err := s.appendTaskActivity(ctx, projectID, taskID, &userID, "COMPLETED", "IN_PROGRESS", "DONE", taskTitle, comment, attachments); err != nil {
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

	var prevStatus string
	var taskTitle string
	if err := s.db.QueryRow(ctx, `
SELECT status, title
FROM tasks
WHERE project_id=$1 AND id=$2
`, projectID, taskID).Scan(&prevStatus, &taskTitle); err != nil {
		return err
	}
	prevStatus = strings.ToUpper(strings.TrimSpace(prevStatus))

	const q = `
UPDATE tasks t
SET assignee_user_id = COALESCE(t.assignee_user_id, $3), status='IN_PROGRESS', updated_at=now()
WHERE t.id=$2
  AND t.project_id=$1
  AND t.status='OPEN'
  AND (
    (
      t.assignee_user_id IS NULL
      AND EXISTS (
        SELECT 1 FROM project_members m
        WHERE m.project_id=$1
          AND m.user_id=$3
          AND m.status='ACTIVE'
          AND m.position_id = t.position_id
      )
    )
    OR t.assignee_user_id = $3
  );
`
	ct, err := s.db.Exec(ctx, q, projectID, taskID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrForbidden
	}
	if err := s.appendTaskActivity(ctx, projectID, taskID, &userID, "CLAIMED", prevStatus, "IN_PROGRESS", taskTitle, "", nil); err != nil {
		return err
	}
	return nil
}
