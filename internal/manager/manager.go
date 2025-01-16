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

type AutodetectSourceType int

const (
	AutodetectSourceForMount AutodetectSourceType = iota
	AutodetectSourceForUmount
)

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

func AutodetectSource(project *project.Project, sourceNameSel string, purpose AutodetectSourceType) ([]string, []string, error) {
	if sourceNameSel != "" {
		if !strings.HasPrefix(sourceNameSel, "./sources/") {
			return nil, nil, fmt.Errorf("source '%s' not found", sourceNameSel)
		}

		sourcePath := filepath.Join(project.WorkingDir, sourceNameSel)
		if fstat, err := os.Stat(sourcePath); err != nil || !fstat.IsDir() {
			return nil, nil, fmt.Errorf("source '%s' not found", sourceNameSel)
		}

		if purpose == AutodetectSourceForUmount {
			alt, ok := project.LocalMounts[sourceNameSel]
			if !ok {
				return nil, nil, fmt.Errorf("source '%s' not found", sourceNameSel)
			}

			sourcePath = alt
		}

		// detect affected services
		affectedServices := make(map[string]bool)
		for _, service := range project.Services {
			for _, volume := range service.Volumes {
				if volume.Source != sourcePath {
					continue
				}

				affectedServices[service.Name] = true
			}
		}

		if len(affectedServices) == 0 {
			return nil, nil, fmt.Errorf("no services found using the detected source")
		}

		return []string{sourceNameSel}, toList(affectedServices), nil
	}

	curDir, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get the top-level directory of the current Git repository
	curGit := git.New(curDir)
	toplevelDir, err := curGit.GetTopLevel(context.TODO())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get git top-level directory: %w", err)
	}

	remoteURL, err := curGit.GetRemote(context.TODO())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get remote url: %w", err)
	}
	remoteURL = normalizeRemoteURL(remoteURL)

	// Get the relative path between the Git top-level directory and the current directory
	relativePath, err := filepath.Rel(toplevelDir, curDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get relative path: %w", err)
	}
	relativePath = strings.ToLower(relativePath) // Normalize the relative path

	// Determine if the Git top-level directory matches any project source
	sources := make(map[string]bool)
	affectedServices := make(map[string]bool)

	for sourceName := range project.Sources {
		sourcePath := filepath.Join(project.WorkingDir, app.SourcesDir, sourceName)
		expectedPath := filepath.Join(sourcePath, relativePath)

		relSourcePath, _ := filepath.Rel(project.WorkingDir, expectedPath)
		relSourcePath = "./" + relSourcePath

		if purpose == AutodetectSourceForUmount {
			alt, ok := project.LocalMounts[relSourcePath]
			if !ok {
				continue
			}

			expectedPath = alt
		}

		sourceGit := git.New(sourcePath)

		sourceRemoteURL, err := sourceGit.GetRemote(context.TODO())
		if err != nil {
			continue
		}
		sourceRemoteURL = normalizeRemoteURL(sourceRemoteURL)

		if remoteURL != sourceRemoteURL {
			continue
		}

		// Check all services using this source for affected services
		for _, service := range project.Services {
			for _, volume := range service.Volumes {
				if volume.Source != expectedPath {
					continue
				}

				sources[relSourcePath] = true
				affectedServices[service.Name] = true
			}
		}
	}

	if len(sources) == 0 {
		return nil, nil, fmt.Errorf("no services found using the detected source and relative path")
	}

	return toList(sources), toList(affectedServices), nil
}

func toList(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func GetLocalMountCandidates(project *project.Project, filter string) []string {
	filter = strings.ToLower(filter)

	results := make(map[string]bool)
	for sourceName := range project.Sources {
		sourcePath := filepath.Join(project.WorkingDir, app.SourcesDir, sourceName)
		for _, service := range project.Services {
			for _, volume := range service.Volumes {
				if volume.Type == "bind" && strings.HasPrefix(volume.Source, sourcePath) {
					relSourcePath, _ := filepath.Rel(project.WorkingDir, volume.Source)
					relSourcePath = "./" + relSourcePath

					if _, ok := project.LocalMounts[relSourcePath]; ok {
						continue
					}

					results[relSourcePath] = true
				}
			}
		}
	}

	r2 := []string{}
	for k := range results {
		if filter != "" && !strings.Contains(strings.ToLower(k), filter) {
			continue
		}

		r2 = append(r2, k)
	}

	return r2
}

func GetLocalMounts(project *project.Project, filter string) []string {
	filter = strings.ToLower(filter)

	results := []string{}
	for k := range project.LocalMounts {
		if filter == "" || strings.Contains(strings.ToLower(k), filter) {
			results = append(results, k)
		}
	}

	return results
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
