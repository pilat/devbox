package project

import (
	"context"
	"fmt"
)

func (p *Project) Umount(ctx context.Context, sources []string) error {
	for _, sourceName := range sources {
		delete(p.LocalMounts, sourceName)
	}

	if err := p.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	if err := p.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	return nil
}
