package kbmodule

import (
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/kb"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Output struct {
	Repo    *postgres.KBRepo
	Service *kb.Service
	Handler *handlers.KBHandler
}

func New(pool *pgxpool.Pool) Output {
	repo := postgres.NewKBRepo(pool)
	svc := kb.NewService(repo)
	h := handlers.NewKBHandler(svc)
	return Output{
		Repo:    repo,
		Service: svc,
		Handler: h,
	}
}
