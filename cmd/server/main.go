package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"

	docs "trading_server/docs"
	"trading_server/internal/config"
	"trading_server/internal/infra/db"
	"trading_server/internal/infra/httpclient"
	"trading_server/internal/infra/repository"
	httptransport "trading_server/internal/transport/http"
	"trading_server/internal/usecase"
)

// @title Forex Trading Server API
// @version 1.0
// @description API for ForexFactory calendar data, live positions, trade history, and analytics.
// @BasePath /api/v1

func main() {
	rootCtx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	docs.SwaggerInfo.Title = "Forex Trading Server API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "API for ForexFactory calendar data, live positions, trade history, and analytics."
	docs.SwaggerInfo.BasePath = "/api/v1"

	gormDB, err := db.Connect(rootCtx, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("underlying sql db: %v", err)
	}
	defer sqlDB.Close()

	migrationsPath := filepath.Join("scripts", "migrations")
	if err := db.ApplyMigrations(rootCtx, gormDB, migrationsPath); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	feed, err := httpclient.NewForexFactoryFeed(cfg.Feed.URL)
	if err != nil {
		log.Fatalf("init feed client: %v", err)
	}

	repo, err := repository.NewGormEventRepository(gormDB)
	if err != nil {
		log.Fatalf("init repository: %v", err)
	}

	service, err := usecase.NewCalendarService(feed, repo)
	if err != nil {
		log.Fatalf("init calendar service: %v", err)
	}

	positionRepo, err := repository.NewGormPositionRepository(gormDB)
	if err != nil {
		log.Fatalf("init position repository: %v", err)
	}
	tradeRepo, err := repository.NewGormTradeRepository(gormDB)
	if err != nil {
		log.Fatalf("init trade repository: %v", err)
	}
	userRepo, err := repository.NewGormUserRepository(gormDB)
	if err != nil {
		log.Fatalf("init user repository: %v", err)
	}
	tradingService, err := usecase.NewTradingService(positionRepo, tradeRepo, userRepo)
	if err != nil {
		log.Fatalf("init trading service: %v", err)
	}

	router := httptransport.New(service, tradingService)

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("init scheduler: %v", err)
	}
	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Printf("scheduler shutdown error: %v", err)
		}
	}()

	_, err = scheduler.NewJob(
		gocron.DurationJob(cfg.Scheduler.Interval),
		gocron.NewTask(func(ctx context.Context) {
			if _, err := service.Sync(ctx); err != nil && err != usecase.ErrNoEvents {
				log.Printf("scheduler sync error: %v", err)
			}
		}),
	)
	if err != nil {
		log.Fatalf("schedule job: %v", err)
	}
	scheduler.Start()

	go func() {
		if _, err := service.Sync(context.Background()); err != nil && err != usecase.ErrNoEvents {
			log.Printf("initial sync error: %v", err)
		}
	}()

	serverErr := make(chan error, 1)
	go func() {
		addr := ":" + cfg.Server.Port
		log.Printf("server listening on %s", addr)
		serverErr <- router.App().Listen(addr)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("fiber server error: %v", err)
		}
	case sig := <-signalCh:
		log.Printf("received signal %s, shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := router.App().ShutdownWithContext(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}
}
