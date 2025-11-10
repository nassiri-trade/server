package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"trading_server/internal/infra/repository"
)

func ApplyMigrations(ctx context.Context, db *gorm.DB, _ string) error {
	if err := db.WithContext(ctx).Exec("PRAGMA foreign_keys = ON;").Error; err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := db.WithContext(ctx).AutoMigrate(
		&repository.CalendarEventModel{},
		&repository.UserModel{},
		&repository.UserPositionModel{},
		&repository.UserPositionSnapshotModel{},
		&repository.UserTradeModel{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}
