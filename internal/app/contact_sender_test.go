package app

import (
	"context"
	"errors"
	"testing"

	"idsai-core-up/internal/config"

	"github.com/stretchr/testify/require"
)

type fakeDirectEmailSender struct {
	to      string
	subject string
	body    string
	err     error
}

func (f *fakeDirectEmailSender) Send(_ context.Context, to, subject, body string) error {
	f.to = to
	f.subject = subject
	f.body = body
	return f.err
}

type fakeContactSender struct {
	err   error
	calls int
}

func (f *fakeContactSender) SendText(_ context.Context, text string) error {
	f.calls++
	if text == "" {
		return errors.New("empty")
	}
	return f.err
}

func TestContactEmailRecipientPrefersExplicitRecipient(t *testing.T) {
	got := contactEmailRecipient(config.Config{
		ContactEmailTo: "owner@example.com",
		SMTPUser:       "smtp@example.com",
		SMTPFrom:       "IDSAI <from@example.com>",
	})
	require.Equal(t, "owner@example.com", got)
}

func TestContactEmailRecipientFallsBackToSMTPUser(t *testing.T) {
	got := contactEmailRecipient(config.Config{
		SMTPUser: "smtp@example.com",
		SMTPFrom: "IDSAI <from@example.com>",
	})
	require.Equal(t, "smtp@example.com", got)
}

func TestContactEmailSenderUsesSMTPRecipient(t *testing.T) {
	fake := &fakeDirectEmailSender{}
	sender := contactEmailSender{
		sender:  fake,
		to:      "owner@example.com",
		subject: "IDSAI: новое сообщение с landing page",
	}

	err := sender.SendText(context.Background(), "hello")
	require.NoError(t, err)
	require.Equal(t, "owner@example.com", fake.to)
	require.Equal(t, "IDSAI: новое сообщение с landing page", fake.subject)
	require.Equal(t, "hello", fake.body)
}

func TestMultiContactSenderFallsBackToNextSender(t *testing.T) {
	first := &fakeContactSender{err: errors.New("telegram down")}
	second := &fakeContactSender{}

	err := multiContactSender{senders: []contactSender{first, second}}.SendText(context.Background(), "hello")
	require.NoError(t, err)
	require.Equal(t, 1, first.calls)
	require.Equal(t, 1, second.calls)
}
