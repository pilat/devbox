package composer

import (
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"

	"github.com/docker/compose/v2/pkg/compose"

	"github.com/docker/compose/v2/pkg/api"
)

func getClient() (api.Service, error) {
	cmd, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	cliOpts := flags.NewClientOptions()
	if err = cmd.Initialize(cliOpts); err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}

	return compose.NewComposeService(cmd), nil
}
