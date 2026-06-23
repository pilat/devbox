package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/spf13/cobra"

	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/project"
)

func init() {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update devbox project sources",
		Long:  "That command will update sources in devbox project",
		Args:  cobra.MinimumNArgs(0),
		ValidArgsFunction: validArgsWrapper(
			func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			},
		),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			// We are attempting to update the project by its name before trying autodetection,
			// as autodetection may fail if the project manifest is damaged.
			updated := runEmergencyProjectUpdate(ctx, projectName)

			p, err := mgr.AutodetectProject(ctx, projectName)
			if err != nil {
				return fmt.Errorf("failed to autodetect project: %w", err)
			}

			if !updated {
				if err := runProjectUpdate(ctx, p); err != nil {
					return fmt.Errorf("failed to update project: %w", err)
				}
			}

			if err := runHostsUpdate(p, true, false); err != nil {
				return fmt.Errorf("failed to update hosts file: %w", err)
			}

			if err := runCertUpdate(p, true); err != nil {
				return fmt.Errorf("failed to update certificates: %w", err)
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

	root.AddCommand(cmd)
}

func runEmergencyProjectUpdate(ctx context.Context, projectName string) bool {
	if projectName == "" {
		return false
	}

	workingDir := filepath.Join(app.AppDir, projectName)
	if _, err := os.Stat(workingDir); err != nil {
		return false
	}

	fakeProject := &project.Project{
		Project: &types.Project{},
	}
	fakeProject.WorkingDir = workingDir
	fakeProject.Name = projectName

	return runProjectUpdate(ctx, fakeProject) == nil
}

func runProjectUpdate(ctx context.Context, p *project.Project) error {
	g := git.New(p.WorkingDir)

	fmt.Println("[*] Updating project...")

	err := g.Pull(ctx) // TODO: consider using git.Sync() to reset it every time
	if err != nil {
		return fmt.Errorf("failed to pull git repo: %w", err)
	}

	_, err = g.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get git info: %w", err)
	}

	if err := p.Reload(ctx, []string{"*"}); err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	return nil
}

func runSourcesUpdate(ctx context.Context, p *project.Project) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	fmt.Println("[*] Updating sources...")

	bus := newProgressBus()
	bus.Start(ctx, "sources")

	err := syncSources(ctx, p,
		func(name string) { bus.On(api.Resource{ID: name, Status: api.Working, Text: "Syncing"}) },
		func(name string) { bus.On(api.Resource{ID: name, Status: api.Done, Text: "Synced"}) },
		func(name string) { bus.On(api.Resource{ID: name, Status: api.Error, Text: "Failed"}) },
	)

	bus.Done("sources", err == nil)

	return err
}

func syncSources(ctx context.Context, p *project.Project, onSyncing, onSynced, onFailed func(name string)) error {
	ctx, cancelSync := context.WithCancel(ctx)
	defer cancelSync()

	const maxConcurrency = 4
	sem := make(chan struct{}, maxConcurrency)

	errCh := make(chan error, len(p.Sources))
	for name, src := range p.Sources {
		go func(name string, src project.SourceConfig) {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}

			onSyncing(name)

			repoDir := filepath.Join(p.WorkingDir, app.SourcesDir, name)

			g := git.New(repoDir)
			err := g.Sync(ctx, src.URL, src.Branch, src.SparseCheckout)
			if err != nil {
				onFailed(name)
				cancelSync()
			} else {
				onSynced(name)
			}

			errCh <- err
		}(name, src)
	}

	var firstErr error
	for range p.Sources {
		if err := <-errCh; err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return fmt.Errorf("failed to sync source: %w", firstErr)
	}

	return nil
}
