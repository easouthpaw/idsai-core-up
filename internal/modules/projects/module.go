package projectsmodule

import (
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/projects"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Output struct {
	Repo    *postgres.ProjectsRepo
	Service *projects.Service
}

func New(pool *pgxpool.Pool, grantor projects.RoleGrantor) Output {
	repo := postgres.NewProjectsRepo(pool)
	svc := projects.NewService(repo, grantor)
	return Output{
		Repo:    repo,
		Service: svc,
	}
}
