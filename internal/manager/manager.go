package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/project"
)

func GetLocalMountCandidates(project *project.Project, filter string) []string {
	filter = strings.ToLower(filter)

	results := make(map[string]bool)
	for sourceName := range project.Sources {
		sourcePath := filepath.Join(project.WorkingDir, app.SourcesDir, sourceName)
		for _, service := range project.Services {
			if service.Build != nil && strings.HasPrefix(service.Build.Context, sourcePath) {
				relSourcePath, _ := filepath.Rel(project.WorkingDir, service.Build.Context)
				relSourcePath = "./" + relSourcePath
				results[relSourcePath] = true

				if _, ok := project.LocalMounts[relSourcePath]; ok {
					continue
				}

				results[relSourcePath] = true
			}

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
		fmt.Sprintf("/%s", app.EnvFile),
	}

	err = git.SetLocalExclude(patterns)
	if err != nil {
		return fmt.Errorf("failed to set local exclude: %w", err)
	}

	for k, content := range map[string]string{
		app.EnvFile:   "",
		app.StateFile: "{}",
	} {
		err = os.WriteFile(filepath.Join(projectFolder, k), []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
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
