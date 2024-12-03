package services

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/cyllective/oauth-labs/oalib/metadata"
	"github.com/cyllective/oauth-labs/oalib/scope"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/cyllective/oauth-labs/lab00/server/internal/config"
	"github.com/cyllective/oauth-labs/lab00/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab00/server/internal/repositories"
)

type AccessTokenService struct {
	accessTokenRepository *repositories.AccessTokenRepository
	clientRepository      *repositories.ClientRepository
	userRepository        *repositories.UserRepository
	meta                  *metadata.ServerMetadata
	cryp                  *CryptoService
	keys                  *JWKService
}

func NewAccessTokenService(accessTokenRepository *repositories.AccessTokenRepository, clientRepository *repositories.ClientRepository, userRepository *repositories.UserRepository, meta *metadata.ServerMetadata, cryp *CryptoService, keys *JWKService) *AccessTokenService {
	return &AccessTokenService{accessTokenRepository, clientRepository, userRepository, meta, cryp, keys}
}

func (a *AccessTokenService) Create(ctx context.Context, request *dto.CreateAccessToken) (*dto.AccessToken, error) {
	id := uuid.NewString()
	iat := time.Now().UTC()
	exp := iat.Add(request.Expiration)
	tok, err := jwt.NewBuilder().
		JwtID(id).
		Issuer(a.meta.Issuer).
		Subject(request.UserID).
		Audience([]string{request.ClientID}).
		IssuedAt(iat).
		NotBefore(iat).
		Expiration(exp).
		Claim("scope", request.Scope).
		Claim("client_id", request.ClientID).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create access_token: jwt builder failed: %w", err)
	}

	headers := jws.NewHeaders()
	_ = headers.Set("typ", "at+jwt")
	privateKey := a.keys.PrivateKey()
	signedJWT, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, privateKey, jws.WithProtectedHeaders(headers)))
	if err != nil {
		return nil, fmt.Errorf("failed to create access_token: signing failed: %w", err)
	}
	key := config.GetJWTEncryptionKey()
	encJWT := a.cryp.Encrypt(signedJWT, key)
	encJWTHex := hex.EncodeToString(encJWT)
	err = a.accessTokenRepository.Create(ctx, &dto.CreateAccessToken{
		ID:           id,
		UserID:       request.UserID,
		ClientID:     request.ClientID,
		EncryptedJWT: encJWTHex,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create access_token: database operation failed: %w", err)
	}
	return &dto.AccessToken{
		SignedToken: string(signedJWT),
		Token:       tok,
		ExpiresIn:   int(exp.Sub(iat).Seconds()),
	}, nil
}

var ErrRevokedAccessToken = errors.New("access_token is revoked")

func (a *AccessTokenService) validator(ctx context.Context) jwt.ValidatorFunc {
	return func(_ context.Context, t jwt.Token) jwt.ValidationError {
		if _, err := a.GetJti(ctx, t); err != nil {
			return jwt.NewValidationError(err)
		}
		if _, err := a.GetSub(ctx, t); err != nil {
			return jwt.NewValidationError(err)
		}
		if _, err := a.GetScope(t); err != nil {
			return jwt.NewValidationError(err)
		}
		aud, err := a.GetAud(ctx, t)
		if err != nil {
			return jwt.NewValidationError(err)
		}
		cid, err := a.GetClientID(ctx, t)
		if err != nil {
			return jwt.NewValidationError(err)
		}
		if !slices.Contains(aud, cid) {
			return jwt.NewValidationError(errors.New("invalid audience"))
		}
		if a.IsRevoked(ctx, t) {
			return jwt.NewValidationError(ErrRevokedAccessToken)
		}

		return nil
	}
}

