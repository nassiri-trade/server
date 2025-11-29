package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"trading_server/internal/domain"
)

type GormPasskeyRepository struct {
	db *gorm.DB
}

func NewGormPasskeyRepository(db *gorm.DB) (*GormPasskeyRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &GormPasskeyRepository{db: db}, nil
}

func (r *GormPasskeyRepository) AddPasskey(ctx context.Context, passkey domain.Passkey) error {
	model := toPasskeyModel(passkey)
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormPasskeyRepository) UpdatePasskeyStatus(ctx context.Context, passkeyID string, enabled bool) error {
	result := r.db.WithContext(ctx).
		Model(&PasskeyModel{}).
		Where("passkey_id = ?", passkeyID).
		Update("enabled", enabled)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("passkey not found")
	}

	return nil
}

func (r *GormPasskeyRepository) PasskeyExists(ctx context.Context, passkeyID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&PasskeyModel{}).
		Where("passkey_id = ?", passkeyID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

