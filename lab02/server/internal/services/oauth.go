package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/cyllective/oauth-labs/oalib"
	"github.com/cyllective/oauth-labs/oalib/metadata"
	"github.com/cyllective/oauth-labs/oalib/redirecturi"
	"github.com/cyllective/oauth-labs/oalib/scope"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/cyllective/oauth-labs/lab02/server/internal/config"
	"github.com/cyllective/oauth-labs/lab02/server/internal/dto"
)

type OAuthService struct {
	meta                     *metadata.ServerMetadata
	authenticationService    *AuthenticationService
	clientService            *ClientService
	consentService           *ConsentService
	tokenService             *TokenService
	jwkService               *JWKService
	authorizationCodeService *AuthorizationCodeService
}

func NewOAuthService(
	meta *metadata.ServerMetadata,
	authenticationService *AuthenticationService,
	clientService *ClientService,
	consentService *ConsentService,
	tokenService *TokenService,
	jwkService *JWKService,
	authorizationCodeService *AuthorizationCodeService,
) *OAuthService {
	return &OAuthService{
		meta,
		authenticationService,
		clientService,
		consentService,
		tokenService,
		jwkService,
		authorizationCodeService,
	}
}

func (oa *OAuthService) Metadata() *metadata.ServerMetadata {
	return oa.meta
}

func (oa *OAuthService) JWKs() jwk.Set {
	return oa.jwkService.PublicKeys()
}

type authorizeContext struct {
	User               *dto.User
	Client             *dto.Client
	RequestRedirectURI *redirecturi.RedirectURI
	AZCRedirectURI     string
	Scope              *scope.Scope
	State              string

	CodeChallenge       string
	CodeChallengeMethod string
}

type AuthorizationCodeResponse struct {
	RedirectURI string
}

var (
	ErrAuthenticationRequired = errors.New("authentication_required")
	ErrConsentRequired        = errors.New("consent_required")
	ErrAbortAuthorize         = errors.New("abort")
)

// Authorize services authorization requests.
func (oa *OAuthService) Authorize(c *gin.Context, request *dto.AuthorizeRequest) (*AuthorizationCodeResponse, error) {
	authorizeCtx := &authorizeContext{State: request.State}

	// user validation - needs to be authenticated
	user, err := oa.authenticationService.GetUserFromSession(c)
	if err != nil {
		log.Printf("[OAuthService.Authorize]: failed to load user from session: %s\n", err.Error())
		return nil, ErrAuthenticationRequired
	}
	authorizeCtx.User = user

	// client validation - needs to exist
	if request.ClientID == "" {
		log.Printf("[OAuthService.Authorize]: client_id validation failed: client_id was not provided")
		return nil, ErrAbortAuthorize
	}
	client, err := oa.clientService.Get(c.Request.Context(), request.ClientID)
	if err != nil {
		log.Printf("[OAuthService.Authorize]: failed to retrieve client: %s\n", err.Error())
		return nil, ErrAbortAuthorize
	}
	authorizeCtx.Client = client

	// redirect_uri validation
	{
		redirectURI, err := validateAuthorizeRedirectURI(request.RedirectURI, client.RedirectURIs)
		if err != nil {
			log.Printf("[OAuthService.Authorize]: redirect_uri validation failed: err=%s, desc=%s\n", err.Err, err.Description)
			return nil, ErrAbortAuthorize
		}
		authorizeCtx.RequestRedirectURI = redirectURI
		if request.RedirectURI != "" {
			authorizeCtx.AZCRedirectURI = authorizeCtx.RequestRedirectURI.String()
		}
		if request.State != "" {
			redirectURI.SetState(request.State)
		}
	}

	// consent validation - needs to be granted
	consent := &dto.Consent{ClientID: client.ID, UserID: user.ID}
	if !oa.consentService.HasConsent(c.Request.Context(), consent) {
		log.Printf("[OAuthService.Authorize]: user=%s did not yet consent to client=%s\n", consent.UserID, consent.ClientID)
		return nil, ErrConsentRequired
	}

	//
	// ... and now for the errors that we have to deliver back to the client ...
	//

	// response_type validation
	if !slices.Contains(client.ResponseTypes, request.ResponseType) {
		log.Println("[OAuthService.Authorize]: response_type validation failed: requested response_type is not supported by the client")
		return nil, newAuthorizeError("unauthorized_client", "response_type must be set to \"code\".", *authorizeCtx.RequestRedirectURI)
	}
	if !oa.meta.ResponseTypesSupported.Contains(request.ResponseType) {
		log.Println("[OAuthService.Authorize]: response_type validation failed: requested response_type is not supported")
		return nil, newAuthorizeError("unsupported_response_type", "response_type must be set to \"code\".", *authorizeCtx.RequestRedirectURI)
	}

	// Scope validation
	{
		cScope := scope.NewWith(authorizeCtx.Client.Scope...)
		rScope := scope.New(request.Scope)
		requestedScope, err := validateAuthorizeScope(rScope, cScope)
		if err != nil {
			log.Printf("[OAuthService.Authorize]: scope validation failed: err=%s desc=%s", err.Err, err.Description)
			return nil, newAuthorizeError(err.Err, err.Description, *authorizeCtx.RequestRedirectURI)
		}
		authorizeCtx.Scope = requestedScope
	}

	// Optional PKCE support and validation if present
	if request.CodeChallenge != "" {
		if err := validateAuthorizePKCE(oa.meta, request.CodeChallengeMethod, request.CodeChallenge); err != nil {
			log.Printf("[OAuthService.Authorize]: PKCE validation failed: err=%s, desc=%s\n", err.Err, err.Description)
			return nil, newAuthorizeError(err.Err, err.Description, *authorizeCtx.RequestRedirectURI)
		}

		authorizeCtx.CodeChallengeMethod = request.CodeChallengeMethod
		authorizeCtx.CodeChallenge = request.CodeChallenge
	}

	switch request.ResponseType {
	case "code":
		return oa.handleAuthorizationCodeResponse(c, request, authorizeCtx)
	default:
		panic("unknown or unsupported response_type")
	}
}

