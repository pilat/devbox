package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/pkg/git"
	"github.com/pilat/devbox/internal/state"
)

func (a *app) autodetect() (string, string, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	projects, err := a.getProjects()
	if err != nil {
		return "", "", fmt.Errorf("failed to get projects: %w", err)
	}

	ambiguous := false

	// if the current dir is a source of one project
	projectName, sourceName := func() (string, string) {
		foundProject := ""
		foundSource := ""

		for _, project := range projects {
			stateFile := filepath.Join(a.homeDir, appFolder, project, ".devboxstate")
			state, err := state.New(stateFile)
			if err != nil {
				continue
			}

			for k, v := range state.Mounts {
				if v == curDir {
					if foundProject != "" {
						ambiguous = true
						return "", "" // ambiguous project
					}

					foundProject = project
					foundSource = k
				}
			}
		}

		return foundProject, foundSource
	}()

	if projectName != "" {
		return projectName, sourceName, nil
	}

	// when the current dir is a git repo
	g := git.New(curDir)
	remoteURL, err := g.GetRemote(context.TODO())
	if err != nil {
		return "", "", nil
	}

	normalizeRemoteURL := func(s string) string { // TODO: improve it to handle cases with auth
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "https://", "")
		s = strings.ReplaceAll(s, "git@", "")
		s = strings.ReplaceAll(s, ":", "/")
		return s
	}
	remoteURL = normalizeRemoteURL(remoteURL)

	toplevelDir, err := g.GetTopLevel(context.TODO())
	if err != nil {
		return "", "", nil
	}

	relativePath, err := filepath.Rel(toplevelDir, curDir)
	if err != nil {
		return "", "", nil
	}
	relativePath = strings.ToLower(relativePath)

	// if that path is source of one project
	projectName, sourceName = func() (string, string) {
		foundProject := ""
		foundSource := ""

		for _, project := range projects {
			manifestFile := filepath.Join(a.homeDir, appFolder, project, "devbox.yaml")
			cfg, err := config.New(manifestFile)
			if err != nil {
				continue
			}

			for _, source := range cfg.Sources {
				sourcePath := filepath.Join(a.homeDir, appFolder, project, sourcesDir, source.Name)
				g := git.New(sourcePath)
				remoteURLCurrent, err := g.GetRemote(context.TODO())
				if err != nil {
					continue
				}

				remoteURLCurrent = normalizeRemoteURL(remoteURLCurrent)
				if remoteURL != remoteURLCurrent {
					continue
				}

				if len(source.SparseCheckout) == 0 {
					if foundProject != "" {
						ambiguous = true
						return "", "" // ambiguous project
					}

					foundProject = project
					foundSource = source.Name
					continue
				}

				// if sparse checkout is set, we need to check if the path is in the sparse checkout list
				for _, v := range source.SparseCheckout {
					if strings.ToLower(v) == relativePath {
						if foundProject != "" {
							ambiguous = true
							return "", "" // ambiguous project
						}

						foundProject = project
						foundSource = source.Name
						break
					}
				}
			}
		}

		return foundProject, foundSource
	}()

	if projectName != "" {
		return projectName, sourceName, nil
	}

	if ambiguous {
		return "", "", fmt.Errorf("ambiguous project, please specify project name")
	}

	return "", "", fmt.Errorf("project not found")
}
