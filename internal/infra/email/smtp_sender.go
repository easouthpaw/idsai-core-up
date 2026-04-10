package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

type SMTPSender struct {
	host         string
	port         string
	user         string
	pass         string
	fromHeader   string
	fromEnvelope string
	timeout      time.Duration
}

func NewSMTPSender(host, port, user, pass, from string) *SMTPSender {
	return NewSMTPSenderWithTimeout(host, port, user, pass, from, 15*time.Second)
}

func NewSMTPSenderWithTimeout(host, port, user, pass, from string, timeout time.Duration) *SMTPSender {
	return &SMTPSender{
		host:         strings.TrimSpace(host),
		port:         strings.TrimSpace(port),
		user:         strings.TrimSpace(user),
		pass:         pass,
		fromHeader:   normalizeFromHeader(from),
		fromEnvelope: normalizeFromEnvelope(from),
		timeout:      timeout,
	}
}

func (s *SMTPSender) Send(ctx context.Context, to, subject, body string) error {
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)
	if s.host == "" || s.port == "" || s.fromEnvelope == "" {
		return errors.New("smtp is not configured")
	}
	if to == "" {
		return errors.New("email recipient is required")
	}
	addr := s.host + ":" + s.port
	msg := buildMessage(s.fromHeader, to, subject, body)

	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}

	if s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return err
		}
	}
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(s.fromEnvelope); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildMessage(from, to, subject, body string) string {
	if body == "" {
		body = "(empty)"
	}
	return fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n",
		from,
		to,
		subject,
		body,
	)
}

func normalizeFromHeader(from string) string {
	from = strings.TrimSpace(from)
	if from == "" {
		return ""
	}
	addr, err := mail.ParseAddress(from)
	if err != nil {
		return from
	}
	if addr.Name == "" {
		return addr.Address
	}
	return addr.String()
}

func normalizeFromEnvelope(from string) string {
	from = strings.TrimSpace(from)
	if from == "" {
		return ""
	}
	addr, err := mail.ParseAddress(from)
	if err == nil && strings.TrimSpace(addr.Address) != "" {
		return strings.TrimSpace(addr.Address)
	}
	// If parsing fails, fall back to raw value when it looks like a plain address.
	if strings.Contains(from, " ") || strings.Contains(from, "<") || strings.Contains(from, ">") {
		return ""
	}
	return from
}
