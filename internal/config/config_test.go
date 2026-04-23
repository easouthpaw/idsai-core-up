package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_ReadsEnvAndAppliesDefaults(t *testing.T) {
	t.Chdir(t.TempDir())

	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://test-db")
	t.Setenv("JWT_SECRET", strings.Repeat("a", 32))
	t.Setenv("RENDER_EXTERNAL_URL", "https://render.example")
	t.Setenv("AUTH_AUTO_VERIFY_REGISTRATIONS", "yes")
	t.Setenv("SMTP_SEND_TIMEOUT_SECONDS", "25")
	t.Setenv("SMTP_USE_SSL", "yes")
	t.Setenv("EMAIL_ENABLED", "off")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("REDIS_DB", "7")
	t.Setenv("BACKGROUND_JOBS_ENABLED", "no")
	t.Setenv("HEALTH_MONITOR_ENABLED", "false")

	cfg := Load()

	if cfg.Addr != ":9090" {
		t.Fatalf("expected port-derived addr, got %q", cfg.Addr)
	}
	if cfg.DatabaseURL != "postgres://test-db" {
		t.Fatalf("unexpected database url: %q", cfg.DatabaseURL)
	}
	if cfg.PublicBaseURL != "https://render.example" {
		t.Fatalf("unexpected public base url: %q", cfg.PublicBaseURL)
	}
	if !cfg.AuthAutoVerifyRegistrants {
		t.Fatalf("expected auto-verify to be enabled")
	}
	if cfg.SMTPSendTimeoutS != 25 {
		t.Fatalf("unexpected smtp timeout: %d", cfg.SMTPSendTimeoutS)
	}
	if !cfg.SMTPUseSSL {
		t.Fatalf("expected smtp SSL to be enabled")
	}
	if cfg.EmailEnable {
		t.Fatalf("expected email to be disabled")
	}
	if !cfg.StorageUseSSL {
		t.Fatalf("expected storage SSL to be enabled")
	}
	if cfg.RedisDB != 7 {
		t.Fatalf("unexpected redis db: %d", cfg.RedisDB)
	}
	if cfg.BackgroundJobsEnable {
		t.Fatalf("expected background jobs to be disabled")
	}
	if cfg.HealthMonitorEnable {
		t.Fatalf("expected health monitor to be disabled")
	}
	if cfg.SMTPPort != "587" {
		t.Fatalf("expected default smtp port, got %q", cfg.SMTPPort)
	}
	if cfg.StorageBucket != "idsai-media" {
		t.Fatalf("expected default bucket, got %q", cfg.StorageBucket)
	}
	if cfg.PhotonCountryCode != "KZ" {
		t.Fatalf("expected default photon country code KZ, got %q", cfg.PhotonCountryCode)
	}
}

func TestLoad_ReadsDotEnvFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	data := strings.Join([]string{
		"ADDR=:7777",
		"PUBLIC_BASE_URL=https://from-dotenv.example",
		"JWT_SECRET=abcdefghijklmnopqrstuvwxyz123456",
		"REDIS_ADDR=localhost:6379",
	}, "\n")
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(data), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg := Load()

	if cfg.Addr != ":7777" {
		t.Fatalf("expected dotenv addr, got %q", cfg.Addr)
	}
	if cfg.PublicBaseURL != "https://from-dotenv.example" {
		t.Fatalf("expected dotenv public base url, got %q", cfg.PublicBaseURL)
	}
	if cfg.JWTSecret != "abcdefghijklmnopqrstuvwxyz123456" {
		t.Fatalf("expected dotenv jwt secret, got %q", cfg.JWTSecret)
	}
	if cfg.RedisAddr != "localhost:6379" {
		t.Fatalf("expected dotenv redis addr, got %q", cfg.RedisAddr)
	}
}

func TestValidate(t *testing.T) {
	valid := Config{
		JWTSecret:     strings.Repeat("z", 32),
		PublicBaseURL: "https://example.local",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}

	shortSecret := Config{
		JWTSecret:     "short",
		PublicBaseURL: "https://example.local",
	}
	if err := shortSecret.Validate(); err == nil {
		t.Fatalf("expected short JWT secret to fail validation")
	}

	missingBaseURL := Config{
		JWTSecret:     strings.Repeat("z", 32),
		PublicBaseURL: "   ",
	}
	if err := missingBaseURL.Validate(); err == nil {
		t.Fatalf("expected missing base url to fail validation")
	}
}

func TestEnvHelpers(t *testing.T) {
	t.Setenv("TEXT_VALUE", "hello")
	if got := getenv("TEXT_VALUE", "fallback"); got != "hello" {
		t.Fatalf("unexpected getenv value: %q", got)
	}
	if got := getenv("MISSING_TEXT_VALUE", "fallback"); got != "fallback" {
		t.Fatalf("unexpected getenv fallback: %q", got)
	}

	t.Setenv("INT_VALUE", "42")
	t.Setenv("INT_INVALID", "oops")
	if got := getenvInt("INT_VALUE", 7); got != 42 {
		t.Fatalf("unexpected getenvInt value: %d", got)
	}
	if got := getenvInt("INT_INVALID", 7); got != 7 {
		t.Fatalf("unexpected getenvInt fallback: %d", got)
	}

	t.Setenv("BOOL_TRUE", "on")
	t.Setenv("BOOL_FALSE", "no")
	t.Setenv("BOOL_INVALID", "maybe")
	if !getenvBool("BOOL_TRUE", false) {
		t.Fatalf("expected truthy bool value")
	}
	if getenvBool("BOOL_FALSE", true) {
		t.Fatalf("expected falsy bool value")
	}
	if got := getenvBool("BOOL_INVALID", true); !got {
		t.Fatalf("expected invalid bool to use fallback")
	}

	t.Setenv("ADDR_VALUE", "127.0.0.1:8080")
	t.Setenv("PORT_ONLY", "5050")
	if got := getenvAddr("ADDR_VALUE", "IGNORED_PORT", ":1"); got != "127.0.0.1:8080" {
		t.Fatalf("unexpected addr value: %q", got)
	}
	if got := getenvAddr("MISSING_ADDR", "PORT_ONLY", ":1"); got != ":5050" {
		t.Fatalf("unexpected port-derived addr: %q", got)
	}
	if got := getenvAddr("MISSING_ADDR", "MISSING_PORT", ":1"); got != ":1" {
		t.Fatalf("unexpected addr fallback: %q", got)
	}

	t.Setenv("FIRST_EMPTY", "   ")
	t.Setenv("SECOND_VALUE", "https://second.example")
	if got := getenvFirstNonEmpty([]string{"FIRST_EMPTY", "SECOND_VALUE"}, "fallback"); got != "https://second.example" {
		t.Fatalf("unexpected first-non-empty value: %q", got)
	}
	if got := getenvFirstNonEmpty([]string{"UNKNOWN_A", "UNKNOWN_B"}, "fallback"); got != "fallback" {
		t.Fatalf("unexpected first-non-empty fallback: %q", got)
	}
}