func (oa *OAuthService) handleAuthorizationCodeResponse(ctx context.Context, _ *dto.AuthorizeRequest, actx *authorizeContext) (*AuthorizationCodeResponse, error) {
	// Create an authorization code.
	code, err := oa.authorizationCodeService.Create(ctx, &dto.CreateAuthorizationCode{
		ClientID:            actx.Client.ID,
		UserID:              actx.User.ID,
		RedirectURI:         actx.AZCRedirectURI,
		Scope:               actx.Scope.SetString(),
		CodeChallengeMethod: actx.CodeChallengeMethod,
		CodeChallenge:       actx.CodeChallenge,
		Expiration:          time.Duration(10) * time.Minute,
	})
	if err != nil {
		panic(err)
	}

	redirectURI := actx.RequestRedirectURI
	redirectURI.SetState(actx.State)
	redirectURI.SetCode(code.Code)
	return &AuthorizationCodeResponse{RedirectURI: redirectURI.String()}, nil
}

func validateAuthorizeRedirectURI(uri string, allowedURIs []string) (*redirecturi.RedirectURI, *oalib.VerboseError) {
	if uri == "" && len(allowedURIs) > 1 {
		return nil, &oalib.VerboseError{
			Err:         "invalid_redirect_uri",
			Description: "unable to determine default redirect_uri, none was passed and the client has more than one allowed redirect_uri.",
		}
	}
	if uri == "" && len(allowedURIs) == 1 {
		uri = allowedURIs[0]
	}
	redirectURI, err := redirecturi.New(uri)
	if err != nil {
		return nil, &oalib.VerboseError{
			Err:         "invalid_redirect_uri",
			Description: "failed to parse redirect_uri",
		}
	}
	if redirectURI.URL().Fragment != "" {
		return nil, &oalib.VerboseError{
			Err:         "invalid_redirect_uri",
			Description: "redirect_uri contained a fragment, which is not allowed.",
		}
	}

	return redirectURI, nil
}

func validateAuthorizeScope(requestedScope *scope.Scope, clientScope *scope.Scope) (*scope.Scope, *oalib.VerboseError) {
	if requestedScope.SliceLength() != requestedScope.SetLength() {
		return nil, &oalib.VerboseError{
			Err:         "invalid_scope",
			Description: "scope contains one or more duplicate values.",
		}
	}
	if requestedScope.SetLength() > clientScope.SetLength() {
		return nil, &oalib.VerboseError{
			Err:         "invalid_scope",
			Description: "requested scopes exceed scopes known to client.",
		}
	}
	if !clientScope.Set().IsSuperset(requestedScope.Set()) {
		return nil, &oalib.VerboseError{
			Err:         "invalid_scope",
			Description: "requested scope is unknown to client.",
		}
	}

	// Default to the full scope if the requested scope was empty.
	if requestedScope.SetLength() == 0 {
		return clientScope.Copy(), nil
	}
	return requestedScope, nil
}

func validateAuthorizePKCE(meta *metadata.ServerMetadata, method string, challenge string) *oalib.VerboseError {
	if method == "" {
		method = "plain"
	}
	if !meta.SupportsCodeChallengeMethod(method) {
		return &oalib.VerboseError{
			Err:         "invalid_request",
			Description: "unsupported code_challenge_method.",
		}
	}
	if challenge == "" {
		return &oalib.VerboseError{
			Err:         "invalid_request",
			Description: "invalid code_challenge.",
		}
	}
	return nil
}

