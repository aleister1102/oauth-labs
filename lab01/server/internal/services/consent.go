package services

import (
	"context"
	"errors"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab01/server/internal/repositories"
)

type ConsentService struct {
	consentRepository *repositories.ConsentRepository
}

func NewConsentService(consentRepository *repositories.ConsentRepository) *ConsentService {
	return &ConsentService{consentRepository}
}

func (cs *ConsentService) Create(ctx context.Context, consent *dto.Consent) error {
	if consent.ClientID == "" || consent.UserID == "" {
		return errors.New("invalid user_id or client_id")
	}
	return cs.consentRepository.Create(ctx, consent)
}

func (cs *ConsentService) Revoke(ctx context.Context, consent *dto.Consent) error {
	if consent.ClientID == "" || consent.UserID == "" {
		return errors.New("invalid user_id or client_id")
	}
	return cs.consentRepository.Delete(ctx, consent)
}

func (cs *ConsentService) Get(ctx context.Context, consent *dto.Consent) (*dto.Consent, error) {
	if consent.ClientID == "" || consent.UserID == "" {
		return nil, errors.New("invalid user_id or client_id")
	}

	return cs.consentRepository.Get(ctx, consent)
}

func (cs *ConsentService) GetAll(ctx context.Context, userID string) (*dto.UserConsents, error) {
	return cs.consentRepository.GetAll(ctx, userID)
}

func (cs *ConsentService) HasConsent(ctx context.Context, consent *dto.Consent) bool {
	return cs.consentRepository.Exists(ctx, consent)
}
