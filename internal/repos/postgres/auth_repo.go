package postgres

import (
	"context"
	"errors"
	"strings"
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
		return uuid.Nil, uuid.Nil, svc.ErrDepartmentNotFound
	}
	return deptID, facultyID, err
}

func (r *AuthRepo) FindDepartmentInFaculty(ctx context.Context, tenantID, facultyID uuid.UUID, departmentCode string) (uuid.UUID, error) {
	const q = `
SELECT d.id
FROM departments d
WHERE d.tenant_id = $1
  AND d.faculty_id = $2
  AND d.code = $3;
`
	var deptID uuid.UUID
	err := r.db.QueryRow(ctx, q, tenantID, facultyID, departmentCode).Scan(&deptID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, svc.ErrDepartmentNotFound
	}
	return deptID, err
}

func (r *AuthRepo) FindSchoolRegistrationScope(ctx context.Context, tenantID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	const q = `
SELECT f.id, d.id
FROM tenants t
JOIN faculties f
  ON f.tenant_id = t.id
 AND f.code = t.code || '_SCHOOL'
JOIN departments d
  ON d.tenant_id = t.id
 AND d.faculty_id = f.id
 AND d.code = 'CLASS'
WHERE t.id = $1;
`
	var facultyID, departmentID uuid.UUID
	err := r.db.QueryRow(ctx, q, tenantID).Scan(&facultyID, &departmentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, svc.ErrSchoolRegistrationUnavailable
	}
	return facultyID, departmentID, err
}

func (r *AuthRepo) FindGroupByCodeInDepartment(ctx context.Context, tenantID, departmentID uuid.UUID, groupCode string) (uuid.UUID, error) {
	const q = `
SELECT id
FROM student_groups
WHERE tenant_id = $1
  AND department_id = $2
  AND UPPER(group_code) = UPPER($3);
`
	var groupID uuid.UUID
	err := r.db.QueryRow(ctx, q, tenantID, departmentID, groupCode).Scan(&groupID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, svc.ErrGroupNotFound
	}
	return groupID, err
}

func (r *AuthRepo) ListFaculties(ctx context.Context, tenantID uuid.UUID) ([]svc.Faculty, error) {
	const q = `
SELECT f.id, f.code, f.name, f.created_at
FROM faculties f
JOIN tenants t ON t.id = f.tenant_id
WHERE f.tenant_id = $1
  AND f.code <> t.code || '_SCHOOL'
ORDER BY f.name ASC;
`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]svc.Faculty, 0, 8)
	for rows.Next() {
		var item svc.Faculty
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *AuthRepo) SuggestKnownInstitutions(ctx context.Context, tenantID uuid.UUID, educationType, query string, limit int) ([]svc.InstitutionSuggestion, error) {
	if limit <= 0 {
		limit = 5
	}

	const q = `
SELECT
  up.institution_provider,
  up.institution_external_id,
  up.institution_name,
  up.institution_address,
  COUNT(*) AS usage_count
FROM user_profiles up
JOIN faculties f
  ON f.id = up.faculty_id
 AND f.tenant_id = up.tenant_id
JOIN tenants t
  ON t.id = up.tenant_id
WHERE up.tenant_id = $1
  AND up.institution_name <> ''
  AND (
    up.institution_name ILIKE '%' || $3::text || '%'
    OR up.institution_address ILIKE '%' || $3::text || '%'
  )
  AND (
    ($2 = 'SCHOOL' AND f.code = t.code || '_SCHOOL')
    OR ($2 <> 'SCHOOL' AND f.code <> t.code || '_SCHOOL')
  )
GROUP BY
  up.institution_provider,
  up.institution_external_id,
  up.institution_name,
  up.institution_address
ORDER BY
  CASE
    WHEN LOWER(up.institution_name) = LOWER($3::text) THEN 0
    WHEN LOWER(up.institution_name) LIKE LOWER($3::text) || '%' THEN 1
    WHEN LOWER(up.institution_address) LIKE LOWER($3::text) || '%' THEN 2
    ELSE 3
  END,
  usage_count DESC,
  up.institution_name ASC
LIMIT $4;
`

	rows, err := r.db.Query(ctx, q, tenantID, strings.ToUpper(strings.TrimSpace(educationType)), strings.TrimSpace(query), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]svc.InstitutionSuggestion, 0, limit)
	for rows.Next() {
		var item svc.InstitutionSuggestion
		var usageCount int
		if err := rows.Scan(&item.Provider, &item.ExternalID, &item.Name, &item.Address, &usageCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *AuthRepo) CreateGroupInDepartment(
	ctx context.Context,
	tenantID, facultyID, departmentID uuid.UUID,
	groupCode string,
	groupNumber int,
) (uuid.UUID, error) {
	const q = `
INSERT INTO student_groups(
  tenant_id,
  faculty_id,
  department_id,
  code,
  name,
  year,
  group_code,
  group_number,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $4, NULL, $4, $5, now(), now())
ON CONFLICT (tenant_id, group_code)
DO UPDATE SET updated_at = student_groups.updated_at
RETURNING id;
`
	var groupID uuid.UUID
	err := r.db.QueryRow(ctx, q, tenantID, facultyID, departmentID, groupCode, groupNumber).Scan(&groupID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return uuid.Nil, svc.ErrGroupMismatch
		}
	}
	return groupID, err
}

