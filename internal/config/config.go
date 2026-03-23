package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr                      string
	DatabaseURL               string
	JWTSecret                 string
	PublicBaseURL             string
	SMTPHost                  string
	SMTPPort                  string
	SMTPUser                  string
	SMTPPass                  string
	SMTPFrom                  string
	EmailEnable               bool
	OutboxPollS               int
	ServerName                string
	TelegramBotToken          string
	TelegramSuperadminChat    string
	TelegramRequestTimeoutS   int
	TelegramDedupeWindowS     int
	HealthcheckPollS          int
	HeartbeatS                int
	AuthAccessTTLMinutes      int
	AuthRefreshTTLHours       int
	PasswordResetTTLMinutes   int
	EmailVerificationTTLHours int
	MaxLoginAttempts          int
	LoginAttemptWindowMinutes int
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Addr:                      getenv("ADDR", ":8080"),
		DatabaseURL:               getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/idsai?sslmode=disable"),
		JWTSecret:                 getenv("JWT_SECRET", ""),
		PublicBaseURL:             getenv("PUBLIC_BASE_URL", "http://localhost:8080"),
		SMTPHost:                  getenv("SMTP_HOST", ""),
		SMTPPort:                  getenv("SMTP_PORT", "587"),
		SMTPUser:                  getenv("SMTP_USER", ""),
		SMTPPass:                  getenv("SMTP_PASS", ""),
		SMTPFrom:                  getenv("SMTP_FROM", "noreply@idsai.local"),
		EmailEnable:               getenvBool("EMAIL_ENABLED", true),
		OutboxPollS:               getenvInt("NOTIFICATIONS_OUTBOX_POLL_SECONDS", 15),
		ServerName:                getenv("SERVER_NAME", "idsai"),
		TelegramBotToken:          getenv("TELEGRAM_BOT_TOKEN", ""),
		TelegramSuperadminChat:    getenv("TELEGRAM_SUPERADMIN_CHAT_ID", ""),
		TelegramRequestTimeoutS:   getenvInt("TELEGRAM_REQUEST_TIMEOUT_SECONDS", 5),
		TelegramDedupeWindowS:     getenvInt("TELEGRAM_ALERT_DEDUPE_SECONDS", 120),
		HealthcheckPollS:          getenvInt("SERVER_HEALTHCHECK_POLL_SECONDS", 20),
		HeartbeatS:                getenvInt("SERVER_HEARTBEAT_SECONDS", 3600),
		AuthAccessTTLMinutes:      getenvInt("AUTH_ACCESS_TTL_MINUTES", 15),
		AuthRefreshTTLHours:       getenvInt("AUTH_REFRESH_TTL_HOURS", 168),
		PasswordResetTTLMinutes:   getenvInt("AUTH_PASSWORD_RESET_TTL_MINUTES", 30),
		EmailVerificationTTLHours: getenvInt("AUTH_EMAIL_VERIFICATION_TTL_HOURS", 24),
		MaxLoginAttempts:          getenvInt("AUTH_MAX_LOGIN_ATTEMPTS", 5),
		LoginAttemptWindowMinutes: getenvInt("AUTH_LOGIN_ATTEMPT_WINDOW_MINUTES", 15),
	}
}

func (c Config) Validate() error {
	if len(strings.TrimSpace(c.JWTSecret)) < 32 {
		return errors.New("JWT_SECRET must be set to at least 32 characters")
	}
	if strings.TrimSpace(c.PublicBaseURL) == "" {
		return errors.New("PUBLIC_BASE_URL must be set")
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
