package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/compose/v2/pkg/progress"
	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update devbox project sources",
		Long:  "That command will update sources in devbox project",
		Args:  cobra.MinimumNArgs(0),
		RunE: runWrapperWithProject(func(ctx context.Context, p *project.Project, cmd *cobra.Command, args []string) error {
			if err := runProjectUpdate(ctx, p); err != nil {
				return fmt.Errorf("failed to update project: %w", err)
			}

			if err := runSourcesUpdate(ctx, p); err != nil {
				return fmt.Errorf("failed to update sources: %w", err)
			}

			if err := runInfo(ctx, p); err != nil {
				return fmt.Errorf("failed to get project info: %w", err)
			}

			return nil
		}),
	}

	cmd.ValidArgsFunction = validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, p *project.Project, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	})

	root.AddCommand(cmd)
}

func runProjectUpdate(ctx context.Context, p *project.Project) error {
	ggg := git.New(p.WorkingDir)

	fmt.Println("[*] Updating project...")

	err := ggg.Pull(ctx) // TODO: consider using git.Sync() to reset it every time
	if err != nil {
		return fmt.Errorf("failed to pull git repo: %w", err)
	}

	_, err = ggg.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get git info: %w", err)
	}

	if err := p.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	return nil
}

func runSourcesUpdate(ctx context.Context, p *project.Project) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	fmt.Println("[*] Updating sources...")

	updateSources := func(ctx context.Context) error {
		cw := progress.ContextWriter(ctx)
		for name := range p.Sources {
			cw.Event(progress.Event{
				ID:         "Source " + name,
				StatusText: "Syncing",
			})
		}

		var errCh = make(chan error, len(p.Sources))
		for name, src := range p.Sources {
			go func(name string, src project.SourceConfig) {
				repoDir := filepath.Join(p.WorkingDir, app.SourcesDir, name)

				git := git.New(repoDir)
				err := git.Sync(ctx, src.URL, src.Branch, src.SparseCheckout)

				cw.Event(progress.Event{
					ID:         "Source " + name,
					StatusText: "Synced",
					Status:     progress.Done,
				})

				errCh <- err
			}(name, src)
		}

		for i := 0; i < len(p.Sources); i++ {
			if err := <-errCh; err != nil {
				return fmt.Errorf("failed to sync source: %w", err)
			}
		}

		return nil
	}

	out := streams.NewOut(os.Stdout)
	if err := progress.RunWithTitle(ctx, updateSources, out, "Updating sources"); err != nil {
		return fmt.Errorf("failed to update sources: %w", err)
	}

	fmt.Println("")

	return nil
}
