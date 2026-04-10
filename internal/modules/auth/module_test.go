package authmodule

import (
	"testing"
	"time"

	"idsai-core-up/internal/config"
)

func TestNew_WiresAuthModule(t *testing.T) {
	out := New(nil, config.Config{
		JWTSecret:                 "super-secret-super-secret-123456",
		PublicBaseURL:             "https://idsai.example",
		AuthAccessTTLMinutes:      10,
		AuthRefreshTTLHours:       48,
		EmailVerificationTTLHours: 36,
		EmailChangeTTLHours:       12,
		PasswordResetTTLMinutes:   20,
		AuthAutoVerifyRegistrants: true,
		MaxLoginAttempts:          4,
		LoginAttemptWindowMinutes: 30,
	})

	if out.Repo == nil || out.Service == nil || out.Handler == nil {
		t.Fatalf("expected auth module output to be fully wired")
	}
	if out.Service.AccessTTL() != 10*time.Minute {
		t.Fatalf("unexpected access ttl: %v", out.Service.AccessTTL())
	}
	if out.Service.RefreshTTL() != 48*time.Hour {
		t.Fatalf("unexpected refresh ttl: %v", out.Service.RefreshTTL())
	}
	if out.Service.PasswordResetTTL() != 20*time.Minute {
		t.Fatalf("unexpected password reset ttl: %v", out.Service.PasswordResetTTL())
	}
	if out.Service.RegistrationRequiresVerification() {
		t.Fatalf("expected auto-verified registrations to disable verification requirement")
	}
}
