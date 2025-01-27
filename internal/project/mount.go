package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func (p *Project) Mount(ctx context.Context, source string, path string) error {
	path = p.absPath(path)

	if !filepath.IsAbs(path) {
		curDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		path = filepath.Join(curDir, path)
	}

	if path == "" {
		curDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		path = curDir
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("failed to get path: %w", err)
	}

	p.LocalMounts[source] = path

	if err := p.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	if err := p.Reload(ctx, []string{"*"}); err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	return nil
}
