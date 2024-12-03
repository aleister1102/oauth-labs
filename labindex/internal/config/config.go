package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/cyllective/oauth-labs/labindex/internal/dto"
)

var config *viper.Viper

func Init() (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigType("yaml")
	cfg.SetDefault("environment", "production")

	cfg.SetDefault("server.host", "127.0.0.1")
	cfg.SetDefault("server.port", 3000)

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

func GetLabs() *dto.Labs {
	type labEntry struct {
		Client string `mapstructure:"client"`
		Server string `mapstructure:"server"`
		Number int    `mapstructure:"number"`
	}
	var labs []labEntry
	if err := config.UnmarshalKey("labs", &labs); err != nil {
		panic(err)
	}
	entries := make(dto.Labs, len(labs))
	for i, le := range labs {
		entries[i] = dto.LabEntry{
			Number: fmt.Sprintf("%02d", le.Number),
			Client: le.Client,
			Server: le.Server,
		}
	}
	return &entries
}
