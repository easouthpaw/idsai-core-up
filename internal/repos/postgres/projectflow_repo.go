package postgres

import (
	"idsai-core-up/internal/services/projectflow"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectFlowRepo struct {
	db *pgxpool.Pool
}

var (
	_ projectflow.StacksRepository     = (*ProjectFlowRepo)(nil)
	_ projectflow.ProjectsRepository   = (*ProjectFlowRepo)(nil)
	_ projectflow.PositionsRepository  = (*ProjectFlowRepo)(nil)
	_ projectflow.MembersRepository    = (*ProjectFlowRepo)(nil)
	_ projectflow.ProfessorsRepository = (*ProjectFlowRepo)(nil)
	_ projectflow.CriteriaRepository   = (*ProjectFlowRepo)(nil)
	_ projectflow.LifecycleRepository  = (*ProjectFlowRepo)(nil)
	_ projectflow.TasksRepository      = (*ProjectFlowRepo)(nil)
	_ projectflow.AccessRepository     = (*ProjectFlowRepo)(nil)
)

func NewProjectFlowRepo(db *pgxpool.Pool) *ProjectFlowRepo {
	return &ProjectFlowRepo{db: db}
}
