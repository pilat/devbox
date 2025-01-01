package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/pkg/utils"
)

func (c *cli) List() error {
	c.log.Debug("Get home dir")
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	devboxDir := filepath.Join(homeDir, ".devbox")
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
		info, err := git.GetInfo()

		if err != nil {
			c.log.Error("Failed to get git info", "folder", folder.Name(), "error", err)
			continue
		}

		t.AppendRow(table.Row{
			folder.Name(), info.Message, info.Author, info.Date,
		})
	}

	renderTable(t, 20, 50, 20, 30)

	return nil
}
