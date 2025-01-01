package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pilat/devbox/internal/git"
)

func (c *app) List() error {
	devboxDir := filepath.Join(c.homeDir, appFolder)
	folders, err := os.ReadDir(devboxDir)
	if err != nil {
		return fmt.Errorf("failed to read .devbox directory: %w", err)
	}

	t := makeTable("Name", "Message", "Author", "Date")
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}

		git := git.New(filepath.Join(devboxDir, folder.Name()))
		info, err := git.GetInfo(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to get git info: %w", err)
		}

		t.AppendRow(table.Row{
			folder.Name(), info.Message, info.Author, info.Date,
		})
	}

	renderTable(t, 20, 50, 20, 30)

	return nil
}
