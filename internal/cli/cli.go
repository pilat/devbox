package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/pkg/utils"
)

const appFolder = ".devbox"

type cli struct {
	log *slog.Logger
}

func New(log *slog.Logger) *cli {
	return &cli{
		log: log,
	}
}

func (c *cli) Init(url, name string, branch string) error {
	if name == "" {
		c.log.Debug("Name is empty, guessing name from url", "url", url)
		name = guessName(url)
	}

	c.log.Debug("Validating name", "name", name)
	if !validateName(name) {
		return fmt.Errorf("invalid project name %s", name)
	}

	c.log.Debug("Getting home dir...")
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	targetPath := filepath.Join(homeDir, appFolder, name)

	c.log.Debug("Checking if target folder exists", "folder", targetPath)
	_, err = os.Stat(homeDir + "/.devbox/" + name)
	if err == nil {
		return fmt.Errorf("project %s already exists", name)
	}

	git := git.New(targetPath)

	c.log.Debug("Cloning git repo", "url", url, "folder", targetPath)
	err = git.Clone(url, branch)
	if err != nil {
		return fmt.Errorf("failed to clone git repo: %w", err)
	}

	c.log.Debug("Customize exclude files", "folder", targetPath)
	excludeFile := filepath.Join(targetPath, ".git/info/exclude")
	file, err := os.OpenFile(excludeFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open exclude file: %w", err)
	}
	defer file.Close()

	patterns := []string{
		"/sources/",
	}

	for _, pattern := range patterns {
		_, err = file.WriteString(pattern + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to exclude file: %w", err)
		}
	}

	return nil
}

func guessName(source string) string {
	elems := strings.Split(source, "/")
	name := elems[len(elems)-1]

	if strings.HasSuffix(strings.ToLower(name), ".git") {
		name = name[:len(name)-4]
	}

	return name
}

func validateName(name string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(name)
}
