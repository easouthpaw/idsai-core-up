package postgres

import (
	"context"
	"errors"
	"time"

	svc "idsai-core-up/internal/services/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, svc.ErrNotFound
	}
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
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, svc.ErrNotFound
	}
	return deptID, facultyID, err
}

func (r *AuthRepo) CreateUser(ctx context.Context, in svc.CreateUserParams) (uuid.UUID, error) {
	const q = `
INSERT INTO users(tenant_id, email, password_hash, status, email_verified_at, password_changed_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;
`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, in.TenantID, in.Email, in.PasswordHash, in.Status, in.EmailVerifiedAt, in.PasswordChangedAt).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, svc.ErrUserExists
		}
	}
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
  COALESCE(u.pending_email, ''),
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
  ) AS is_professor,
  u.email_verified_at,
  u.pending_email_requested_at,
  COALESCE(u.avatar_key, ''),
  u.avatar_updated_at
FROM users u
JOIN user_profiles p ON p.user_id=u.id
WHERE u.tenant_id = $1
  AND u.email = $2;
`
	return r.scanUser(ctx, q, tenantID, email)
}

func (r *AuthRepo) FindUserByID(ctx context.Context, tenantID, userID uuid.UUID) (svc.User, error) {
	const q = `
SELECT
  u.id,
  u.tenant_id,
  u.email,
  COALESCE(u.pending_email, ''),
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
  ) AS is_professor,
  u.email_verified_at,
  u.pending_email_requested_at,
  COALESCE(u.avatar_key, ''),
  u.avatar_updated_at
FROM users u
JOIN user_profiles p ON p.user_id=u.id
WHERE u.tenant_id = $1
  AND u.id = $2;
`
	return r.scanUser(ctx, q, tenantID, userID)
}

func (r *AuthRepo) UpdateUserPasswordHash(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string, changedAt time.Time) error {
	tag, err := r.db.Exec(ctx, `
UPDATE users
SET password_hash = $3,
    password_changed_at = $4
WHERE tenant_id = $1
  AND id = $2;
`, tenantID, userID, passwordHash, changedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrNotFound
	}
	return nil
}

func (r *AuthRepo) UpdateUserProfileFullName(ctx context.Context, tenantID, userID uuid.UUID, fullName string) error {
	tag, err := r.db.Exec(ctx, `
UPDATE user_profiles
SET full_name = $3
WHERE tenant_id = $1
  AND user_id = $2;
`, tenantID, userID, fullName)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrNotFound
	}
	return nil
}

func (r *AuthRepo) MarkUserEmailVerified(ctx context.Context, tenantID, userID uuid.UUID, verifiedAt time.Time) error {
	tag, err := r.db.Exec(ctx, `
UPDATE users
SET email_verified_at = COALESCE(email_verified_at, $3),
    status = CASE WHEN status = 'PENDING' THEN 'ACTIVE' ELSE status END
WHERE tenant_id = $1
  AND id = $2;
`, tenantID, userID, verifiedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrNotFound
	}
	return nil
}

func (r *AuthRepo) IsEmailInUse(ctx context.Context, tenantID, excludeUserID uuid.UUID, email string) (bool, error) {
	var inUse bool
	err := r.db.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
  FROM users
  WHERE tenant_id = $1
    AND id <> $2
    AND (
      email = $3
      OR pending_email = $3
    )
) AS in_use;
`, tenantID, excludeUserID, email).Scan(&inUse)
	return inUse, err
}

func (r *AuthRepo) SetPendingEmail(ctx context.Context, tenantID, userID uuid.UUID, pendingEmail string, requestedAt time.Time) error {
	tag, err := r.db.Exec(ctx, `
UPDATE users
SET pending_email = $3,
    pending_email_requested_at = $4
WHERE tenant_id = $1
  AND id = $2;
`, tenantID, userID, pendingEmail, requestedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrNotFound
	}
	return nil
}

