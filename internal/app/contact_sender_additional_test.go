package app

import (
	"context"
	"errors"
	"testing"

	"idsai-core-up/internal/config"

	"github.com/stretchr/testify/require"
)

func TestNewPublicContactSenderAndEmailHelpers(t *testing.T) {
	require.Nil(t, newPublicContactSender(config.Config{}))
	require.Nil(t, newContactEmailSender(config.Config{
		SMTPHost: "smtp.example.com",
		SMTPPort: "587",
		SMTPFrom: "",
	}))

	require.Equal(t, "from@example.com", contactEmailRecipient(config.Config{
		SMTPFrom: "IDSAI <from@example.com>",
	}))
	require.Equal(t, "plain@example.com", contactEmailRecipient(config.Config{
		SMTPFrom: "plain@example.com",
	}))
	require.Empty(t, contactEmailRecipient(config.Config{
		SMTPFrom: "broken <address",
	}))
}

func TestContactAndMultiSenderErrorPaths(t *testing.T) {
	err := contactEmailSender{
		sender:  &fakeDirectEmailSender{},
		to:      "owner@example.com",
		subject: "subject",
	}.SendText(context.Background(), "   ")
	require.EqualError(t, err, "contact email is empty")

	err = multiContactSender{}.SendText(context.Background(), "hello")
	require.EqualError(t, err, "contact delivery unavailable")

	first := &fakeContactSender{err: errors.New("telegram down")}
	second := &fakeContactSender{err: errors.New("email down")}
	err = multiContactSender{senders: []contactSender{first, nil, second}}.SendText(context.Background(), "hello")
	require.EqualError(t, err, "contact delivery failed: telegram down; email down")
}
