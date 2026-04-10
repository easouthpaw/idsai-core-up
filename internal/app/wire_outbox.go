package app

import (
	"context"
	"log"
	"strings"
	"time"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/infra/email"
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

	host := strings.TrimSpace(cfg.SMTPHost)
	port := strings.TrimSpace(cfg.SMTPPort)
	from := strings.TrimSpace(cfg.SMTPFrom)
	if host == "" || port == "" || from == "" {
		log.Printf("email outbox dispatcher disabled: SMTP config incomplete (SMTP_HOST/SMTP_PORT/SMTP_FROM)")
		return
	}
	if looksLikePlaceholder(cfg.SMTPUser) || looksLikePlaceholder(cfg.SMTPPass) || looksLikePlaceholder(cfg.SMTPFrom) {
		log.Printf("email config appears to use placeholder values; update SMTP_USER/SMTP_PASS/SMTP_FROM")
	}

	sendTimeout := time.Duration(cfg.SMTPSendTimeoutS) * time.Second
	emailSender := email.NewSMTPSenderWithTimeout(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom, sendTimeout)
	dispatcher := notifications.NewOutboxDispatcher(repo, emailSender)
	dispatcher.SetSendTimeout(sendTimeout)
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
