package project

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/format"
	"github.com/pilat/devbox/internal/app"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

// Replacement of composer service with our state keeper. Another extended service (with client inside) will be used.
type Project struct {
	*types.Project

	Sources     SourceConfigs
	Scenarios   ScenarioConfigs
	LocalMounts map[string]string

	envFiles []string
}

func New(ctx context.Context, projectName string) (*Project, error) {
	projectFolder := filepath.Join(app.AppDir, projectName)

	if _, err := os.Stat(projectFolder); os.IsNotExist(err) {
		return nil, fmt.Errorf("project '%s' not found", projectName)
	}

	o, err := cli.NewProjectOptions(
		[]string{},
		cli.WithoutEnvironmentResolution, // the app performs Validate() later
		cli.WithWorkingDirectory(projectFolder),
		cli.WithDefaultConfigPath,
		cli.WithName(projectName),
		cli.WithInterpolation(true),
		cli.WithResolvedPaths(true),
		cli.WithExtension("x-devbox-sources", SourceConfigs{}),
		cli.WithExtension("x-devbox-scenarios", ScenarioConfigs{}),
		cli.WithExtension("x-devbox-default-stop-grace-period", Duration(0)),
		cli.WithExtension("x-devbox-volumes", AlternativeVolumes{}), // Experimental feature
		cli.WithExtension("x-devbox-init-subpath", false),
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
		setupGracePeriod,
		volumesWithSubpaths,
		initVolumes,
		applyLabels,
		remountSourceVolumes,
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
		Mounts: p.LocalMounts,
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

func (p *Project) Reload(ctx context.Context) error {
	p2, err := New(ctx, p.Name)
	if err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	*p = *p2

	return nil
}

func (p *Project) Validate() error {
	project, err := p.WithServicesEnvironmentResolved(false)
	if err != nil {
		return fmt.Errorf("failed to resolve services environment: %w", err)
	}

	p.Project = project

	return nil
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

func volumesWithSubpaths(p *Project) error {
	// extended inline volumes with subpath support
	for name, s := range p.Services {
		if e, ok := s.Extensions["x-devbox-volumes"]; ok {
			s.Volumes = []types.ServiceVolumeConfig{}
			altVolumes := e.(AlternativeVolumes) // nolint: forcetypeassert
			for _, volume := range altVolumes {
				v, err := format.ParseVolume(volume)
				if err != nil {
					return fmt.Errorf("failed to parse volume: %w", err)
				}

				if v.Target != "" {
					v.Target = path.Clean(v.Target)
				}

				if v.Type == types.VolumeTypeVolume && v.Source != "" { // non anonymous volumes
					elements := strings.Split(v.Source, "/")
					v.Source = elements[0]
					if len(elements) > 1 {
						v.Volume.Subpath = strings.Join(elements[1:], "/")
					}
				}

				v.Source = absVolumeMount(p.WorkingDir, v.Source)

				s.Volumes = append(s.Volumes, v)
			}

			p.Services[name] = s
		}
	}

	return nil
}

func initVolumes(p *Project) error {
	if _, ok := p.Extensions["x-devbox-init-subpath"]; !ok {
		return nil
	}

	// collect subpaths from volumes and services that use them
	subpaths := map[string]map[string]struct{}{}
	services := map[string]struct{}{}

	for _, s := range p.Services {
		for _, v := range s.Volumes {
			if v.Type == types.VolumeTypeVolume && v.Source != "" && v.Volume != nil && v.Volume.Subpath != "" {
				if _, ok := subpaths[v.Source]; !ok {
					subpaths[v.Source] = map[string]struct{}{}
				}

				subpaths[v.Source][v.Volume.Subpath] = struct{}{}
				services[s.Name] = struct{}{}
			}
		}
	}

	// create volumes-init service
	volumes := []types.ServiceVolumeConfig{}
	for source := range subpaths {
		volumes = append(volumes, types.ServiceVolumeConfig{
			Type:   types.VolumeTypeVolume,
			Source: source,
			Target: fmt.Sprintf("/volume/%s", source),
		})
	}

	// construct the initialization command in one execution
	initCommands := []string{}
	for source, paths := range subpaths {
		for path := range paths {
			// TODO: escape paths
			cmd := fmt.Sprintf("mkdir -p /volume/%s/%s && printf 'Created %s/%s\n'", source, path, source, path)
			initCommands = append(initCommands, cmd)
		}
	}

	// create volumes-init service
	initService := types.ServiceConfig{
		Name:    "volumes-init",
		Image:   "docker.io/library/busybox:latest",
		Volumes: volumes,
		Command: []string{"sh", "-c", strings.Join(initCommands, " && ")},
	}

	p.Services["volumes-init"] = initService

	// add volumes-init service to all services that use volumes as a dependency
	for name, s := range p.Services {
		if _, ok := services[name]; !ok {
			continue
		}

		if name == "volumes-init" {
			continue
		}

		for _, v := range s.Volumes {
			if v.Type == types.VolumeTypeVolume {
				if s.DependsOn == nil {
					s.DependsOn = map[string]types.ServiceDependency{}
				}

				s.DependsOn["volumes-init"] = types.ServiceDependency{
					Condition: types.ServiceConditionCompletedSuccessfully,
				}
				break
			}
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

func remountSourceVolumes(p *Project) error {
	fullPathToSources := filepath.Join(p.WorkingDir, app.SourcesDir)

	fullPathToSources += "/"

	for _, service := range p.Services {
		for i := range service.Volumes {
			volume := &service.Volumes[i]

			if volume.Type != "bind" || !strings.HasPrefix(volume.Source, fullPathToSources) {
				continue
			}

			sourceName := strings.TrimPrefix(volume.Source, fullPathToSources)
			sourceName = strings.Split(sourceName, "/")[0]

			altMountPath, ok := p.LocalMounts[sourceName]
			if !ok {
				continue
			}

			volume.Source = altMountPath
		}
	}

	return nil
}

// Convert a relative path to an absolute path
func absVolumeMount(workingDir, source string) string {
	prefix := ""

	switch {
	case strings.HasPrefix(source, "~"):
		if home, err := os.UserHomeDir(); err != nil {
			prefix = home
		}
	case strings.HasPrefix(source, "."):
		prefix = workingDir
	}

	return filepath.Join(prefix, source)
}
