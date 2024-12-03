package victim

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cyllective/oauth-labs/configure/internal/constants"
	"gopkg.in/yaml.v3"
)

type serverConfig struct {
	AdminPassword string `yaml:"admin_password"`
}

type configWrapper struct {
	Server serverConfig `yaml:"server"`
}

func Configure() {
	dockerVictimDir := filepath.Join(constants.DockerDir, "victim")
	if err := os.MkdirAll(dockerVictimDir, 0o750); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			panic(err)
		}
	}

	lab02ConfigFile := filepath.Join(constants.DockerDir, "lab02", "server.config.yaml")
	lab02Config := unmarshalConfig(lab02ConfigFile)
	lab03ConfigFile := filepath.Join(constants.DockerDir, "lab03", "server.config.yaml")
	lab03Config := unmarshalConfig(lab03ConfigFile)
	config := fmt.Sprintf(`
server:
  host: '0.0.0.0'
  port: 3000

victims:
  2:
    server_url: 'https://server-02.oauth.labs'
    client_url: 'https://client-02.oauth.labs'
    username: 'admin'
    password: '%s'

  3:
    server_url: 'https://server-03.oauth.labs'
    client_url: 'https://client-03.oauth.labs'
    username: 'admin'
    password: '%s'
`, lab02Config.AdminPassword, lab03Config.AdminPassword)

	configFile := filepath.Join(dockerVictimDir, "config.yaml")
	fh, err := os.OpenFile(configFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
	w := bufio.NewWriter(fh)
	w.WriteString(config)
	_ = w.Flush()
}

func unmarshalConfig(file string) serverConfig {
	fh, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	rawConfig, err := io.ReadAll(fh)
	if err != nil {
		panic(err)
	}
	var cfg configWrapper
	if err := yaml.Unmarshal(rawConfig, &cfg); err != nil {
		panic(fmt.Errorf("failed to unmarshal config %s: %s", file, err.Error()))
	}
	return cfg.Server
}
