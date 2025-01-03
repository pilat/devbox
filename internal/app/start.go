package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/pilat/devbox/internal/pkg/container"
	"github.com/pilat/devbox/internal/pkg/depgraph"
	"github.com/pilat/devbox/internal/runners"
)

func (a *app) Start() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	d, err := container.New()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer d.Close()

	err = d.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("failed to ping docker: %w", err)
	}

	plan, err := a.getPlan(d)
	if err != nil {
		return fmt.Errorf("failed to get plan: %w", err)
	}

	pw := createProgress()

	trackersMap := make(map[string]*progress.Tracker)
	for _, round := range plan {
		for _, step := range round {
			roc := slices.Contains([]runners.ServiceType{
				runners.TypeVolume,
				runners.TypeNetwork,
				runners.TypePull,
				runners.TypeImage,
			}, step.Type())

			t := addTracker(pw, step.Ref(), roc)
			trackersMap[step.Ref()] = t
		}
	}

	ctx := context.Background()
	err = depgraph.Exec(ctx, plan, func(ctx context.Context, r runners.Runner) error {
		t := trackersMap[r.Ref()]

		t.Start()
		err := r.Start(ctx)

		if err != nil {
			t.MarkAsErrored()
		} else {
			t.MarkAsDone()
		}

		return err
	})

	stopProgress(pw)

	if err != nil {
		return fmt.Errorf("failed to execute steps: %w", err)
	}

	return nil
}
