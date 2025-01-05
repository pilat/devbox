package composer

import (
	"context"
	"fmt"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

func Load(ctx context.Context, projectPath, name string) (*types.Project, error) {
	o, err := cli.NewProjectOptions(
		[]string{},
		cli.WithWorkingDirectory(projectPath),
		cli.WithDefaultConfigPath,
		cli.WithName(name),
		cli.WithInterpolation(true),
		cli.WithResolvedPaths(true),
		cli.WithExtension("x-devbox-sources", SourceConfigs{}),
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

	return project, nil
}
