package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/table"
)

func (a *app) List() error {
	projects, err := a.getProjects()
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}

	t := table.New("Name", "Message", "Author", "Date")
	for _, project := range projects {
		git := git.New(filepath.Join(a.homeDir, appFolder, project))
		info, err := git.GetInfo(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to get git info: %w", err)
		}

		t.AppendRow(project, info.Message, info.Author, info.Date)
	}

	t.Render()

	return nil
}

func (a *app) getProjects() ([]string, error) {
	dirs := make([]string, 0)

	devboxDir := filepath.Join(a.homeDir, appFolder)
	folders, err := os.ReadDir(devboxDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read .devbox directory: %w", err)
	}

	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}

		dirs = append(dirs, folder.Name())
	}

	return dirs, nil
}
