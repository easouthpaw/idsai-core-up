package alerts

import (
	"context"
	"errors"
	"log"
	"time"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthMonitor struct {
	pinger           Pinger
	notifier         *TelegramNotifier
	healthCheckEvery time.Duration
	heartbeatEvery   time.Duration
}

func NewHealthMonitor(pinger Pinger, notifier *TelegramNotifier, healthCheckEvery, heartbeatEvery time.Duration) *HealthMonitor {
	if healthCheckEvery <= 0 {
		healthCheckEvery = 20 * time.Second
	}
	return &HealthMonitor{
		pinger:           pinger,
		notifier:         notifier,
		healthCheckEvery: healthCheckEvery,
		heartbeatEvery:   heartbeatEvery,
	}
}

func (m *HealthMonitor) Start(ctx context.Context) {
	if m == nil || m.pinger == nil || m.notifier == nil || !m.notifier.Enabled() {
		return
	}

	healthTicker := time.NewTicker(m.healthCheckEvery)
	defer healthTicker.Stop()

	var hbTicker *time.Ticker
	if m.heartbeatEvery > 0 {
		hbTicker = time.NewTicker(m.heartbeatEvery)
		defer hbTicker.Stop()
	}

	wasFailing := false
	for {
		select {
		case <-ctx.Done():
			return
		case <-healthTicker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err := m.pinger.Ping(pingCtx)
			cancel()

			if err != nil {
				if !wasFailing {
					wasFailing = true
					if sendErr := m.notifier.NotifyDBFailure(ctx, err); sendErr != nil {
						log.Printf("telegram notify db failure failed: %v", sendErr)
					}
				}
				continue
			}

			if wasFailing {
				wasFailing = false
				if sendErr := m.notifier.NotifyRecovered(ctx, "database connectivity restored"); sendErr != nil {
					log.Printf("telegram notify recovered failed: %v", sendErr)
				}
			}
		case <-heartbeat(hbTicker):
			if sendErr := m.notifier.NotifyHeartbeat(ctx); sendErr != nil {
				if !errors.Is(sendErr, context.Canceled) {
					log.Printf("telegram heartbeat notify failed: %v", sendErr)
				}
			}
		}
	}
}

func heartbeat(ticker *time.Ticker) <-chan time.Time {
	if ticker == nil {
		return nil
	}
	return ticker.C
}
