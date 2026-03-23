package authmodule

import (
	"time"

	"idsai-core-up/internal/config"
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

func New(pool *pgxpool.Pool, cfg config.Config) Output {
	repo := postgres.NewAuthRepo(pool)
	svc := auth.NewService(repo, auth.Config{
		JWTSecret:              cfg.JWTSecret,
		PublicBaseURL:          cfg.PublicBaseURL,
		AccessTTL:              time.Duration(cfg.AuthAccessTTLMinutes) * time.Minute,
		RefreshTTL:             time.Duration(cfg.AuthRefreshTTLHours) * time.Hour,
		VerificationTTL:        time.Duration(cfg.EmailVerificationTTLHours) * time.Hour,
		PasswordResetTTL:       time.Duration(cfg.PasswordResetTTLMinutes) * time.Minute,
		MaxFailedLoginAttempts: cfg.MaxLoginAttempts,
		LoginAttemptWindow:     time.Duration(cfg.LoginAttemptWindowMinutes) * time.Minute,
	})
	h := handlers.NewAuthHandler(svc)
	return Output{
		Repo:    repo,
		Service: svc,
		Handler: h,
	}
}
