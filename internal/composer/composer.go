package composer

import (
	"context"
	"fmt"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/docker/compose/v2/pkg/api"
)

func New(ctx context.Context, projectPath, name string) (*Project, error) {
	o, err := cli.NewProjectOptions(
		[]string{},
		cli.WithWorkingDirectory(projectPath),
		cli.WithDefaultConfigPath,
		cli.WithName(name),
		cli.WithInterpolation(true),
		cli.WithResolvedPaths(true),
		cli.WithExtension("x-devbox-sources", SourceConfigs{}),
		cli.WithExtension("x-devbox-default-stop-grace-period", Duration(0)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose project options: %w", err)
	}

	project, err := cli.ProjectFromOptions(ctx, o)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose project: %w", err)
	}

	for name, s := range project.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  project.WorkingDir,
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False",
		}
		if len(o.EnvFiles) != 0 {
			s.CustomLabels[api.EnvironmentFileLabel] = strings.Join(o.EnvFiles, ",")
		}
		project.Services[name] = s
	}

	p := &Project{
		Project: project,
	}

	if s, ok := project.Extensions["x-devbox-sources"]; ok {
		p.Sources = s.(SourceConfigs)
	}

	if s, ok := project.Extensions["x-devbox-default-stop-grace-period"]; ok {
		v := s.(Duration)
		p.DefaultStopGracePeriod = &v
	}

	// apply default grace period to all services
	for name, s := range project.Services {
		if s.StopGracePeriod != nil {
			continue
		}

		if p.DefaultStopGracePeriod != nil {
			s.StopGracePeriod = p.DefaultStopGracePeriod
		}

		project.Services[name] = s
	}

	return p, nil
}
