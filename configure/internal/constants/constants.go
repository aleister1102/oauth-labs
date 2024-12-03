package constants

import (
	"path/filepath"

	"github.com/cyllective/oauth-labs/configure/internal/utils"
)

var (
	RootDir     = utils.GetLabRoot()
	DockerDir   = filepath.Join(RootDir, "docker")
	SQLInitFile = filepath.Join(DockerDir, "db", "init.sql")
)
