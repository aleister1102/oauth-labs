package dto

import (
	"time"
)

type CreateAuthorizationCode struct {
	ClientID            string
	UserID              string
	RedirectURI         string
	Scope               string
	CodeChallengeMethod string
	CodeChallenge       string
	Expiration          time.Duration
}

type GetAuthorizationCode struct {
	Code string
}
