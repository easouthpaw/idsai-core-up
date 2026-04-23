package alerts

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type sequencePinger struct {
	mu    sync.Mutex
	calls int
}

func (p *sequencePinger) Ping(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.calls++
	if p.calls == 1 {
		return errors.New("database unavailable")
	}
	return nil
}

func TestHealthMonitorDefaultsAndEarlyReturns(t *testing.T) {
	monitor := NewHealthMonitor(&sequencePinger{}, nil, 0, 0)
	if monitor.healthCheckEvery != 20*time.Second {
		t.Fatalf("expected default health interval, got %s", monitor.healthCheckEvery)
	}
	if heartbeat(nil) != nil {
		t.Fatal("nil heartbeat ticker should return nil channel")
	}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	if heartbeat(ticker) != ticker.C {
		t.Fatal("heartbeat should expose ticker channel")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	(*HealthMonitor)(nil).Start(ctx)
	NewHealthMonitor(nil, NewTelegramNotifier("token", "chat", "test", time.Millisecond, 0), time.Millisecond, 0).Start(ctx)
	NewHealthMonitor(&sequencePinger{}, nil, time.Millisecond, 0).Start(ctx)
	NewHealthMonitor(&sequencePinger{}, NewTelegramNotifier("", "", "test", time.Millisecond, 0), time.Millisecond, 0).Start(ctx)
}

func TestHealthMonitorSendsFailureRecoveryAndHeartbeat(t *testing.T) {
	var mu sync.Mutex
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	notifier := NewTelegramNotifier("token", "chat", "test", 100*time.Millisecond, 0)
	notifier.baseURL = server.URL
	notifier.httpClient = server.Client()
	notifier.ipv4Client = nil

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	monitor := NewHealthMonitor(&sequencePinger{}, notifier, time.Millisecond, time.Millisecond)
	monitor.Start(ctx)

	mu.Lock()
	defer mu.Unlock()
	if requests == 0 {
		t.Fatal("expected health monitor to send at least one telegram request")
	}
}
