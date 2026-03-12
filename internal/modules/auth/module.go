package authmodule

import (
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Output struct {
	Repo    *postgres.AuthRepo
	Service *auth.Service
	Handler *handlers.AuthHandler
}

func New(pool *pgxpool.Pool, jwtSecret string) Output {
	repo := postgres.NewAuthRepo(pool)
	svc := auth.NewService(repo, jwtSecret)
	h := handlers.NewAuthHandler(svc)
	return Output{
		Repo:    repo,
		Service: svc,
		Handler: h,
	}
}