func (r *AuthRepo) ActivatePendingEmail(ctx context.Context, tenantID, userID uuid.UUID, activatedAt time.Time) (string, error) {
	var email string
	err := r.db.QueryRow(ctx, `
UPDATE users
SET email = pending_email,
    pending_email = NULL,
    pending_email_requested_at = NULL,
    email_verified_at = $3
WHERE tenant_id = $1
  AND id = $2
  AND pending_email IS NOT NULL
RETURNING email;
`, tenantID, userID, activatedAt).Scan(&email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", svc.ErrNoPendingEmail
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", svc.ErrEmailInUse
		}
		return "", err
	}
	return email, nil
}

func (r *AuthRepo) UpdateUserAvatarKey(ctx context.Context, tenantID, userID uuid.UUID, avatarKey *string, updatedAt time.Time) error {
	tag, err := r.db.Exec(ctx, `
UPDATE users
SET avatar_key = $3,
    avatar_updated_at = $4
WHERE tenant_id = $1
  AND id = $2;
`, tenantID, userID, avatarKey, updatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrNotFound
	}
	return nil
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
WHERE token_hash = $1;
`
	var tenantID uuid.UUID
	var userID uuid.UUID
	var exp time.Time
	var revokedAt *time.Time
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(&tenantID, &userID, &exp, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, time.Time{}, nil, svc.ErrNotFound
	}
	return tenantID, userID, exp, revokedAt, err
}

func (r *AuthRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	const q = `
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = $1
  AND revoked_at IS NULL;
`
	_, err := r.db.Exec(ctx, q, tokenHash)
	return err
}

func (r *AuthRepo) RevokeUserRefreshTokens(ctx context.Context, tenantID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
UPDATE refresh_tokens
SET revoked_at = now()
WHERE tenant_id = $1
  AND user_id = $2
  AND revoked_at IS NULL;
`, tenantID, userID)
	return err
}

func (r *AuthRepo) InsertAuthToken(ctx context.Context, tenantID, userID uuid.UUID, purpose, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO auth_tokens(tenant_id, user_id, purpose, token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5);
`, tenantID, userID, purpose, tokenHash, expiresAt)
	return err
}

func (r *AuthRepo) FindAuthToken(ctx context.Context, purpose, tokenHash string) (svc.AuthTokenRecord, error) {
	var record svc.AuthTokenRecord
	err := r.db.QueryRow(ctx, `
SELECT id, tenant_id, user_id, purpose, token_hash, expires_at, consumed_at
FROM auth_tokens
WHERE purpose = $1
  AND token_hash = $2;
`, purpose, tokenHash).Scan(
		&record.ID,
		&record.TenantID,
		&record.UserID,
		&record.Purpose,
		&record.TokenHash,
		&record.ExpiresAt,
		&record.ConsumedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return svc.AuthTokenRecord{}, svc.ErrNotFound
	}
	return record, err
}

func (r *AuthRepo) ConsumeAuthToken(ctx context.Context, tokenID uuid.UUID, consumedAt time.Time) error {
	tag, err := r.db.Exec(ctx, `
UPDATE auth_tokens
SET consumed_at = $2
WHERE id = $1
  AND consumed_at IS NULL;
`, tokenID, consumedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return svc.ErrNotFound
	}
	return nil
}

func (r *AuthRepo) InvalidateAuthTokens(ctx context.Context, tenantID, userID uuid.UUID, purpose string) error {
	_, err := r.db.Exec(ctx, `
UPDATE auth_tokens
SET consumed_at = now()
WHERE tenant_id = $1
  AND user_id = $2
  AND purpose = $3
  AND consumed_at IS NULL;
`, tenantID, userID, purpose)
	return err
}

func (r *AuthRepo) scanUser(ctx context.Context, q string, args ...any) (svc.User, error) {
	var out svc.User
	err := r.db.QueryRow(ctx, q, args...).Scan(
		&out.ID,
		&out.TenantID,
		&out.Email,
		&out.PendingEmail,
		&out.PasswordHash,
		&out.Status,
		&out.FacultyID,
		&out.DepartmentID,
		&out.FullName,
		&out.IsAdmin,
		&out.IsProfessor,
		&out.EmailVerifiedAt,
		&out.PendingEmailAt,
		&out.AvatarKey,
		&out.AvatarUpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return svc.User{}, svc.ErrNotFound
	}
	return out, err
}
