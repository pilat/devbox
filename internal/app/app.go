package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/pilat/devbox/internal/composer"
	"github.com/pilat/devbox/internal/state"
	"github.com/pilat/devbox/internal/sys"
)

const (
	appFolder  = ".devbox"
	sourcesDir = "sources"
)

var (
	ErrProjectIsNotSet = fmt.Errorf("project is not set")
)

type app struct { // TODO: rename to app
	homeDir string

	// only for project
	project *types.Project
	sources composer.SourceConfigs

	projectName string
	projectPath string
	state       *state.State
}

func New() (*app, error) {
	homeDir, err := sys.GetHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	return &app{
		homeDir: homeDir,
	}, nil
}

func (a *app) Clone() *app {
	return &app{
		homeDir: a.homeDir,
	}
}

func (a *app) WithProject(name string) error {
	if a.projectName != "" {
		panic("project already set")
	}

	if name == "" {
		projectName, _, err := a.autodetect()
		if err != nil {
			return fmt.Errorf("failed to autodetect project name: %w", err)
		} else {
			name = projectName
		}
	}

	a.projectName = name
	a.projectPath = filepath.Join(a.homeDir, appFolder, name)

	return nil
}

func (a *app) LoadProject() error {
	if !a.isProjectExists() {
		return fmt.Errorf("failed to get project path")
	}

	project, err := composer.Load(context.Background(), a.projectPath, a.projectName)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}
	a.project = project

	if s, ok := a.project.Extensions["x-devbox-sources"]; ok {
		coercedSources, ok := s.(composer.SourceConfigs)
		if !ok {
			return fmt.Errorf("failed to coerce sources")
		}

		a.sources = coercedSources
	}

	stateFile := filepath.Join(a.projectPath, ".devboxstate")
	state, err := state.New(stateFile)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}
	a.state = state

	// Replace "mounted" sources
	fullPathToSources := filepath.Join(a.projectPath, sourcesDir) + "/"
	for _, service := range a.project.Services {
		for i := range service.Volumes {
			volume := &service.Volumes[i]

			if volume.Type != "bind" || !strings.HasPrefix(volume.Source, fullPathToSources) {
				continue
			}

			sourceName := strings.TrimPrefix(volume.Source, fullPathToSources)
			sourceName = strings.Split(sourceName, "/")[0]

			altMountPath, ok := a.state.Mounts[sourceName]
			if !ok {
				continue
			}

			volume.Source = altMountPath
		}
	}

	return nil
}

func (a *app) servicesAffectedByMounts(path string) []string {
	affectedServices := []string{}

	for _, service := range a.project.Services {
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

func (a *app) isProjectExists() bool {
	_, err := os.Stat(a.projectPath)
	return err == nil
}
