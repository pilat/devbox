package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/compose/v2/pkg/progress"
	"github.com/pilat/devbox/internal/composer"
	"github.com/pilat/devbox/internal/git"
)

func (a *app) UpdateProject() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	if !isProjectExists(a.projectPath) {
		return fmt.Errorf("failed to get project path")
	}

	fmt.Println("Updating project...")
	git := git.New(a.projectPath)

	err := git.Pull(context.TODO()) // TODO: consider using git.Sync() to reset it every time
	if err != nil {
		return fmt.Errorf("failed to pull git repo: %w", err)
	}

	_, err = git.GetInfo(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get git info: %w", err)
	}

	if err := a.LoadProject(a.project.Name); err != nil {
		return err
	}

	fmt.Println("")

	return nil
}

func (a *app) UpdateSources() error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute) // TODO
	defer cancel()

	fmt.Println("Updating sources...")

	out := streams.NewOut(os.Stdout)
	if err := progress.RunWithTitle(ctx, a.updateSources, out, "Updating sources"); err != nil {
		return fmt.Errorf("failed to update sources: %w", err)
	}

	fmt.Println("")

	return nil
}

func (a *app) updateSources(ctx context.Context) error {
	cw := progress.ContextWriter(ctx)
	for name := range a.project.Sources {
		cw.Event(progress.Event{
			ID:         "Source " + name,
			StatusText: "Syncing",
		})
	}

	var errCh = make(chan error, len(a.project.Sources))
	for name, src := range a.project.Sources {
		go func(name string, src composer.SourceConfig) {
			targetPath := filepath.Join(a.projectPath, sourcesDir, name)

			git := git.New(targetPath)
			err := git.Sync(ctx, src.URL, src.Branch, src.SparseCheckout)

			cw.Event(progress.Event{
				ID:         "Source " + name,
				StatusText: "Synced",
				Status:     progress.Done,
			})

			errCh <- err
		}(name, src)
	}

	for i := 0; i < len(a.project.Sources); i++ {
		if err := <-errCh; err != nil {
			return fmt.Errorf("failed to sync source: %w", err)
		}
	}

	return nil
}
