package postgres

import (
	"context"
	"time"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MembersRepo struct {
	db *pgxpool.Pool
}

func NewMembersRepo(db *pgxpool.Pool) *MembersRepo {
	return &MembersRepo{db: db}
}

// Apply creates member with status APPLIED (idempotent per (project_id,user_id))
func (r *MembersRepo) Apply(ctx context.Context, projectID, userID uuid.UUID) (uuid.UUID, error) {
	const q = `
INSERT INTO project_members(project_id, user_id, status)
VALUES ($1, $2, 'APPLIED')
ON CONFLICT (project_id, user_id) DO UPDATE SET project_id = EXCLUDED.project_id
RETURNING id;
`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, projectID, userID).Scan(&id)
	return id, err
}

// Approve sets status ACTIVE, assigns position_id, sets joined_at
func (r *MembersRepo) Approve(ctx context.Context, projectID, userID, positionID uuid.UUID) error {
	now := time.Now().UTC()
	const q = `
UPDATE project_members
SET status='ACTIVE', position_id=$3, joined_at=$4
WHERE project_id=$1 AND user_id=$2;
`
	_, err := r.db.Exec(ctx, q, projectID, userID, positionID, now)
	return err
}

func (r *MembersRepo) GetByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (domain.ProjectMember, error) {
	const q = `
SELECT id, project_id, user_id, position_id, status, joined_at, created_at
FROM project_members
WHERE project_id=$1 AND user_id=$2;
`
	var m domain.ProjectMember
	var posID *uuid.UUID
	var joinedAt *time.Time
	err := r.db.QueryRow(ctx, q, projectID, userID).Scan(
		&m.ID,
		&m.ProjectID,
		&m.UserID,
		&posID,
		&m.Status,
		&joinedAt,
		&m.CreatedAt,
	)
	if err != nil {
		return domain.ProjectMember{}, err
	}
	m.PositionID = posID
	m.JoinedAt = joinedAt
	return m, nil
}
