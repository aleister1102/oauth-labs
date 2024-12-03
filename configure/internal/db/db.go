package db

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cyllective/oauth-labs/configure/internal/constants"
	"github.com/cyllective/oauth-labs/configure/internal/utils"
	"gopkg.in/yaml.v3"
)

type databaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type configWrapper struct {
	Database databaseConfig `yaml:"database"`
}

func Configure() {
	dockerDBDir := filepath.Join(constants.DockerDir, "db")
	if err := os.MkdirAll(dockerDBDir, 0o750); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			panic(err)
		}
	}

	pwFile := filepath.Join(constants.DockerDir, "db", "root_password.prod.txt")
	fh, err := os.OpenFile(pwFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
	w := bufio.NewWriter(fh)
	if _, err := w.WriteString(utils.RandomPassword(32)); err != nil {
		panic(err)
	}
	_ = w.Flush()

	// Extract each components' database credentials to generate init.sql
	dirs := []string{
		"lab00",
		"lab01",
		"lab02",
		"lab03",
		"lab04",
		"lab05",
	}
	sqlInit := new(strings.Builder)
	for _, labDir := range dirs {
		dockerLabDir := filepath.Join(constants.DockerDir, labDir)
		configFile := filepath.Join(dockerLabDir, "server.config.yaml")
		serverConfig := unmarshalConfig(configFile)
		configFile = filepath.Join(dockerLabDir, "client.config.yaml")
		clientConfig := unmarshalConfig(configFile)
		WriteSQLBlock(sqlInit, labDir, serverConfig, clientConfig)
	}

	sqlInitFile := filepath.Join(constants.DockerDir, "db", "init.prod.sql")
	fh, err = os.OpenFile(sqlInitFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
	w = bufio.NewWriter(fh)
	if _, err := w.WriteString(sqlInit.String()); err != nil {
		panic(err)
	}
	_ = w.Flush()
}

func WriteSQLBlock(wr io.Writer, comment string, sc databaseConfig, cc databaseConfig) {
	tmpl, err := template.New("sql_block").Parse(`-- {{ .comment }}
CREATE USER '{{ .sc.Username }}'@'%' IDENTIFIED BY '{{ .sc.Password }}';
CREATE DATABASE {{ .sc.Name }};
GRANT ALL PRIVILEGES ON {{ .sc.Name }}.* TO '{{ .sc.Username }}'@'%';
CREATE USER '{{ .cc.Username }}'@'%' IDENTIFIED BY '{{ .cc.Password }}';
CREATE DATABASE {{ .cc.Name }};
GRANT ALL PRIVILEGES ON {{ .cc.Name }}.* TO '{{ .cc.Username }}'@'%';
`)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(wr, map[string]any{
		"comment": comment,
		"sc":      sc,
		"cc":      cc,
	})
	if err != nil {
		panic(err)
	}
}

func unmarshalConfig(file string) databaseConfig {
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
	return cfg.Database
}
