package auth

import (
	"sync"
	"time"
)

const (
	maxLimiterEntries     = 10000
	cleanupThrottleInterval = 60 * time.Second
)

type attemptLimiter struct {
	mu          sync.Mutex
	window      time.Duration
	maxFails    int
	now         func() time.Time
	entries     map[string]attemptWindow
	lastCleanup time.Time
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

	l.cleanupLocked()

	entry, ok := l.currentEntry(key)
	if !ok {
		return true
	}
	return entry.failures < l.maxFails
}

func (l *attemptLimiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanupLocked()

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

// cleanupLocked removes expired entries to prevent unbounded memory growth.
// Runs at most once per cleanupThrottleInterval. Must be called with l.mu held.
func (l *attemptLimiter) cleanupLocked() {
	now := l.now()
	if now.Sub(l.lastCleanup) < cleanupThrottleInterval {
		return
	}
	l.lastCleanup = now

	for key, entry := range l.entries {
		if !entry.resetAt.After(now) {
			delete(l.entries, key)
		}
	}

	// Hard cap: if map is still too large after expiry cleanup, evict oldest entries
	if len(l.entries) > maxLimiterEntries {
		excess := len(l.entries) - maxLimiterEntries
		for key := range l.entries {
			if excess <= 0 {
				break
			}
			delete(l.entries, key)
			excess--
		}
	}
}
