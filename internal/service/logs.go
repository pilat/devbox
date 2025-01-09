package service

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/pilat/devbox/internal/project"
)

func (a *Service) Logs(ctx context.Context, p *project.Project, services []string) error {
	opts := project.LogOptions{
		Project:  p.Project,
		Services: services,
		Tail:     "500",
		Follow:   true,
	}

	outStream := streams.NewOut(os.Stdout)
	errStream := streams.NewOut(os.Stderr)

	consumer := formatter.NewLogConsumer(ctx, outStream, errStream, true, true, false)
	if err := a.service.Logs(ctx, p.Name, consumer, opts); err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}
