package metadata

import (
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type ServerMetadata struct {
	Issuer                                             string             `json:"issuer"`
	AuthorizationEndpoint                              *string            `json:"authorization_endpoint,omitempty"`
	TokenEndpoint                                      *string            `json:"token_endpoint,omitempty"`
	JwksURI                                            *string            `json:"jwks_uri,omitempty"`
	RegistrationEndpoint                               *string            `json:"registration_endpoint,omitempty"`
	ScopesSupported                                    mapset.Set[string] `json:"scopes_supported,omitempty"`
	ResponseTypesSupported                             mapset.Set[string] `json:"response_types_supported"`
	ResponseModesSupported                             mapset.Set[string] `json:"response_modes_supported,omitempty"`
	GrantTypesSupported                                mapset.Set[string] `json:"grant_types_supported,omitempty"` // "... if omitted, the default value is ["authorization_code", "implicit"]..."
	TokenEndpointAuthMethodsSupported                  mapset.Set[string] `json:"token_endpoint_auth_methods_supported,omitempty"`
	TokenEndpointAuthSigningAlgValuesSupported         mapset.Set[string] `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`
	ServiceDocumentation                               *string            `json:"service_documentation,omitempty"`
	UILocalesSupported                                 mapset.Set[string] `json:"ui_locales_supported,omitempty"`
	OpPolicyURI                                        *string            `json:"op_policy_uri,omitempty"`
	OpTosURI                                           *string            `json:"op_tos_uri,omitempty"`
	RevocationEndpoint                                 *string            `json:"revocation_endpoint,omitempty"`
	RevocationEndpointAuthMethodsSupported             mapset.Set[string] `json:"revocation_endpoint_auth_methods_supported,omitempty"`
	RevocationEndpointAuthSigningAlgValuesSupported    mapset.Set[string] `json:"revocation_endpoint_auth_signing_alg_values_supported,omitempty"`
	IntrospectionEndpoint                              *string            `json:"introspection_endpoint,omitempty"`
	IntrospectionEndpointAuthMethodsSupported          mapset.Set[string] `json:"introspection_endpoint_auth_methods_supported,omitempty"`
	IntrospectionEndpointAuthSigningAlgValuesSupported mapset.Set[string] `json:"introspection_endpoint_auth_signing_alg_values_supported,omitempty"`
	CodeChallengeMethodsSupported                      mapset.Set[string] `json:"code_challenge_methods_supported,omitempty"`
}

func (s *ServerMetadata) SupportsCodeChallengeMethod(method string) bool {
	if s.CodeChallengeMethodsSupported != nil {
		return s.CodeChallengeMethodsSupported.Contains(method)
	}
	return false
}

func (s *ServerMetadata) SupportsPKCE() bool {
	if s.CodeChallengeMethodsSupported == nil {
		return false
	}
	if s.CodeChallengeMethodsSupported.Cardinality() <= 0 {
		return false
	}
	return true
}

func New(issuer string, responseTypes ...string) *ServerMetadata {
	iss := strings.TrimSuffix(issuer, "/")
	meta := &ServerMetadata{
		Issuer:                            iss,
		AuthorizationEndpoint:             nil,
		TokenEndpoint:                     nil,
		JwksURI:                           nil,
		RegistrationEndpoint:              nil,
		ScopesSupported:                   nil,
		ResponseTypesSupported:            nil,
		ResponseModesSupported:            nil,
		GrantTypesSupported:               nil,
		TokenEndpointAuthMethodsSupported: nil,
		TokenEndpointAuthSigningAlgValuesSupported:         nil,
		ServiceDocumentation:                               nil,
		UILocalesSupported:                                 nil,
		OpPolicyURI:                                        nil,
		OpTosURI:                                           nil,
		RevocationEndpoint:                                 nil,
		RevocationEndpointAuthMethodsSupported:             nil,
		RevocationEndpointAuthSigningAlgValuesSupported:    nil,
		IntrospectionEndpoint:                              nil,
		IntrospectionEndpointAuthMethodsSupported:          nil,
		IntrospectionEndpointAuthSigningAlgValuesSupported: nil,
		CodeChallengeMethodsSupported:                      nil,
	}
	if len(responseTypes) > 0 {
		meta = meta.WithResponseTypes(responseTypes...)
	}
	return meta
}

type Endpoints struct {
	JwksURI               string
	RegistrationEndpoint  string
	AuthorizationEndpoint string
	TokenEndpoint         string
	RevocationEndpoint    string
}

// WithEndpoints populates endpoints and URLs, optionally prepending relative endpoints with the issuer.
func (s *ServerMetadata) WithEndpoints(endpoints *Endpoints) *ServerMetadata {
	prefixIssuer := func(endpoint *string) *string {
		if strings.HasPrefix(*endpoint, "/") {
			*endpoint = s.Issuer + *endpoint
			return endpoint
		}
		return endpoint
	}
	if endpoints.JwksURI != "" {
		s.JwksURI = prefixIssuer(&endpoints.JwksURI)
	}
	if endpoints.RegistrationEndpoint != "" {
		s.RegistrationEndpoint = prefixIssuer(&endpoints.RegistrationEndpoint)
	}
	if endpoints.AuthorizationEndpoint != "" {
		s.AuthorizationEndpoint = prefixIssuer(&endpoints.AuthorizationEndpoint)
	}
	if endpoints.TokenEndpoint != "" {
		s.TokenEndpoint = prefixIssuer(&endpoints.TokenEndpoint)
	}
	if endpoints.RevocationEndpoint != "" {
		s.RevocationEndpoint = prefixIssuer(&endpoints.RevocationEndpoint)
	}

	return s
}

// WithGrantType sets the grant_types_supported metadata field.
// Also adds the corresponding response_types_supported entry if applicable.
func (s *ServerMetadata) WithGrantTypes(grantTypes ...string) *ServerMetadata {
	if s.GrantTypesSupported == nil {
		s.GrantTypesSupported = mapset.NewSet[string]()
	}
	if s.ResponseTypesSupported == nil {
		s.ResponseTypesSupported = mapset.NewSet[string]()
	}
	for _, grantType := range grantTypes {
		switch grantType {
		case "authorization_code":
			s.GrantTypesSupported.Add(grantType)
			s.ResponseTypesSupported.Add("code")
		case "implicit":
			s.GrantTypesSupported.Add(grantType)
			s.ResponseTypesSupported.Add("token")
		case "refresh_token":
			s.GrantTypesSupported.Add(grantType)
		default:
			panic(fmt.Errorf("unknown grant_types_supported value %#v", grantType))
		}
	}
	return s
}

// WithResponseTypes sets the response_types_supported metadata field.
// Also adds the corresponding grant_types_supported entry if applicable.
func (s *ServerMetadata) WithResponseTypes(responseTypes ...string) *ServerMetadata {
	if s.ResponseTypesSupported == nil {
		s.ResponseTypesSupported = mapset.NewSet[string]()
	}
	if s.GrantTypesSupported == nil {
		s.GrantTypesSupported = mapset.NewSet[string]()
	}
	for _, responseType := range responseTypes {
		switch responseType {
		case "code":
			s.ResponseTypesSupported.Add(responseType)
			s.GrantTypesSupported.Add("authorization_code")
		case "token":
			s.ResponseTypesSupported.Add(responseType)
			s.GrantTypesSupported.Add("implicit")

		default:
			panic(fmt.Errorf("unknown response_types_supported value %#v", responseType))
		}
	}
	return s
}

// WithCodeChallengeMethods sets the code_challenge_methods_supported metadata field. (PKCE support)
func (s *ServerMetadata) WithCodeChallengeMethods(codeChallengeMethods ...string) *ServerMetadata {
	if s.CodeChallengeMethodsSupported == nil {
		s.CodeChallengeMethodsSupported = mapset.NewSet[string]()
	}
	for _, codeChallengeMethod := range codeChallengeMethods {
		switch codeChallengeMethod {
		case "plain":
			s.CodeChallengeMethodsSupported.Add(codeChallengeMethod)
		case "S256":
			s.CodeChallengeMethodsSupported.Add(codeChallengeMethod)
		default:
			panic(fmt.Errorf("unknown code_challenge_methods_supported value %#v", codeChallengeMethod))
		}
	}

	return s
}

// WithTokenEndpointAuthMethodsSupported sets the token_endpoint_auth_methods_supported metadata field.
func (s *ServerMetadata) WithTokenEndpointAuthMethodsSupported(values ...string) *ServerMetadata {
	if s.TokenEndpointAuthMethodsSupported == nil {
		s.TokenEndpointAuthMethodsSupported = mapset.NewSet[string]()
	}
	for _, v := range values {
		switch v {
		case "client_secret_basic":
			s.TokenEndpointAuthMethodsSupported.Add(v)
		case "client_secret_post":
			s.TokenEndpointAuthMethodsSupported.Add(v)
		default:
			panic(fmt.Errorf("unknown token_endpoint_auth_methods_supported value %#v", v))
		}
	}

	return s
}

func (s *ServerMetadata) WithRevocationEndpointAuthMethodsSupported(values ...string) *ServerMetadata {
	if s.RevocationEndpointAuthMethodsSupported == nil {
		s.RevocationEndpointAuthMethodsSupported = mapset.NewSet[string]()
	}
	for _, v := range values {
		switch v {
		case "client_secret_basic":
			s.RevocationEndpointAuthMethodsSupported.Add(v)
		case "client_secret_post":
			s.RevocationEndpointAuthMethodsSupported.Add(v)
		default:
			panic(fmt.Errorf("unknown revocation_endpoint_auth_methods_supported value %#v", v))
		}
	}

	return s
}

func (s *ServerMetadata) WithScopes(scopes ...string) *ServerMetadata {
	s.ScopesSupported = mapset.NewSet(scopes...)
	return s
}
