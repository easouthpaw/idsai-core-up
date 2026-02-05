package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"idsai-core-up/internal/app"
)

// @title IDSAI Core API
// @version 0.1
// @description Core platform for IDSAI projects (RBAC-driven).
// @BasePath /
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
		log.Printf("listening on %s", a.Cfg.Addr)
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