func (r *AuthRepo) ListDepartments(ctx context.Context, tenantID uuid.UUID, educationType string) ([]svc.Department, error) {
	const q = `
SELECT d.id, d.faculty_id, d.code, d.name, d.created_at
FROM departments d
JOIN faculties f ON f.id = d.faculty_id AND f.tenant_id = d.tenant_id
JOIN tenants t ON t.id = d.tenant_id
WHERE d.tenant_id = $1
  AND (
    ($2 = 'SCHOOL' AND f.code = t.code || '_SCHOOL')
    OR ($2 <> 'SCHOOL' AND f.code <> t.code || '_SCHOOL')
  )
ORDER BY d.name ASC;
`
	rows, err := r.db.Query(ctx, q, tenantID, educationType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]svc.Department, 0, 16)
	for rows.Next() {
		var item svc.Department
		if err := rows.Scan(&item.ID, &item.FacultyID, &item.Code, &item.Name, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.ShortCode = item.Code
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *AuthRepo) ListGroupsByDepartmentCode(ctx context.Context, tenantID uuid.UUID, departmentCode string) ([]svc.StudentGroup, error) {
	const q = `
SELECT sg.id, sg.department_id, sg.group_code, sg.group_number, sg.created_at, sg.updated_at
FROM student_groups sg
JOIN departments d ON d.id = sg.department_id
WHERE sg.tenant_id = $1
  AND d.tenant_id = $1
  AND UPPER(d.code) = UPPER($2)
ORDER BY sg.group_number ASC, sg.group_code ASC;
`
	rows, err := r.db.Query(ctx, q, tenantID, departmentCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]svc.StudentGroup, 0, 32)
	for rows.Next() {
		var item svc.StudentGroup
		if err := rows.Scan(
			&item.ID,
			&item.DepartmentID,
			&item.GroupCode,
			&item.GroupNumber,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
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

func (r *AuthRepo) CreateProfile(ctx context.Context, tenantID, userID uuid.UUID, fullName string, facultyID, departmentID uuid.UUID, groupID *uuid.UUID, institution svc.InstitutionSelection) error {
	const q = `
INSERT INTO user_profiles(
  tenant_id,
  user_id,
  full_name,
  faculty_id,
  department_id,
  group_id,
  institution_provider,
  institution_external_id,
  institution_name,
  institution_address
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
`
	_, err := r.db.Exec(
		ctx,
		q,
		tenantID,
		userID,
		fullName,
		facultyID,
		departmentID,
		groupID,
		institution.Provider,
		institution.ExternalID,
		institution.Name,
		institution.Address,
	)
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
	u.password_changed_at,
	u.status,
	p.faculty_id,
  COALESCE(f.code, ''),
  p.department_id,
  COALESCE(d.code, ''),
  p.group_id,
  COALESCE(sg.group_code, ''),
  sg.group_number,
  COALESCE(p.institution_provider, ''),
  COALESCE(p.institution_external_id, ''),
  COALESCE(p.institution_name, ''),
  COALESCE(p.institution_address, ''),
  p.full_name,
  COALESCE(p.headline, ''),
  COALESCE(p.about, ''),
  COALESCE(p.preferred_role, ''),
  COALESCE(p.semester, ''),
  COALESCE(p.availability, ''),
  COALESCE(p.goals, ''),
  COALESCE(p.github_url, ''),
  COALESCE(p.telegram_username, ''),
  COALESCE(p.portfolio_url, ''),
  COALESCE(p.stacks, ARRAY[]::TEXT[]),
  COALESCE(p.interests, ARRAY[]::TEXT[]),
  p.updated_at,
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
JOIN faculties f ON f.id = p.faculty_id
JOIN departments d ON d.id = p.department_id
LEFT JOIN student_groups sg ON sg.id = p.group_id
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
	u.password_changed_at,
	u.status,
	p.faculty_id,
  COALESCE(f.code, ''),
  p.department_id,
  COALESCE(d.code, ''),
  p.group_id,
  COALESCE(sg.group_code, ''),
  sg.group_number,
  COALESCE(p.institution_provider, ''),
  COALESCE(p.institution_external_id, ''),
  COALESCE(p.institution_name, ''),
  COALESCE(p.institution_address, ''),
  p.full_name,
  COALESCE(p.headline, ''),
  COALESCE(p.about, ''),
  COALESCE(p.preferred_role, ''),
  COALESCE(p.semester, ''),
  COALESCE(p.availability, ''),
  COALESCE(p.goals, ''),
  COALESCE(p.github_url, ''),
  COALESCE(p.telegram_username, ''),
  COALESCE(p.portfolio_url, ''),
  COALESCE(p.stacks, ARRAY[]::TEXT[]),
  COALESCE(p.interests, ARRAY[]::TEXT[]),
  p.updated_at,
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
JOIN faculties f ON f.id = p.faculty_id
JOIN departments d ON d.id = p.department_id
LEFT JOIN student_groups sg ON sg.id = p.group_id
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

func (r *AuthRepo) GetUserAuthState(ctx context.Context, tenantID, userID uuid.UUID) (svc.UserAuthState, error) {
	var state svc.UserAuthState
	err := r.db.QueryRow(ctx, `
SELECT password_changed_at, status, email_verified_at
FROM users
WHERE tenant_id = $1
  AND id = $2;
`, tenantID, userID).Scan(&state.PasswordChangedAt, &state.Status, &state.EmailVerifiedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return svc.UserAuthState{}, svc.ErrNotFound
	}
	return state, err
}

func (r *AuthRepo) UpdateUserProfile(ctx context.Context, tenantID, userID uuid.UUID, in svc.ProfileUpdate, updatedAt time.Time) error {
	tag, err := r.db.Exec(ctx, `
UPDATE user_profiles
SET full_name = $3,
    headline = $4,
    about = $5,
    preferred_role = $6,
    semester = $7,
    availability = $8,
    goals = $9,
    github_url = $10,
    telegram_username = $11,
    portfolio_url = $12,
    stacks = $13,
    interests = $14,
    updated_at = $15
WHERE tenant_id = $1
  AND user_id = $2;
`, tenantID, userID, in.FullName, in.Headline, in.About, in.PreferredRole, in.Semester, in.Availability, in.Goals, in.GithubURL, in.Telegram, in.PortfolioURL, in.Stacks, in.Interests, updatedAt)
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

func (r *AuthRepo) RevokeAndReturnRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, uuid.UUID, time.Time, error) {
	const q = `
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = $1
  AND revoked_at IS NULL
RETURNING tenant_id, user_id, expires_at;
`
	var tenantID, userID uuid.UUID
	var exp time.Time
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(&tenantID, &userID, &exp)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, time.Time{}, svc.ErrNotFound
	}
	return tenantID, userID, exp, err
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

func (r *AuthRepo) InsertGroupChangeRequest(
	ctx context.Context,
	tenantID, studentID, currentGroupID, requestedGroupID uuid.UUID,
	createdAt time.Time,
) (svc.GroupChangeRequest, error) {
	const q = `
INSERT INTO group_change_requests(
  tenant_id,
  student_id,
  current_group_id,
  requested_group_id,
  status,
  created_at
)
VALUES ($1, $2, $3, $4, 'PENDING', $5)
RETURNING id;
`
	var requestID uuid.UUID
	err := r.db.QueryRow(ctx, q, tenantID, studentID, currentGroupID, requestedGroupID, createdAt).Scan(&requestID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return svc.GroupChangeRequest{}, svc.ErrPendingGroupRequestExists
			}
			if pgErr.Code == "23503" {
				return svc.GroupChangeRequest{}, svc.ErrGroupMismatch
			}
		}
		return svc.GroupChangeRequest{}, err
	}
	return r.getGroupChangeRequestByID(ctx, tenantID, requestID)
}

func (r *AuthRepo) ListOwnGroupChangeRequests(
	ctx context.Context,
	tenantID, studentID uuid.UUID,
	limit int,
) ([]svc.GroupChangeRequest, error) {
	const q = `
SELECT
  gcr.id,
  gcr.student_id,
  COALESCE(up.full_name, '') AS student_name,
  COALESCE(u.email, '') AS student_email,
  gcr.current_group_id,
  COALESCE(curr.group_code, '') AS current_group_code,
  gcr.requested_group_id,
  COALESCE(req.group_code, '') AS requested_group_code,
  gcr.status,
  COALESCE(gcr.admin_comment, '') AS admin_comment,
  gcr.created_at,
  gcr.reviewed_at,
  gcr.reviewed_by,
  COALESCE(reviewer.full_name, '') AS reviewed_by_name
FROM group_change_requests gcr
JOIN users u ON u.id = gcr.student_id
LEFT JOIN user_profiles up ON up.user_id = gcr.student_id
LEFT JOIN student_groups curr ON curr.id = gcr.current_group_id
LEFT JOIN student_groups req ON req.id = gcr.requested_group_id
LEFT JOIN user_profiles reviewer ON reviewer.user_id = gcr.reviewed_by
WHERE gcr.tenant_id = $1
  AND gcr.student_id = $2
ORDER BY gcr.created_at DESC
LIMIT $3;
`
	rows, err := r.db.Query(ctx, q, tenantID, studentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanGroupChangeRequests(rows)
}

func (r *AuthRepo) ListGroupChangeRequests(
	ctx context.Context,
	tenantID uuid.UUID,
	status, search string,
	limit int,
) ([]svc.GroupChangeRequest, error) {
	const q = `
SELECT
  gcr.id,
  gcr.student_id,
  COALESCE(up.full_name, '') AS student_name,
  COALESCE(u.email, '') AS student_email,
  gcr.current_group_id,
  COALESCE(curr.group_code, '') AS current_group_code,
  gcr.requested_group_id,
  COALESCE(req.group_code, '') AS requested_group_code,
  gcr.status,
  COALESCE(gcr.admin_comment, '') AS admin_comment,
  gcr.created_at,
  gcr.reviewed_at,
  gcr.reviewed_by,
  COALESCE(reviewer.full_name, '') AS reviewed_by_name
FROM group_change_requests gcr
JOIN users u ON u.id = gcr.student_id
LEFT JOIN user_profiles up ON up.user_id = gcr.student_id
LEFT JOIN student_groups curr ON curr.id = gcr.current_group_id
LEFT JOIN student_groups req ON req.id = gcr.requested_group_id
LEFT JOIN user_profiles reviewer ON reviewer.user_id = gcr.reviewed_by
WHERE gcr.tenant_id = $1
  AND ($2::text = '' OR gcr.status = $2::text)
  AND (
    $3::text = ''
    OR COALESCE(up.full_name, '') ILIKE '%' || $3::text || '%'
    OR COALESCE(u.email, '') ILIKE '%' || $3::text || '%'
    OR COALESCE(curr.group_code, '') ILIKE '%' || $3::text || '%'
    OR COALESCE(req.group_code, '') ILIKE '%' || $3::text || '%'
  )
ORDER BY gcr.created_at DESC
LIMIT $4;
`
	rows, err := r.db.Query(ctx, q, tenantID, status, search, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanGroupChangeRequests(rows)
}

func (r *AuthRepo) ReviewGroupChangeRequest(
	ctx context.Context,
	tenantID, requestID, reviewerID uuid.UUID,
	decision, comment string,
	reviewedAt time.Time,
) (svc.GroupChangeRequest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return svc.GroupChangeRequest{}, err
	}
	defer tx.Rollback(ctx)

	var (
		status             string
		studentID          uuid.UUID
		requestedGroupID   uuid.UUID
		requestedDeptID    uuid.UUID
		requestedFacultyID uuid.UUID
	)
	err = tx.QueryRow(ctx, `
SELECT gcr.status, gcr.student_id, gcr.requested_group_id, sg.department_id, d.faculty_id
FROM group_change_requests gcr
JOIN student_groups sg
  ON sg.id = gcr.requested_group_id
 AND sg.tenant_id = gcr.tenant_id
JOIN departments d
  ON d.id = sg.department_id
 AND d.tenant_id = gcr.tenant_id
WHERE gcr.tenant_id = $1
  AND gcr.id = $2
FOR UPDATE;
`, tenantID, requestID).Scan(&status, &studentID, &requestedGroupID, &requestedDeptID, &requestedFacultyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return svc.GroupChangeRequest{}, svc.ErrGroupRequestNotFound
	}
	if err != nil {
		return svc.GroupChangeRequest{}, err
	}
	if strings.ToUpper(strings.TrimSpace(status)) != "PENDING" {
		return svc.GroupChangeRequest{}, svc.ErrGroupRequestReviewed
	}

	tag, err := tx.Exec(ctx, `
UPDATE group_change_requests
SET status = $3,
    admin_comment = $4,
    reviewed_at = $5,
    reviewed_by = $6
WHERE tenant_id = $1
  AND id = $2
  AND status = 'PENDING';
`, tenantID, requestID, decision, comment, reviewedAt, reviewerID)
	if err != nil {
		return svc.GroupChangeRequest{}, err
	}
	if tag.RowsAffected() == 0 {
		return svc.GroupChangeRequest{}, svc.ErrGroupRequestReviewed
	}

	if decision == "APPROVED" {
		tag, err = tx.Exec(ctx, `
UPDATE user_profiles
SET group_id = $3
    , department_id = $4
    , faculty_id = $5
WHERE tenant_id = $1
  AND user_id = $2;
`, tenantID, studentID, requestedGroupID, requestedDeptID, requestedFacultyID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return svc.GroupChangeRequest{}, svc.ErrGroupMismatch
			}
			return svc.GroupChangeRequest{}, err
		}
		if tag.RowsAffected() == 0 {
			return svc.GroupChangeRequest{}, svc.ErrNotFound
		}

		if _, err := tx.Exec(ctx, `
UPDATE refresh_tokens
SET revoked_at = now()
WHERE tenant_id = $1
  AND user_id = $2
  AND revoked_at IS NULL;
`, tenantID, studentID); err != nil {
			return svc.GroupChangeRequest{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return svc.GroupChangeRequest{}, err
	}
	return r.getGroupChangeRequestByID(ctx, tenantID, requestID)
}

func (r *AuthRepo) ListDepartmentGroupsTree(
	ctx context.Context,
	tenantID uuid.UUID,
	departmentCode, search, educationType string,
) ([]svc.DepartmentGroupsTree, error) {
	const q = `
SELECT
  d.id,
  d.code,
  d.name,
  sg.id,
  sg.group_code,
  sg.group_number,
  u.id,
  COALESCE(up.full_name, '') AS student_name,
  COALESCE(u.email, '') AS student_email,
  COALESCE(u.avatar_key, '') AS avatar_key,
  COALESCE(u.status, '') AS student_status,
  COALESCE(primary_role.role_code, '') AS role_code
FROM departments d
JOIN faculties f
  ON f.id = d.faculty_id
 AND f.tenant_id = d.tenant_id
JOIN tenants t
  ON t.id = d.tenant_id
LEFT JOIN student_groups sg
  ON sg.department_id = d.id
 AND sg.tenant_id = d.tenant_id
LEFT JOIN user_profiles up
  ON up.group_id = sg.id
 AND up.tenant_id = d.tenant_id
LEFT JOIN users u
  ON u.id = up.user_id
 AND u.tenant_id = d.tenant_id
LEFT JOIN LATERAL (
  SELECT r.code AS role_code
  FROM role_assignments ra
  JOIN roles r ON r.id = ra.role_id
  WHERE ra.user_id = u.id
    AND ra.tenant_id = d.tenant_id
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
WHERE d.tenant_id = $1
  AND ($2::text = '' OR UPPER(d.code) = UPPER($2))
  AND (
    $3::text = ''
    OR COALESCE(sg.group_code, '') ILIKE '%' || $3::text || '%'
    OR COALESCE(up.full_name, '') ILIKE '%' || $3::text || '%'
    OR COALESCE(u.email, '') ILIKE '%' || $3::text || '%'
  )
  AND (
    ($4 = 'SCHOOL' AND f.code = t.code || '_SCHOOL')
    OR ($4 <> 'SCHOOL' AND f.code <> t.code || '_SCHOOL')
  )
ORDER BY d.name ASC, sg.group_number ASC, sg.group_code ASC, up.full_name ASC, u.email ASC;
`
	rows, err := r.db.Query(ctx, q, tenantID, departmentCode, search, educationType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type groupKey struct {
		departmentID uuid.UUID
		groupID      uuid.UUID
	}
	departments := make([]svc.DepartmentGroupsTree, 0, 16)
	depIndex := make(map[uuid.UUID]int)
	groupIndex := make(map[groupKey]int)

	for rows.Next() {
		var (
			depID       uuid.UUID
			depCode     string
			depName     string
			groupID     *uuid.UUID
			groupCode   *string
			groupNumber *int
			studentID   *uuid.UUID
			studentName string
			studentMail string
			avatarKey   string
			studentStat string
			roleCode    string
		)

		if err := rows.Scan(
			&depID,
			&depCode,
			&depName,
			&groupID,
			&groupCode,
			&groupNumber,
			&studentID,
			&studentName,
			&studentMail,
			&avatarKey,
			&studentStat,
			&roleCode,
		); err != nil {
			return nil, err
		}

		di, ok := depIndex[depID]
		if !ok {
			departments = append(departments, svc.DepartmentGroupsTree{
				ID:        depID,
				Code:      depCode,
				ShortCode: depCode,
				Name:      depName,
				Groups:    make([]svc.GroupNode, 0, 8),
			})
			di = len(departments) - 1
			depIndex[depID] = di
		}

		if groupID == nil || *groupID == uuid.Nil {
			continue
		}
		gk := groupKey{departmentID: depID, groupID: *groupID}
		gi, ok := groupIndex[gk]
		if !ok {
			group := svc.GroupNode{
				ID:            *groupID,
				GroupCode:     strings.TrimSpace(derefString(groupCode)),
				GroupNumber:   derefInt(groupNumber),
				TotalStudents: 0,
				Students:      make([]svc.GroupStudent, 0, 16),
			}
			departments[di].Groups = append(departments[di].Groups, group)
			gi = len(departments[di].Groups) - 1
			groupIndex[gk] = gi
		}

		if studentID == nil || *studentID == uuid.Nil {
			continue
		}

		student := svc.GroupStudent{
			UserID:    *studentID,
			FullName:  strings.TrimSpace(studentName),
			Email:     strings.TrimSpace(studentMail),
			AvatarURL: strings.TrimSpace(avatarKey),
			Status:    strings.TrimSpace(studentStat),
			Role:      strings.TrimSpace(roleCode),
		}
		departments[di].Groups[gi].Students = append(departments[di].Groups[gi].Students, student)
		departments[di].Groups[gi].TotalStudents++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return departments, nil
}

func (r *AuthRepo) getGroupChangeRequestByID(ctx context.Context, tenantID, requestID uuid.UUID) (svc.GroupChangeRequest, error) {
	const q = `
SELECT
  gcr.id,
  gcr.student_id,
  COALESCE(up.full_name, '') AS student_name,
  COALESCE(u.email, '') AS student_email,
  gcr.current_group_id,
  COALESCE(curr.group_code, '') AS current_group_code,
  gcr.requested_group_id,
  COALESCE(req.group_code, '') AS requested_group_code,
  gcr.status,
  COALESCE(gcr.admin_comment, '') AS admin_comment,
  gcr.created_at,
  gcr.reviewed_at,
  gcr.reviewed_by,
  COALESCE(reviewer.full_name, '') AS reviewed_by_name
FROM group_change_requests gcr
JOIN users u ON u.id = gcr.student_id
LEFT JOIN user_profiles up ON up.user_id = gcr.student_id
LEFT JOIN student_groups curr ON curr.id = gcr.current_group_id
LEFT JOIN student_groups req ON req.id = gcr.requested_group_id
LEFT JOIN user_profiles reviewer ON reviewer.user_id = gcr.reviewed_by
WHERE gcr.tenant_id = $1
  AND gcr.id = $2;
`
	row := r.db.QueryRow(ctx, q, tenantID, requestID)
	item, err := scanGroupChangeRequest(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return svc.GroupChangeRequest{}, svc.ErrGroupRequestNotFound
	}
	return item, err
}

func scanGroupChangeRequests(rows pgx.Rows) ([]svc.GroupChangeRequest, error) {
	items := make([]svc.GroupChangeRequest, 0, 32)
	for rows.Next() {
		item, err := scanGroupChangeRequest(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanGroupChangeRequest(row scanner) (svc.GroupChangeRequest, error) {
	var item svc.GroupChangeRequest
	err := row.Scan(
		&item.ID,
		&item.StudentID,
		&item.StudentName,
		&item.StudentEmail,
		&item.CurrentGroupID,
		&item.CurrentGroupCode,
		&item.RequestedGroupID,
		&item.RequestedCode,
		&item.Status,
		&item.AdminComment,
		&item.CreatedAt,
		&item.ReviewedAt,
		&item.ReviewedBy,
		&item.ReviewedByName,
	)
	return item, err
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func (r *AuthRepo) scanUser(ctx context.Context, q string, args ...any) (svc.User, error) {
	var out svc.User
	err := r.db.QueryRow(ctx, q, args...).Scan(
		&out.ID,
		&out.TenantID,
		&out.Email,
		&out.PendingEmail,
		&out.PasswordHash,
		&out.PasswordChangedAt,
		&out.Status,
		&out.FacultyID,
		&out.FacultyCode,
		&out.DepartmentID,
		&out.DepartmentCode,
		&out.GroupID,
		&out.GroupCode,
		&out.GroupNumber,
		&out.Institution.Provider,
		&out.Institution.ExternalID,
		&out.Institution.Name,
		&out.Institution.Address,
		&out.FullName,
		&out.Headline,
		&out.About,
		&out.PreferredRole,
		&out.Semester,
		&out.Availability,
		&out.Goals,
		&out.GithubURL,
		&out.Telegram,
		&out.PortfolioURL,
		&out.Stacks,
		&out.Interests,
		&out.ProfileUpdatedAt,
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
