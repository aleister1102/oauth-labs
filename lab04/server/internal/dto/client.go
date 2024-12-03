package dto

type Client struct {
	ID                      string   `json:"id"`
	Secret                  string   `json:"secret"`
	Name                    string   `json:"name"`
	URL                     string   `json:"url"`
	LogoURI                 string   `json:"logo_uri"`
	Scope                   []string `json:"scope"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}

type ClientRegister struct {
	ID                      string   `json:"id" binding:"required"`
	Secret                  string   `json:"secret" binding:"required"`
	Name                    string   `json:"name" binding:"required"`
	URL                     string   `json:"url"`
	LogoURI                 string   `json:"logo_uri"`
	Scope                   []string `json:"scope" binding:"required"`
	RedirectURIs            []string `json:"redirect_uris" binding:"required"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}
