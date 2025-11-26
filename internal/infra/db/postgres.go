package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	applogger "trading_server/internal/infra/logger"
)

// zerologWriter adapts zerolog.Logger to gorm logger.Writer interface
type zerologWriter struct {
	logger zerolog.Logger
}

func (w *zerologWriter) Printf(format string, v ...interface{}) {
	w.logger.Warn().Msg(fmt.Sprintf(format, v...))
}

func Connect(ctx context.Context, dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn required")
	}

	gormLogger := applogger.Logger.With().Str("component", "gorm").Logger()
	writer := &zerologWriter{logger: gormLogger}

	newLogger := logger.New(
		writer,
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("underlying db: %w", err)
	}

	// Configure connection pool before ping
	configurePool(sqlDB)

	// Retry ping with backoff
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if err := sqlDB.PingContext(pingCtx); err == nil {
			return gormDB, nil
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
		}
	}

	_ = sqlDB.Close()
	return nil, fmt.Errorf("ping database: failed after %d retries", maxRetries)
}

func configurePool(db *sql.DB) {
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(time.Hour)
}
