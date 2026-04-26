package authmodule

import (
	"time"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/http/handlers"
	"idsai-core-up/internal/infra/kzschools"
	"idsai-core-up/internal/infra/kzuniversities"
	"idsai-core-up/internal/infra/photon"
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
		EmailChangeTTL:         time.Duration(cfg.EmailChangeTTLHours) * time.Hour,
		PasswordResetTTL:       time.Duration(cfg.PasswordResetTTLMinutes) * time.Minute,
		AutoVerifyRegistrants:  cfg.AuthAutoVerifyRegistrants,
		MaxFailedLoginAttempts: cfg.MaxLoginAttempts,
		LoginAttemptWindow:     time.Duration(cfg.LoginAttemptWindowMinutes) * time.Minute,
	})
	h := handlers.NewAuthHandler(svc)
	h.SetInstitutionSuggester(auth.NewInstitutionSuggester(repo, auth.NewInstitutionAutocompleteProvider(
		kzschools.New(),
		kzuniversities.New(),
		photon.New(photon.Config{
			BaseURL:        cfg.PhotonBaseURL,
			Lang:           cfg.PhotonLang,
			CountryCode:    cfg.PhotonCountryCode,
			DefaultLon:     cfg.PhotonDefaultLon,
			DefaultLat:     cfg.PhotonDefaultLat,
			RequestTimeout: time.Duration(cfg.PhotonRequestTimeoutS) * time.Second,
		}))))
	return Output{
		Repo:    repo,
		Service: svc,
		Handler: h,
	}
}
