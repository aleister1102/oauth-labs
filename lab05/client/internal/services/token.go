package services

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab05/client/internal/client"
	"github.com/cyllective/oauth-labs/lab05/client/internal/config"
	"github.com/cyllective/oauth-labs/lab05/client/internal/session"
)

type TokenService struct {
	ctx       context.Context
	oaclient  *client.OAuthClient
	jkuClient *http.Client
}

func NewTokenService() *TokenService {
	ctx := context.Background()
	oac := client.NewOAuthClient()
	t := http.DefaultTransport.(*http.Transport)
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c := &http.Client{
		Timeout:   time.Duration(3) * time.Second,
		Transport: t,
	}
	return &TokenService{ctx, oac, c}
}

func (t TokenService) KeyProviderfunc(ctx context.Context, sink jws.KeySink, sig *jws.Signature, msg *jws.Message) error {
	headers := sig.ProtectedHeaders()
	set, err := jwk.Fetch(ctx, headers.JWKSetURL(), jwk.WithHTTPClient(t.jkuClient))
	if err != nil {
		return errors.New("failed to fetch jku")
	}
	alg := headers.Algorithm()
	key, ok := set.LookupKeyID(headers.KeyID())
	if !ok {
		return errors.New("invalid signature")
	}
	sink.Key(alg, key)
	return nil
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
		jwt.WithRequiredClaim("scope"),
		jwt.WithRequiredClaim("sub"),
		jwt.WithKeyProvider(jws.KeyProviderFunc(t.KeyProviderfunc)),
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
