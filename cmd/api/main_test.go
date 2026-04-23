package main

import (
	"errors"
	"testing"
)

func TestResolveBaseURL(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{name: "http", addr: "http://127.0.0.1:8080/", want: "http://127.0.0.1:8080"},
		{name: "https", addr: "https://idsai.example/", want: "https://idsai.example"},
		{name: "port", addr: ":8080", want: "http://localhost:8080"},
		{name: "host port", addr: "127.0.0.1:9090", want: "http://127.0.0.1:9090"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveBaseURL(tt.addr); got != tt.want {
				t.Fatalf("resolveBaseURL(%q) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}

func TestRunGuardExecutesAndRecoversWithNilNotifier(t *testing.T) {
	called := false
	runGuard(nil, "worker", func() {
		called = true
	})
	if !called {
		t.Fatal("guarded function was not called")
	}

	runGuard(nil, "worker", func() {
		panic("boom")
	})
}

func TestNotifyHelpersIgnoreNilNotifier(t *testing.T) {
	notifyStarted(nil)
	notifyCritical(nil, errors.New("critical"))
	notifyDBFailure(nil, errors.New("db"))
	notifyUnhandledPanic(nil, "panic", []byte("stack"))
	notifyStopped(nil, "done")
}
