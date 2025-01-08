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

func Autodetect() (string, string, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	projectNames := ListProjects("")

	projects := make([]*project.Project, 0)
	for _, projectName := range projectNames {
		project, err := project.New(context.Background(), projectName)
		if err != nil {
			continue
		}

		projects = append(projects, project)
	}

	ambiguous := false

	// if the current dir is a source of one project
	projectName, sourceName := func() (string, string) {
		foundProject := ""
		foundSource := ""

		for _, project := range projects {
			for k, v := range project.LocalMounts {
				if v == curDir {
					if foundProject != "" {
						ambiguous = true
						return "", "" // ambiguous project
					}

					foundProject = project.Name
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
					if foundProject != "" {
						ambiguous = true
						return "", "" // ambiguous project
					}

					foundProject = project.Name
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

						foundProject = project.Name
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
