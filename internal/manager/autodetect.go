package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/project"
)

type AutodetectSourceType int

// AutodetectProject validates project name (if provided) or performs autodetection
func AutodetectProject(name string) (*project.Project, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load projects
	projects := make([]*project.Project, 0)
	projectNames := ListProjects("")
	for _, projectName := range projectNames {
		project, err := project.New(context.Background(), projectName, []string{"*"})
		if err != nil {
			continue
		}
		projects = append(projects, project)
	}

	// If name was provided we are validating it and immediately return the result or error
	if name != "" {
		for _, project := range projects {
			if project.Name == name {
				return project, nil
			}
		}

		return nil, fmt.Errorf("project '%s' not found", name)
	}

	// Examine all the projects and expect the only one match
	var foundProject *project.Project
	for _, project := range projects {
		source, err := findSourceMatch(project, curDir)
		if err != nil {
			return nil, err
		}

		if source == "" {
			continue
		}

		if foundProject != nil && project != foundProject {
			return nil, fmt.Errorf("ambiguous project, please specify project name")
		}

		foundProject = project
	}

	if foundProject == nil {
		return nil, fmt.Errorf("unknown project")
	}

	return foundProject, nil
}

func AutodetectSource(project *project.Project, sourceNameSel string) (string, []string, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// fake curDir to make it look like we are in the source so algorithms below will be able to detect it
	if sourceNameSel != "" {
		curDir = filepath.Join(project.WorkingDir, sourceNameSel)
	}

	source, err := findSourceMatch(project, curDir)
	if err != nil {
		return "", nil, err
	}

	if source == "" {
		return "", nil, fmt.Errorf("unknown source")
	}

	affectedServices := detectAffectedServices(project, source)

	return source, toList(affectedServices), nil
}

// checkProjectByGitRemote checks if the given project matches the Git remote URL
func checkProjectByGitRemote(project *project.Project, normalizedURL string) bool {
	projectGit := git.New(project.WorkingDir)
	remoteURLCurrent, err := projectGit.GetRemote(context.TODO())
	if err != nil {
		return false
	}

	currentNormalized := normalizeRemoteURL(remoteURLCurrent)
	matches := currentNormalized == normalizedURL
	return matches
}

// checkSourceByGitRemote checks if the given source matches the Git remote URL
func checkSourceByGitRemote(sourcePath string, normalizedURL string) bool {
	g := git.New(sourcePath)
	remoteURLCurrent, err := g.GetRemote(context.TODO())
	if err != nil {
		return false
	}

	currentNormalized := normalizeRemoteURL(remoteURLCurrent)
	matches := currentNormalized == normalizedURL
	return matches
}

// detectAffectedServices returns a map of service names that use the given path
func detectAffectedServices(project *project.Project, path string) map[string]bool {
	affected := make(map[string]bool)
	for _, service := range project.Services {
		if service.Build != nil && service.Build.Context == path {
			affected[service.Name] = true
		}

		for _, volume := range service.Volumes {
			if volume.Source == path {
				affected[service.Name] = true
			}
		}
	}
	return affected
}

func toList(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func isProjectExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func validateName(name string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name)
}

func normalizeRemoteURL(s string) string {
	s = strings.ToLower(s)

	// Remove protocol
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "git@")
	s = strings.TrimPrefix(s, "ssh://")

	// Remove auth info if present (user:pass@ or token@)
	if idx := strings.Index(s, "@"); idx != -1 {
		s = s[idx+1:]
	}

	// Convert git SSH format to path format, but preserve port numbers
	// First, find if there's a port number
	parts := strings.Split(s, ":")
	if len(parts) == 2 {
		// Check if the second part starts with a number (port)
		if len(parts[1]) > 0 && (parts[1][0] >= '0' && parts[1][0] <= '9') {
			// This is a port number, keep the colon
			s = strings.Join(parts, ":")
		} else {
			// This is a path separator in git format, replace with slash
			s = strings.Join(parts, "/")
		}
	}

	// Remove .git suffix
	s = strings.TrimSuffix(s, ".git")

	return s
}
