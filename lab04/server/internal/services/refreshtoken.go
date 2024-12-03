package services

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jws"

	"github.com/cyllective/oauth-labs/lab04/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab04/server/internal/entities"
	"github.com/cyllective/oauth-labs/lab04/server/internal/repositories"
)

type RefreshTokenService struct {
	refreshTokenRepository *repositories.RefreshTokenRepository
	cryp                   *CryptoService
	keys                   *JWKService
}

func NewRefreshTokenService(refreshTokenRepository *repositories.RefreshTokenRepository, cryp *CryptoService, keys *JWKService) *RefreshTokenService {
	return &RefreshTokenService{refreshTokenRepository, cryp, keys}
}

func (r *RefreshTokenService) Create(ctx context.Context, request *dto.CreateRefreshToken) (*dto.RefreshToken, error) {
	rid := uuid.NewString()
	payload := &dto.RawRefreshToken{
		ID:        rid,
		UserID:    request.UserID,
		ClientID:  request.ClientID,
		Scope:     request.Scope,
		CreatedAt: int(time.Now().UTC().Unix()),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh_token: json.marshal failed: %w", err)
	}
	headers := jws.NewHeaders()
	_ = headers.Set("typ", "rt+jwt")
	signedPayload, err := jws.Sign(jsonPayload, jws.WithKey(jwa.RS256, r.keys.PrivateKey(), jws.WithProtectedHeaders(headers)))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh_token: jws.sign failed: %w", err)
	}
	encPayload, err := jwe.Encrypt(signedPayload, jwe.WithKey(jwa.RSA_OAEP, r.keys.PublicKey()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh_token: jwe.encrypt failed: %w", err)
	}
	encPayloadHex := hex.EncodeToString(encPayload)
	err = r.refreshTokenRepository.Create(ctx, &entities.RefreshToken{
		ID:       rid,
		ClientID: request.ClientID,
		UserID:   request.UserID,
		Data:     encPayloadHex,
		Revoked:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh_token: database operation failed: %w", err)
	}
	signedToken := base64.RawURLEncoding.EncodeToString(encPayload)
	return &dto.RefreshToken{
		SignedToken: signedToken,
		ID:          rid,
		UserID:      request.UserID,
		ClientID:    request.ClientID,
		Scope:       request.Scope,
		CreatedAt:   payload.CreatedAt,
	}, nil
}

var ErrRevokedRefreshToken = errors.New("refresh_token is revoked")

func (r *RefreshTokenService) Get(ctx context.Context, signedToken string) (*dto.RefreshToken, error) {
	encPayload, err := base64.RawURLEncoding.DecodeString(signedToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh_token: base64url decoding failed: %w", err)
	}
	signedPayload, err := jwe.Decrypt(encPayload, jwe.WithKey(jwa.RSA_OAEP, r.keys.PrivateKey()))
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh_token: jwe.decrypt failed: %w", err)
	}
	headers := jws.NewHeaders()
	_ = headers.Set("typ", "rt+jwt")
	jsonPayload, err := jws.Verify(signedPayload, jws.WithKey(jwa.RS256, r.keys.PublicKey(), jws.WithProtectedHeaders(headers)))
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh_token: jws.verify failed: %w", err)
	}
	var payload dto.RawRefreshToken
	if err := json.Unmarshal(jsonPayload, &payload); err != nil {
		return nil, fmt.Errorf("failed to get refresh_token: json.unmarshal failed: %w", err)
	}
	dbToken, err := r.refreshTokenRepository.Get(ctx, payload.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh_token: database query failed: %w", err)
	}
	if dbToken.Revoked {
		return nil, ErrRevokedRefreshToken
	}
	return &dto.RefreshToken{
		SignedToken: signedToken,
		ID:          payload.ID,
		UserID:      payload.UserID,
		ClientID:    payload.ClientID,
		Scope:       payload.Scope,
	}, nil
}

func (r *RefreshTokenService) RevokeAll(ctx context.Context, request *dto.RevokeRefreshTokens) error {
	return r.refreshTokenRepository.RevokeAll(ctx, request)
}
