package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pilat/devbox/internal/pkg/utils"
)

type svc struct {
	targetPath string
}

func New(targetFolder string) *svc {
	return &svc{
		targetPath: targetFolder,
	}
}

func (s *svc) Reset() error {
	_ = os.Remove(filepath.Join(s.targetPath, ".git/index.lock"))

	out, err := utils.Exec("git", "-C", s.targetPath, "reset", "--hard")
	if err != nil {
		return fmt.Errorf("failed to reset: %s %w", out, err)
	}

	out, err = utils.Exec("git", "-C", s.targetPath, "clean", "-fd")
	if err != nil {
		return fmt.Errorf("failed to clean: %s %w", out, err)
	}

	return nil
}

func (s *svc) Clone(url, branch string) error {
	cmds := []string{"clone", url, s.targetPath}
	if branch != "" {
		cmds = append(cmds, "--branch", branch)
	}

	out, err := utils.Exec("git", cmds...)
	if err != nil {
		return fmt.Errorf("failed to clone: %s %w", out, err)
	}

	return nil
}

func (s *svc) Sync(url, branch string, sparseCheckout []string) error {
	isExist := false
	if _, err := os.Stat(s.targetPath); err == nil {
		isExist = true
	}

	if isExist {
		_ = os.Remove(filepath.Join(s.targetPath, ".git/index.lock"))

		out, err := utils.Exec("git", "-C", s.targetPath, "reset", "--hard")
		if err != nil {
			return fmt.Errorf("failed to reset: %s %w", out, err)
		}

		out, err = utils.Exec("git", "-C", s.targetPath, "clean", "-fd")
		if err != nil {
			return fmt.Errorf("failed to clean: %s %w", out, err)
		}
	} else {
		_ = os.MkdirAll(s.targetPath, os.ModePerm)
		out, err := utils.Exec("git", "clone", "--no-checkout", "--depth", "1", url, s.targetPath)
		if err != nil {
			return fmt.Errorf("failed to clone: %s %w", out, err)
		}
	}

	if len(sparseCheckout) > 0 {
		out, err := utils.Exec("git", "-C", s.targetPath, "sparse-checkout", "init", "--cone")
		if err != nil {
			return fmt.Errorf("failed to init sparse-checkout: %s %w", out, err)
		}

		out, err = utils.Exec("git", append([]string{"-C", s.targetPath, "sparse-checkout", "set"}, sparseCheckout...)...)
		if err != nil {
			return fmt.Errorf("failed to set sparse-checkout: %s %w", out, err)
		}
	} else {
		out, err := utils.Exec("git", "-C", s.targetPath, "sparse-checkout", "disable")
		if err != nil {
			return fmt.Errorf("failed to disable sparse-checkout: %s %w", out, err)
		}
	}

	out, err := utils.Exec("git", "-C", s.targetPath, "checkout", branch)
	if err != nil {
		return fmt.Errorf("failed to checkout: %s %w", out, err)
	}

	return nil
}

func (s *svc) Pull() error {
	out, err := utils.Exec("git", "-C", s.targetPath, "pull")
	if err != nil {
		return fmt.Errorf("failed to pull: %s %w", out, err)
	}

	return nil
}

func (s *svc) GetInfo() (*commitInfo, error) {
	out, err := utils.Exec("git", "-C", s.targetPath, "log", "-1", "--pretty=format:%H%n%aN%n%ad%n%s")
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %s", out)
	}

	parts := strings.Split(out, "\n")
	if len(parts) != 4 {
		return nil, fmt.Errorf("failed to parse commit info: %s", out)
	}

	return &commitInfo{
		Hash:    parts[0],
		Author:  parts[1],
		Date:    parts[2],
		Message: parts[3],
	}, nil
}
