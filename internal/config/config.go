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
	PhotonBaseURL             string
	PhotonLang                string
	PhotonCountryCode         string
	PhotonDefaultLon          string
	PhotonDefaultLat          string
	PhotonRequestTimeoutS     int
	AuthAutoVerifyRegistrants bool
	SMTPHost                  string
	SMTPPort                  string
	SMTPUser                  string
	SMTPPass                  string
	SMTPFrom                  string
	SMTPSendTimeoutS          int
	ContactEmailTo            string
	EmailEnable               bool
	OutboxPollS               int
	DBMaxConns                int
	DBMinConns                int
	DBHealthcheckS            int
	DBPingTimeoutS            int
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
	EmailChangeTTLHours       int
	MaxLoginAttempts          int
	LoginAttemptWindowMinutes int
	StorageEndpoint           string
	StorageAccessKey          string
	StorageSecretKey          string
	StorageBucket             string
	StorageUseSSL             bool
	StoragePublicBaseURL      string
	LocalStorageDir           string
	RedisAddr                 string
	RedisPassword             string
	RedisDB                   int
	BackgroundJobsEnable      bool
	HealthMonitorEnable       bool
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Addr:                      getenvAddr("ADDR", "PORT", ":8080"),
		DatabaseURL:               getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/idsai?sslmode=disable"),
		JWTSecret:                 getenv("JWT_SECRET", ""),
		PublicBaseURL:             getenvFirstNonEmpty([]string{"PUBLIC_BASE_URL", "RENDER_EXTERNAL_URL"}, "http://localhost:8080"),
		PhotonBaseURL:             getenv("PHOTON_BASE_URL", "https://photon.komoot.io"),
		PhotonLang:                getenv("PHOTON_LANG", "ru"),
		PhotonCountryCode:         getenv("PHOTON_COUNTRY_CODE", "KZ"),
		PhotonDefaultLon:          getenv("PHOTON_DEFAULT_LON", ""),
		PhotonDefaultLat:          getenv("PHOTON_DEFAULT_LAT", ""),
		PhotonRequestTimeoutS:     getenvInt("PHOTON_REQUEST_TIMEOUT_SECONDS", 4),
		AuthAutoVerifyRegistrants: getenvBool("AUTH_AUTO_VERIFY_REGISTRATIONS", false),
		SMTPHost:                  getenv("SMTP_HOST", ""),
		SMTPPort:                  getenv("SMTP_PORT", "587"),
		SMTPUser:                  getenv("SMTP_USER", ""),
		SMTPPass:                  getenv("SMTP_PASS", ""),
		SMTPFrom:                  getenv("SMTP_FROM", "noreply@idsai.local"),
		SMTPSendTimeoutS:          getenvInt("SMTP_SEND_TIMEOUT_SECONDS", 15),
		ContactEmailTo:            getenv("CONTACT_EMAIL_TO", ""),
		EmailEnable:               getenvBool("EMAIL_ENABLED", true),
		OutboxPollS:               getenvInt("NOTIFICATIONS_OUTBOX_POLL_SECONDS", 15),
		DBMaxConns:                getenvInt("DB_MAX_CONNS", 10),
		DBMinConns:                getenvInt("DB_MIN_CONNS", 1),
		DBHealthcheckS:            getenvInt("DB_HEALTHCHECK_SECONDS", 30),
		DBPingTimeoutS:            getenvInt("DB_PING_TIMEOUT_SECONDS", 2),
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
		EmailChangeTTLHours:       getenvInt("AUTH_EMAIL_CHANGE_TTL_HOURS", 24),
		MaxLoginAttempts:          getenvInt("AUTH_MAX_LOGIN_ATTEMPTS", 5),
		LoginAttemptWindowMinutes: getenvInt("AUTH_LOGIN_ATTEMPT_WINDOW_MINUTES", 15),
		StorageEndpoint:           getenv("MINIO_ENDPOINT", ""),
		StorageAccessKey:          getenv("MINIO_ACCESS_KEY", ""),
		StorageSecretKey:          getenv("MINIO_SECRET_KEY", ""),
		StorageBucket:             getenv("MINIO_BUCKET", "idsai-media"),
		StorageUseSSL:             getenvBool("MINIO_USE_SSL", false),
		StoragePublicBaseURL:      getenv("MINIO_PUBLIC_BASE_URL", ""),
		LocalStorageDir:           getenv("LOCAL_STORAGE_DIR", ""),
		RedisAddr:                 getenv("REDIS_ADDR", ""),
		RedisPassword:             getenv("REDIS_PASSWORD", ""),
		RedisDB:                   getenvInt("REDIS_DB", 0),
		BackgroundJobsEnable:      getenvBool("BACKGROUND_JOBS_ENABLED", true),
		HealthMonitorEnable:       getenvBool("HEALTH_MONITOR_ENABLED", true),
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

func getenvAddr(addrKey, portKey, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(addrKey)); v != "" {
		return v
	}
	if port := strings.TrimSpace(os.Getenv(portKey)); port != "" {
		return ":" + strings.TrimPrefix(port, ":")
	}
	return fallback
}

func getenvFirstNonEmpty(keys []string, fallback string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return fallback
}
