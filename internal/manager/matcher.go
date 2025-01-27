package manager

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/project"
)

// findSourceMatch checks if the given directory matches a specific project's source
func findSourceMatch(proj *project.Project, curDir string) (string, error) {
	gitInfo, err := getGitInfo(curDir)
	if err != nil {
		return "", err
	}

	// 1. Check if current dir is already mounted to this project
	for k, v := range proj.LocalMounts {
		if v == curDir {
			return k, nil
		}
	}

	// 2. Check sources of the project
	for sourceName, source := range proj.Sources {
		sourcePath := filepath.Join(proj.WorkingDir, app.SourcesDir, sourceName)
		relSourcePath, err := filepath.Rel(proj.WorkingDir, sourcePath)
		if err != nil {
			return "", fmt.Errorf("failed to get relative path: %w", err)
		}
		relSourcePath = "./" + relSourcePath

		if !checkSourceByGitRemote(sourcePath, gitInfo.normalizedURL) {
			continue
		}

		if len(source.SparseCheckout) == 0 {
			return relSourcePath, nil
		}

		// if sparse checkout is set, check if the path is in the sparse checkout list
		sparseCheckouts := make([]string, 0, len(source.SparseCheckout))
		for _, v := range source.SparseCheckout {
			if v == "." {
				sparseCheckouts = append(sparseCheckouts, "./")
			} else {
				sparseCheckouts = append(sparseCheckouts, "./"+v)
			}
		}
		sort.Slice(sparseCheckouts, func(i, j int) bool {
			return len(sparseCheckouts[i]) > len(sparseCheckouts[j])
		})

		// Check paths from longest to shortest
		for _, sparseCheckoutPath := range sparseCheckouts {
			if strings.HasPrefix(gitInfo.relativePath, sparseCheckoutPath) {
				return relSourcePath, nil
			}
		}
	}

	// 3. Check if current dir is a git repo of the project itself
	if checkProjectByGitRemote(proj, gitInfo.normalizedURL) {
		return gitInfo.relativePath, nil
	}

	return "", nil
}
