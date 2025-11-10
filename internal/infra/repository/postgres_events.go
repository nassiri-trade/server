package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"trading_server/internal/domain"
)

type GormEventRepository struct {
	db *gorm.DB
}

func NewGormEventRepository(db *gorm.DB) (*GormEventRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}

	return &GormEventRepository{db: db}, nil
}

func (r *GormEventRepository) UpsertEvents(ctx context.Context, events []domain.CalendarEvent) error {
	if len(events) == 0 {
		return nil
	}

	models := make([]CalendarEventModel, len(events))
	for i, ev := range events {
		models[i] = toCalendarEventModel(ev)
	}

	assignments := clause.Assignments(map[string]interface{}{
		"event_date": gorm.Expr("EXCLUDED.event_date"),
		"currency":   gorm.Expr("EXCLUDED.currency"),
		"impact":     gorm.Expr("EXCLUDED.impact"),
		"detail":     gorm.Expr("EXCLUDED.detail"),
		"actual":     gorm.Expr("EXCLUDED.actual"),
		"forecast":   gorm.Expr("EXCLUDED.forecast"),
		"previous":   gorm.Expr("EXCLUDED.previous"),
		"source_url": gorm.Expr("EXCLUDED.source_url"),
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	})

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "event_hash"}},
			DoUpdates: assignments,
		}).
		Create(&models).Error
}

func (r *GormEventRepository) ListEvents(ctx context.Context, limit int) ([]domain.CalendarEvent, error) {
	var models []CalendarEventModel
	query := r.db.WithContext(ctx).Order("event_date DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	events := make([]domain.CalendarEvent, len(models))
	for i, model := range models {
		events[i] = model.toDomain()
	}

	return events, nil
}
