package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	svc "idsai-core-up/internal/services/admin"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepo struct {
	db *pgxpool.Pool
}

func NewAdminRepo(db *pgxpool.Pool) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) ListUsers(ctx context.Context, roleCode, search string) ([]svc.User, error) {
	const q = `
SELECT
  u.id,
  p.full_name,
  u.email,
  COALESCE(primary_role.role_code, '') AS role_code,
  u.status,
  f.code AS faculty_code,
  d.code AS department_code
FROM users u
JOIN user_profiles p ON p.user_id = u.id
JOIN faculties f ON f.id = p.faculty_id
JOIN departments d ON d.id = p.department_id
LEFT JOIN LATERAL (
  SELECT r.code AS role_code
  FROM role_assignments ra
  JOIN roles r ON r.id = ra.role_id
  WHERE ra.user_id = u.id
    AND (ra.expires_at IS NULL OR ra.expires_at > now())
  ORDER BY
    CASE r.code
      WHEN 'SUPER_ADMIN' THEN 1
      WHEN 'PROFESSOR' THEN 2
      WHEN 'STUDENT' THEN 3
      ELSE 9
    END,
    ra.created_at DESC
  LIMIT 1
) AS primary_role ON TRUE
WHERE ($1::text = '' OR COALESCE(primary_role.role_code, '') = $1::text)
  AND (
    $2::text = ''
    OR u.email ILIKE '%' || $2::text || '%'
    OR p.full_name ILIKE '%' || $2::text || '%'
  )
ORDER BY p.full_name ASC, u.created_at DESC;
`
	rows, err := r.db.Query(ctx, q, roleCode, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]svc.User, 0, 64)
	for rows.Next() {
		var u svc.User
		if err := rows.Scan(
			&u.ID,
			&u.FullName,
			&u.Email,
			&u.RoleCode,
			&u.Status,
			&u.FacultyCode,
			&u.DepartmentCode,
		); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *AdminRepo) ListProjects(ctx context.Context, status, search string) ([]svc.Project, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT
  p.id,
  p.title,
  p.description,
  p.status,
  p.visibility,
  p.is_public,
  p.created_by,
  COALESCE(up.full_name, '') AS author_name,
  COALESCE(u.email, '') AS author_email,
  COALESCE(f.code, '') AS faculty_code,
  COALESCE(d.code, '') AS department_code,
  p.created_at,
  p.updated_at
FROM projects p
LEFT JOIN users u ON u.id = p.created_by AND u.tenant_id = p.tenant_id
LEFT JOIN user_profiles up ON up.user_id = p.created_by AND up.tenant_id = p.tenant_id
LEFT JOIN faculties f ON f.id = p.faculty_id AND f.tenant_id = p.tenant_id
LEFT JOIN departments d ON d.id = up.department_id AND d.tenant_id = p.tenant_id
WHERE p.tenant_id = $1
  AND ($2::text = '' OR p.status = $2::text)
  AND (
    $3::text = ''
    OR p.title ILIKE '%' || $3::text || '%'
    OR p.description ILIKE '%' || $3::text || '%'
    OR COALESCE(up.full_name, '') ILIKE '%' || $3::text || '%'
    OR COALESCE(u.email, '') ILIKE '%' || $3::text || '%'
  )
ORDER BY p.updated_at DESC, p.created_at DESC;
`
	rows, err := r.db.Query(ctx, q, tenantID, status, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]svc.Project, 0, 64)
	for rows.Next() {
		var p svc.Project
		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Status,
			&p.Visibility,
			&p.IsPublic,
			&p.CreatedBy,
			&p.AuthorName,
			&p.AuthorEmail,
			&p.FacultyCode,
			&p.DepartmentCode,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *AdminRepo) CreateUser(ctx context.Context, in svc.CreateUserParams) (svc.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return svc.User{}, err
	}
	defer tx.Rollback(ctx)

	var departmentID uuid.UUID
	var facultyID uuid.UUID
	var tenantID uuid.UUID
	var groupID *uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT
  d.id,
  d.faculty_id,
  d.tenant_id,
  (
    SELECT sg.id
    FROM student_groups sg
    WHERE sg.department_id = d.id
      AND sg.tenant_id = d.tenant_id
    ORDER BY sg.group_number ASC, sg.created_at ASC
    LIMIT 1
  ) AS group_id
FROM departments d
WHERE UPPER(d.code) = UPPER($1);
`, in.DepartmentCode).Scan(&departmentID, &facultyID, &tenantID, &groupID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return svc.User{}, svc.ErrDepartmentNotFound
		}
		return svc.User{}, err
	}
	if in.RoleCode == svc.RoleProfessor {
		groupID = nil
	}

	var userID uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO users(tenant_id, email, password_hash, status, email_verified_at, password_changed_at)
