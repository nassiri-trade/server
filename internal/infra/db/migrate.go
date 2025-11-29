package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"trading_server/internal/infra/repository"
)

func ApplyMigrations(ctx context.Context, db *gorm.DB, _ string) error {
	if err := db.WithContext(ctx).AutoMigrate(
		&repository.CalendarEventModel{},
		&repository.UserModel{},
		&repository.UserPositionModel{},
		&repository.UserPositionSnapshotModel{},
		&repository.UserTradeModel{},
		&repository.PasskeyModel{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}
