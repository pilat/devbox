package service

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"

	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
)

type Service struct {
	service   api.Service
	apiClient client.APIClient
}

func New() (*Service, error) {
	dockerCLI, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	cliOpts := flags.NewClientOptions()
	if err = dockerCLI.Initialize(cliOpts); err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}

	return &Service{
		service:   compose.NewComposeService(dockerCLI),
		apiClient: dockerCLI.Client(),
	}, nil
}
