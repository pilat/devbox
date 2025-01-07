package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pilat/devbox/internal/git"
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

	apps := make([]*app, 0)
	for _, projectName := range projects {
		app, err := a.getAppByName(projectName)
		if err != nil {
			continue
		}

		apps = append(apps, app)
	}

	ambiguous := false

	// if the current dir is a source of one project
	projectName, sourceName := func() (string, string) {
		foundProject := ""
		foundSource := ""

		for _, app := range apps {
			for k, v := range app.state.Mounts {
				if v == curDir {
					if foundProject != "" {
						ambiguous = true
						return "", "" // ambiguous project
					}

					foundProject = app.project.Name
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
		s = strings.TrimPrefix(s, "https://")
		s = strings.TrimPrefix(s, "git@")
		s = strings.ReplaceAll(s, ":", "/")
		s = strings.TrimSuffix(s, ".git")
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

		for _, app := range apps {
			for name, source := range app.project.Sources {
				sourcePath := filepath.Join(app.projectPath, sourcesDir, name)
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

					foundProject = app.project.Name
					foundSource = name
					continue
				}

				// if sparse checkout is set, we need to check if the path is in the sparse checkout list
				for _, v := range source.SparseCheckout {
					if strings.ToLower(v) == relativePath {
						if foundProject != "" {
							ambiguous = true
							return "", "" // ambiguous project
						}

						foundProject = app.project.Name
						foundSource = name
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

func (a *app) getAppByName(projectName string) (*app, error) {
	if projectName == "" {
		return nil, fmt.Errorf("project name should be set")
	}

	a2 := a.Clone()
	err := a2.LoadProject(projectName)
	if err != nil {
		return nil, err
	}

	return a2, nil
}
