package dto

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

type CreateAccessToken struct {
	ID           string
	UserID       string
	ClientID     string
	Scope        string
	EncryptedJWT string
	Expiration   time.Duration
}

type DeleteAccessTokens struct {
	UserID   string
	ClientID string
}

type RevokeAccessTokens struct {
	UserID   string
	ClientID string
}

type AccessToken struct {
	Token       jwt.Token
	ID          string
	UserID      string
	ClientID    string
	SignedToken string
	ExpiresIn   int
}
