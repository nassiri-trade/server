package usecase

import (
	"context"
	"errors"
	"fmt"

	"trading_server/internal/domain"
)

type PasskeyService struct {
	passkeyRepo domain.PasskeyRepository
}

func NewPasskeyService(passkeyRepo domain.PasskeyRepository) (*PasskeyService, error) {
	if passkeyRepo == nil {
		return nil, errors.New("passkey repository required")
	}
	return &PasskeyService{
		passkeyRepo: passkeyRepo,
	}, nil
}

func (s *PasskeyService) AddPasskey(ctx context.Context, passkeyID string) error {
	if passkeyID == "" {
		return errors.New("passkey id required")
	}

	// Check if passkey already exists
	exists, err := s.passkeyRepo.PasskeyExists(ctx, passkeyID)
	if err != nil {
		return fmt.Errorf("failed to check passkey existence: %w", err)
	}
	if exists {
		return errors.New("passkey already exists")
	}

	passkey := domain.Passkey{
		PasskeyID: passkeyID,
		Enabled:   true,
	}

	return s.passkeyRepo.AddPasskey(ctx, passkey)
}

func (s *PasskeyService) UpdatePasskeyStatus(ctx context.Context, passkeyID string, enabled bool) error {
	if passkeyID == "" {
		return errors.New("passkey id required")
	}

	return s.passkeyRepo.UpdatePasskeyStatus(ctx, passkeyID, enabled)
}

func (s *PasskeyService) PasskeyExists(ctx context.Context, passkeyID string) (bool, error) {
	if passkeyID == "" {
		return false, errors.New("passkey id required")
	}

	return s.passkeyRepo.PasskeyExists(ctx, passkeyID)
}
