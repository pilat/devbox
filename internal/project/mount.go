package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pilat/devbox/internal/app"
)

func (p *Project) Mount(ctx context.Context, sourceName, path string) ([]string, error) {
	if path == "" {
		curDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		path = curDir
	}

	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("failed to get path: %w", err)
	}

	if _, ok := p.LocalMounts[sourceName]; ok {
		return nil, fmt.Errorf("source %s already mounted", sourceName)
	}

	p.LocalMounts[sourceName] = path

	if err := p.SaveState(); err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	if err := p.Reload(ctx); err != nil {
		return nil, fmt.Errorf("failed to reload project: %w", err)
	}

	fullPathToSources := filepath.Join(p.WorkingDir, app.SourcesDir, sourceName)
	affectedServices := p.servicesAffectedByMounts(fullPathToSources)

	return affectedServices, nil
}

func (p *Project) servicesAffectedByMounts(path string) []string {
	affectedServices := []string{}

	for _, service := range p.Services {
		isAffected := false
		for i := range service.Volumes {
			volume := &service.Volumes[i]

			// has prefix is using because mount path can be .devbox/sources/sourceName/sub/path
			if volume.Type == "bind" && strings.HasPrefix(volume.Source, path) {
				isAffected = true
				break
			}
		}

		if isAffected {
			affectedServices = append(affectedServices, service.Name)
		}
	}

	return affectedServices
}
