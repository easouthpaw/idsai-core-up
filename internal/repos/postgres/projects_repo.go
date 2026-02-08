package postgres

import (
	"context"

	"idsai-core-up/internal/domain"

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
		&p.GroupID,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return domain.Project{}, err
	}
	p.ProfessorID = professorID
	return p, nil
}
