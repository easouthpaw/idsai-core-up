package app

import (
	"strings"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/infra/email"
)

func newEmailSender(cfg config.Config) directEmailSender {
	from := strings.TrimSpace(cfg.SMTPFrom)
	if key := strings.TrimSpace(cfg.ResendAPIKey); key != "" && from != "" {
		return email.NewResendSender(key, from)
	}

	host := strings.TrimSpace(cfg.SMTPHost)
	port := strings.TrimSpace(cfg.SMTPPort)
	if host == "" || port == "" || from == "" {
		return nil
	}
	return email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
}

func emailSenderMode(cfg config.Config) string {
	if strings.TrimSpace(cfg.ResendAPIKey) != "" && strings.TrimSpace(cfg.SMTPFrom) != "" {
		return "resend"
	}
	if strings.TrimSpace(cfg.SMTPHost) != "" && strings.TrimSpace(cfg.SMTPPort) != "" && strings.TrimSpace(cfg.SMTPFrom) != "" {
		return "smtp"
	}
	return ""
}
