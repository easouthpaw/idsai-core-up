package app

import (
	"testing"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/infra/email"

	"github.com/stretchr/testify/require"
)

func TestNewEmailSenderPrefersResend(t *testing.T) {
	sender := newEmailSender(config.Config{
		ResendAPIKey: "re_test_key",
		SMTPFrom:     "Acme <noreply@example.com>",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     "587",
	})

	_, ok := sender.(*email.ResendSender)
	require.True(t, ok)
	require.Equal(t, "resend", emailSenderMode(config.Config{
		ResendAPIKey: "re_test_key",
		SMTPFrom:     "Acme <noreply@example.com>",
	}))
}

func TestNewEmailSenderFallsBackToSMTP(t *testing.T) {
	sender := newEmailSender(config.Config{
		SMTPHost: "smtp.example.com",
		SMTPPort: "587",
		SMTPUser: "smtp-user",
		SMTPPass: "smtp-pass",
		SMTPFrom: "Acme <noreply@example.com>",
	})

	_, ok := sender.(*email.SMTPSender)
	require.True(t, ok)
	require.Equal(t, "smtp", emailSenderMode(config.Config{
		SMTPHost: "smtp.example.com",
		SMTPPort: "587",
		SMTPFrom: "Acme <noreply@example.com>",
	}))
}

func TestNewEmailSenderReturnsNilWhenConfigIncomplete(t *testing.T) {
	require.Nil(t, newEmailSender(config.Config{}))
	require.Equal(t, "", emailSenderMode(config.Config{}))
}
