package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ResendSender struct {
	apiKey     string
	fromHeader string
	baseURL    string
	httpClient *http.Client
}

type resendSendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

type resendErrorResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{
		apiKey:     strings.TrimSpace(apiKey),
		fromHeader: normalizeFromHeader(from),
		baseURL:    "https://api.resend.com",
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *ResendSender) Send(ctx context.Context, to, subject, body string) error {
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)
	body = strings.TrimSpace(body)
	if s.apiKey == "" || s.fromHeader == "" {
		return errors.New("resend is not configured")
	}
	if to == "" {
		return errors.New("email recipient is required")
	}
	if body == "" {
		body = "(empty)"
	}

	payload, err := json.Marshal(resendSendRequest{
		From:    s.fromHeader,
		To:      []string{to},
		Subject: subject,
		Text:    body,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(s.baseURL, "/")+"/emails", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	var apiErr resendErrorResponse
	if len(respBody) > 0 && json.Unmarshal(respBody, &apiErr) == nil {
		msg := strings.TrimSpace(apiErr.Message)
		if msg != "" {
			return fmt.Errorf("resend http status %d: %s", resp.StatusCode, msg)
		}
	}

	msg := strings.TrimSpace(string(respBody))
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}
	return fmt.Errorf("resend http status %d: %s", resp.StatusCode, msg)
}
