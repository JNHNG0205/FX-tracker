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
	"fx-tracker/internal/price"
	"fx-tracker/internal/repository"
	"fx-tracker/internal/router"
	"fx-tracker/internal/service"
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

	// 10s timeout so a hung upstream can't stall a rate fetch indefinitely.
	httpClient := &http.Client{Timeout: 10 * time.Second}
	cache := fx.NewCache(httpClient, cfg.FxBaseURL)

	convRepo := repository.NewConversionRepository(db)
	convService := service.NewConversionService(convRepo)
	settingsRepo := repository.NewSettingsRepository(db)

	priceCache := price.NewCache(httpClient, cfg.FinnhubAPIKey)
	holdingRepo := repository.NewHoldingRepository(db)
	holdingService := service.NewHoldingService(holdingRepo)
	portfolioService := service.NewPortfolioService(holdingRepo, priceCache, convRepo, cache)

	h := handler.New(cache, convService, settingsRepo, holdingService, portfolioService)
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
