package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type svc struct {
	targetPath string
}

func New(targetFolder string) *svc {
	return &svc{
		targetPath: targetFolder,
	}
}

func (s *svc) Clone(ctx context.Context, url, branch string) error {
	cmds := []string{"clone", url, s.targetPath}
	if branch != "" {
		cmds = append(cmds, "--branch", branch)
	}

	out, err := s.exec(ctx, "git", cmds...)
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
	// if there is no `.git` directory we should not to try to reset because it will try to lock a repo above
	if _, err := os.Stat(filepath.Join(s.targetPath, ".git")); os.IsNotExist(err) {
		os.RemoveAll(s.targetPath)
	}

	isExist := false
	if _, err := os.Stat(s.targetPath); err == nil {
		isExist = true
	}

	if isExist {
		if err := s.reset(ctx); err != nil {
			return fmt.Errorf("failed to reset repo %s: %w", s.targetPath, err)
		}
	} else {
		_ = os.MkdirAll(s.targetPath, os.ModePerm)
		out, err := s.exec(ctx, "git", "clone", "--no-checkout", "--depth", "1", url, s.targetPath)
		if err != nil {
			return fmt.Errorf("failed to clone: %s %w", out, err)
		}
	}

	if len(sparseCheckout) > 0 {
		out, err := s.exec(ctx, "git", "-C", s.targetPath, "sparse-checkout", "init", "--cone")
		if err != nil {
			return fmt.Errorf("failed to init sparse-checkout: %s %w", out, err)
		}

		out, err = s.exec(ctx, "git", append([]string{"-C", s.targetPath, "sparse-checkout", "set"}, sparseCheckout...)...)
		if err != nil {
			return fmt.Errorf("failed to set sparse-checkout: %s %w", out, err)
		}
	} else {
		out, err := s.exec(ctx, "git", "-C", s.targetPath, "sparse-checkout", "disable")
		if err != nil {
			return fmt.Errorf("failed to disable sparse-checkout: %s %w", out, err)
		}
	}

	out, err := s.exec(ctx, "git", "-C", s.targetPath, "checkout", branch)
	if err != nil {
		return fmt.Errorf("failed to checkout: %s %w", out, err)
	}

	return s.Pull(ctx)
}

func (s *svc) Pull(ctx context.Context) error {
	if err := s.reset(ctx); err != nil {
		return fmt.Errorf("failed to reset repo: %w", err)
	}

	out, err := s.exec(ctx, "git", "-C", s.targetPath, "pull", "--rebase")
	if err != nil {
		return fmt.Errorf("failed to pull: %s %w", out, err)
	}

	return nil
}

func (s *svc) GetInfo(ctx context.Context) (*CommitInfo, error) {
	out, err := s.exec(ctx, "git", "-C", s.targetPath, "log", "-1", "--pretty=format:%H%n%aN%n%ad%n%s")
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %s", out)
	}

	parts := strings.Split(out, "\n")
	if len(parts) != 4 {
		return nil, fmt.Errorf("failed to parse commit info: %s", out)
	}

	return &CommitInfo{
		Hash:    parts[0],
		Author:  parts[1],
		Date:    parts[2],
		Message: parts[3],
	}, nil
}

func (s *svc) GetRemote(ctx context.Context) (string, error) {
	out, err := s.exec(ctx, "git", "-C", s.targetPath, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %s %w", out, err)
	}

	return strings.TrimSpace(out), nil
}

func (s *svc) GetTopLevel(ctx context.Context) (string, error) {
	out, err := s.exec(ctx, "git", "-C", s.targetPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get top level: %s %w", out, err)
	}

	return strings.TrimSpace(out), nil
}

func (s *svc) reset(ctx context.Context) error {
	_ = os.Remove(filepath.Join(s.targetPath, ".git/index.lock"))

	out, err := s.exec(ctx, "git", "-C", s.targetPath, "reset", "--hard")
	if err != nil {
		return fmt.Errorf("failed to reset: %s %w", out, err)
	}

	out, err = s.exec(ctx, "git", "-C", s.targetPath, "clean", "-fd")
	if err != nil {
		return fmt.Errorf("failed to clean: %s %w", out, err)
	}

	return nil
}

func (s *svc) exec(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}
