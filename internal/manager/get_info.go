package manager

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pilat/devbox/internal/git"
)

// gitInfo contains common Git repository information needed for detection
type gitInfo struct {
	remoteURL     string
	toplevelDir   string
	relativePath  string
	normalizedURL string
}

// getGitInfo retrieves common Git information for the given directory
func getGitInfo(dir string) (*gitInfo, error) {
	g := git.New(dir)

	remoteURL, err := g.GetRemote(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get remote url: %w", err)
	}

	toplevelDir, err := g.GetTopLevel(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get top level directory: %w", err)
	}

	relativePath, err := filepath.Rel(toplevelDir, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	// Normalize relative path to always have "./" prefix
	normalizedPath := "./"
	if relativePath != "." {
		normalizedPath = "./" + relativePath
	}

	normalizedURL := normalizeRemoteURL(remoteURL)

	return &gitInfo{
		remoteURL:     remoteURL,
		toplevelDir:   toplevelDir,
		relativePath:  normalizedPath,
		normalizedURL: normalizedURL,
	}, nil
}
