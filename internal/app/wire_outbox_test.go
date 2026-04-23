package app

import (
	"context"
	"testing"

	"idsai-core-up/internal/config"
)

func TestLooksLikePlaceholder(t *testing.T) {
	for _, value := range []string{"change-me", "changeme", "your_password", "admin@example.com"} {
		if !looksLikePlaceholder(value) {
			t.Fatalf("expected %q to be treated as placeholder", value)
		}
	}
	for _, value := range []string{"", "smtp-user", "alerts@idsai.local"} {
		if looksLikePlaceholder(value) {
			t.Fatalf("did not expect %q to be treated as placeholder", value)
		}
	}
}

func TestStartEmailOutboxDispatcherSkipsWhenDisabledOrIncomplete(t *testing.T) {
	ctx := context.Background()

	startEmailOutboxDispatcher(ctx, config.Config{
		BackgroundJobsEnable: false,
		EmailEnable:          true,
	}, nil)

	startEmailOutboxDispatcher(ctx, config.Config{
		BackgroundJobsEnable: true,
		EmailEnable:          false,
	}, nil)

	startEmailOutboxDispatcher(ctx, config.Config{
		BackgroundJobsEnable: true,
		EmailEnable:          true,
		SMTPHost:             "smtp.example.local",
		SMTPPort:             "",
		SMTPFrom:             "noreply@idsai.local",
	}, nil)

	startEmailOutboxDispatcher(ctx, config.Config{
		BackgroundJobsEnable: true,
		EmailEnable:          true,
		SMTPHost:             "smtp.example.local",
		SMTPPort:             "587",
		SMTPUser:             "change-me",
		SMTPPass:             "your_password",
		SMTPFrom:             "admin@example.com",
		SMTPSendTimeoutS:     1,
		OutboxPollS:          1,
	}, nil)
}
