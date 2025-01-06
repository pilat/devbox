package composer

type SourceConfigs map[string]SourceConfig

type AlternativeVolumes []string

type SourceConfig struct {
	URL            string   `yaml:"url"`
	Branch         string   `yaml:"branch"`
	SparseCheckout []string `yaml:"sparseCheckout"`
	Environment    []string `yaml:"environment"`
}
