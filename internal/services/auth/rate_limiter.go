package auth

import (
	"sync"
	"time"
)

type attemptLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	maxFails int
	now      func() time.Time
	entries  map[string]attemptWindow
}

type attemptWindow struct {
	failures int
	resetAt  time.Time
}

func newAttemptLimiter(window time.Duration, maxFails int) *attemptLimiter {
	if window <= 0 {
		window = 15 * time.Minute
	}
	if maxFails <= 0 {
		maxFails = 5
	}

	return &attemptLimiter{
		window:   window,
		maxFails: maxFails,
		now:      time.Now,
		entries:  make(map[string]attemptWindow),
	}
}

func (l *attemptLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.currentEntry(key)
	if !ok {
		return true
	}
	return entry.failures < l.maxFails
}

func (l *attemptLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	entry, ok := l.currentEntry(key)
	if !ok {
		l.entries[key] = attemptWindow{
			failures: 1,
			resetAt:  now.Add(l.window),
		}
		return
	}

	entry.failures++
	l.entries[key] = entry
}

func (l *attemptLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

func (l *attemptLimiter) currentEntry(key string) (attemptWindow, bool) {
	entry, ok := l.entries[key]
	if !ok {
		return attemptWindow{}, false
	}
	if !entry.resetAt.After(l.now()) {
		delete(l.entries, key)
		return attemptWindow{}, false
	}
	return entry, true
}
