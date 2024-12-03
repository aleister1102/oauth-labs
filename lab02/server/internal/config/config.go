package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	"github.com/cyllective/oauth-labs/lab02/server/internal/constants"
	"github.com/cyllective/oauth-labs/lab02/server/internal/utils"
)

var config *viper.Viper

func Init() (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigType("yaml")
	cfg.SetDefault("environment", "production")

	cfg.SetDefault("server.host", "127.0.0.1")
	cfg.SetDefault("server.port", 3000)
	cfg.SetDefault("server.bcrypt_cost", 12)
	cfg.SetDefault("server.admin_password", utils.RandomPassword(32))

	cfg.SetDefault("database.host", "127.0.0.1")
	cfg.SetDefault("database.port", 3306)
	cfg.SetDefault("database.name", fmt.Sprintf("server%s", constants.LabNumber))
	cfg.SetDefault("database.username", fmt.Sprintf("server%s", constants.LabNumber))
	cfg.SetDefault("database.password", "")

	cfg.SetDefault("oauth.issuer", "http://127.0.0.1:3000")
	cfg.SetDefault("oauth.expiration_seconds", 3600)
	cfg.SetDefault("oauth.allowed_clients", []string{})
	cfg.SetDefault("oauth.registration_secret", utils.RandomHex(32))
	cfg.SetDefault("oauth.encryption_key", utils.RandomHex(32))
	cfg.SetDefault("oauth.private_key", "undefined")

	cfg.SetDefault("cookie.name", fmt.Sprintf("server-%s", constants.LabNumber))
	cfg.SetDefault("cookie.secret", utils.RandomHex(32))
	cfg.SetDefault("cookie.path", "/")
	cfg.SetDefault("cookie.domain", "127.0.0.1")
	cfg.SetDefault("cookie.max_age", 86400)
	cfg.SetDefault("cookie.secure", false)
	cfg.SetDefault("cookie.http_only", true)
	cfg.SetDefault("cookie.samesite", "lax")

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
		panic(err)
	}
	return s
}

func GetJWTPrivateKey() []byte {
	cfg := Get()
	return []byte(cfg.GetString("oauth.private_key"))
}

func GetDatabaseURI() string {
	cfg := Get()
	uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4,utf8&collation=utf8mb4_general_ci",
		cfg.GetString("database.username"),
		cfg.GetString("database.password"),
		cfg.GetString("database.host"),
		cfg.GetInt("database.port"),
		cfg.GetString("database.name"),
	)
	return uri
}

func GetRedisConfig() *redis.Options {
	cfg := Get()
	addr := fmt.Sprintf("%s:%d", cfg.GetString("redis.host"), cfg.GetInt("redis.port"))
	opts := &redis.Options{
		Addr:       addr,
		Network:    config.GetString("redis.network"),
		DB:         cfg.GetInt("redis.db"),
		ClientName: fmt.Sprintf("oauth-labs-server-%s", constants.LabNumber),
	}
	if username := cfg.GetString("redis.username"); username != "" {
		opts.Username = username
	}
	if password := cfg.GetString("redis.password"); password != "" {
		opts.Password = password
	}
	return opts
}

func GetJWTEncryptionKey() []byte {
	cfg := Get()
	key := cfg.GetString("oauth.encryption_key")
	s, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}
	if len(s) != 32 {
		panic(errors.New("invalid oauth.encryption_key, length must be 32 bytes"))
	}
	return s
}

func GetAccessTokenExpiration() time.Duration {
	cfg := Get()
	seconds := cfg.GetInt("oauth.expiration_seconds")
	return time.Duration(seconds) * time.Second
}

func GetSessionOptions() sessions.Options {
	cfg := Get()
	opts := sessions.Options{
		Path:     cfg.GetString("cookie.path"),
		Domain:   cfg.GetString("cookie.domain"),
		MaxAge:   cfg.GetInt("cookie.max_age"),
		Secure:   cfg.GetBool("cookie.secure"),
		HttpOnly: cfg.GetBool("cookie.http_only"),
		SameSite: http.SameSiteLaxMode,
	}
	return opts
}