// Token services token requests.
func (oa *OAuthService) Token(c *gin.Context, request *dto.TokenRequest) (*oalib.TokenResponse, *oalib.VerboseError) {
	ctx := c.Request.Context()
	credentials, err := oa.clientService.ParseCredentials(c)
	if err != nil {
		return nil, err
	}
	client, err := oa.clientService.Authenticate(ctx, credentials)
	if err != nil {
		return nil, err
	}

	// grant_type validation.
	if request.GrantType == "" {
		return nil, &oalib.VerboseError{
			Err:         "invalid_request",
			Description: "missing grant_type.",
		}
	}
	if !slices.Contains(client.GrantTypes, request.GrantType) {
		return nil, &oalib.VerboseError{
			Err:         "unauthorized_client",
			Description: fmt.Sprintf("client has unsupported grant_type: %#v", request.GrantType),
		}
	}
	if !oa.meta.GrantTypesSupported.Contains(request.GrantType) {
		return nil, &oalib.VerboseError{
			Err:         "invalid_grant",
			Description: fmt.Sprintf("server does not support grant_type: %#v", request.GrantType),
		}
	}

	switch request.GrantType {
	case "authorization_code":
		return oa.handleAuthorizationCodeGrant(ctx, request, client)
	case "refresh_token":
		return oa.handleRefreshTokenGrant(ctx, request, client)
	default:
		panic("unknown or unhandled grant_type")
	}
}

func (oa *OAuthService) handleRefreshTokenGrant(ctx context.Context, request *dto.TokenRequest, client *dto.Client) (*oalib.TokenResponse, *oalib.VerboseError) {
	if request.RefreshToken == "" {
		return nil, &oalib.VerboseError{
			Err:         "invalid_request",
			Description: "missing refresh_token.",
		}
	}
	newTokens, err := oa.tokenService.RefreshTokens(ctx, &dto.RefreshTokens{
		RefreshToken: request.RefreshToken,
		ClientID:     client.ID,
	})
	if err != nil {
		return nil, &oalib.VerboseError{
			Err:         "invalid_grant",
			Description: err.Error(),
		}
	}

	return &oalib.TokenResponse{
		AccessToken:  newTokens.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    newTokens.ExpiresIn,
		RefreshToken: newTokens.RefreshToken,
		Scope:        newTokens.Scope,
	}, nil
}

func (oa *OAuthService) handleAuthorizationCodeGrant(ctx context.Context, request *dto.TokenRequest, client *dto.Client) (*oalib.TokenResponse, *oalib.VerboseError) {
	code, err := oa.authorizationCodeService.Get(ctx, request.Code)
	if err != nil {
		return nil, &oalib.VerboseError{
			Err:         "invalid_grant",
			Description: err.Error(),
		}
	}
	// Check if the code was issued for the given client_id.
	if code.ClientID != client.ID {
		return nil, &oalib.VerboseError{
			Err:         "invalid_grant",
			Description: "client requested a code not issued to them.",
		}
	}

	// Check if the user's consent is still valid.
	if !oa.consentService.HasConsent(ctx, &dto.Consent{UserID: code.UserID, ClientID: code.ClientID}) {
		return nil, &oalib.VerboseError{
			Err:         "invalid_grant",
			Description: "consent has been revoked.",
		}
	}

	// Custom scope validation.
	actualScope := code.Scope
	if request.Scope != "" {
		reqScope := scope.New(request.Scope)
		if reqScope.SetLength() != reqScope.SliceLength() {
			return nil, &oalib.VerboseError{
				Err:         "invalid_scope",
				Description: "requested scope included one or more invalid scopes",
			}
		}
		codeScope := scope.New(code.Scope)
		if reqScope.SetLength() > codeScope.SetLength() {
			return nil, &oalib.VerboseError{
				Err:         "invalid_scope",
				Description: "requested scope exceeds code scope",
			}
		}
		if !reqScope.Set().IsSubset(codeScope.Set()) {
			return nil, &oalib.VerboseError{
				Err:         "invalid_scope",
				Description: "requested scope includes invalid or exceeding scopes",
			}
		}

		actualScope = reqScope.SetString()
	}

	// Exchange the authorization code for tokens.
	tokens := oa.tokenService.CreateTokens(ctx, &dto.CreateTokens{
		UserID:     code.UserID,
		ClientID:   code.ClientID,
		Scope:      actualScope,
		Expiration: config.GetAccessTokenExpiration(),
	})
	if err := oa.authorizationCodeService.Delete(ctx, code.Code); err != nil {
		log.Printf("warning: failed to delete authorization code: %s", err.Error())
	}
	return &oalib.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
		Scope:        actualScope,
	}, nil
}

