package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"trading_server/internal/domain"
)

type GormPositionRepository struct {
	db *gorm.DB
}

func NewGormPositionRepository(db *gorm.DB) (*GormPositionRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &GormPositionRepository{db: db}, nil
}

func (r *GormPositionRepository) UpsertPosition(ctx context.Context, position domain.UserPosition) error {
	model := toUserPositionModel(position)

	assignments := clause.Assignments(map[string]interface{}{
		"user_id":       gorm.Expr("EXCLUDED.user_id"),
		"platform":      gorm.Expr("EXCLUDED.platform"),
		"account_name":  gorm.Expr("EXCLUDED.account_name"),
		"account_login": gorm.Expr("EXCLUDED.account_login"),
		"broker_server": gorm.Expr("EXCLUDED.broker_server"),
		"broker_name":   gorm.Expr("EXCLUDED.broker_name"),
		"balance":       gorm.Expr("EXCLUDED.balance"),
		"status":        gorm.Expr("EXCLUDED.status"),
		"symbol":        gorm.Expr("EXCLUDED.symbol"),
		"side":          gorm.Expr("EXCLUDED.side"),
		"volume":        gorm.Expr("EXCLUDED.volume"),
		"entry_time":    gorm.Expr("EXCLUDED.entry_time"),
		"entry_price":   gorm.Expr("EXCLUDED.entry_price"),
		"stop_loss":     gorm.Expr("EXCLUDED.stop_loss"),
		"take_profit":   gorm.Expr("EXCLUDED.take_profit"),
		"profit":        gorm.Expr("EXCLUDED.profit"),
		"risk_percent":  gorm.Expr("EXCLUDED.risk_percent"),
		"tick_value":    gorm.Expr("EXCLUDED.tick_value"),
		"tick_size":     gorm.Expr("EXCLUDED.tick_size"),
		"magic":         gorm.Expr("EXCLUDED.magic"),
		"reason":        gorm.Expr("EXCLUDED.reason"),
		"comment":       gorm.Expr("EXCLUDED.comment"),
		"last_update":   gorm.Expr("EXCLUDED.last_update"),
		"raw_payload":   gorm.Expr("EXCLUDED.raw_payload"),
		"updated_at":    gorm.Expr("CURRENT_TIMESTAMP"),
	})

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "ticket"}},
			DoUpdates: assignments,
		}).
		Create(&model).Error
}

func (r *GormPositionRepository) ClosePosition(ctx context.Context, ticket int64, closeTime time.Time, profit float64) error {
	result := r.db.WithContext(ctx).Model(&UserPositionModel{}).
		Where("ticket = ?", ticket).
		Updates(map[string]interface{}{
			"status":      domain.PositionStatusClosed,
			"profit":      profit,
			"last_update": closeTime,
			"updated_at":  gorm.Expr("CURRENT_TIMESTAMP"),
		})

	return result.Error
}

func (r *GormPositionRepository) ListOpenPositions(ctx context.Context, userID string) ([]domain.UserPosition, error) {
	var models []UserPositionModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.PositionStatusOpen).
		Order("last_update DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	positions := make([]domain.UserPosition, len(models))
	for i, model := range models {
		positions[i] = model.toDomain()
	}

	return positions, nil
}

func (r *GormPositionRepository) InsertPositionSnapshot(ctx context.Context, position domain.UserPosition) error {
	model := toUserPositionSnapshotModel(position)
	return r.db.WithContext(ctx).Create(&model).Error
}
