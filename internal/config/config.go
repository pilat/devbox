package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func New(projectPath string) (*Config, error) {
	configFile := filepath.Join(projectPath, "devbox.yaml")
	stateFile := filepath.Join(projectPath, ".devboxstate")

	cfg := &Config{}
	configContent, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configContent, cfg)
	if err != nil {
		return nil, err
	}

	stateContent, err := os.ReadFile(stateFile)
	if os.IsNotExist(err) {
		stateContent = []byte("{}")
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(stateContent, &cfg.State)
	if err != nil {
		return nil, err
	}

	cfg.State.filename = stateFile
	if cfg.State.Mounts == nil {
		cfg.State.Mounts = make(map[string]string)
	}

	return cfg, nil
}
