package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(ctx context.Context, dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("sqlite dsn required")
	}

	if err := ensureDirectory(dsn); err != nil {
		return nil, err
	}

	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("underlying db: %w", err)
	}

	configurePool(sqlDB)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return gormDB, nil
}

func configurePool(db *sql.DB) {
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(0)
}

func ensureDirectory(dsn string) error {
	path := sqliteFilePath(dsn)
	if path == "" {
		return nil
	}

	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create sqlite directory: %w", err)
	}
	return nil
}

func sqliteFilePath(dsn string) string {
	if dsn == ":memory:" {
		return ""
	}

	trimmed := strings.TrimPrefix(dsn, "file:")
	trimmed = strings.TrimPrefix(trimmed, "//")

	if idx := strings.IndexRune(trimmed, '?'); idx >= 0 {
		trimmed = trimmed[:idx]
	}

	if trimmed == "" || trimmed == ":memory:" {
		return ""
	}

	return trimmed
}
