package dto

import "github.com/lestrrat-go/jwx/v2/jwk"

type AuthorizeRequest struct {
	ResponseType string `form:"response_type"`
	ClientID     string `form:"client_id"`
	RedirectURI  string `form:"redirect_uri"`
	Scope        string `form:"scope"`
	State        string `form:"state"`

	CodeChallengeMethod string `form:"code_challenge_method"`
	CodeChallenge       string `form:"code_challenge"`
}

type TokenRequest struct {
	GrantType    string `form:"grant_type"`
	Code         string `form:"code"`
	RefreshToken string `form:"refresh_token"`
	RedirectURI  string `form:"redirect_uri"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	Scope        string `form:"scope"`
	CodeVerifier string `form:"code_verifier"`
}

type OAuthRegisterClient struct {
	JWKs                    jwk.Set `json:"jwks"`
	ClientURI               string  `json:"client_uri"`
	JwksURI                 string  `json:"jwks_uri"`
	LogoURI                 string  `json:"logo_uri"`
	TokenEndpointAuthMethod string  `json:"token_endpoint_auth_method"`
	Scope                   string  `json:"scope" binding:"required"`
	RegisterSecret          string
	ClientName              string `json:"client_name"`
	SoftwareVersion         string `json:"software_version"`
	SoftwareID              string `json:"software_id"`
	ClientSecret            string
	ClientID                string
	TosURI                  string   `json:"tos_uri"`
	PolicyURI               string   `json:"policy_uri"`
	ResponseTypes           []string `json:"response_types" binding:"required"`
	GrantTypes              []string `json:"grant_types" binding:"required"`
	RedirectURIs            []string `json:"redirect_uris" binding:"required"`
	Contacts                []string `json:"contacts"`
}

type OAuthRevoke struct {
	Token         string `form:"token" binding:"required"`
	TokenTypeHint string `form:"token_type_hint"`
}
