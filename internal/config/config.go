package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Name        string `yaml:"name"`
	NetworkName string `yaml:"network_name"`

	Configs    map[string]ConfigFile `yaml:"configs"`
	Sources    []SourceConfig        `yaml:"sources"`
	Containers []ContainerConfig     `yaml:"containers"`
	Services   []ServiceConfig       `yaml:"services"`
	Actions    []ActionConfig        `yaml:"actions"`
}

func New() (*Config, error) {
	cfg := &Config{}

	file, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
