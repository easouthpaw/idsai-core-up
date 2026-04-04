package app

import (
	"context"
	"log"
	"strings"
	"time"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/services/notifications"
)

func startEmailOutboxDispatcher(ctx context.Context, cfg config.Config, repo notifications.OutboxRepository) {
	if !cfg.BackgroundJobsEnable {
		log.Printf("background jobs disabled by BACKGROUND_JOBS_ENABLED=false")
		return
	}
	if !cfg.EmailEnable {
		log.Printf("email notifications disabled by EMAIL_ENABLED=false")
		return
	}

	emailSender := newEmailSender(cfg)
	mode := emailSenderMode(cfg)
	if emailSender == nil {
		log.Printf("email outbox dispatcher disabled: email config incomplete (RESEND_API_KEY + SMTP_FROM, or SMTP_HOST/SMTP_PORT/SMTP_FROM)")
		return
	}
	if mode == "smtp" && (looksLikePlaceholder(cfg.SMTPUser) || looksLikePlaceholder(cfg.SMTPPass) || looksLikePlaceholder(cfg.SMTPFrom)) {
		log.Printf("email config appears to use placeholder values; update SMTP_USER/SMTP_PASS/SMTP_FROM")
	}
	log.Printf("email outbox dispatcher using %s transport", mode)
	dispatcher := notifications.NewOutboxDispatcher(repo, emailSender)
	pollEvery := time.Duration(cfg.OutboxPollS) * time.Second
	go dispatcher.Start(ctx, pollEvery)
}

func looksLikePlaceholder(v string) bool {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		return false
	}
	return strings.Contains(s, "change-me") ||
		strings.Contains(s, "changeme") ||
		strings.Contains(s, "your_") ||
		strings.Contains(s, "example.com")
}
