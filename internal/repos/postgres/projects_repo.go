package postgres

import (
	"context"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projects"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectsRepo struct {
	db *pgxpool.Pool
}

func NewProjectsRepo(db *pgxpool.Pool) *ProjectsRepo {
	return &ProjectsRepo{db: db}
}

func (r *ProjectsRepo) Create(ctx context.Context, title, description string, facultyID uuid.UUID, visibility string, groupID *uuid.UUID, createdBy uuid.UUID) (uuid.UUID, error) {
	const q = `
INSERT INTO projects(title, description, status, is_public, created_by, faculty_id, visibility, group_id)
VALUES ($1, $2, 'DRAFT', ($4 = 'PUBLIC'), $5, $3, $4, $6)
RETURNING id;
`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, title, description, facultyID, visibility, createdBy, groupID).Scan(&id)
	return id, err
}

func (r *ProjectsRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error) {
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

	err := r.db.QueryRow(ctx, q, id).Scan(
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

func (r *ProjectsRepo) ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]domain.Project, error) {
	const q = `
SELECT id, title, description, status, is_public, created_by, professor_id,
       faculty_id, visibility, group_id,
       created_at, updated_at
FROM projects
WHERE created_by = $1
ORDER BY created_at DESC;
`
	rows, err := r.db.Query(ctx, q, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]domain.Project, 0, 16)
	for rows.Next() {
		var p domain.Project
		var professorID *uuid.UUID
		var groupID *uuid.UUID

		if err := rows.Scan(
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
		); err != nil {
			return nil, err
		}

		p.ProfessorID = professorID
		p.GroupID = groupID
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *ProjectsRepo) ListPublic(ctx context.Context) ([]domain.Project, error) {
	const q = `
SELECT id, title, description, status, is_public, created_by, professor_id,
       faculty_id, visibility, group_id,
       created_at, updated_at
FROM projects
WHERE is_public = TRUE
ORDER BY created_at DESC;
`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Project, 0, 16)
	for rows.Next() {
		var p domain.Project
		var professorID *uuid.UUID
		var groupID *uuid.UUID

		if err := rows.Scan(
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
		); err != nil {
			return nil, err
		}

		p.ProfessorID = professorID
		p.GroupID = groupID
		items = append(items, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ProjectsRepo) FindGroupIDByCode(ctx context.Context, facultyID uuid.UUID, code string) (uuid.UUID, error) {
	const q = `
SELECT id
FROM student_groups
WHERE faculty_id = $1
  AND code = $2;
`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, facultyID, code).Scan(&id)
	return id, err
}

func (r *ProjectsRepo) ListGroupsByFaculty(ctx context.Context, facultyID uuid.UUID) ([]projects.Group, error) {
	const q = `
SELECT id, code, name
FROM student_groups
WHERE faculty_id = $1
ORDER BY code;
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
