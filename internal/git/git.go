package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pilat/devbox/internal/sys"
)

type svc struct {
	targetPath string
}

func New(targetFolder string) *svc {
	return &svc{
		targetPath: targetFolder,
	}
}

func (s *svc) Reset(ctx context.Context) error {
	_ = os.Remove(filepath.Join(s.targetPath, ".git/index.lock"))

	out, err := sys.Exec(ctx, "git", "-C", s.targetPath, "reset", "--hard")
	if err != nil {
		return fmt.Errorf("failed to reset: %s %w", out, err)
	}

	out, err = sys.Exec(ctx, "git", "-C", s.targetPath, "clean", "-fd")
	if err != nil {
		return fmt.Errorf("failed to clean: %s %w", out, err)
	}

	return nil
}

func (s *svc) Clone(ctx context.Context, url, branch string) error {
	cmds := []string{"clone", url, s.targetPath}
	if branch != "" {
		cmds = append(cmds, "--branch", branch)
	}

	out, err := sys.Exec(ctx, "git", cmds...)
	if err != nil {
		return fmt.Errorf("failed to clone: %s %w", out, err)
	}

	return nil
}

func (s *svc) SetLocalExclude(patterns []string) error {
	excludeFile := filepath.Join(s.targetPath, ".git/info/exclude")
	file, err := os.OpenFile(excludeFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open exclude file: %w", err)
	}
	defer file.Close()

	for _, pattern := range patterns {
		_, err = file.WriteString(pattern + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to exclude file: %w", err)
		}
	}

	return nil
}

func (s *svc) Sync(ctx context.Context, url, branch string, sparseCheckout []string) error {
	isExist := false
	if _, err := os.Stat(s.targetPath); err == nil {
		isExist = true
	}

	if isExist {
		_ = os.Remove(filepath.Join(s.targetPath, ".git/index.lock"))

		out, err := sys.Exec(ctx, "git", "-C", s.targetPath, "reset", "--hard")
		if err != nil {
			return fmt.Errorf("failed to reset: %s %w", out, err)
		}

		out, err = sys.Exec(ctx, "git", "-C", s.targetPath, "clean", "-fd")
		if err != nil {
			return fmt.Errorf("failed to clean: %s %w", out, err)
		}
	} else {
		_ = os.MkdirAll(s.targetPath, os.ModePerm)
		out, err := sys.Exec(ctx, "git", "clone", "--no-checkout", "--depth", "1", url, s.targetPath)
		if err != nil {
			return fmt.Errorf("failed to clone: %s %w", out, err)
		}
	}

	if len(sparseCheckout) > 0 {
		out, err := sys.Exec(ctx, "git", "-C", s.targetPath, "sparse-checkout", "init", "--cone")
		if err != nil {
			return fmt.Errorf("failed to init sparse-checkout: %s %w", out, err)
		}

		out, err = sys.Exec(ctx, "git", append([]string{"-C", s.targetPath, "sparse-checkout", "set"}, sparseCheckout...)...)
		if err != nil {
			return fmt.Errorf("failed to set sparse-checkout: %s %w", out, err)
		}
	} else {
		out, err := sys.Exec(ctx, "git", "-C", s.targetPath, "sparse-checkout", "disable")
		if err != nil {
			return fmt.Errorf("failed to disable sparse-checkout: %s %w", out, err)
		}
	}

	out, err := sys.Exec(ctx, "git", "-C", s.targetPath, "checkout", branch)
	if err != nil {
		return fmt.Errorf("failed to checkout: %s %w", out, err)
	}

	return s.Pull(ctx)
}

func (s *svc) Pull(ctx context.Context) error {
	out, err := sys.Exec(ctx, "git", "-C", s.targetPath, "pull", "--rebase")
	if err != nil {
		return fmt.Errorf("failed to pull: %s %w", out, err)
	}

	return nil
}

func (s *svc) GetInfo(ctx context.Context) (*commitInfo, error) {
	out, err := sys.Exec(ctx, "git", "-C", s.targetPath, "log", "-1", "--pretty=format:%H%n%aN%n%ad%n%s")
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

func (s *svc) GetRemote(ctx context.Context) (string, error) {
	out, err := sys.Exec(ctx, "git", "-C", s.targetPath, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %s %w", out, err)
	}

	return strings.TrimSpace(out), nil
}

func (s *svc) GetTopLevel(ctx context.Context) (string, error) {
	out, err := sys.Exec(ctx, "git", "-C", s.targetPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get top level: %s %w", out, err)
	}

	return strings.TrimSpace(out), nil
}
