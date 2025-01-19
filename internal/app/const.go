package app

import (
	"os"
	"path/filepath"
)

var (
	AppDir string = "/opt/devbox"
)

const (
	SourcesDir = "sources"
	StateFile  = ".devboxstate"
	EnvFile    = ".env"
)

func init() {
	if home, err := os.UserHomeDir(); err == nil {
		AppDir = filepath.Join(home, ".devbox")
	}
}
