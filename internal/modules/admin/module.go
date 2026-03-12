package adminmodule

import (
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/admin"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Output struct {
	Repo    *postgres.AdminRepo
	Service *admin.Service
	Handler *handlers.AdminHandler
}

func New(pool *pgxpool.Pool) Output {
	repo := postgres.NewAdminRepo(pool)
	svc := admin.NewService(repo)
	h := handlers.NewAdminHandler(svc)
	return Output{
		Repo:    repo,
		Service: svc,
		Handler: h,
	}
}