// Register services dynamic client registration.
// We don't authenticate clients as defined in rfc7591 (OAuth 2.0 Dynamic
// Client Registration Protocol). We only implement it as an easy way to
// configure the respective client(s). Thus, instead of having an open
// registration or using pre-defined access tokens for the clients we use a
// pre-defined allowlist of client_ids and a pre-shared secret in a custom
// header.
func (oa *OAuthService) Register(ctx context.Context, request dto.OAuthRegisterClient) (*oalib.ClientInformationResponse, *oalib.VerboseError) {
	cfg := config.Get()
	if request.RegisterSecret != cfg.GetString("oauth.registration_secret") {
		log.Println("register request failed: invalid x-register-key")
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: "invalid x-register-key",
		}
	}
	allowedIDs := cfg.GetStringSlice("oauth.allowed_clients")
	if !slices.Contains(allowedIDs, request.ClientID) {
		log.Println("register request failed: client_id not allowed")
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: "invalid client_id",
		}
	}

	// Ensure we have at least one redirect_uri.
	if len(request.RedirectURIs) == 0 {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: "invalid redirect_uris",
		}
	}

	// Ensure that the grant_types are supported
	gs := mapset.NewSet(request.GrantTypes...)
	if !oa.meta.GrantTypesSupported.IsSuperset(gs) {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: "invalid grant_type",
		}
	}

	// Ensure that the response_types are supported
	rs := mapset.NewSet(request.ResponseTypes...)
	if !oa.meta.ResponseTypesSupported.IsSuperset(rs) {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: "invalid response_type",
		}
	}

	// Ensure that grant_types and corresponding response_types exist in case they are defined.
	// check relationship: grant_type=authorization_code and response_type=code
	if (gs.Contains("authorization_code") && !rs.Contains("code")) || (!gs.Contains("authorization_code") && rs.Contains("code")) {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: "invalid grant_types and response_types combination.",
		}
	}
	// check relationship: grant_type=implicit and response_type=token
	if (gs.Contains("implicit") && !rs.Contains("token")) || (!gs.Contains("implicit") && rs.Contains("token")) {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: "invalid grant_types and response_types combination.",
		}
	}

	// Ensure the token_endpoint_auth_method is supported
	if !oa.meta.TokenEndpointAuthMethodsSupported.Contains(request.TokenEndpointAuthMethod) {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: "invalid token_endpoint_auth_method",
		}
	}

	// Make sure that the client included scopes that we know about and support.
	scopes := scope.New(request.Scope).Set()
	if !oa.meta.ScopesSupported.IsSuperset(scopes) {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: "invalid scopes",
		}
	}

	err := oa.clientService.Register(ctx, &dto.ClientRegister{
		ID:                      request.ClientID,
		Secret:                  request.ClientSecret,
		Name:                    request.ClientName,
		URL:                     request.ClientURI,
		LogoURI:                 request.LogoURI,
		Scope:                   scopes.ToSlice(),
		RedirectURIs:            request.RedirectURIs,
		TokenEndpointAuthMethod: request.TokenEndpointAuthMethod,
		GrantTypes:              request.GrantTypes,
		ResponseTypes:           request.ResponseTypes,
	})
	if err != nil {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client_metadata",
			Description: err.Error(),
		}
	}

	return &oalib.ClientInformationResponse{
		ClientID:         request.ClientID,
		ClientSecret:     request.ClientSecret,
		ClientIDIssuedAt: int(time.Now().UTC().Unix()),
	}, nil
}

// Revoke services token revocation.
// https://www.rfc-editor.org/rfc/rfc7009
func (oa *OAuthService) Revoke(c *gin.Context, request *dto.OAuthRevoke) *oalib.VerboseError {
	ctx := c.Request.Context()
	credentials, verr := oa.clientService.ParseCredentials(c)
	if verr != nil {
		return verr
	}
	client, verr := oa.clientService.Authenticate(ctx, credentials)
	if verr != nil {
		return verr
	}
	return oa.tokenService.Revoke(ctx, &dto.RevokeTokens{
		Token:    request.Token,
		ClientID: client.ID,
	})
}

type AuthorizeError struct {
	Err         string
	Description string
	RedirectURI string
}

func (a AuthorizeError) Error() string { return a.Err }

func newAuthorizeError(err string, description string, redirectURI redirecturi.RedirectURI) *AuthorizeError {
	if err != "" {
		redirectURI.SetError(err)
	}
	return &AuthorizeError{
		Err:         err,
		Description: description,
		RedirectURI: redirectURI.String(),
	}
}
