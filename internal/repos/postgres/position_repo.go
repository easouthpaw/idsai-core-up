package postgres

import (
	"context"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PositionsRepo struct {
	db *pgxpool.Pool
}

func NewPositionsRepo(db *pgxpool.Pool) *PositionsRepo {
	return &PositionsRepo{db: db}
}

func (r *PositionsRepo) Create(ctx context.Context, projectID uuid.UUID, code, name string, capacity int) (uuid.UUID, error) {
	const q = `
INSERT INTO project_positions(project_id, code, name, capacity)
VALUES ($1, $2, $3, $4)
RETURNING id;
`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, projectID, code, name, capacity).Scan(&id)
	return id, err
}

func (r *PositionsRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.ProjectPosition, error) {
	const q = `
SELECT id, project_id, code, name, capacity, created_at
FROM project_positions
WHERE project_id = $1
ORDER BY created_at ASC;
`
	rows, err := r.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.ProjectPosition
	for rows.Next() {
		var p domain.ProjectPosition
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Code, &p.Name, &p.Capacity, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
