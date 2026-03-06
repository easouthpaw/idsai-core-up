// @title IDSAI Core API
// @version 0.1
// @description Core platform for IDSAI projects (RBAC-driven).
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"idsai-core-up/internal/app"
	"idsai-core-up/internal/config"
	"idsai-core-up/internal/infra/alerts"

	"github.com/gin-gonic/gin"
)

func main() {
	if strings.TrimSpace(os.Getenv("GIN_MODE")) == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg := config.Load()
	notifier := alerts.NewTelegramNotifier(
		cfg.TelegramBotToken,
		cfg.TelegramSuperadminChat,
		cfg.ServerName,
		time.Duration(cfg.TelegramRequestTimeoutS)*time.Second,
		time.Duration(cfg.TelegramDedupeWindowS)*time.Second,
	)

	defer func() {
		if recovered := recover(); recovered != nil {
			stack := debug.Stack()
			notifyUnhandledPanic(notifier, recovered, stack)
			panic(recovered)
		}
	}()

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(rootCtx)
	if err != nil {
		log.Printf("app init failed: %v", err)
		notifyDBFailure(notifier, err)
		os.Exit(1)
	}
	defer a.DB.Close()

	srv := &http.Server{
		Addr:              a.Cfg.Addr,
		Handler:           a.HTTP,
		ReadHeaderTimeout: 5 * time.Second,
	}

	healthMonitor := alerts.NewHealthMonitor(
		a.DB,
		notifier,
		time.Duration(a.Cfg.HealthcheckPollS)*time.Second,
		time.Duration(a.Cfg.HeartbeatS)*time.Second,
	)
	go runGuard(notifier, "health monitor", func() {
		healthMonitor.Start(rootCtx)
	})

	ln, err := net.Listen("tcp", a.Cfg.Addr)
	if err != nil {
		log.Printf("server listen failed: %v", err)
		notifyCritical(notifier, fmt.Errorf("listen failed: %w", err))
		a.DB.Close()
		os.Exit(1)
	}

	serveErrCh := make(chan error, 1)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				stack := debug.Stack()
				notifyUnhandledPanic(notifier, recovered, stack)
				select {
				case serveErrCh <- fmt.Errorf("http server panic: %v", recovered):
				default:
				}
			}
		}()

		baseURL := resolveBaseURL(a.Cfg.Addr)
		log.Printf("listening on %s", a.Cfg.Addr)
		log.Printf("authorization: %s/dev/login", baseURL)
		notifyStarted(notifier)

		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case serveErrCh <- err:
			default:
			}
		}
	}()

	exitCode := 0
	gracefulStop := false
	select {
	case <-rootCtx.Done():
		gracefulStop = true
	case err := <-serveErrCh:
		log.Printf("server failed: %v", err)
		notifyCritical(notifier, err)
		exitCode = 1
	}

	if gracefulStop {
		notifyStopped(notifier, "signal received")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	if exitCode != 0 {
		a.DB.Close()
		os.Exit(exitCode)
	}
}

func resolveBaseURL(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return strings.TrimRight(addr, "/")
	}
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	return "http://" + strings.TrimRight(addr, "/")
}

func runGuard(notifier *alerts.TelegramNotifier, component string, fn func()) {
	defer func() {
		if recovered := recover(); recovered != nil {
			stack := debug.Stack()
			notifyUnhandledPanic(notifier, fmt.Sprintf("%s panic: %v", component, recovered), stack)
		}
	}()
	fn()
}

func notifyStarted(notifier *alerts.TelegramNotifier) {
	if notifier == nil || !notifier.Enabled() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := notifier.NotifyStarted(ctx); err != nil {
		log.Printf("telegram notify started failed: %v", err)
	}
}

func notifyCritical(notifier *alerts.TelegramNotifier, err error) {
	if notifier == nil || !notifier.Enabled() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if sendErr := notifier.NotifyCritical(ctx, err); sendErr != nil {
		log.Printf("telegram notify critical failed: %v", sendErr)
	}
}

func notifyDBFailure(notifier *alerts.TelegramNotifier, err error) {
	if notifier == nil || !notifier.Enabled() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if sendErr := notifier.NotifyDBFailure(ctx, err); sendErr != nil {
		log.Printf("telegram notify db failure failed: %v", sendErr)
	}
}

func notifyUnhandledPanic(notifier *alerts.TelegramNotifier, recovered any, stack []byte) {
	if notifier == nil || !notifier.Enabled() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := notifier.NotifyUnhandledPanic(ctx, recovered, stack); err != nil {
		log.Printf("telegram notify panic failed: %v", err)
	}
}

func notifyStopped(notifier *alerts.TelegramNotifier, reason string) {
	if notifier == nil || !notifier.Enabled() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := notifier.NotifyStopped(ctx, reason); err != nil {
		log.Printf("telegram notify stopped failed: %v", err)
	}
}
