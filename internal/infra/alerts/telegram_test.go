package alerts

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNotifyCritical_DeduplicatesByError(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	n := NewTelegramNotifier("token", "chat", "idsai", time.Second, time.Minute)
	n.baseURL = srv.URL

	if err := n.NotifyCritical(context.Background(), errors.New("db timeout")); err != nil {
		t.Fatalf("first notify failed: %v", err)
	}
	if err := n.NotifyCritical(context.Background(), errors.New("db timeout")); err != nil {
		t.Fatalf("second notify failed: %v", err)
	}

	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("expected 1 telegram request, got %d", got)
	}
}

func TestNotifyHeartbeat_NoDeduplication(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	n := NewTelegramNotifier("token", "chat", "idsai", time.Second, time.Minute)
	n.baseURL = srv.URL

	if err := n.NotifyHeartbeat(context.Background()); err != nil {
		t.Fatalf("first heartbeat failed: %v", err)
	}
	if err := n.NotifyHeartbeat(context.Background()); err != nil {
		t.Fatalf("second heartbeat failed: %v", err)
	}

	if got := atomic.LoadInt32(&hits); got != 2 {
		t.Fatalf("expected 2 telegram requests, got %d", got)
	}
}

func TestDisabledNotifier_NoRequests(t *testing.T) {
	n := NewTelegramNotifier("", "", "idsai", time.Second, time.Minute)
	if n.Enabled() {
		t.Fatalf("notifier must be disabled without token and chat")
	}
	if err := n.NotifyStarted(context.Background()); err != nil {
		t.Fatalf("disabled notifier should return nil, got %v", err)
	}
}

func TestNotifyStarted_RetriesViaIPv4WhenIPv6Unavailable(t *testing.T) {
	var primaryHits int32
	var ipv4Hits int32

	n := NewTelegramNotifier("token", "chat", "idsai", time.Second, time.Minute)
	n.httpClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&primaryHits, 1)
			return nil, errors.New(`Post "https://api.telegram.org/bottoken/sendMessage": dial tcp [2001:67c:4e8:f004::9]:443: connect: network is unreachable`)
		}),
	}
	n.ipv4Client = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&ipv4Hits, 1)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"ok": true}`)),
			}, nil
		}),
	}

	if err := n.NotifyStarted(context.Background()); err != nil {
		t.Fatalf("notify with ipv4 retry failed: %v", err)
	}

	if got := atomic.LoadInt32(&primaryHits); got != 1 {
		t.Fatalf("expected 1 primary attempt, got %d", got)
	}
	if got := atomic.LoadInt32(&ipv4Hits); got != 1 {
		t.Fatalf("expected 1 ipv4 retry, got %d", got)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
