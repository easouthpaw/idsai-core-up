package notificationsmodule

import (
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/notifications"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Output struct {
	Repo    *postgres.NotificationsRepo
	Service *notifications.Service
	Handler *handlers.NotificationsHandler
}

func New(pool *pgxpool.Pool) Output {
	repo := postgres.NewNotificationsRepo(pool)
	svc := notifications.NewService(repo)
	h := handlers.NewNotificationsHandler(svc)
	return Output{
		Repo:    repo,
		Service: svc,
		Handler: h,
	}
}
