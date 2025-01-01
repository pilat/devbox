package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/pkg/utils"
)

func (c *cli) Info(name string) error {
	c.log.Debug("Update project", "name", name)
	err := c.update(name)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	c.log.Debug("Get home dir")
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	targetPath := filepath.Join(homeDir, appFolder, name)

	c.log.Debug("Reading configuration", "target", targetPath)
	cfg, err := config.New(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	cfg.Name = name

	// Load state to read mounts
	sourcesTable := makeTable("Name", "Message", "Author", "Date")
	mountsTable := makeTable("Name", "Local path")
	for _, source := range cfg.Sources {
		c.log.Debug("Get git info", "folder", source.Name)
		git := git.New(filepath.Join(homeDir, appFolder, name, "sources", source.Name))
		info, err := git.GetInfo()
		if err != nil {
			c.log.Error("Failed to get git info", "folder", source.Name, "error", err)
			continue
		}

		sourcesTable.AppendRow(table.Row{
			source.Name, info.Message, info.Author, info.Date,
		})

		if localPath, ok := cfg.State.Mounts[source.Name]; ok {
			mountsTable.AppendRow(table.Row{
				source.Name, localPath,
			})
			continue
		}
	}

	fmt.Println("")
	fmt.Println(" Sources:")
	renderTable(sourcesTable, 20, 50, 20, 30)

	fmt.Println("")
	fmt.Println(" Mounts:")
	renderTable(mountsTable, 20, 106)

	return nil
}

func makeTable(fields ...string) table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = true

	rows := make(table.Row, len(fields))
	for i, f := range fields {
		rows[i] = f
	}

	t.AppendHeader(rows)

	return t
}

func renderTable(t table.Writer, width ...int) {
	t.SortBy([]table.SortBy{
		{Name: "Name", Mode: table.Asc},
	})

	configs := make([]table.ColumnConfig, len(width))
	for i, w := range width {
		configs[i] = table.ColumnConfig{
			Number:      i + 1,
			AlignHeader: text.AlignCenter,
			AutoMerge:   true,
			WidthMin:    w,
			WidthMax:    w,
		}
	}
	t.SetColumnConfigs(configs)
	t.Render()
}
