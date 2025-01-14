package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/project"
)

// AutodetectSource detects the source name of the current directory in a context of a project.
func AutodetectSource(project *project.Project, onlyMounted bool) (string, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// If the current directory is mounted to some source in the project
	for k, v := range project.LocalMounts {
		if v == curDir {
			return k, nil
		}
	}

	// When the current dir is a git repo
	g := git.New(curDir)
	remoteURL, err := g.GetRemote(context.TODO())
	if err != nil {
		return "", nil
	}

	remoteURL = normalizeRemoteURL(remoteURL)

	toplevelDir, err := g.GetTopLevel(context.TODO())
	if err != nil {
		return "", nil
	}

	relativePath, err := filepath.Rel(toplevelDir, curDir)
	if err != nil {
		return "", nil
	}
	relativePath = strings.ToLower(relativePath) // a/b or "."

	foundSource := ""
	ambiguous := false

	for name, source := range project.Sources {
		if onlyMounted && project.LocalMounts[name] == "" {
			continue
		}

		sourcePath := filepath.Join(project.WorkingDir, app.SourcesDir, name)
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
			if foundSource != "" && foundSource != name {
				ambiguous = true
			}

			foundSource = name
		} else {
			// if sparse checkout is set, we need to check if the path is in the sparse checkout list
			for _, v := range source.SparseCheckout {
				if strings.ToLower(v) == relativePath {
					if foundSource != "" && foundSource != name {
						ambiguous = true
					}

					foundSource = name
				}
			}
		}
	}

	if foundSource != "" && !ambiguous {
		return foundSource, nil
	}

	if ambiguous {
		return "", fmt.Errorf("ambiguous source, please specify source name")
	}

	return "", fmt.Errorf("source not found")
}

// AutodetectProject validates project name (if provided) and tries to autodetect the project by comparing
// the current directory with the project sources and local mounts. If not successful or ambiguous, it returns an error.
func AutodetectProject(name string) (*project.Project, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	projects := make([]*project.Project, 0)

	projectNames := ListProjects("")
	for _, projectName := range projectNames {
		project, err := project.New(context.Background(), projectName)
		if err != nil {
			return nil, fmt.Errorf("failed to load project: %w", err)
		}

		projects = append(projects, project)
	}

	if name != "" {
		for _, project := range projects {
			if project.Name == name {
				return project, nil
			}
		}

		return nil, fmt.Errorf("project '%s' not found", name)
	}

	var foundProject *project.Project
	ambiguous := false

	// if the current dir is a source of one project
	for _, project := range projects {
		for _, v := range project.LocalMounts {
			if v == curDir {
				if foundProject != nil && foundProject != project {
					ambiguous = true
				}

				foundProject = project
			}
		}
	}

	if foundProject != nil && !ambiguous {
		return foundProject, nil
	}

	foundProject = nil

	// when the current dir is a git repo
	g := git.New(curDir)
	remoteURL, err := g.GetRemote(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get remote url: %w", err)
	}

	remoteURL = normalizeRemoteURL(remoteURL)

	toplevelDir, err := g.GetTopLevel(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get top level directory: %w", err)
	}

	relativePath, err := filepath.Rel(toplevelDir, curDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}
	relativePath = strings.ToLower(relativePath) // a/b or "."

	// if that path is source of one project
	for _, project := range projects {
		for name, source := range project.Sources {
			sourcePath := filepath.Join(project.WorkingDir, app.SourcesDir, name)
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
				if foundProject != nil && foundProject != project {
					ambiguous = true
				}

				foundProject = project
				continue
			}

			// if sparse checkout is set, we need to check if the path is in the sparse checkout list
			for _, v := range source.SparseCheckout {
				if strings.ToLower(v) == relativePath {
					if foundProject != nil && foundProject != project {
						ambiguous = true
					}

					foundProject = project
					break
				}
			}
		}
	}

	if foundProject != nil && !ambiguous {
		return foundProject, nil
	}

	if ambiguous {
		return nil, fmt.Errorf("ambiguous project, please specify project name")
	}

	return foundProject, fmt.Errorf("project is unknown")
}

func Init(name, url string, branch string) error {
	if !validateName(name) {
		return fmt.Errorf("invalid project name: %s", name)
	}

	projectFolder := filepath.Join(app.AppDir, name)

	if isProjectExists(projectFolder) {
		return fmt.Errorf("project already exists")
	}

	git := git.New(projectFolder)
	err := git.Clone(context.TODO(), url, branch)
	if err != nil {
		return fmt.Errorf("failed to clone git repo: %w", err)
	}

	patterns := []string{
		fmt.Sprintf("/%s/", app.SourcesDir),
		fmt.Sprintf("/%s", app.StateFile),
	}

	err = git.SetLocalExclude(patterns)
	if err != nil {
		return fmt.Errorf("failed to set local exclude: %w", err)
	}

	return nil
}

func Destroy(ctx context.Context, project *project.Project) error {
	if !isProjectExists(project.WorkingDir) {
		return fmt.Errorf("project %s does not exist", project.Name)
	}

	err := os.RemoveAll(project.WorkingDir)
	if err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}

	return nil
}

func ListProjects(filter string) []string {
	filter = strings.ToLower(filter)

	results := make([]string, 0)

	folders, err := os.ReadDir(app.AppDir)
	if err != nil {
		return []string{}
	}

	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}

		name := folder.Name()
		if !strings.HasPrefix(name, filter) {
			continue
		}

		results = append(results, folder.Name())
	}

	return results
}

func isProjectExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func validateName(name string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name)
}

func normalizeRemoteURL(s string) string { // TODO: improve it to handle cases with auth
	s = strings.ToLower(s)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "git@")
	s = strings.ReplaceAll(s, ":", "/")
	s = strings.TrimSuffix(s, ".git")
	return s
}
