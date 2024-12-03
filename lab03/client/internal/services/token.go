package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab03/client/internal/client"
	"github.com/cyllective/oauth-labs/lab03/client/internal/config"
	"github.com/cyllective/oauth-labs/lab03/client/internal/session"
)

type TokenService struct {
	ctx      context.Context
	pubkeys  jwk.Set
	oaclient *client.OAuthClient
}

func NewTokenService() *TokenService {
	cfg := config.Get()
	ctx := context.Background()
	jwkURL := cfg.GetString("authorization_server.jwk_uri")
	cache := buildJWKCache(ctx, jwkURL)
	oac := client.NewOAuthClient()
	return &TokenService{ctx, cache, oac}
}

func buildJWKCache(ctx context.Context, jwkURL string) jwk.Set {
	c := jwk.NewCache(ctx)
	jwkWhitelist := jwk.NewMapWhitelist().Add(jwkURL)
	err := c.Register(jwkURL, jwk.WithMinRefreshInterval(time.Duration(4)*time.Hour), jwk.WithFetchWhitelist(jwkWhitelist))
	if err != nil {
		panic(fmt.Errorf("failed to build jwk cache: %w", err))
	}
	return jwk.NewCachedSet(c, jwkURL)
}

func (t *TokenService) Parse(accessToken string) (jwt.Token, error) {
	cfg := config.Get()
	iss := cfg.GetString("authorization_server.issuer")
	clientID := cfg.GetString("client.id")
	token, err := jwt.Parse(
		[]byte(accessToken),
		jwt.WithIssuer(iss),
		jwt.WithAudience(clientID),
		jwt.WithClaimValue("client_id", clientID),
		jwt.WithKeySet(t.pubkeys, jws.WithInferAlgorithmFromKey(true)),
		jwt.WithRequiredClaim("scope"),
		jwt.WithRequiredClaim("sub"),
	)
	if err != nil {
		if strings.Contains(err.Error(), `"exp" not satisfied`) {
			return nil, ErrAccessTokenExpired
		}
		return nil, err
	}
	return token, nil
}

type Tokens struct {
	AccessTokenJWT jwt.Token
	AccessToken    string
	RefreshToken   string
}

var ErrAccessTokenExpired = errors.New("access_token expired")

func (t *TokenService) Get(s sessions.Session) (*oauth2.Token, error) {
	refreshToken, ok := session.GetString(s, "refresh_token")
	if !ok {
		return nil, errors.New("refresh_token not found")
	}
	accessToken, ok := session.GetString(s, "access_token")
	if !ok {
		return nil, errors.New("access_token not found")
	}
	accessTokenJWT, err := t.Parse(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decode access_token: %w", err)
	}
	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		Expiry:       accessTokenJWT.Expiration(),
	}, nil
}

func (t *TokenService) Revoke(ctx context.Context, tokens *oauth2.Token) []error {
	return t.oaclient.RevokeWithContext(ctx, tokens)
}
