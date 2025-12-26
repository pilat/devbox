package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Service interface {
	Clone(ctx context.Context, url, branch string) error
	SetLocalExclude(patterns []string) error
	Sync(ctx context.Context, url, branch string, sparseCheckout []string) error
	Pull(ctx context.Context) error
	GetInfo(ctx context.Context) (*CommitInfo, error)
	GetRemote(ctx context.Context) (string, error)
	GetTopLevel(ctx context.Context) (string, error)
}

var _ Service = (*svc)(nil)

type svc struct {
	targetPath string
	runner     CommandRunner
}

func New(targetFolder string) Service {
	return &svc{
		targetPath: targetFolder,
		runner:     &defaultRunner{},
	}
}

func (s *svc) Clone(ctx context.Context, url, branch string) error {
	args := []string{"clone", url, s.targetPath}
	if branch != "" {
		args = append(args, "--branch", branch)
	}

	out, err := s.runner.RunWithTTY(ctx, "git", args...)
	if err != nil {
		return fmt.Errorf("failed to clone: %s\n%s\n%w", out, gitConfigHint(url), err)
	}

	return nil
}

func (s *svc) SetLocalExclude(patterns []string) error {
	excludeFile := filepath.Join(s.targetPath, ".git/info/exclude")
	file, err := os.OpenFile(excludeFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open exclude file: %w", err)
	}
	defer func() { _ = file.Close() }()

	for _, pattern := range patterns {
		_, err = file.WriteString(pattern + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to exclude file: %w", err)
		}
	}

	return nil
}

func (s *svc) Sync(ctx context.Context, url, branch string, sparseCheckout []string) error {
	// if there is no `.git` directory we should not try to reset because it will try to lock a repo above
	if _, err := os.Stat(filepath.Join(s.targetPath, ".git")); os.IsNotExist(err) {
		_ = os.RemoveAll(s.targetPath)
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
		out, err := s.runner.RunWithTTY(ctx, "git", "clone", "--no-checkout", "--depth", "1", url, s.targetPath)
		if err != nil {
			return fmt.Errorf("failed to clone: %s\n%s\n%w", out, gitConfigHint(url), err)
		}
	}

	if len(sparseCheckout) > 0 {
		out, err := s.runner.Run(ctx, "git", "-C", s.targetPath, "sparse-checkout", "init", "--cone")
		if err != nil {
			return fmt.Errorf("failed to init sparse-checkout: %s %w", out, err)
		}

		out, err = s.runner.Run(ctx, "git", append([]string{"-C", s.targetPath, "sparse-checkout", "set"}, sparseCheckout...)...)
		if err != nil {
			return fmt.Errorf("failed to set sparse-checkout: %s %w", out, err)
		}
	} else {
		out, err := s.runner.Run(ctx, "git", "-C", s.targetPath, "sparse-checkout", "disable")
		if err != nil {
			return fmt.Errorf("failed to disable sparse-checkout: %s %w", out, err)
		}
	}

	out, err := s.runner.Run(ctx, "git", "-C", s.targetPath, "checkout", branch)
	if err != nil {
		return fmt.Errorf("failed to checkout: %s %w", out, err)
	}

	return s.Pull(ctx)
}

func (s *svc) Pull(ctx context.Context) error {
	if err := s.reset(ctx); err != nil {
		return fmt.Errorf("failed to reset repo: %w", err)
	}

	out, err := s.runner.RunWithTTY(ctx, "git", "-C", s.targetPath, "pull", "--rebase")
	if err != nil {
		return fmt.Errorf("failed to pull: %s %w", out, err)
	}

	return nil
}

func (s *svc) GetInfo(ctx context.Context) (*CommitInfo, error) {
	out, err := s.runner.Run(ctx, "git", "-C", s.targetPath, "log", "-1", "--pretty=format:%H%n%aN%n%ad%n%s")
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
	out, err := s.runner.Run(ctx, "git", "-C", s.targetPath, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %s %w", out, err)
	}

	return strings.TrimSpace(out), nil
}

func (s *svc) GetTopLevel(ctx context.Context) (string, error) {
	out, err := s.runner.Run(ctx, "git", "-C", s.targetPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get top level: %s %w", out, err)
	}

	return strings.TrimSpace(out), nil
}

func (s *svc) reset(ctx context.Context) error {
	_ = os.Remove(filepath.Join(s.targetPath, ".git/index.lock"))

	out, err := s.runner.Run(ctx, "git", "-C", s.targetPath, "reset", "--hard")
	if err != nil {
		return fmt.Errorf("failed to reset: %s %w", out, err)
	}

	out, err = s.runner.Run(ctx, "git", "-C", s.targetPath, "clean", "-fd")
	if err != nil {
		return fmt.Errorf("failed to clean: %s %w", out, err)
	}

	return nil
}

// gitConfigHint returns a helpful message suggesting git URL rewrite configuration.
func gitConfigHint(u string) string {
	if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
		host := strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")
		if idx := strings.Index(host, "/"); idx > 0 {
			host = host[:idx]
		}
		return fmt.Sprintf(`Tip: If you use SSH keys instead of HTTPS, configure git:
  git config --global url."git@%s:".insteadOf "https://%s/"`, host, host)
	}

	if strings.HasPrefix(u, "git@") {
		host := strings.TrimPrefix(u, "git@")
		if idx := strings.Index(host, ":"); idx > 0 {
			host = host[:idx]
		}
		return fmt.Sprintf(`Tip: If you use HTTPS tokens instead of SSH, configure git:
  git config --global url."https://%s/".insteadOf "git@%s:"`, host, host)
	}

	return ""
}