func (a *AccessTokenService) Get(ctx context.Context, signedToken string) (*dto.AccessToken, error) {
	token, err := jwt.Parse(
		[]byte(signedToken),
		jwt.WithKeySet(a.keys.Keys()),
		jwt.WithIssuer(a.meta.Issuer),
		jwt.WithRequiredClaim("jti"),
		jwt.WithRequiredClaim("aud"),
		jwt.WithRequiredClaim("sub"),
		jwt.WithRequiredClaim("scope"),
		jwt.WithRequiredClaim("client_id"),
		jwt.WithValidator(a.validator(ctx)),
	)
	if err != nil {
		return nil, err
	}
	return &dto.AccessToken{
		ID:          token.JwtID(),
		SignedToken: signedToken,
		Token:       token,
		UserID:      token.Subject(),
		ClientID:    token.PrivateClaims()["client_id"].(string),
	}, nil
}

func (a *AccessTokenService) Delete(ctx context.Context, id string) error {
	return a.accessTokenRepository.Delete(ctx, id)
}

func (a *AccessTokenService) DeleteAll(ctx context.Context, request *dto.DeleteAccessTokens) error {
	return a.accessTokenRepository.DeleteAll(ctx, request)
}

// Extract client_id from a jwt.Token; errors are thrown if the token did not contain a client_id string
func (a *AccessTokenService) GetClientID(ctx context.Context, token jwt.Token) (string, error) {
	cid, ok := token.Get("client_id")
	if !ok {
		return "", errors.New("client_id not found in token")
	}
	cids, ok := cid.(string)
	if !ok {
		return "", errors.New("failed to cast client_id to string")
	}
	if !a.clientRepository.Exists(ctx, cids) {
		return "", errors.New("client not found in database")
	}
	return cids, nil
}

// Extract audience from a jwt.Token; errors are thrown if the token did not contain a known audience
func (a *AccessTokenService) GetAud(ctx context.Context, token jwt.Token) ([]string, error) {
	aud := token.Audience()
	if len(aud) != 1 {
		return nil, errors.New("invalid audience")
	}
	if !a.clientRepository.Exists(ctx, aud[0]) {
		return nil, errors.New("client not found in database")
	}
	return aud, nil
}

// Extract sub from a jwt.Token; errors are thrown if the token did not contain a sub
func (a *AccessTokenService) GetSub(ctx context.Context, token jwt.Token) (string, error) {
	sub := token.Subject()
	if sub == "" {
		return "", errors.New("empty subject")
	}
	if !a.userRepository.Exists(ctx, sub) {
		return "", errors.New("subject (user) not found in database")
	}
	return sub, nil
}

// Extract jti from a jwt.Token; errors are thrown if the token did not contain a jti
func (a *AccessTokenService) GetJti(ctx context.Context, token jwt.Token) (string, error) {
	jti := token.JwtID()
	if jti == "" {
		return "", errors.New("empty jti")
	}
	if !a.accessTokenRepository.Exists(ctx, jti) {
		return "", errors.New("jti not found in database")
	}
	return jti, nil
}

// Extract scope from a jwt.Token; errors are thrown if the token did not contain a scope
func (a *AccessTokenService) GetScope(token jwt.Token) (*scope.Scope, error) {
	maybeScopeString, ok := token.Get("scope")
	if !ok {
		return nil, errors.New("empty scope")
	}
	scopeString, ok := maybeScopeString.(string)
	if !ok {
		return nil, errors.New("failed to cast scope to scope.Scope")
	}
	s := scope.New(scopeString)
	if !a.meta.ScopesSupported.IsSuperset(s.Set()) {
		return nil, errors.New("scope contains one or more invalid values")
	}
	return scope.New(scopeString), nil
}

func (a *AccessTokenService) HasRequiredScopes(token jwt.Token, required *scope.Scope) bool {
	tokenScopes, err := a.GetScope(token)
	if err != nil {
		return false
	}
	return required.Set().IsSubset(tokenScopes.Set())
}

func (a *AccessTokenService) IsRevoked(ctx context.Context, token jwt.Token) bool {
	return a.accessTokenRepository.IsRevoked(ctx, token.JwtID())
}

func (a *AccessTokenService) RevokeAll(ctx context.Context, request *dto.RevokeAccessTokens) error {
	return a.accessTokenRepository.RevokeAll(ctx, request)
}
