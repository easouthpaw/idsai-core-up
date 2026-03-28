package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projects"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectsRepo struct {
	db *pgxpool.Pool
}

func NewProjectsRepo(db *pgxpool.Pool) *ProjectsRepo {
	return &ProjectsRepo{db: db}
}

func (r *ProjectsRepo) Create(ctx context.Context, title, description string, facultyID uuid.UUID, visibility string, groupID *uuid.UUID, createdBy uuid.UUID) (uuid.UUID, error) {
	const qCreateProject = `
INSERT INTO projects(tenant_id, title, description, status, is_public, created_by, faculty_id, visibility, group_id, default_cover_variant)
VALUES (
  (SELECT tenant_id FROM faculties WHERE id = $3),
  $1,
  $2,
  'DRAFT',
  ($4 = 'PUBLIC'),
  $5,
  $3,
  $4,
  $6,
  1 + (ABS(hashtext(COALESCE($1, '') || ':' || COALESCE($2, ''))) % 6)
)
RETURNING id;
`
	const qCreateLeadMember = `
INSERT INTO project_members(tenant_id, project_id, user_id, status, joined_at)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, 'ACTIVE', now())
ON CONFLICT (project_id, user_id)
DO UPDATE SET status='ACTIVE', joined_at=COALESCE(project_members.joined_at, now());
`
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	err = tx.QueryRow(ctx, qCreateProject, title, description, facultyID, visibility, createdBy, groupID).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	if _, err := tx.Exec(ctx, qCreateLeadMember, id, createdBy); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *ProjectsRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return domain.Project{}, err
	}

	const q = `
SELECT
  p.id,
  p.title,
  p.description,
  p.status,
  p.is_public,
  p.created_by,
  COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), p.created_by::text) AS created_by_name,
  COALESCE(u.email, '') AS created_by_email,
  p.professor_id,
  COALESCE(p.professor_review_status, 'NONE') AS professor_review_status,
  p.faculty_id,
  p.visibility,
  p.group_id,
  COALESCE(p.image_key, '') AS image_key,
  p.default_cover_variant,
  p.image_updated_at,
  p.created_at,
  p.updated_at
FROM projects p
LEFT JOIN users u ON u.id = p.created_by
LEFT JOIN user_profiles up ON up.user_id = p.created_by
WHERE p.tenant_id = $1
  AND p.id = $2;
`
	var p domain.Project
	var professorID *uuid.UUID
	var groupID *uuid.UUID
	var imageUpdatedAt *time.Time

	err = r.db.QueryRow(ctx, q, tenantID, id).Scan(
		&p.ID,
		&p.Title,
		&p.Description,
		&p.Status,
		&p.IsPublic,
		&p.CreatedBy,
		&p.CreatedByName,
		&p.CreatedByEmail,
		&professorID,
		&p.ProfessorReviewStatus,
		&p.FacultyID,
		&p.Visibility,
		&groupID,
		&p.ImageKey,
		&p.DefaultCoverVariant,
		&imageUpdatedAt,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, projects.ErrNotFound
		}
		return domain.Project{}, err
	}
	p.ProfessorID = professorID
	p.GroupID = groupID
	p.ImageUpdatedAt = imageUpdatedAt
	return p, nil
}

func (r *ProjectsRepo) HasProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permissionCode string) (bool, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return false, err
	}

	const q = `
SELECT EXISTS (
  SELECT 1
  FROM role_assignments ra
  JOIN users u ON u.id = ra.user_id
  JOIN projects p ON p.id = $2
  JOIN role_permissions rp ON rp.role_id = ra.role_id
  JOIN permissions perm ON perm.id = rp.permission_id
  WHERE ra.user_id = $1
    AND u.tenant_id = $4
    AND u.tenant_id = ra.tenant_id
    AND p.tenant_id = $4
    AND p.tenant_id = ra.tenant_id
    AND ra.scope_type = 'PROJECT'
    AND ra.scope_id = $2
    AND (ra.expires_at IS NULL OR ra.expires_at > now())
    AND perm.code = $3
) AS ok;
`
	var ok bool
	err = r.db.QueryRow(ctx, q, userID, projectID, permissionCode, tenantID).Scan(&ok)
	return ok, err
}

