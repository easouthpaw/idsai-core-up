package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResendSenderSendPostsExpectedPayload(t *testing.T) {
	var gotAuth string
	var gotReq resendSendRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/emails", r.URL.Path)
		gotAuth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotReq))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"email_123"}`))
	}))
	defer server.Close()

	sender := NewResendSender("re_test_key", "Acme <noreply@example.com>")
	sender.baseURL = server.URL
	sender.httpClient = server.Client()

	err := sender.Send(context.Background(), "student@example.com", "Reset password", "Code: 123456")
	require.NoError(t, err)
	require.Equal(t, "Bearer re_test_key", gotAuth)
	require.Equal(t, "\"Acme\" <noreply@example.com>", gotReq.From)
	require.Equal(t, []string{"student@example.com"}, gotReq.To)
	require.Equal(t, "Reset password", gotReq.Subject)
	require.Equal(t, "Code: 123456", gotReq.Text)
}

func TestResendSenderSendReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Invalid API key"}`))
	}))
	defer server.Close()

	sender := NewResendSender("re_bad_key", "Acme <noreply@example.com>")
	sender.baseURL = server.URL
	sender.httpClient = server.Client()

	err := sender.Send(context.Background(), "student@example.com", "Reset password", "Code: 123456")
	require.EqualError(t, err, "resend http status 401: Invalid API key")
}
