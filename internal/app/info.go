package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pilat/devbox/internal/pkg/git"
)

func (a *app) Info() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
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

		name := source.Name
		additionalInfo := strings.Join(source.SparseCheckout, ", ")
		if additionalInfo != "" {
			name = fmt.Sprintf("%s (%s)", name, additionalInfo)
		}

		sourcesTable.AppendRow(table.Row{
			name, info.Message, info.Author, info.Date,
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
	renderTable(sourcesTable)

	if hasMounts {
		fmt.Println("")
		fmt.Println(" Mounts:")
		renderTable(mountsTable)
	}

	return nil
}
