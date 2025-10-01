package project

type SourceConfigs map[string]SourceConfig
type SourceConfig struct {
	URL            string   `yaml:"url"`
	Branch         string   `yaml:"branch"`
	SparseCheckout []string `yaml:"sparseCheckout"`
	Environment    []string `yaml:"environment"`
}

type ScenarioConfigs map[string]ScenarioConfig
type ScenarioConfig struct {
	Service     string   `yaml:"service"`
	Description string   `yaml:"description"`
	Command     []string `yaml:"command"`
	Entrypoint  []string `yaml:"entrypoint"`
	Tty         *bool    `yaml:"tty"`        // default: auto-detect
	Interactive *bool    `yaml:"stdin_open"` // default: true
	WorkingDir  string   `yaml:"working_dir"`
	User        string   `yaml:"user"`
}

type HostConfigs []HostConfig
type HostConfig struct {
	IP    string   `yaml:"ip"`
	Hosts []string `yaml:"hosts"`
}

type CertConfig struct {
	Domains  []string `yaml:"domains"`
	KeyFile  string   `yaml:"keyFile"`
	CertFile string   `yaml:"certFile"`
}
