package lab03

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/cyllective/oauth-labs/configure/internal/constants"
	"github.com/cyllective/oauth-labs/configure/internal/utils"
)

func Configure() {
	dockerLabDir := filepath.Join(constants.DockerDir, "lab03")
	if err := os.MkdirAll(dockerLabDir, 0o750); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			panic(err)
		}
	}
	clientID := utils.RandomClientID()
	registrationSecret := utils.RandomHex(32)

	clientConfigFile := filepath.Join(dockerLabDir, "client.config.yaml")
	t, err := template.New("client_config").Parse(clientConfigTemplate)
	if err != nil {
		panic(err)
	}
	utils.WriteTemplateConfig(clientConfigFile, t, map[string]string{
		"database_password":   utils.RandomPassword(32),
		"client_id":           clientID,
		"client_secret":       utils.RandomPassword(32),
		"registration_secret": registrationSecret,
		"cookie_secret":       utils.RandomHex(32),
	})

	serverConfigFile := filepath.Join(dockerLabDir, "server.config.yaml")
	t, err = template.New("server_config").Parse(serverConfigTemplate)
	if err != nil {
		panic(err)
	}
	utils.WriteTemplateConfig(serverConfigFile, t, map[string]string{
		"admin_password":      utils.RandomPassword(32),
		"database_password":   utils.RandomPassword(32),
		"cookie_secret":       utils.RandomHex(32),
		"registration_secret": registrationSecret,
		"allowed_client":      clientID,
		"encryption_key":      utils.RandomHex(32),
		"private_key":         utils.IndentPEM(utils.NewRSAPrivateKey(), 4),
	})
}

var clientConfigTemplate = `
server:
  host: '0.0.0.0'
  port: 3000

database:
  host: 'db'
  port: 3306
  name: 'client03'
  username: 'client03'
  password: '{{ .database_password }}'

client:
  id: '{{ .client_id }}'
  name: 'client-03'
  secret: '{{ .client_secret }}'
  scopes:
    - 'read:profile'
  uri: 'https://client-03.oauth.labs'
  logo_uri: 'https://client-03.oauth.labs/static/img/logo.png'
  redirect_uri: 'https://client-03.oauth.labs/callback'

authorization_server:
  issuer: 'https://server-03.oauth.labs'
  authorize_uri: 'https://server-03.oauth.labs/oauth/authorize'
  token_uri: 'http://server-03:3000/oauth/token'
  jwk_uri: 'http://server-03:3000/.well-known/jwks.json'
  revocation_uri: 'http://server-03:3000/oauth/revoke'
  register_uri: 'http://server-03:3000/oauth/register'
  registration_secret: '{{ .registration_secret }}'

resource_server:
  base_url: 'http://server-03:3000'

cookie:
  name: 'client-03'
  secret: '{{ .cookie_secret }}'
  domain: 'client-03.oauth.labs'
  path: '/'
  max_age: 80400
  http_only: true
  secure: true
  samesite: 'strict'

redis:
  host: 'valkey'
  port: 6379
  database: 0
`

var serverConfigTemplate = `
server:
  host: '0.0.0.0'
  port: 3000
  admin_password: '{{ .admin_password }}'

database:
  host: 'db'
  port: 3306
  name: 'server03'
  username: 'server03'
  password: '{{ .database_password }}'

cookie:
  name: 'server-03'
  secret: '{{ .cookie_secret }}'
  domain: 'server-03.oauth.labs'
  path: '/'
  max_age: 80400
  http_only: true
  secure: true
  samesite: 'strict'

redis:
  host: 'valkey'
  port: 6379
  database: 0

oauth:
  issuer: 'https://server-03.oauth.labs'
  registration_secret: '{{ .registration_secret }}'
  allowed_clients:
    - '{{ .allowed_client }}'
  encryption_key: '{{ .encryption_key }}'
  private_key: |
    {{ .private_key }}
`
