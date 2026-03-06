package alerts

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