func (r *ProjectsRepo) HasResolvedProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permissionCode string) (bool, error) {
	const q = resolvedScopesCTE + `
SELECT EXISTS (
  SELECT 1
  FROM role_assignments ra
  JOIN users u ON u.id = ra.user_id
  JOIN resolved_scopes rs
    ON rs.scope_type = ra.scope_type
   AND rs.tenant_id = ra.tenant_id
   AND (
     (rs.scope_id IS NULL AND ra.scope_id IS NULL)
     OR (rs.scope_id = ra.scope_id)
   )
  JOIN role_permissions rp ON rp.role_id = ra.role_id
  JOIN permissions p ON p.id = rp.permission_id
  WHERE ra.user_id = $1
    AND u.tenant_id = ra.tenant_id
    AND (ra.expires_at IS NULL OR ra.expires_at > now())
    AND p.code = $4
) AS ok;
`
	var ok bool
	err := r.db.QueryRow(ctx, q, userID, "PROJECT", projectID, permissionCode).Scan(&ok)
	return ok, err
}

func (r *ProjectsRepo) GetProjectReviewSummary(ctx context.Context, projectID uuid.UUID) (*projects.ReviewSummary, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT
  COUNT(c.id) AS total,
  COALESCE(SUM(CASE WHEN r.is_met = TRUE THEN 1 ELSE 0 END), 0) AS met,
  MAX(r.updated_at) AS reviewed_at,
  COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), 'Преподаватель') AS reviewer
FROM projects p
LEFT JOIN users u ON u.id = p.professor_id
LEFT JOIN user_profiles up ON up.user_id = p.professor_id
LEFT JOIN project_criteria c ON c.project_id = p.id
LEFT JOIN project_criterion_reviews r
  ON r.project_id = p.id
 AND r.criterion_id = c.id
 AND r.professor_id = p.professor_id
WHERE p.tenant_id = $1
  AND p.id = $2
GROUP BY reviewer;
`

	var (
		total      int
		met        int
		reviewedAt *time.Time
		reviewer   string
	)
	if err := r.db.QueryRow(ctx, q, tenantID, projectID).Scan(&total, &met, &reviewedAt, &reviewer); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		if isUndefinedRelationErr(err, "project_criterion_reviews") {
			return nil, nil
		}
		return nil, err
	}

	score := "0.0"
	passPercent := 0
	if total > 0 {
		passPercent = int(math.Round(float64(met*100) / float64(total)))
		score = fmt.Sprintf("%.1f", float64(met*5)/float64(total))
	}

	return &projects.ReviewSummary{
		Score:       score,
		PassPercent: passPercent,
		Met:         met,
		Total:       total,
		ReviewedAt:  reviewedAt,
		Reviewer:    reviewer,
	}, nil
}

func (r *ProjectsRepo) SetProjectImage(ctx context.Context, projectID uuid.UUID, imageKey string, updatedAt time.Time) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	tag, err := r.db.Exec(ctx, `
UPDATE projects
SET image_key = $2,
    image_updated_at = $3
WHERE tenant_id = $1
  AND id = $2;
`, tenantID, projectID, imageKey, updatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return projects.ErrNotFound
	}
	return nil
}

func (r *ProjectsRepo) ClearProjectImage(ctx context.Context, projectID uuid.UUID) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	tag, err := r.db.Exec(ctx, `
UPDATE projects
SET image_key = NULL,
    image_updated_at = now()
WHERE tenant_id = $1
  AND id = $2;
`, tenantID, projectID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return projects.ErrNotFound
	}
	return nil
}

func (r *ProjectsRepo) ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]domain.Project, error) {
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
  p.is_public,
  p.created_by,
  COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), p.created_by::text) AS created_by_name,
  COALESCE(u.email, '') AS created_by_email,
  p.professor_id,
  COALESCE(p.professor_review_status, 'NONE') AS professor_review_status,
  p.faculty_id,
  p.visibility,
  p.group_id,
  COALESCE(p.image_key, '') AS image_key,
  p.default_cover_variant,
  p.image_updated_at,
  p.created_at,
  p.updated_at
FROM projects p
LEFT JOIN users u ON u.id = p.created_by
LEFT JOIN user_profiles up ON up.user_id = p.created_by
WHERE p.tenant_id = $1
  AND p.created_by = $2
ORDER BY p.created_at DESC;
`
	rows, err := r.db.Query(ctx, q, tenantID, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]domain.Project, 0, 16)
	for rows.Next() {
		var p domain.Project
		var professorID *uuid.UUID
		var groupID *uuid.UUID
		var imageUpdatedAt *time.Time

		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Status,
			&p.IsPublic,
			&p.CreatedBy,
			&p.CreatedByName,
			&p.CreatedByEmail,
			&professorID,
			&p.ProfessorReviewStatus,
			&p.FacultyID,
			&p.Visibility,
			&groupID,
			&p.ImageKey,
			&p.DefaultCoverVariant,
			&imageUpdatedAt,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		p.ProfessorID = professorID
		p.GroupID = groupID
		p.ImageUpdatedAt = imageUpdatedAt
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *ProjectsRepo) ListByFaculty(ctx context.Context, facultyID uuid.UUID) ([]domain.Project, error) {
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
  p.is_public,
  p.created_by,
  COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), p.created_by::text) AS created_by_name,
  COALESCE(u.email, '') AS created_by_email,
  p.professor_id,
  COALESCE(p.professor_review_status, 'NONE') AS professor_review_status,
  p.faculty_id,
  p.visibility,
  p.group_id,
  COALESCE(p.image_key, '') AS image_key,
  p.default_cover_variant,
  p.image_updated_at,
  p.created_at,
  p.updated_at
