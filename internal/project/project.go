package project

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/pilat/devbox/internal/app"
	"golang.org/x/net/idna"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

// Replacement of composer service with our state keeper. Another extended service (with client inside) will be used.
type Project struct {
	*types.Project

	Sources      SourceConfigs
	Scenarios    ScenarioConfigs
	HostEntities []string // IP hostname1 [hostname2] [hostname3] ...
	CertConfig   CertConfig

	LocalMounts map[string]string // some service's full mount path -> local path
	HasHosts    bool

	hostConfigs HostConfigs
	envFiles    []string
}

// New creates a new project. We init it always with all profiles by using "*"
func New(ctx context.Context, projectName string, profiles []string) (*Project, error) {
	projectFolder := filepath.Join(app.AppDir, projectName)

	if _, err := os.Stat(projectFolder); os.IsNotExist(err) {
		return nil, fmt.Errorf("project '%s' not found", projectName)
	}

	o, err := cli.NewProjectOptions(
		[]string{},
		cli.WithWorkingDirectory(projectFolder),
		cli.WithDefaultConfigPath,
		cli.WithName(projectName),
		cli.WithInterpolation(true),
		cli.WithResolvedPaths(true),
		cli.WithProfiles(profiles),
		cli.WithExtension("x-devbox-sources", SourceConfigs{}),
		cli.WithExtension("x-devbox-scenarios", ScenarioConfigs{}),
		cli.WithExtension("x-devbox-hosts", HostConfigs{}),
		cli.WithExtension("x-devbox-cert", CertConfig{}),
		cli.WithExtension("x-devbox-default-stop-grace-period", Duration(0)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load compose project options: %w", err)
	}

	project, err := cli.ProjectFromOptions(ctx, o)
	if err != nil {
		return nil, fmt.Errorf("failed to load compose project: %w", err)
	}

	p := &Project{
		Project:     project,
		envFiles:    o.EnvFiles,
		LocalMounts: make(map[string]string),
	}

	allFuncs := []func(p *Project) error{
		loadState,
		applySources,
		applyScenarios,
		applyHosts,
		applyCert,
		setupGracePeriod,
		applyLabels,
		mountSourceVolumes,
	}

	for _, f := range allFuncs {
		if err := f(p); err != nil {
			return nil, fmt.Errorf("failed to open project '%s': %w", projectName, err)
		}
	}

	return p, nil
}

func (p *Project) WithSelectedServices(names []string, options ...types.DependencyOption) (*Project, error) {
	p2, err := p.Project.WithSelectedServices(names, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to select services: %w", err)
	}

	return &Project{
		Project: p2,
		Sources: p.Sources,
	}, nil
}

func (p *Project) SaveState() error {
	state := &stateFileStruct{
		Mounts:   p.LocalMounts,
		HasHosts: p.HasHosts,
	}

	json, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	filename := filepath.Join(p.WorkingDir, app.StateFile)
	err = os.WriteFile(filename, json, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (p *Project) Reload(ctx context.Context, profiles []string) error {
	p2, err := New(ctx, p.Name, profiles)
	if err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	*p = *p2

	return nil
}

func (p *Project) absPath(source string) string {
	prefix := ""

	switch {
	case strings.HasPrefix(source, "~"):
		if home, err := os.UserHomeDir(); err != nil {
			prefix = home
		}
	case strings.HasPrefix(source, "."):
		prefix = p.WorkingDir
	}

	return filepath.Join(prefix, source)
}

func loadState(p *Project) error {
	filename := filepath.Join(p.WorkingDir, app.StateFile)

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get state file: %w", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	state := &stateFileStruct{}
	err = json.Unmarshal(content, state)
	if err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	if state.Mounts == nil {
		return nil
	}

	p.LocalMounts = state.Mounts

	return nil
}

func applySources(p *Project) error {
	if s, ok := p.Extensions["x-devbox-sources"]; ok {
		p.Sources = s.(SourceConfigs) // nolint: forcetypeassert
	}

	return nil
}

func applyScenarios(p *Project) error {
	if s, ok := p.Extensions["x-devbox-scenarios"]; ok {
		p.Scenarios = s.(ScenarioConfigs) // nolint: forcetypeassert
	}

	return nil
}

func applyHosts(p *Project) error {
	if s, ok := p.Extensions["x-devbox-hosts"]; ok {
		hostConfigs := s.(HostConfigs) // nolint: forcetypeassert

		ipToHosts := make(map[string][]string)
		for _, item := range hostConfigs {
			ip := net.ParseIP(item.IP)
			item.IP = ip.String()

			for _, hostname := range item.Hosts {
				hostname, err := idna.Lookup.ToASCII(hostname)
				if err != nil {
					return fmt.Errorf("failed to convert hostname to ASCII: %w", err)
				}

				if _, ok := ipToHosts[item.IP]; !ok {
					ipToHosts[item.IP] = []string{}
				}

				ipToHosts[item.IP] = append(ipToHosts[item.IP], hostname)
			}
		}

		entities := []string{}
		for ip, hosts := range ipToHosts {
			entities = append(entities, fmt.Sprintf("%s %s", ip, strings.Join(hosts, " ")))
		}

		p.HostEntities = entities
	}

	return nil
}

func applyCert(p *Project) error {
	if s, ok := p.Extensions["x-devbox-cert"]; ok {
		p.CertConfig = s.(CertConfig) // nolint: forcetypeassert

		if p.CertConfig.CertFile != "" {
			p.CertConfig.CertFile = p.absPath(p.CertConfig.CertFile)
		}
		if p.CertConfig.KeyFile != "" {
			p.CertConfig.KeyFile = p.absPath(p.CertConfig.KeyFile)
		}
	}

	return nil
}

func setupGracePeriod(p *Project) error {
	var defaultStopGracePeriod *Duration

	if s, ok := p.Extensions["x-devbox-default-stop-grace-period"]; ok {
		v := s.(Duration) // nolint: forcetypeassert
		defaultStopGracePeriod = &v
	}

	// apply default grace period to all services
	for name, s := range p.Services {
		if s.StopGracePeriod != nil {
			continue
		}

		if defaultStopGracePeriod != nil {
			s.StopGracePeriod = defaultStopGracePeriod
		}

		p.Services[name] = s
	}

	return nil
}

func applyLabels(p *Project) error {
	for name, s := range p.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     p.Name,
			api.ServiceLabel:     name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  p.WorkingDir,
			api.ConfigFilesLabel: strings.Join(p.ComposeFiles, ","),
			api.OneoffLabel:      "False",
		}

		if len(p.envFiles) != 0 {
			s.CustomLabels[api.EnvironmentFileLabel] = strings.Join(p.envFiles, ",")
		}

		p.Services[name] = s
	}

	return nil
}

func mountSourceVolumes(p *Project) error {
	fullPathToSources := filepath.Join(p.WorkingDir)

	for name, service := range p.Services {
		envPrefix := fmt.Sprintf("DEVBOX_%s_", convertToEnvName(service.Name))

		for i := range service.Volumes {
			volume := &service.Volumes[i]

			if volume.Type != "bind" {
				continue
			}

			sourceName := "." + strings.TrimPrefix(volume.Source, fullPathToSources)
			if !strings.HasPrefix(sourceName, "./sources/") {
				continue
			}

			altMountPath, ok := p.LocalMounts[sourceName]
			if !ok {
				continue
			}

			volume.Source = altMountPath

			if service.Environment == nil {
				service.Environment = types.MappingWithEquals{}
			}

			value := "mounted"
			sourcePostfix := convertToEnvName(sourceName)
			service.Environment[envPrefix+sourcePostfix] = &value
		}

		p.Services[name] = service
	}

	return nil
}

func convertToEnvName(name string) string {
	var result strings.Builder
	prevWasUnderscore := false

	for _, char := range name {
		// Check if the character is valid: letter, digit, or underscore
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			result.WriteRune(unicode.ToUpper(char))
			prevWasUnderscore = false
		} else {
			// Replace invalid characters with underscores, avoiding consecutive underscores
			if !prevWasUnderscore {
				result.WriteRune('_')
				prevWasUnderscore = true
			}
		}
	}

	// Ensure the name starts with a letter or underscore
	finalName := result.String()
	if len(finalName) > 0 && !unicode.IsLetter(rune(finalName[0])) {
		finalName = "_" + finalName
	}

	return finalName
}
