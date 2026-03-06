package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr                    string
	DatabaseURL             string
	JWTSecret               string
	SMTPHost                string
	SMTPPort                string
	SMTPUser                string
	SMTPPass                string
	SMTPFrom                string
	EmailEnable             bool
	OutboxPollS             int
	ServerName              string
	TelegramBotToken        string
	TelegramSuperadminChat  string
	TelegramRequestTimeoutS int
	TelegramDedupeWindowS   int
	HealthcheckPollS        int
	HeartbeatS              int
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Addr:                    getenv("ADDR", ":8080"),
		DatabaseURL:             getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/idsai?sslmode=disable"),
		JWTSecret:               getenv("JWT_SECRET", "dev-jwt-secret"),
		SMTPHost:                getenv("SMTP_HOST", ""),
		SMTPPort:                getenv("SMTP_PORT", "587"),
		SMTPUser:                getenv("SMTP_USER", ""),
		SMTPPass:                getenv("SMTP_PASS", ""),
		SMTPFrom:                getenv("SMTP_FROM", "noreply@idsai.local"),
		EmailEnable:             getenvBool("EMAIL_ENABLED", true),
		OutboxPollS:             getenvInt("NOTIFICATIONS_OUTBOX_POLL_SECONDS", 15),
		ServerName:              getenv("SERVER_NAME", "idsai"),
		TelegramBotToken:        getenv("TELEGRAM_BOT_TOKEN", ""),
		TelegramSuperadminChat:  getenv("TELEGRAM_SUPERADMIN_CHAT_ID", ""),
		TelegramRequestTimeoutS: getenvInt("TELEGRAM_REQUEST_TIMEOUT_SECONDS", 5),
		TelegramDedupeWindowS:   getenvInt("TELEGRAM_ALERT_DEDUPE_SECONDS", 120),
		HealthcheckPollS:        getenvInt("SERVER_HEALTHCHECK_POLL_SECONDS", 20),
		HeartbeatS:              getenvInt("SERVER_HEARTBEAT_SECONDS", 3600),
	}
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