FROM projects p
LEFT JOIN users u ON u.id = p.created_by
LEFT JOIN user_profiles up ON up.user_id = p.created_by
WHERE p.tenant_id = $1
  AND p.faculty_id = $2
ORDER BY p.updated_at DESC, p.created_at DESC;
`
	rows, err := r.db.Query(ctx, q, tenantID, facultyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Project, 0, 16)
	for rows.Next() {
		var p domain.Project
		var professorID *uuid.UUID
		var groupID *uuid.UUID
		var imageUpdatedAt *time.Time

		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Status,
			&p.IsPublic,
			&p.CreatedBy,
			&p.CreatedByName,
			&p.CreatedByEmail,
			&professorID,
			&p.ProfessorReviewStatus,
			&p.FacultyID,
			&p.Visibility,
			&groupID,
			&p.ImageKey,
			&p.DefaultCoverVariant,
			&imageUpdatedAt,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		p.ProfessorID = professorID
		p.GroupID = groupID
		p.ImageUpdatedAt = imageUpdatedAt
		items = append(items, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ProjectsRepo) ListPublic(ctx context.Context) ([]domain.Project, error) {
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
  p.is_public,
  p.created_by,
  COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), p.created_by::text) AS created_by_name,
  COALESCE(u.email, '') AS created_by_email,
  p.professor_id,
  COALESCE(p.professor_review_status, 'NONE') AS professor_review_status,
  p.faculty_id,
  p.visibility,
  p.group_id,
  COALESCE(p.image_key, '') AS image_key,
  p.default_cover_variant,
  p.image_updated_at,
  p.created_at,
  p.updated_at
FROM projects p
LEFT JOIN users u ON u.id = p.created_by
LEFT JOIN user_profiles up ON up.user_id = p.created_by
WHERE p.tenant_id = $1
  AND p.is_public = TRUE
ORDER BY p.created_at DESC;
`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Project, 0, 16)
	for rows.Next() {
		var p domain.Project
		var professorID *uuid.UUID
		var groupID *uuid.UUID
		var imageUpdatedAt *time.Time

		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.Status,
			&p.IsPublic,
			&p.CreatedBy,
			&p.CreatedByName,
			&p.CreatedByEmail,
			&professorID,
			&p.ProfessorReviewStatus,
			&p.FacultyID,
			&p.Visibility,
			&groupID,
			&p.ImageKey,
			&p.DefaultCoverVariant,
			&imageUpdatedAt,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		p.ProfessorID = professorID
		p.GroupID = groupID
		p.ImageUpdatedAt = imageUpdatedAt
		items = append(items, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ProjectsRepo) FindGroupIDByCode(ctx context.Context, facultyID uuid.UUID, code string) (uuid.UUID, error) {
	const q = `
SELECT sg.id
FROM student_groups sg
JOIN departments d ON d.id = sg.department_id
WHERE d.faculty_id = $1
  AND UPPER(sg.group_code) = UPPER($2);
`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, facultyID, code).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, projects.ErrGroupNotFound
	}
	return id, err
}

func (r *ProjectsRepo) ListGroupsByFaculty(ctx context.Context, facultyID uuid.UUID) ([]projects.Group, error) {
	const q = `
SELECT sg.id, sg.group_code, COALESCE(sg.group_code, '')
FROM student_groups sg
JOIN departments d ON d.id = sg.department_id
WHERE d.faculty_id = $1
ORDER BY sg.group_number ASC, sg.group_code ASC;
`
	rows, err := r.db.Query(ctx, q, facultyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]projects.Group, 0, 16)
	for rows.Next() {
		var g projects.Group
		if err := rows.Scan(&g.ID, &g.Code, &g.Name); err != nil {
			return nil, err
		}
		items = append(items, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
