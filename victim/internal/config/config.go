package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var config *viper.Viper

func Init() (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigType("yaml")
	cfg.SetDefault("environment", "production")
	cfg.SetDefault("server.host", "127.0.0.1")
	cfg.SetDefault("server.port", 7777)

	cfg.SetDefault("browser.globalTimeout", 10)
	cfg.SetDefault("browser.bin", "auto")

	cfg.SetDefault("victims", map[string]interface{}{})

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
