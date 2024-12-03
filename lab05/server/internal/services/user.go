package services

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/cyllective/oauth-labs/lab05/server/internal/config"
	"github.com/cyllective/oauth-labs/lab05/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab05/server/internal/repositories"
)

type UserService struct {
	userRepository *repositories.UserRepository
}

func NewUserService(repository *repositories.UserRepository) *UserService {
	return &UserService{repository}
}

func (u UserService) Get(ctx context.Context, id string) (*dto.User, error) {
	return u.userRepository.Get(ctx, id)
}

func (u UserService) GetByUsername(ctx context.Context, username string) (*dto.User, error) {
	return u.userRepository.GetByUsername(ctx, username)
}

func (u UserService) Register(ctx context.Context, request *dto.RegisterRequest) error {
	if u.Exists(ctx, request.Username) {
		return errors.New("Username taken, choose a different one")
	}
	cfg := config.Get()
	bcryptCost := cfg.GetInt("server.bcrypt_cost")
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcryptCost)
	if err != nil {
		panic(fmt.Errorf("failed to generate password hash: %w", err))
	}
	hash := string(hashBytes)
	return u.userRepository.Create(ctx, request.Username, hash)
}

func (u UserService) Exists(ctx context.Context, username string) bool {
	return u.userRepository.ExistsUsername(ctx, username)
}
