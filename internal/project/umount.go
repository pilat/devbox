package project

import (
	"context"
	"fmt"
)

func (p *Project) Umount(ctx context.Context, sourceName string) ([]string, error) {
	curPath, ok := p.LocalMounts[sourceName]
	if !ok {
		return nil, fmt.Errorf("source %s is not mounted", sourceName)
	}

	affectedServices := p.servicesAffectedByMounts(curPath)

	delete(p.LocalMounts, sourceName)

	if err := p.SaveState(); err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	if err := p.Reload(ctx); err != nil {
		return nil, fmt.Errorf("failed to reload project: %w", err)
	}

	return affectedServices, nil
}
