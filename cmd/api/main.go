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
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"idsai-core-up/internal/app"
)

func main() {
	a, err := app.New(context.Background())
	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}
	defer a.DB.Close()

	srv := &http.Server{
		Addr:              a.Cfg.Addr,
		Handler:           a.HTTP,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		baseURL := resolveBaseURL(a.Cfg.Addr)
		log.Printf("listening on %s", a.Cfg.Addr)
		log.Printf("authorization: %s/dev/login", baseURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
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
