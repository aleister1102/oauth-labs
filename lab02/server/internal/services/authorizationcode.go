package services

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cyllective/oauth-labs/oalib"
	"github.com/redis/go-redis/v9"

	"github.com/cyllective/oauth-labs/lab02/server/internal/constants"
	"github.com/cyllective/oauth-labs/lab02/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab02/server/internal/utils"
)

var authorizationCodePrefix = fmt.Sprintf("server%s:authorization_code:", constants.LabNumber)

type AuthorizationCodeService struct {
	rdb  *redis.Client
	keys *JWKService
}

func NewAuthorizationCodeService(rdb *redis.Client, keys *JWKService) *AuthorizationCodeService {
	return &AuthorizationCodeService{rdb, keys}
}

type rawAuthorizationCode struct {
	ClientID            string `redis:"client_id" json:"client_id"`
	UserID              string `redis:"user_id" json:"user_id"`
	RedirectURI         string `redis:"redirect_uri" json:"redirect_uri"`
	Scope               string `redis:"scope" json:"scope"`
	CodeChallenge       string `redis:"code_challenge" json:"code_challenge"`
	CodeChallengeMethod string `redis:"code_challenge_method" json:"code_challenge_method"`
	CreatedAt           int    `redis:"created_at" json:"created_at"`
}

func (s *AuthorizationCodeService) Create(ctx context.Context, request *dto.CreateAuthorizationCode) (*oalib.AuthorizationCode, error) {
	codeBytes := utils.RandomBytes(32)
	key := authorizationCodePrefix + hex.EncodeToString(codeBytes)
	code := base64.RawURLEncoding.EncodeToString(codeBytes)

	createdAt := int(time.Now().Unix())
	rawCode := &rawAuthorizationCode{
		ClientID:    request.ClientID,
		UserID:      request.UserID,
		RedirectURI: request.RedirectURI,
		Scope:       request.Scope,
		CreatedAt:   createdAt,
	}
	_, err := s.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key, rawCode)
		pipe.Expire(ctx, key, request.Expiration)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create code: redis pipeline failed: %w", err)
	}

	return &oalib.AuthorizationCode{
		Code:                code,
		ClientID:            request.ClientID,
		UserID:              request.UserID,
		RedirectURI:         request.RedirectURI,
		Scope:               request.Scope,
		CreatedAt:           createdAt,
		CodeChallenge:       request.CodeChallenge,
		CodeChallengeMethod: request.CodeChallengeMethod,
	}, nil
}

func (s *AuthorizationCodeService) Get(ctx context.Context, code string) (*oalib.AuthorizationCode, error) {
	codeBytes, err := base64.RawURLEncoding.DecodeString(code)
	if err != nil {
		return nil, err
	}
	key := authorizationCodePrefix + hex.EncodeToString(codeBytes)

	var rawCode rawAuthorizationCode
	if err := s.rdb.HGetAll(ctx, key).Scan(&rawCode); err != nil {
		return nil, err
	}
	return &oalib.AuthorizationCode{
		Code:                hex.EncodeToString(codeBytes),
		ClientID:            rawCode.ClientID,
		UserID:              rawCode.UserID,
		RedirectURI:         rawCode.RedirectURI,
		Scope:               rawCode.Scope,
		CreatedAt:           rawCode.CreatedAt,
		CodeChallenge:       rawCode.CodeChallenge,
		CodeChallengeMethod: rawCode.CodeChallengeMethod,
	}, nil
}

func (s *AuthorizationCodeService) Delete(ctx context.Context, code string) error {
	key := authorizationCodePrefix + code
	return s.rdb.Del(ctx, key).Err()
}
