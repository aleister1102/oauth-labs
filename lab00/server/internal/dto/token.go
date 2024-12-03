package dto

import "time"

type CreateTokens struct {
	UserID     string
	ClientID   string
	Scope      string
	Expiration time.Duration
}

type RefreshTokens struct {
	RefreshToken string
	ClientID     string
}

type RevokeTokens struct {
	Token    string
	ClientID string
}
