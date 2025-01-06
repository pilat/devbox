package composer

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/format"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

func New(ctx context.Context, projectPath, name string) (*Project, error) {
	o, err := cli.NewProjectOptions(
		[]string{},
		cli.WithoutEnvironmentResolution, // the app performs Validate() later
		cli.WithWorkingDirectory(projectPath),
		cli.WithDefaultConfigPath,
		cli.WithName(name),
		cli.WithInterpolation(true),
		cli.WithResolvedPaths(true),
		cli.WithExtension("x-devbox-sources", SourceConfigs{}),
		cli.WithExtension("x-devbox-default-stop-grace-period", Duration(0)),
		cli.WithExtension("x-devbox-volumes", AlternativeVolumes{}),
		cli.WithExtension("x-devbox-init-subpath", false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose project options: %w", err)
	}

	project, err := cli.ProjectFromOptions(ctx, o)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose project: %w", err)
	}

	p := &Project{
		Project:  project,
		envFiles: o.EnvFiles,
	}

	if s, ok := project.Extensions["x-devbox-sources"]; ok {
		p.Sources = s.(SourceConfigs)
	}

	allFuncs := []func() error{
		p.setupGracePeriod,
		p.volumesWithSubpaths,
		p.initVolumes,
		p.applyLabels,
	}

	for _, f := range allFuncs {
		if err := f(); err != nil {
			return nil, fmt.Errorf("failed to process project: %w", err)
		}
	}

	return p, nil
}

func (p *Project) Validate() error {
	project, err := p.WithServicesEnvironmentResolved(false)
	if err != nil {
		return fmt.Errorf("failed to resolve services environment: %w", err)
	}

	p.Project = project

	return nil
}

func (p *Project) setupGracePeriod() error {
	var defaultStopGracePeriod *Duration

	if s, ok := p.Extensions["x-devbox-default-stop-grace-period"]; ok {
		v := s.(Duration)
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

func (p *Project) volumesWithSubpaths() error {
	// extended inline volumes with subpath support
	for name, s := range p.Services {
		if e, ok := s.Extensions["x-devbox-volumes"]; ok {
			s.Volumes = []types.ServiceVolumeConfig{}
			for _, volume := range e.(AlternativeVolumes) {
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

func (p *Project) initVolumes() error {
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

	// create post-start commands
	postStart := []types.ServiceHook{}
	for source, paths := range subpaths {
		for p := range paths {
			postStart = append(postStart, types.ServiceHook{
				// TODO: validate names
				Command: []string{"sh", "-c", fmt.Sprintf("mkdir -p /volume/%s/%s || true", source, p)},
			})
		}
	}

	postStart = append(postStart, types.ServiceHook{
		Command: []string{"touch", "/tmp/ok"},
	})

	// create volumes-init service
	initService := types.ServiceConfig{
		Name:      "volumes-init",
		Image:     "docker.io/library/busybox:latest",
		Volumes:   volumes,
		Command:   []string{"sh", "-c", "while [ ! -f /tmp/ok ]; do sleep 0.1; done; exit 0"},
		PostStart: postStart,
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

func (p *Project) applyLabels() error {
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
