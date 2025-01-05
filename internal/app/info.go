package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/table"
)

func (a *app) Info() error {
	if a.projectPath == "" {
		return ErrProjectIsNotSet
	}

	hasMounts := false
	sourcesTable := table.New("Name", "Message", "Author", "Date")
	sourcesTable.SortBy([]table.SortBy{
		{Name: "Message", Mode: table.Asc},
		{Name: "Name", Mode: table.Asc},
	})

	mountsTable := table.New("Name", "Local path")
	for name, source := range a.sources {
		repoPath := filepath.Join(a.projectPath, sourcesDir, name)
		git := git.New(repoPath)
		info, err := git.GetInfo(context.TODO())
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

		if localPath, ok := a.state.Mounts[name]; ok {
			hasMounts = true
			mountsTable.AppendRow(name, localPath)
		}
	}

	fmt.Println(" Sources:")
	sourcesTable.Render()

	if hasMounts {
		fmt.Println("")
		fmt.Println(" Mounts:")
		mountsTable.Render()
	}

	return nil
}
