package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"trading_server/internal/domain"
)

type GormTradeRepository struct {
	db *gorm.DB
}

func NewGormTradeRepository(db *gorm.DB) (*GormTradeRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &GormTradeRepository{db: db}, nil
}

func (r *GormTradeRepository) AddTrade(ctx context.Context, trade domain.UserTrade) error {
	model := toUserTradeModel(trade)

	assignments := clause.Assignments(map[string]interface{}{
		"position_ticket": gorm.Expr("EXCLUDED.position_ticket"),
		"platform":        gorm.Expr("EXCLUDED.platform"),
		"symbol":          gorm.Expr("EXCLUDED.symbol"),
		"side":            gorm.Expr("EXCLUDED.side"),
		"entry_type":      gorm.Expr("EXCLUDED.entry_type"),
		"volume":          gorm.Expr("EXCLUDED.volume"),
		"entry_time":      gorm.Expr("EXCLUDED.entry_time"),
		"entry_price":     gorm.Expr("EXCLUDED.entry_price"),
		"stop_loss":       gorm.Expr("EXCLUDED.stop_loss"),
		"take_profit":     gorm.Expr("EXCLUDED.take_profit"),
		"profit":          gorm.Expr("EXCLUDED.profit"),
		"risk_percent":    gorm.Expr("EXCLUDED.risk_percent"),
		"tick_value":      gorm.Expr("EXCLUDED.tick_value"),
		"tick_size":       gorm.Expr("EXCLUDED.tick_size"),
		"magic":           gorm.Expr("EXCLUDED.magic"),
		"reason":          gorm.Expr("EXCLUDED.reason"),
		"comment":         gorm.Expr("EXCLUDED.comment"),
		"raw_payload":     gorm.Expr("EXCLUDED.raw_payload"),
		"updated_at":      gorm.Expr("CURRENT_TIMESTAMP"),
	})

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "ticket"}},
			DoUpdates: assignments,
		}).
		Create(&model).Error
}

func (r *GormTradeRepository) ListTrades(ctx context.Context, userID string, limit int) ([]domain.UserTrade, error) {
	var models []UserTradeModel
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("entry_time DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	trades := make([]domain.UserTrade, len(models))
	for i, model := range models {
		trades[i] = model.toDomain()
	}

	return trades, nil
}