VALUES ($1, $2, $3, 'ACTIVE', now(), now())
RETURNING id;
`, tenantID, in.Email, in.PasswordHash).Scan(&userID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return svc.User{}, svc.ErrUserExists
		}
		return svc.User{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO user_profiles(tenant_id, user_id, full_name, faculty_id, department_id, group_id)
VALUES ($1, $2, $3, $4, $5, $6);
`, tenantID, userID, in.FullName, facultyID, departmentID, groupID); err != nil {
		return svc.User{}, err
	}

	tag, err := tx.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT $1, $2, r.id, 'FACULTY', $3
FROM roles r
WHERE r.code = $4;
`, tenantID, userID, facultyID, in.RoleCode)
	if err != nil {
		return svc.User{}, err
	}
	if tag.RowsAffected() == 0 {
		return svc.User{}, fmt.Errorf("role not found: %s", in.RoleCode)
	}

	if err := tx.Commit(ctx); err != nil {
		return svc.User{}, err
	}

	return r.GetUserByID(ctx, userID)
}

func (r *AdminRepo) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	tag, err := r.db.Exec(ctx, `
UPDATE users
SET status = $2
WHERE id = $1;
`, userID, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrUserNotFound
	}
	return nil
}

func (r *AdminRepo) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleCode string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var tenantID uuid.UUID
	var facultyID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT u.tenant_id, p.faculty_id
FROM users u
JOIN user_profiles p ON p.user_id = u.id
WHERE u.id = $1;
`, userID).Scan(&tenantID, &facultyID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return svc.ErrUserNotFound
		}
		return err
	}

	_, err = tx.Exec(ctx, `
DELETE FROM role_assignments ra
USING roles r
WHERE ra.user_id = $1
  AND ra.tenant_id = $2
  AND ra.scope_type = 'FACULTY'
  AND ra.scope_id = $3
  AND ra.role_id = r.id
  AND r.code IN ('STUDENT', 'PROFESSOR');
`, userID, tenantID, facultyID)
	if err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT $1, $2, r.id, 'FACULTY', $3
FROM roles r
WHERE r.code = $4;
`, tenantID, userID, facultyID, roleCode)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("role not found: %s", roleCode)
	}

	return tx.Commit(ctx)
}

func (r *AdminRepo) UpdateUserPasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	tag, err := r.db.Exec(ctx, `
UPDATE users
SET password_hash = $2,
    password_changed_at = now()
WHERE id = $1;
`, userID, passwordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrUserNotFound
	}
	return nil
}

func (r *AdminRepo) RevokeUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
UPDATE refresh_tokens
SET revoked_at = now()
WHERE user_id = $1
  AND revoked_at IS NULL;
`, userID)
	return err
}

func (r *AdminRepo) UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status string) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	tag, err := r.db.Exec(ctx, `
UPDATE projects
SET status = $2, updated_at = now()
WHERE id = $1
  AND tenant_id = $3;
`, projectID, status, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrProjectNotFound
	}
	return nil
}

