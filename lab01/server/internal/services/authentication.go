package services

import (
	"context"
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab01/server/internal/session"
)

type AuthenticationService struct {
	userService *UserService
}

func NewAuthenticationService(userService *UserService) *AuthenticationService {
	return &AuthenticationService{userService}
}

func (a *AuthenticationService) Login(c *gin.Context, request *dto.LoginRequest) error {
	user, err := a.userService.GetByUsername(c.Request.Context(), request.Username)
	if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		return err
	}

	s := sessions.Default(c)
	s.Clear()
	s.Set("user_id", user.ID)
	return s.Save()
}

func (a *AuthenticationService) Register(ctx context.Context, request *dto.RegisterRequest) error {
	return a.userService.Register(ctx, request)
}

func (a *AuthenticationService) GetUserFromSession(c *gin.Context) (*dto.User, error) {
	s := sessions.Default(c)
	if !session.IsAuthenticated(s) {
		return nil, errors.New("user is not authenticated")
	}
	uid, _ := session.GetString(s, "user_id")
	return a.userService.Get(c.Request.Context(), uid)
}
