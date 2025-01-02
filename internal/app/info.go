package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pilat/devbox/internal/git"
)

func (a *app) Info() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	err := a.update()
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	hasMounts := false
	sourcesTable := makeTable("Name", "Message", "Author", "Date")
	mountsTable := makeTable("Name", "Local path")
	for _, source := range a.cfg.Sources {
		repoPath := filepath.Join(a.projectPath, sourcesDir, source.Name)
		git := git.New(repoPath)
		info, err := git.GetInfo(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to get git info for %s: %w", source.Name, err)
		}

		sourcesTable.AppendRow(table.Row{
			source.Name, info.Message, info.Author, info.Date,
		})

		if localPath, ok := a.state.Mounts[source.Name]; ok {
			hasMounts = true
			mountsTable.AppendRow(table.Row{
				source.Name, localPath,
			})
		}
	}

	fmt.Println("")
	fmt.Println(" Sources:")
	renderTable(sourcesTable, 20, 50, 20, 30)

	if hasMounts {
		fmt.Println("")
		fmt.Println(" Mounts:")
		renderTable(mountsTable, 20, 106)
	}

	return nil
}