func (r *AdminRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Keep RBAC clean because role_assignments.user_id has no FK to users.
	if _, err := tx.Exec(ctx, `
DELETE FROM role_assignments
WHERE user_id = $1;
`, userID); err != nil {
		return err
	}

	// Keep team memberships clean; tasks FK will set assignee_user_id to NULL.
	if _, err := tx.Exec(ctx, `
DELETE FROM project_members
WHERE user_id = $1;
`, userID); err != nil {
		return err
	}

	// Unlink user from assigned professor field in projects.
	if _, err := tx.Exec(ctx, `
UPDATE projects
SET professor_id = NULL, updated_at = now()
WHERE professor_id = $1;
`, userID); err != nil {
		return err
	}

	// project_files has uploaded_by FK ON DELETE RESTRICT.
	if _, err := tx.Exec(ctx, `
DELETE FROM project_files
WHERE uploaded_by = $1;
`, userID); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
DELETE FROM users
WHERE id = $1;
`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrUserNotFound
	}

	return tx.Commit(ctx)
}

func (r *AdminRepo) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Keep RBAC clean because role_assignments.scope_id has no FK to projects.
	if _, err := tx.Exec(ctx, `
DELETE FROM role_assignments
WHERE tenant_id = $1
  AND scope_type = 'PROJECT'
  AND scope_id = $2;
`, tenantID, projectID); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
DELETE FROM projects
WHERE tenant_id = $1
  AND id = $2;
`, tenantID, projectID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrProjectNotFound
	}

	return tx.Commit(ctx)
}

func (r *AdminRepo) GetProjectByID(ctx context.Context, projectID uuid.UUID) (svc.Project, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return svc.Project{}, err
	}

	const q = `
SELECT
  p.id,
  p.title,
  p.description,
  p.status,
  p.visibility,
  p.is_public,
  p.created_by,
  COALESCE(up.full_name, '') AS author_name,
  COALESCE(u.email, '') AS author_email,
  COALESCE(f.code, '') AS faculty_code,
  COALESCE(d.code, '') AS department_code,
  p.created_at,
  p.updated_at
FROM projects p
LEFT JOIN users u ON u.id = p.created_by AND u.tenant_id = p.tenant_id
LEFT JOIN user_profiles up ON up.user_id = p.created_by AND up.tenant_id = p.tenant_id
LEFT JOIN faculties f ON f.id = p.faculty_id AND f.tenant_id = p.tenant_id
LEFT JOIN departments d ON d.id = up.department_id AND d.tenant_id = p.tenant_id
WHERE p.tenant_id = $1
  AND p.id = $2;
`
	var p svc.Project
	err = r.db.QueryRow(ctx, q, tenantID, projectID).Scan(
		&p.ID,
		&p.Title,
		&p.Description,
		&p.Status,
		&p.Visibility,
		&p.IsPublic,
		&p.CreatedBy,
		&p.AuthorName,
		&p.AuthorEmail,
		&p.FacultyCode,
		&p.DepartmentCode,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return svc.Project{}, svc.ErrProjectNotFound
	}
	return p, err
}

