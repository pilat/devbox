package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/pkg/utils"
	"github.com/pilat/devbox/internal/state"
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
	projectName string
	projectPath string
	cfg         *config.Config
	state       *state.State
}

func New() (*app, error) {
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	return &app{
		homeDir: homeDir,
	}, nil
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

	configFile := filepath.Join(a.projectPath, "devbox.yaml")
	cfg, err := config.New(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	a.cfg = cfg
	a.cfg.Name = a.projectName
	a.cfg.NetworkName = fmt.Sprintf("devbox-%s", a.cfg.Name)

	stateFile := filepath.Join(a.projectPath, ".devboxstate")
	state, err := state.New(stateFile)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}
	a.state = state

	return nil
}

func (a *app) UpdateSources() error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	pw := createProgress()

	var errCh = make(chan error, len(a.cfg.Sources))
	for _, src := range a.cfg.Sources {
		t := addTracker(pw, " Syncing "+src.Name, true)
		t.Start()
		go func(src config.SourceConfig) {
			targetPath := filepath.Join(a.projectPath, sourcesDir, src.Name)

			git := git.New(targetPath)
			err := git.Sync(ctx, src.URL, src.Branch, src.SparseCheckout)

			if err != nil {
				t.MarkAsErrored()
			} else {
				t.MarkAsDone()
			}

			errCh <- err
		}(src)
	}

	for i := 0; i < len(a.cfg.Sources); i++ {
		if err := <-errCh; err != nil {
			return fmt.Errorf("failed to sync source: %w", err)
		}
	}

	stopProgress(pw)

	return nil
}

func (a *app) isProjectExists() bool {
	_, err := os.Stat(a.projectPath)
	return err == nil
}
