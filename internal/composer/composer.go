package composer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

func Load(ctx context.Context, projectPath, name string) (*types.Project, error) {
	dockerComposeFile := filepath.Join(projectPath, "docker-compose.yaml")
	if _, err := os.Stat(dockerComposeFile); err != nil {
		return nil, fmt.Errorf("failed to find Compose file %s: %w", dockerComposeFile, err)
	}

	opts, err := cli.NewProjectOptions(
		[]string{dockerComposeFile},
		cli.WithName(name),
		cli.WithInterpolation(true),
		cli.WithResolvedPaths(true),
		cli.WithWorkingDirectory(projectPath),
		cli.WithExtension("x-devbox-sources", SourceConfigs{}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose project options: %w", err)
	}

	project, err := cli.ProjectFromOptions(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose project: %w", err)
	}

	for i, service := range project.Services {
		service.CustomLabels = map[string]string{
			api.ProjectLabel:    project.Name,
			api.ServiceLabel:    service.Name,
			api.WorkingDirLabel: project.WorkingDir,
			api.OneoffLabel:     "False",
		}
		project.Services[i] = service
	}

	return project, nil
}

func getServiceName(name string, project *types.Project) string {
	name = strings.TrimPrefix(name, project.Name+api.Separator)

	if rIdx := strings.LastIndex(name, "-"); rIdx != -1 {
		name = name[:rIdx]
	}

	return name
}
