package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"trading_server/internal/domain"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) (*GormUserRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &GormUserRepository{db: db}, nil
}

func (r *GormUserRepository) UpsertUser(ctx context.Context, user domain.User) error {
	model := toUserModel(user)

	assignments := clause.Assignments(map[string]interface{}{
		"platform":   gorm.Expr("COALESCE(EXCLUDED.platform, users.platform)"),
		"name":       gorm.Expr("COALESCE(EXCLUDED.name, users.name)"),
		"login":      gorm.Expr("COALESCE(EXCLUDED.login, users.login)"),
		"server":     gorm.Expr("COALESCE(EXCLUDED.server, users.server)"),
		"broker":     gorm.Expr("COALESCE(EXCLUDED.broker, users.broker)"),
		"metadata":   gorm.Expr("COALESCE(EXCLUDED.metadata, users.metadata)"),
		"last_seen":  gorm.Expr("GREATEST(EXCLUDED.last_seen, users.last_seen)"),
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	})

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: assignments,
		}).
		Create(&model).Error
}

func (r *GormUserRepository) GetUser(ctx context.Context, userID string) (domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, err
		}
		return domain.User{}, err
	}

	return model.toDomain(), nil
}

func (r *GormUserRepository) GetUserByLogin(ctx context.Context, login string) (domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).
		Where("login = ?", login).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, err
		}
		return domain.User{}, err
	}

	return model.toDomain(), nil
}

func (r *GormUserRepository) ListUsers(ctx context.Context, limit int) ([]domain.User, error) {
	var models []UserModel

	query := r.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	users := make([]domain.User, len(models))
	for i, model := range models {
		users[i] = model.toDomain()
	}

	return users, nil
}
