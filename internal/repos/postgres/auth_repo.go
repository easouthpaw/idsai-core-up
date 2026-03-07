package postgres

import (
	"context"
	"time"

	svc "idsai-core-up/internal/services/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepo struct {
	db *pgxpool.Pool
}

func NewAuthRepo(db *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{db: db}
}

func (r *AuthRepo) FindTenantByCode(ctx context.Context, tenantCode string) (uuid.UUID, error) {
	const q = `
SELECT id
FROM tenants
WHERE code = $1
  AND status = 'ACTIVE';
`
	var tenantID uuid.UUID
	err := r.db.QueryRow(ctx, q, tenantCode).Scan(&tenantID)
	return tenantID, err
}

func (r *AuthRepo) FindDepartment(ctx context.Context, tenantID uuid.UUID, departmentCode string) (uuid.UUID, uuid.UUID, error) {
	const q = `
SELECT d.id, d.faculty_id
FROM departments d
WHERE d.tenant_id = $1
  AND d.code = $2;
`
	var deptID, facultyID uuid.UUID
	err := r.db.QueryRow(ctx, q, tenantID, departmentCode).Scan(&deptID, &facultyID)
	return deptID, facultyID, err
}

func (r *AuthRepo) CreateUser(ctx context.Context, tenantID uuid.UUID, email, passwordHash, status string) (uuid.UUID, error) {
	const q = `
INSERT INTO users(tenant_id, email, password_hash, status)
VALUES ($1, $2, $3, $4)
RETURNING id;
`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, tenantID, email, passwordHash, status).Scan(&id)
	return id, err
}

func (r *AuthRepo) CreateProfile(ctx context.Context, tenantID, userID uuid.UUID, fullName string, facultyID, departmentID uuid.UUID) error {
	const q = `
INSERT INTO user_profiles(tenant_id, user_id, full_name, faculty_id, department_id)
VALUES ($1, $2, $3, $4, $5);
`
	_, err := r.db.Exec(ctx, q, tenantID, userID, fullName, facultyID, departmentID)
	return err
}

func (r *AuthRepo) GrantStudentFacultyRole(ctx context.Context, tenantID, userID, facultyID uuid.UUID) error {
	const q = `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT $1, $2, r.id, 'FACULTY', $3
FROM roles r
WHERE r.code = 'STUDENT'
  AND NOT EXISTS (
    SELECT 1
    FROM role_assignments ra
    WHERE ra.tenant_id = $1
      AND ra.user_id = $2
      AND ra.role_id = r.id
      AND ra.scope_type = 'FACULTY'
      AND ra.scope_id = $3
      AND ra.expires_at IS NULL
  );
`
	_, err := r.db.Exec(ctx, q, tenantID, userID, facultyID)
	return err
}

func (r *AuthRepo) FindUserByEmail(ctx context.Context, tenantID uuid.UUID, email string) (svc.User, error) {
	const q = `
SELECT
  u.id,
  u.tenant_id,
  u.email,
  u.password_hash,
  u.status,
  p.faculty_id,
  p.department_id,
  p.full_name,
  EXISTS (
    SELECT 1
    FROM role_assignments ra
    JOIN roles r ON r.id = ra.role_id
    WHERE ra.tenant_id = u.tenant_id
      AND ra.user_id = u.id
      AND r.code = 'SUPER_ADMIN'
      AND ra.scope_type = 'SYSTEM'
      AND ra.scope_id IS NULL
      AND (ra.expires_at IS NULL OR ra.expires_at > now())
  ) AS is_admin,
  EXISTS (
    SELECT 1
    FROM role_assignments ra
    JOIN roles r ON r.id = ra.role_id
    WHERE ra.tenant_id = u.tenant_id
      AND ra.user_id = u.id
      AND r.code = 'PROFESSOR'
      AND (ra.expires_at IS NULL OR ra.expires_at > now())
  ) AS is_professor
FROM users u
JOIN user_profiles p ON p.user_id=u.id
WHERE u.tenant_id = $1
  AND u.email = $2;
`
	var out svc.User
	err := r.db.QueryRow(ctx, q, tenantID, email).Scan(
		&out.ID, &out.TenantID, &out.Email, &out.PasswordHash, &out.Status,
		&out.FacultyID, &out.DepartmentID, &out.FullName, &out.IsAdmin, &out.IsProfessor,
	)
	return out, err
}

func (r *AuthRepo) FindUserByID(ctx context.Context, tenantID, userID uuid.UUID) (svc.User, error) {
	const q = `
SELECT
  u.id,
  u.tenant_id,
  u.email,
  u.password_hash,
  u.status,
  p.faculty_id,
  p.department_id,
  p.full_name,
  EXISTS (
    SELECT 1
    FROM role_assignments ra
    JOIN roles r ON r.id = ra.role_id
    WHERE ra.tenant_id = u.tenant_id
      AND ra.user_id = u.id
      AND r.code = 'SUPER_ADMIN'
      AND ra.scope_type = 'SYSTEM'
      AND ra.scope_id IS NULL
      AND (ra.expires_at IS NULL OR ra.expires_at > now())
  ) AS is_admin,
  EXISTS (
    SELECT 1
    FROM role_assignments ra
    JOIN roles r ON r.id = ra.role_id
    WHERE ra.tenant_id = u.tenant_id
      AND ra.user_id = u.id
      AND r.code = 'PROFESSOR'
      AND (ra.expires_at IS NULL OR ra.expires_at > now())
  ) AS is_professor
FROM users u
JOIN user_profiles p ON p.user_id=u.id
WHERE u.tenant_id = $1
  AND u.id = $2;
`
	var out svc.User
	err := r.db.QueryRow(ctx, q, tenantID, userID).Scan(
		&out.ID, &out.TenantID, &out.Email, &out.PasswordHash, &out.Status,
		&out.FacultyID, &out.DepartmentID, &out.FullName, &out.IsAdmin, &out.IsProfessor,
	)
	return out, err
}

func (r *AuthRepo) InsertRefreshToken(ctx context.Context, tenantID, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	const q = `
INSERT INTO refresh_tokens(tenant_id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4);
`
	_, err := r.db.Exec(ctx, q, tenantID, userID, tokenHash, expiresAt)
	return err
}

func (r *AuthRepo) FindRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, uuid.UUID, time.Time, *time.Time, error) {
	const q = `
SELECT tenant_id, user_id, expires_at, revoked_at
FROM refresh_tokens
WHERE token_hash=$1;
`
	var tenantID uuid.UUID
	var userID uuid.UUID
	var exp time.Time
	var revokedAt *time.Time
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(&tenantID, &userID, &exp, &revokedAt)
	return tenantID, userID, exp, revokedAt, err
}

func (r *AuthRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	const q = `
UPDATE refresh_tokens
SET revoked_at=now()
WHERE token_hash=$1 AND revoked_at IS NULL;
`
	_, err := r.db.Exec(ctx, q, tokenHash)
	return err
}
