package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab04/client/internal/constants"
	"github.com/cyllective/oauth-labs/lab04/client/internal/utils"
)

var config *viper.Viper

func Init() (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigType("yaml")
	cfg.SetDefault("environment", "production")

	cfg.SetDefault("server.host", "127.0.0.1")
	cfg.SetDefault("server.port", 3001)

	cfg.SetDefault("database.host", "127.0.0.1")
	cfg.SetDefault("database.port", 3306)
	cfg.SetDefault("database.name", fmt.Sprintf("client%s", constants.LabNumber))
	cfg.SetDefault("database.username", fmt.Sprintf("client%s", constants.LabNumber))
	cfg.SetDefault("database.password", "")

	cfg.SetDefault("cookie.secret", utils.RandomHex(32))
	cfg.SetDefault("cookie.path", "/")
	cfg.SetDefault("cookie.domain", "127.0.0.1")
	cfg.SetDefault("cookie.max_age", 86400)
	cfg.SetDefault("cookie.secure", false)
	cfg.SetDefault("cookie.http_only", true)
	cfg.SetDefault("cookie.samesite", "strict")

	cfg.SetDefault("client.id", "5cdad30c-09b3-4317-9290-10e1462d88ea")
	cfg.SetDefault("client.secret", "undefined")
	cfg.SetDefault("client.scopes", []string{"read:profile"})
	cfg.SetDefault("client.redirect_uri", "http://127.0.0.1:3001/callback")
	cfg.SetDefault("client.uri", "http://127.0.0.1:3001/")
	cfg.SetDefault("client.logo_uri", "http://127.0.0.1:3001/static/img/logo.png")

	cfg.SetDefault("authorization_server.issuer", "http://127.0.0.1:3000")
	cfg.SetDefault("authorization_server.authorize_uri", "http://127.0.0.1:3000/oauth/authorize")
	cfg.SetDefault("authorization_server.token_uri", "http://127.0.0.1:3000/oauth/token")
	cfg.SetDefault("authorization_server.jwk_uri", "http://127.0.0.1:3000/.well-known/jwks.json")
	cfg.SetDefault("authorization_server.revocation_uri", "http://127.0.0.1:3000/oauth/revoke")
	cfg.SetDefault("authorization_server.register_uri", "http://127.0.0.1:3000/oauth/register")
	cfg.SetDefault("authorization_server.registration_secret", utils.RandomHex(32))

	cfg.SetDefault("resource_server.base_url", "http://127.0.0.1:3000")

	cfg.SetDefault("redis.network", "tcp")
	cfg.SetDefault("redis.host", "127.0.0.1")
	cfg.SetDefault("redis.port", 6379)
	cfg.SetDefault("redis.database", 0)
	cfg.SetDefault("redis.username", "")
	cfg.SetDefault("redis.password", "")

	config = cfg
	return config, nil
}

func InitFrom(path string) (*viper.Viper, error) {
	cfg, err := Init()
	if err != nil {
		return nil, err
	}
	fh, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration from %s: %w", path, err)
	}
	defer fh.Close()
	err = cfg.MergeConfig(fh)
	if err != nil {
		return nil, fmt.Errorf("failed to merge configuration from %s: %w", path, err)
	}
	config = cfg
	return config, nil
}

func Get() *viper.Viper {
	return config
}

func GetSessionSecret() []byte {
	cfg := Get()
	s, err := hex.DecodeString(cfg.GetString("cookie.secret"))
	if err != nil {
		panic(fmt.Errorf("failed to retrieve session secret: %w", err))
	}
	return s
}

func GetOAuthConfig() *oauth2.Config {
	cfg := Get()
	return &oauth2.Config{
		ClientID:     cfg.GetString("client.id"),
		ClientSecret: cfg.GetString("client.secret"),
		Scopes:       cfg.GetStringSlice("client.scopes"),
		RedirectURL:  cfg.GetString("client.redirect_uri"),
		Endpoint: oauth2.Endpoint{
			AuthURL:   cfg.GetString("authorization_server.authorize_uri"),
			TokenURL:  cfg.GetString("authorization_server.token_uri"),
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
}

func GetRedisConfig() *redis.Options {
	cfg := Get()
	addr := fmt.Sprintf("%s:%d", cfg.GetString("redis.host"), cfg.GetInt("redis.port"))
	opts := &redis.Options{
		Addr:       addr,
		Network:    config.GetString("redis.network"),
		DB:         cfg.GetInt("redis.db"),
		ClientName: fmt.Sprintf("oauth-labs-client-%s", constants.LabNumber),
	}
	if username := cfg.GetString("redis.username"); username != "" {
		opts.Username = username
	}
	if password := cfg.GetString("redis.password"); password != "" {
		opts.Password = password
	}
	return opts
}

func GetDatabaseURI() string {
	cfg := Get()
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4,utf8&collation=utf8mb4_general_ci",
		cfg.GetString("database.username"),
		cfg.GetString("database.password"),
		cfg.GetString("database.host"),
		cfg.GetInt("database.port"),
		cfg.GetString("database.name"),
	)
}

func GetJWTEncryptionKey() []byte {
	cfg := Get()
	key := cfg.GetString("jwt.encryption_key")
	s, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}
	if len(s) != 32 {
		panic(errors.New("invalid jwt.encryption_key length, must be 32 bytes"))
	}
	return s
}

func GetSessionOptions() sessions.Options {
	cfg := Get()
	opts := sessions.Options{
		Path:     cfg.GetString("cookie.path"),
		Domain:   cfg.GetString("cookie.domain"),
		MaxAge:   cfg.GetInt("cookie.max_age"),
		Secure:   cfg.GetBool("cookie.secure"),
		HttpOnly: cfg.GetBool("cookie.http_only"),
	}
	switch strings.ToLower(cfg.GetString("cookie.samesite")) {
	case "lax":
		opts.SameSite = http.SameSiteLaxMode
	case "none":
		opts.SameSite = http.SameSiteLaxMode
	default:
		opts.SameSite = http.SameSiteStrictMode
	}
	return opts
}
