package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func New(filename string) (*Config, error) {
	_, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get config file: %w", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(content, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
