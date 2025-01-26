package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/pilat/devbox/internal/table"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Info devbox projects",
		Long:  "That command returns an info about a particular devbox project",
		Args:  cobra.MinimumNArgs(0),
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if err := runInfo(ctx, p); err != nil {
				return fmt.Errorf("failed to get project info: %w", err)
			}

			return nil
		}),
	}

	root.AddCommand(cmd)
}

func runInfo(ctx context.Context, p *project.Project) error {
	sourcesTable := table.New("Name", "Message", "Author", "Date")
	sourcesTable.SortBy([]table.SortBy{
		{Name: "Message", Mode: table.Asc},
		{Name: "Name", Mode: table.Asc},
	})

	for name, source := range p.Sources {
		repoDir := filepath.Join(p.WorkingDir, app.SourcesDir, name)

		g := git.New(repoDir)
		info, err := g.GetInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get git info for %s: %w", name, err)
		}

		name := name
		nameToDisplay := name
		additionalInfo := strings.Join(source.SparseCheckout, ", ")
		if additionalInfo != "" {
			nameToDisplay = fmt.Sprintf("%s (%s)", nameToDisplay, additionalInfo)
		}

		sourcesTable.AppendRow(nameToDisplay, info.Message, info.Author, info.Date)
	}

	mountsTable := table.New("Mount path", "Local path")
	for name, localPath := range p.LocalMounts {
		mountsTable.AppendRow(name, localPath)
	}

	fmt.Println("Current project:", p.Name)

	if len(p.Sources) > 0 {
		fmt.Println("")
		fmt.Println(" Sources:")
		sourcesTable.Render()
	}

	if len(p.LocalMounts) > 0 {
		fmt.Println("")
		fmt.Println(" Mounts:")
		mountsTable.Render()
	}

	if len(p.Sources) == 0 && len(p.LocalMounts) == 0 {
		fmt.Println("Project has no services or mounts")
	}

	return nil
}
