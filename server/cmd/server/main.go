package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"fx-tracker/internal/config"
	"fx-tracker/internal/fx"
	"fx-tracker/internal/handler"
	"fx-tracker/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := config.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	if err := config.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 10s timeout so a hung upstream can't stall the refresh loop indefinitely.
	httpClient := &http.Client{Timeout: 10 * time.Second}
	cache := fx.NewCache(httpClient, cfg.FxBaseURL)
	go cache.Run(ctx, 10*time.Minute)

	history := fx.NewHistoryCache(httpClient, cfg.FxBaseURL)
	// Historical series changes at most daily; refresh once a day.
	go history.Run(ctx, 24*time.Hour)

	h := handler.New(cache, history)
	r := router.New(h)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("server stopped")
}
