package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Name        string `yaml:"-"`
	NetworkName string `yaml:"-"`

	Sources    []SourceConfig    `yaml:"sources"`
	Containers []ContainerConfig `yaml:"containers"`
	Services   []ServiceConfig   `yaml:"services"`
	Actions    []ActionConfig    `yaml:"actions"`
}

func New(filename string) (*Config, error) {
	cfg := &Config{}

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
