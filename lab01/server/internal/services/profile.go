package services

import (
	"context"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab01/server/internal/repositories"
)

type ProfileService struct {
	userRepository *repositories.UserRepository
}

func NewProfileService(userRepository *repositories.UserRepository) *ProfileService {
	return &ProfileService{userRepository}
}

func (p *ProfileService) Update(ctx context.Context, request *dto.UpdateProfile) error {
	return p.userRepository.Update(ctx, request)
}
