package project

type (
	SourceConfigs map[string]SourceConfig
	SourceConfig  struct {
		URL            string   `yaml:"url"`
		Branch         string   `yaml:"branch"`
		SparseCheckout []string `yaml:"sparseCheckout"`
		Environment    []string `yaml:"environment"`
	}
)

type (
	ScenarioConfigs map[string]ScenarioConfig
	ScenarioConfig  struct {
		Service     string   `yaml:"service"`
		Description string   `yaml:"description"`
		Command     []string `yaml:"command"`
		Entrypoint  []string `yaml:"entrypoint"`
		Tty         *bool    `yaml:"tty"`        // default: auto-detect
		Interactive *bool    `yaml:"stdin_open"` // default: true
		WorkingDir  string   `yaml:"working_dir"`
		User        string   `yaml:"user"`
	}
)

type (
	HostConfigs []HostConfig
	HostConfig  struct {
		IP    string   `yaml:"ip"`
		Hosts []string `yaml:"hosts"`
	}
)

type CertConfig struct {
	Domains  []string `yaml:"domains"`
	KeyFile  string   `yaml:"keyFile"`
	CertFile string   `yaml:"certFile"`
}
