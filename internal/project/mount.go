package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func (p *Project) Mount(ctx context.Context, sourceName, path string) ([]string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		path = filepath.Join(homeDir, path[1:])
	}

	if !filepath.IsAbs(path) {
		curDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		path = filepath.Join(curDir, path)
	}

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

	affectedServices := p.servicesAffectedByMounts(path)

	return affectedServices, nil
}

func (p *Project) GetLocalMountCandidates(filter string) []string {
	filter = strings.ToLower(filter)

	localMounts := p.GetLocalMounts("")

	results := []string{}
	for k := range p.Sources {
		if slices.Contains(localMounts, k) {
			continue
		}

		if filter == "" || strings.Contains(strings.ToLower(k), filter) {
			results = append(results, k)
		}
	}

	return results
}

func (p *Project) GetLocalMounts(filter string) []string {
	filter = strings.ToLower(filter)

	results := []string{}
	for k := range p.LocalMounts {
		if filter == "" || strings.Contains(strings.ToLower(k), filter) {
			results = append(results, k)
		}
	}

	return results
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
