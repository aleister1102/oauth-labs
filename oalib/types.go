package oalib

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// https://www.rfc-editor.org/rfc/rfc6749#section-4.1.1
type AuthorizationCodeRequest struct {
	ResponseType string `form:"response_type"` // required
	ClientID     string `form:"client_id"`     // required
	RedirectURI  string `form:"client_id"`     // optional
	Scope        string `form:"scope"`         // optional
	State        string `form:"state"`         // recommended

	CodeChallengeMethod string `form:"code_challenge_method"` // optional
	CodeChallenge       string `form:"code_challenge_method"` // optional
}

// https://www.rfc-editor.org/rfc/rfc6749#section-4.1.3
type TokenCodeRequest struct {
	GrantType    string `form:"grant_type"`   // required
	Code         string `form:"code"`         // required
	RedirectURI  string `form:"redirect_uri"` // optional
	ClientID     string `form:"client_id"`    // optional
	ClientSecret string `form:"client_secret"`

	CodeVerifier string `form:"code_verifier"` // optional

	BasicAuthUsername string
	BasicAuthPassword string
}

// https://www.rfc-editor.org/rfc/rfc6749#section-5.1
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// https://www.rfc-editor.org/rfc/rfc6749#section-5.2
type TokenCodeError struct {
	Err         string `json:"error"`
	Description string `json:"error_description,omitempty"`
	URI         string `json:"error_uri,omitempty"`
}

func (t TokenCodeError) Error() string { return t.Err }

type AuthorizeError struct {
	Err         string
	Description string
	RedirectURI string
}

func (a AuthorizeError) Error() string { return a.Err }

type AuthorizationCode struct {
	Code        string `redis:"code" json:"code"`
	ClientID    string `redis:"client_id" json:"client_id"`
	UserID      string `redis:"user_id" json:"user_id"`
	RedirectURI string `redis:"redirect_uri" json:"redirect_uri"`
	Scope       string `redis:"scope" json:"scope"`
	CreatedAt   int    `redis:"created_at" json:"created_at"`

	// PKCE
	CodeChallenge       string `redis:"code_challenge" json:"code_challenge"`
	CodeChallengeMethod string `redis:"code_challenge_method" json:"code_challenge_method"`
}

type ClientCredentials struct {
	ID     string
	Secret string
}

// https://www.rfc-editor.org/rfc/rfc7591#section-2
type ClientMetadata struct {
	RedirectURIs            []string `json:"redirect_uris,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	ClientName              string   `json:"client_name,omitempty"`
	ClientURI               string   `json:"client_uri,omitempty"`
	LogoURI                 string   `json:"logo_uri,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
	Contacts                []string `json:"contacts,omitempty"`
	TosURI                  string   `json:"tos_uri,omitempty"`
	PolicyURI               string   `json:"policy_uri,omitempty"`
	JwksURI                 string   `json:"jwks_uri,omitempty"`
	JWKs                    jwk.Set  `json:"jwks,omitempty"`
	SoftwareID              string   `json:"software_id,omitempty"`
	SoftwareVersion         string   `json:"software_version,omitempty"`
}

// https://www.rfc-editor.org/rfc/rfc7591#section-3.2.1
type ClientInformationResponse struct {
	ClientID              string `json:"client_id"`
	ClientSecret          string `json:"client_secret,omitempty"`
	ClientIDIssuedAt      int    `json:"client_id_issued_at,omitempty"`
	ClientSecretExpiresAt int    `json:"client_secret_expires_at,omitempty"`
}

// https://www.rfc-editor.org/rfc/rfc7591#section-3.2.2
type ClientRegistrationErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// https://datatracker.ietf.org/doc/html/rfc7009#section-2.1
// Token revocation
type RevocationRequest struct {
	Token         string `json:"token"`
	TokenTypeHint string `json:"token_type_hint"`
}
