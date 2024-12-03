package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/cyllective/oauth-labs/oalib"
	"github.com/cyllective/oauth-labs/oalib/metadata"
	"github.com/cyllective/oauth-labs/oalib/scope"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/cyllective/oauth-labs/lab03/server/internal/config"
	"github.com/cyllective/oauth-labs/lab03/server/internal/dto"
)

type TokenService struct {
	db   *sql.DB
	meta *metadata.ServerMetadata
	keys *JWKService
	cryp *CryptoService
	cons *ConsentService
	rts  *RefreshTokenService
	ats  *AccessTokenService
}

func NewTokenService(db *sql.DB, meta *metadata.ServerMetadata, keys *JWKService, cryp *CryptoService, cons *ConsentService, ats *AccessTokenService, rts *RefreshTokenService) *TokenService {
	return &TokenService{db, meta, keys, cryp, cons, rts, ats}
}

func (t *TokenService) RevokeAll(ctx context.Context, clientID string, userID string) error {
	err := t.ats.RevokeAll(ctx, &dto.RevokeAccessTokens{ClientID: clientID, UserID: userID})
	if err != nil {
		return err
	}
	err = t.rts.RevokeAll(ctx, &dto.RevokeRefreshTokens{ClientID: clientID, UserID: userID})
	if err != nil {
		return err
	}
	return nil
}

type CreatedTokens struct {
	AccessToken  string
	RefreshToken string
	Scope        string
	ExpiresIn    int
}

func (t *TokenService) CreateTokens(ctx context.Context, request *dto.CreateTokens) *CreatedTokens {
	at, err := t.ats.Create(ctx, &dto.CreateAccessToken{
		UserID:     request.UserID,
		ClientID:   request.ClientID,
		Scope:      request.Scope,
		Expiration: request.Expiration,
	})
	if err != nil {
		panic(err)
	}
	rt, err := t.rts.Create(ctx, &dto.CreateRefreshToken{
		UserID:   request.UserID,
		ClientID: request.ClientID,
		Scope:    request.Scope,
	})
	if err != nil {
		panic(err)
	}
	return &CreatedTokens{
		AccessToken:  at.SignedToken,
		RefreshToken: rt.SignedToken,
		ExpiresIn:    at.ExpiresIn,
		Scope:        request.Scope,
	}
}

func (t *TokenService) ParseAccessToken(ctx context.Context, signedToken string) (*dto.AccessToken, error) {
	return t.ats.Get(ctx, signedToken)
}

func (t *TokenService) GetFromRequest(c *gin.Context) (*dto.AccessToken, error) {
	signedToken := c.GetHeader("Authorization")
	signedToken = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(signedToken, "Bearer"), "bearer"))
	decoded, err := t.ats.Get(c.Request.Context(), signedToken)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func (t *TokenService) GetUnverifiedFromRequest(c *gin.Context) jwt.Token {
	tok, err := jwt.ParseHeader(c.Request.Header, "Authorization", jwt.WithValidate(false), jwt.WithVerify(false))
	if err != nil {
		panic(err)
	}
	return tok
}

func (t *TokenService) HasRequiredScopes(accessToken jwt.Token, required *scope.Scope) bool {
	return t.ats.HasRequiredScopes(accessToken, required)
}

// RefreshTokens refreshes an access token.
func (t *TokenService) RefreshTokens(ctx context.Context, request *dto.RefreshTokens) (*CreatedTokens, error) {
	rt, err := t.rts.Get(ctx, request.RefreshToken)
	if err != nil {
		return nil, err
	}
	err = t.ats.DeleteAll(ctx, &dto.DeleteAccessTokens{
		UserID:   rt.UserID,
		ClientID: rt.ClientID,
	})
	if err != nil {
		log.Printf("warning: failed to delete all `access_token`s for user=%s and client=%s\n", rt.UserID, rt.ClientID)
	}

	if !t.cons.HasConsent(ctx, &dto.Consent{ClientID: rt.ClientID, UserID: rt.UserID}) {
		return nil, errors.New("consent revoked")
	}

	at, err := t.ats.Create(ctx, &dto.CreateAccessToken{
		UserID:     rt.UserID,
		ClientID:   rt.ClientID,
		Scope:      rt.Scope,
		Expiration: config.GetAccessTokenExpiration(),
	})
	if err != nil {
		return nil, err
	}
	return &CreatedTokens{
		AccessToken:  at.SignedToken,
		RefreshToken: request.RefreshToken,
		Scope:        rt.Scope,
		ExpiresIn:    at.ExpiresIn,
	}, nil
}

// Revoke services token revocations.
func (t *TokenService) Revoke(ctx context.Context, request *dto.RevokeTokens) *oalib.VerboseError {
	if request.Token == "" {
		// Invalid (or missing) tokens are silently ignored.
		return nil
	}

	// Handle refresh_token revocation if the request specified a refresh_token.
	rt, err := t.rts.Get(ctx, request.Token)
	if err != nil && errors.Is(err, ErrRevokedRefreshToken) {
		return nil
	}
	if err == nil && rt.ClientID != request.ClientID {
		return &oalib.VerboseError{
			Err:         "invalid_client",
			Description: "client asked to revoke a refresh_token not issued to them",
		}
	}
	if err == nil && rt.ClientID == request.ClientID {
		req := &dto.RevokeRefreshTokens{ClientID: rt.ClientID, UserID: rt.UserID}
		if err = t.rts.RevokeAll(ctx, req); err != nil {
			log.Printf("warning: failed to revoke refresh_tokens for user=%s by client=%s: %s", rt.UserID, rt.ClientID, err.Error())
		}
		return nil
	}

	// Handle access_token revocation if the request specified an access_token.
	at, err := t.ats.Get(ctx, request.Token)
	if err != nil && errors.Is(err, ErrRevokedAccessToken) {
		return nil
	}
	if err == nil && at.ClientID != request.ClientID {
		return &oalib.VerboseError{
			Err:         "invalid_client",
			Description: "client asked to revoke an access_token not issued to them",
		}
	}
	if err == nil && at.ClientID == request.ClientID {
		req := &dto.RevokeAccessTokens{ClientID: at.ClientID, UserID: at.UserID}
		if err := t.ats.RevokeAll(ctx, req); err != nil {
			log.Printf("warning: failed to revoke access_tokens for user=%s by client=%s: %s", at.UserID, at.ClientID, err.Error())
		}
		return nil
	}

	return nil
}
