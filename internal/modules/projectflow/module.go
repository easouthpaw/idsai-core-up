package projectflowmodule

import (
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/projectflow"
	"idsai-core-up/internal/services/rbac"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Output struct {
	Repo    *postgres.ProjectFlowRepo
	Service *projectflow.Service
	Handler *handlers.ProjectFlowHandler
}

func New(pool *pgxpool.Pool, authorizer *rbac.Service, grantor projectflow.RoleGrantor) Output {
	repo := postgres.NewProjectFlowRepo(pool)
	svc := projectflow.NewService(
		authorizer,
		grantor,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
	)
	h := handlers.NewProjectFlowHandler(svc)
	return Output{
		Repo:    repo,
		Service: svc,
		Handler: h,
	}
}
