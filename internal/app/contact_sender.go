package app

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"idsai-core-up/internal/config"
	"idsai-core-up/internal/infra/alerts"
	"idsai-core-up/internal/infra/email"
)

type contactSender interface {
	SendText(ctx context.Context, text string) error
}

type directEmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type contactEmailSender struct {
	sender  directEmailSender
	to      string
	subject string
}

type multiContactSender struct {
	senders []contactSender
}

func newPublicContactSender(cfg config.Config) contactSender {
	senders := make([]contactSender, 0, 2)

	telegram := alerts.NewTelegramNotifier(
		cfg.TelegramBotToken,
		cfg.TelegramSuperadminChat,
		cfg.ServerName,
		time.Duration(cfg.TelegramRequestTimeoutS)*time.Second,
		0,
	)
	if telegram.Enabled() {
		senders = append(senders, telegram)
	}

	if emailSender := newContactEmailSender(cfg); emailSender != nil {
		senders = append(senders, emailSender)
	}

	switch len(senders) {
	case 0:
		return nil
	case 1:
		return senders[0]
	default:
		return multiContactSender{senders: senders}
	}
}

func newContactEmailSender(cfg config.Config) contactSender {
	host := strings.TrimSpace(cfg.SMTPHost)
	port := strings.TrimSpace(cfg.SMTPPort)
	from := strings.TrimSpace(cfg.SMTPFrom)
	to := contactEmailRecipient(cfg)
	if host == "" || port == "" || from == "" || to == "" {
		return nil
	}

	return contactEmailSender{
		sender:  email.NewSMTPSenderWithOptions(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom, time.Duration(cfg.SMTPSendTimeoutS)*time.Second, cfg.SMTPUseSSL),
		to:      to,
		subject: "IDSAI: новое сообщение с landing page",
	}
}

func contactEmailRecipient(cfg config.Config) string {
	if to := strings.TrimSpace(cfg.ContactEmailTo); to != "" {
		return to
	}
	if to := strings.TrimSpace(cfg.SMTPUser); to != "" {
		return to
	}
	from := strings.TrimSpace(cfg.SMTPFrom)
	if from == "" {
		return ""
	}
	addr, err := mail.ParseAddress(from)
	if err != nil {
		if strings.Contains(from, " ") || strings.Contains(from, "<") || strings.Contains(from, ">") {
			return ""
		}
		return from
	}
	return strings.TrimSpace(addr.Address)
}

func (s contactEmailSender) SendText(ctx context.Context, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("contact email is empty")
	}
	return s.sender.Send(ctx, s.to, s.subject, text)
}

func (s multiContactSender) SendText(ctx context.Context, text string) error {
	var errs []string
	for _, sender := range s.senders {
		if sender == nil {
			continue
		}
		if err := sender.SendText(ctx, text); err == nil {
			return nil
		} else {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return errors.New("contact delivery unavailable")
	}
	return fmt.Errorf("contact delivery failed: %s", strings.Join(errs, "; "))
}
