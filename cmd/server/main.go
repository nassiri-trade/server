package main

import (
	"context"
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
	applogger "trading_server/internal/infra/logger"
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

	applogger.Init("info") // Initialize with default level first
	logger := applogger.Logger

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("load config")
	}

	applogger.Init(cfg.Logging.Level)
	logger = applogger.Logger
	logger.Info().Str("level", cfg.Logging.Level).Msg("logger initialized")

	docs.SwaggerInfo.Title = "Forex Trading Server API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "API for ForexFactory calendar data, live positions, trade history, and analytics."
	docs.SwaggerInfo.BasePath = "/api/v1"

	logger.Info().Str("dsn", maskDSN(cfg.Database.DSN)).Msg("connecting to database")
	gormDB, err := db.Connect(rootCtx, cfg.Database.DSN)
	if err != nil {
		logger.Fatal().Err(err).Msg("connect database")
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		logger.Fatal().Err(err).Msg("underlying sql db")
	}
	defer sqlDB.Close()
	logger.Info().Msg("database connected successfully")

	migrationsPath := filepath.Join("scripts", "migrations")
	logger.Info().Str("path", migrationsPath).Msg("applying migrations")
	if err := db.ApplyMigrations(rootCtx, gormDB, migrationsPath); err != nil {
		logger.Fatal().Err(err).Msg("apply migrations")
	}
	logger.Info().Msg("migrations applied successfully")

	logger.Info().Str("url", cfg.Feed.URL).Msg("initializing feed client")
	feed, err := httpclient.NewForexFactoryFeed(cfg.Feed.URL)
	if err != nil {
		logger.Fatal().Err(err).Msg("init feed client")
	}

	repo, err := repository.NewGormEventRepository(gormDB)
	if err != nil {
		logger.Fatal().Err(err).Msg("init repository")
	}

	service, err := usecase.NewCalendarService(feed, repo)
	if err != nil {
		logger.Fatal().Err(err).Msg("init calendar service")
	}

	positionRepo, err := repository.NewGormPositionRepository(gormDB)
	if err != nil {
		logger.Fatal().Err(err).Msg("init position repository")
	}
	tradeRepo, err := repository.NewGormTradeRepository(gormDB)
	if err != nil {
		logger.Fatal().Err(err).Msg("init trade repository")
	}
	userRepo, err := repository.NewGormUserRepository(gormDB)
	if err != nil {
		logger.Fatal().Err(err).Msg("init user repository")
	}
	tradingService, err := usecase.NewTradingService(positionRepo, tradeRepo, userRepo)
	if err != nil {
		logger.Fatal().Err(err).Msg("init trading service")
	}

	passkeyRepo, err := repository.NewGormPasskeyRepository(gormDB)
	if err != nil {
		logger.Fatal().Err(err).Msg("init passkey repository")
	}
	passkeyService, err := usecase.NewPasskeyService(passkeyRepo)
	if err != nil {
		logger.Fatal().Err(err).Msg("init passkey service")
	}

	logger.Info().Msg("all services initialized")

	router := httptransport.New(service, tradingService, passkeyService)

	logger.Info().Dur("interval", cfg.Scheduler.Interval).Msg("initializing scheduler")
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logger.Fatal().Err(err).Msg("init scheduler")
	}
	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			logger.Error().Err(err).Msg("scheduler shutdown error")
		}
	}()

	_, err = scheduler.NewJob(
		gocron.DurationJob(cfg.Scheduler.Interval),
		gocron.NewTask(func(ctx context.Context) {
			logger.Info().Msg("scheduled calendar sync started")
			count, err := service.Sync(ctx)
			if err != nil && err != usecase.ErrNoEvents {
				logger.Error().Err(err).Msg("scheduler sync error")
			} else if err == nil {
				logger.Info().Int("count", count).Msg("scheduled calendar sync completed")
			}
		}),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("schedule job")
	}
	scheduler.Start()
	logger.Info().Msg("scheduler started")

	go func() {
		logger.Info().Msg("initial calendar sync started")
		count, err := service.Sync(context.Background())
		if err != nil && err != usecase.ErrNoEvents {
			logger.Error().Err(err).Msg("initial sync error")
		} else if err == nil {
			logger.Info().Int("count", count).Msg("initial calendar sync completed")
		}
	}()

	serverErr := make(chan error, 1)
	go func() {
		addr := ":" + cfg.Server.Port
		logger.Info().Str("addr", addr).Msg("server listening")
		serverErr <- router.App().Listen(addr)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			logger.Fatal().Err(err).Msg("fiber server error")
		}
	case sig := <-signalCh:
		logger.Info().Str("signal", sig.String()).Msg("received signal, shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := router.App().ShutdownWithContext(ctx); err != nil {
			logger.Error().Err(err).Msg("server shutdown error")
		}
		logger.Info().Msg("server shutdown complete")
	}
}

func maskDSN(dsn string) string {
	// Simple masking to hide password in logs
	// For postgres://user:pass@host:port/db format
	if len(dsn) > 20 {
		return dsn[:10] + "***" + dsn[len(dsn)-10:]
	}
	return "***"
}