func (r *AdminRepo) GetProjectObservation(ctx context.Context, projectID uuid.UUID) (svc.ProjectObservation, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return svc.ProjectObservation{}, err
	}

	project, err := r.GetProjectByID(ctx, projectID)
	if err != nil {
		return svc.ProjectObservation{}, err
	}

	ob := svc.ProjectObservation{
		Project:   project,
		Positions: make([]svc.ProjectPosition, 0, 8),
		Members:   make([]svc.ProjectMember, 0, 16),
		Tasks:     make([]svc.ProjectTask, 0, 32),
		Criteria:  make([]svc.ProjectCriterion, 0, 16),
	}

	posRows, err := r.db.Query(ctx, `
SELECT id, code, name, capacity
FROM project_positions
WHERE tenant_id = $1
  AND project_id = $2
ORDER BY created_at ASC;
`, tenantID, projectID)
	if err != nil {
		return svc.ProjectObservation{}, err
	}
	for posRows.Next() {
		var p svc.ProjectPosition
		if err := posRows.Scan(&p.ID, &p.Code, &p.Name, &p.Capacity); err != nil {
			posRows.Close()
			return svc.ProjectObservation{}, err
		}
		ob.Positions = append(ob.Positions, p)
	}
	if err := posRows.Err(); err != nil {
		posRows.Close()
		return svc.ProjectObservation{}, err
	}
	posRows.Close()

	memberRows, err := r.db.Query(ctx, `
SELECT
  pm.user_id,
  COALESCE(up.full_name, '') AS full_name,
  COALESCE(u.email, '') AS email,
  COALESCE(primary_role.role_code, '') AS role_code,
  pm.status,
  COALESCE(pp.code, '') AS position_code,
  COALESCE(pp.name, '') AS position_name,
  pm.joined_at,
  pm.responded_at
FROM project_members pm
LEFT JOIN users u ON u.id = pm.user_id AND u.tenant_id = pm.tenant_id
LEFT JOIN user_profiles up ON up.user_id = pm.user_id AND up.tenant_id = pm.tenant_id
LEFT JOIN project_positions pp ON pp.id = pm.position_id AND pp.tenant_id = pm.tenant_id
LEFT JOIN LATERAL (
  SELECT r.code AS role_code
  FROM role_assignments ra
  JOIN roles r ON r.id = ra.role_id
  WHERE ra.user_id = pm.user_id
    AND ra.tenant_id = pm.tenant_id
    AND (ra.expires_at IS NULL OR ra.expires_at > now())
  ORDER BY
    CASE r.code
      WHEN 'SUPER_ADMIN' THEN 1
      WHEN 'PROFESSOR' THEN 2
      WHEN 'STUDENT' THEN 3
      WHEN 'TEAM_LEAD' THEN 4
      WHEN 'MEMBER' THEN 5
      ELSE 9
    END,
    ra.created_at DESC
  LIMIT 1
) AS primary_role ON TRUE
WHERE pm.tenant_id = $1
  AND pm.project_id = $2
ORDER BY pm.created_at ASC;
`, tenantID, projectID)
	if err != nil {
		return svc.ProjectObservation{}, err
	}
	for memberRows.Next() {
		var m svc.ProjectMember
		if err := memberRows.Scan(
			&m.UserID,
			&m.FullName,
			&m.Email,
			&m.RoleCode,
			&m.Status,
			&m.PositionCode,
			&m.PositionName,
			&m.JoinedAt,
			&m.RespondedAt,
		); err != nil {
			memberRows.Close()
			return svc.ProjectObservation{}, err
		}
		ob.Members = append(ob.Members, m)
	}
	if err := memberRows.Err(); err != nil {
		memberRows.Close()
		return svc.ProjectObservation{}, err
	}
	memberRows.Close()

	taskRows, err := r.db.Query(ctx, `
SELECT
  t.id,
  t.title,
  t.status,
  pp.code AS position_code,
  t.assignee_user_id,
  COALESCE(assignee.full_name, '') AS assignee_name,
  t.due_at,
  t.updated_at
FROM tasks t
JOIN project_positions pp ON pp.id = t.position_id AND pp.tenant_id = t.tenant_id
LEFT JOIN user_profiles assignee ON assignee.user_id = t.assignee_user_id AND assignee.tenant_id = t.tenant_id
WHERE t.tenant_id = $1
  AND t.project_id = $2
ORDER BY t.created_at DESC
LIMIT 200;
`, tenantID, projectID)
	if err != nil {
		return svc.ProjectObservation{}, err
	}
	for taskRows.Next() {
		var t svc.ProjectTask
		if err := taskRows.Scan(
			&t.ID,
			&t.Title,
			&t.Status,
			&t.PositionCode,
			&t.AssigneeUserID,
			&t.AssigneeName,
			&t.DueAt,
			&t.UpdatedAt,
		); err != nil {
			taskRows.Close()
			return svc.ProjectObservation{}, err
		}
		ob.Tasks = append(ob.Tasks, t)
	}
	if err := taskRows.Err(); err != nil {
		taskRows.Close()
		return svc.ProjectObservation{}, err
	}
	taskRows.Close()

	criteriaQuery := `
SELECT
  c.id,
  c.title,
  c.weight,
  c.created_by,
  c.created_at,
  cr.is_met,
  COALESCE(cr.comment, '') AS comment,
  cr.updated_at
FROM project_criteria c
LEFT JOIN projects p ON p.id = c.project_id AND p.tenant_id = c.tenant_id
LEFT JOIN project_criterion_reviews cr
  ON cr.project_id = c.project_id
 AND cr.criterion_id = c.id
 AND cr.professor_id = p.professor_id
 AND cr.tenant_id = c.tenant_id
WHERE c.tenant_id = $1
  AND c.project_id = $2
ORDER BY c.created_at ASC;
`
	criteriaRows, err := r.db.Query(ctx, criteriaQuery, tenantID, projectID)
	if err != nil {
		if !isUndefinedRelation(err, "project_criterion_reviews") {
			return svc.ProjectObservation{}, err
		}
		criteriaRows, err = r.db.Query(ctx, `
SELECT id, title, weight, created_by, created_at
FROM project_criteria
WHERE tenant_id = $1
  AND project_id = $2
ORDER BY created_at ASC;
`, tenantID, projectID)
		if err != nil {
			return svc.ProjectObservation{}, err
		}
		for criteriaRows.Next() {
			var c svc.ProjectCriterion
			if err := criteriaRows.Scan(&c.ID, &c.Title, &c.Weight, &c.CreatedBy, &c.CreatedAt); err != nil {
				criteriaRows.Close()
				return svc.ProjectObservation{}, err
			}
			ob.Criteria = append(ob.Criteria, c)
		}
		if err := criteriaRows.Err(); err != nil {
			criteriaRows.Close()
			return svc.ProjectObservation{}, err
		}
		criteriaRows.Close()
	} else {
		for criteriaRows.Next() {
			var c svc.ProjectCriterion
			if err := criteriaRows.Scan(
				&c.ID,
				&c.Title,
				&c.Weight,
				&c.CreatedBy,
				&c.CreatedAt,
				&c.IsMet,
				&c.Comment,
				&c.UpdatedAt,
			); err != nil {
				criteriaRows.Close()
				return svc.ProjectObservation{}, err
			}
			ob.Criteria = append(ob.Criteria, c)
		}
		if err := criteriaRows.Err(); err != nil {
			criteriaRows.Close()
			return svc.ProjectObservation{}, err
		}
		criteriaRows.Close()
	}

	ob.Summary = svc.ProjectObservationSummary{
		MembersTotal:  len(ob.Members),
		TasksTotal:    len(ob.Tasks),
		CriteriaTotal: len(ob.Criteria),
	}
	for _, m := range ob.Members {
		switch strings.ToUpper(strings.TrimSpace(m.Status)) {
		case "ACTIVE":
			ob.Summary.MembersActive++
		case "APPLIED":
			ob.Summary.MembersApplied++
		case "INVITED":
			ob.Summary.MembersInvited++
		}
	}
	for _, task := range ob.Tasks {
		if strings.EqualFold(task.Status, "DONE") {
			ob.Summary.TasksDone++
		}
	}

	return ob, nil
}

