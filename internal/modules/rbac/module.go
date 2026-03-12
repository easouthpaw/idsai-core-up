package rbacmodule

import (
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/rbac"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Output struct {
	Repo    *postgres.RBACRepo
	Service *rbac.Service
}

func New(pool *pgxpool.Pool) Output {
	repo := postgres.NewRBACRepo(pool)
	svc := rbac.NewService(repo)
	return Output{
		Repo:    repo,
		Service: svc,
	}
}
