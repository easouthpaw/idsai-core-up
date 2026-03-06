package postgres

import (
	"context"
	"fmt"

	svc "idsai-core-up/internal/services/admin"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
LEFT JOIN users u ON u.id = p.created_by
LEFT JOIN user_profiles up ON up.user_id = p.created_by
LEFT JOIN faculties f ON f.id = p.faculty_id
LEFT JOIN departments d ON d.id = up.department_id
WHERE ($1::text = '' OR p.status = $1::text)
  AND (
    $2::text = ''
    OR p.title ILIKE '%' || $2::text || '%'
    OR p.description ILIKE '%' || $2::text || '%'
    OR COALESCE(up.full_name, '') ILIKE '%' || $2::text || '%'
    OR COALESCE(u.email, '') ILIKE '%' || $2::text || '%'
  )
ORDER BY p.updated_at DESC, p.created_at DESC;
`
	rows, err := r.db.Query(ctx, q, status, search)
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
	if err := tx.QueryRow(ctx, `
SELECT d.id, d.faculty_id
FROM departments d
WHERE UPPER(d.code) = UPPER($1);
`, in.DepartmentCode).Scan(&departmentID, &facultyID); err != nil {
		return svc.User{}, err
	}

	var userID uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO users(email, password_hash, status)
VALUES ($1, $2, 'ACTIVE')
RETURNING id;
`, in.Email, in.PasswordHash).Scan(&userID); err != nil {
		return svc.User{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO user_profiles(user_id, full_name, faculty_id, department_id)
VALUES ($1, $2, $3, $4);
`, userID, in.FullName, facultyID, departmentID); err != nil {
		return svc.User{}, err
	}

	tag, err := tx.Exec(ctx, `
INSERT INTO role_assignments(user_id, role_id, scope_type, scope_id)
SELECT $1, r.id, 'FACULTY', $2
FROM roles r
WHERE r.code = $3;
`, userID, facultyID, in.RoleCode)
	if err != nil {
		return svc.User{}, err
	}
	if tag.RowsAffected() == 0 {
		return svc.User{}, fmt.Errorf("role not found: %s", in.RoleCode)
	}

	if err := tx.Commit(ctx); err != nil {
		return svc.User{}, err
	}

	return r.getUserByID(ctx, userID)
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
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AdminRepo) UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status string) error {
	tag, err := r.db.Exec(ctx, `
UPDATE projects
SET status = $2, updated_at = now()
WHERE id = $1;
`, projectID, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AdminRepo) getUserByID(ctx context.Context, userID uuid.UUID) (svc.User, error) {
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
	return u, err
}
