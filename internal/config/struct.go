package config

import (
	"time"

	"github.com/pilat/devbox/internal/docker"
)

type ConfigFile struct {
	Content string  `yaml:"content"`
	UID     string  `yaml:"uid"`
	GID     string  `yaml:"gid"`
	Mode    *uint32 `yaml:"mode"`
}

type SourceConfig struct {
	Name           string   `yaml:"name"`
	URL            string   `yaml:"url"`
	Branch         string   `yaml:"branch"`
	SparseCheckout []string `yaml:"sparseCheckout"`
	Environment    []string `yaml:"environment"`
	EnvFile        []string `yaml:"env_file"`
}

type ContainerConfig struct {
	Image      string `yaml:"image"`
	Dockerfile string `yaml:"dockerfile"`
}

type ServiceHealthcheckConfig struct {
	Test          []string      `yaml:""`
	Interval      time.Duration `yaml:""`
	Timeout       time.Duration `yaml:""`
	StartPeriod   time.Duration `yaml:""`
	StartInterval time.Duration `yaml:""`
	Retries       int           `yaml:""`
}

type ServiceConfig struct {
	Name        string                     `yaml:"name"`
	Hostname    string                     `yaml:"hostname"`
	Image       string                     `yaml:"image"`
	Command     []string                   `yaml:"command"`    // override the default command
	Entrypoint  *[]string                  `yaml:"entrypoint"` // override the default entrypoint
	Environment []string                   `yaml:"environment"`
	EnvFile     []string                   `yaml:"env_file"`
	User        string                     `yaml:"user"`
	Volumes     []string                   `yaml:"volumes"`
	Ports       []docker.ServicePortConfig `yaml:"-"`
	WorkingDir  string                     `yaml:"working_dir"`
	Healthcheck []string                   `yaml:"healthcheck"`
	HostAliases []string                   `yaml:"host_aliases"`
	DependsOn   []string                   `yaml:"depends_on"`
}

type ActionConfig struct {
	Name        string     `yaml:"name"`
	Image       string     `yaml:"image"`
	Commands    [][]string `yaml:"commands"`   // list of strings, "command" parser below
	Entrypoint  *[]string  `yaml:"entrypoint"` // override the default entrypoint
	Environment []string   `yaml:"environment"`
	EnvFile     []string   `yaml:"env_file"`
	User        string     `yaml:"user"`
	Volumes     []string   `yaml:"volumes"`
	WorkingDir  string     `yaml:"working_dir"`
	DependsOn   []string   `yaml:"depends_on"`
}

func (s *ServiceConfig) UnmarshalYAML(unmarshal func(any) error) error {
	type serviceConfig ServiceConfig
	var sc serviceConfig
	if err := unmarshal(&sc); err != nil {
		return err
	}

	type serviceConfigPorts struct {
		Ports []string `yaml:"ports"`
	}
	var sc2 serviceConfigPorts
	if err := unmarshal(&sc2); err != nil {
		return err
	}

	sc.Ports = []docker.ServicePortConfig{}
	for _, port := range sc2.Ports {
		pp, err := docker.ParsePortConfig(port)
		if err != nil {
			return err
		}

		for _, p1 := range pp {
			sc.Ports = append(sc.Ports, p1)
		}
	}

	*s = ServiceConfig(sc)
	return nil
}

func (s *ActionConfig) UnmarshalYAML(unmarshal func(any) error) error {
	type actionConfig ActionConfig
	var sc actionConfig
	if err := unmarshal(&sc); err != nil {
		return err
	}

	type actionConfigCommand struct {
		Command []string `yaml:"command"`
	}
	var sc3 actionConfigCommand
	if err := unmarshal(&sc3); err != nil {
		return err
	}

	if len(sc3.Command) > 0 {
		sc.Commands = [][]string{sc3.Command}
	}

	*s = ActionConfig(sc)
	return nil
}