func (r *AdminRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (svc.User, error) {
	const q = `
SELECT
  u.id,
  p.full_name,
  u.email,
  COALESCE(primary_role.role_code, '') AS role_code,
  u.status,
  f.code AS faculty_code,
  d.code AS department_code
FROM users u
JOIN user_profiles p ON p.user_id = u.id
JOIN faculties f ON f.id = p.faculty_id
JOIN departments d ON d.id = p.department_id
LEFT JOIN LATERAL (
  SELECT r.code AS role_code
  FROM role_assignments ra
  JOIN roles r ON r.id = ra.role_id
  WHERE ra.user_id = u.id
    AND (ra.expires_at IS NULL OR ra.expires_at > now())
  ORDER BY
    CASE r.code
      WHEN 'SUPER_ADMIN' THEN 1
      WHEN 'PROFESSOR' THEN 2
      WHEN 'STUDENT' THEN 3
      ELSE 9
    END,
    ra.created_at DESC
  LIMIT 1
) AS primary_role ON TRUE
WHERE u.id = $1;
`

	var u svc.User
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&u.ID,
		&u.FullName,
		&u.Email,
		&u.RoleCode,
		&u.Status,
		&u.FacultyCode,
		&u.DepartmentCode,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return svc.User{}, svc.ErrUserNotFound
	}
	return u, err
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
