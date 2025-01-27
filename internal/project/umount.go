package project

import (
	"context"
	"fmt"
)

func (p *Project) Umount(ctx context.Context, source string) error {
	delete(p.LocalMounts, source)

	if err := p.SaveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	if err := p.Reload(ctx, []string{"*"}); err != nil {
		return fmt.Errorf("failed to reload project: %w", err)
	}

	return nil
}
