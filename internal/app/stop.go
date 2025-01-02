package app

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/pkg/depgraph"
	"github.com/pilat/devbox/internal/runners"
)

func (a *app) Stop() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	d, err := docker.New()
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
	defer stopProgress(pw)

	trackersMap := make(map[string]*progress.Tracker)
	for _, round := range plan {
		for _, step := range round {
			t := addTracker(pw, step.Ref(), false)
			trackersMap[step.Ref()] = t
		}
	}

	ctx := context.Background()
	err = depgraph.ExecReverse(ctx, plan, func(ctx context.Context, r runners.Runner) error {
		t := trackersMap[r.Ref()]

		t.Start()
		err := r.Stop(ctx)

		if err != nil {
			t.MarkAsErrored()
		} else {
			t.MarkAsDone()
		}

		return err
	})
	if err != nil {
		return fmt.Errorf("failed to execute steps: %v", err)
	}

	return nil
}
