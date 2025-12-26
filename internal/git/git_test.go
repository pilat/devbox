package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
)

func newSvcWithRunner(targetFolder string, runner CommandRunner) *svc {
	return &svc{targetPath: targetFolder, runner: runner}
}

// ============================================================================
// Clone tests
// ============================================================================

func TestClone(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		branch     string
		output     string
		err        error
		wantErr    bool
		wantHint   bool
		expectArgs []string
	}{
		{
			name:       "success without branch",
			url:        "https://github.com/org/repo.git",
			branch:     "",
			output:     "",
			err:        nil,
			wantErr:    false,
			expectArgs: []string{"clone", "https://github.com/org/repo.git", "/tmp/test"},
		},
		{
			name:       "success with branch",
			url:        "https://github.com/org/repo.git",
			branch:     "develop",
			output:     "",
			err:        nil,
			wantErr:    false,
			expectArgs: []string{"clone", "https://github.com/org/repo.git", "/tmp/test", "--branch", "develop"},
		},
		{
			name:       "failure with HTTPS URL shows hint",
			url:        "https://github.com/org/repo.git",
			branch:     "main",
			output:     "authentication failed",
			err:        errors.New("exit status 128"),
			wantErr:    true,
			wantHint:   true,
			expectArgs: []string{"clone", "https://github.com/org/repo.git", "/tmp/test", "--branch", "main"},
		},
		{
			name:       "failure with SSH URL shows hint",
			url:        "git@github.com:org/repo.git",
			branch:     "main",
			output:     "permission denied",
			err:        errors.New("exit status 128"),
			wantErr:    true,
			wantHint:   true,
			expectArgs: []string{"clone", "git@github.com:org/repo.git", "/tmp/test", "--branch", "main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewMockCommandRunner(t)

			// Build expectation with variadic args
			args := make([]any, len(tt.expectArgs))
			for i, a := range tt.expectArgs {
				args[i] = a
			}
			runner.EXPECT().
				RunWithTTY(mock.Anything, "git", args...).
				Return(tt.output, tt.err)

			svc := newSvcWithRunner("/tmp/test", runner)
			err := svc.Clone(context.Background(), tt.url, tt.branch)

			if (err != nil) != tt.wantErr {
				t.Errorf("Clone() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantHint && err != nil {
				if !strings.Contains(err.Error(), "Tip:") {
					t.Error("expected error to contain hint")
				}
			}
		})
	}
}

// ============================================================================
// Pull tests
// ============================================================================

func TestPull(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *MockCommandRunner)
		wantErr    bool
		errContain string
	}{
		{
			name: "success",
			setupMock: func(m *MockCommandRunner) {
				m.EXPECT().Run(mock.Anything, "git", "-C", "/tmp/test", "reset", "--hard").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", "/tmp/test", "clean", "-fd").Return("", nil)
				m.EXPECT().RunWithTTY(mock.Anything, "git", "-C", "/tmp/test", "pull", "--rebase").Return("", nil)
			},
			wantErr: false,
		},
		{
			name: "reset fails",
			setupMock: func(m *MockCommandRunner) {
				m.EXPECT().Run(mock.Anything, "git", "-C", "/tmp/test", "reset", "--hard").Return("error output", errors.New("reset failed"))
			},
			wantErr:    true,
			errContain: "failed to reset",
		},
		{
			name: "clean fails",
			setupMock: func(m *MockCommandRunner) {
				m.EXPECT().Run(mock.Anything, "git", "-C", "/tmp/test", "reset", "--hard").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", "/tmp/test", "clean", "-fd").Return("error output", errors.New("clean failed"))
			},
			wantErr:    true,
			errContain: "failed to clean",
		},
		{
			name: "pull fails",
			setupMock: func(m *MockCommandRunner) {
				m.EXPECT().Run(mock.Anything, "git", "-C", "/tmp/test", "reset", "--hard").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", "/tmp/test", "clean", "-fd").Return("", nil)
				m.EXPECT().RunWithTTY(mock.Anything, "git", "-C", "/tmp/test", "pull", "--rebase").Return("conflict", errors.New("pull failed"))
			},
			wantErr:    true,
			errContain: "failed to pull",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewMockCommandRunner(t)
			tt.setupMock(runner)

			svc := newSvcWithRunner("/tmp/test", runner)
			err := svc.Pull(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Pull() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errContain != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
				}
			}
		})
	}
}

// ============================================================================
// GetInfo tests
// ============================================================================

func TestGetInfo(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		err        error
		wantErr    bool
		wantHash   string
		wantAuthor string
		wantMsg    string
	}{
		{
			name:       "success",
			output:     "abc123\nJohn Doe\nMon Jan 1 12:00:00 2024\nInitial commit",
			err:        nil,
			wantErr:    false,
			wantHash:   "abc123",
			wantAuthor: "John Doe",
			wantMsg:    "Initial commit",
		},
		{
			name:    "command fails",
			output:  "fatal: not a git repository",
			err:     errors.New("exit status 128"),
			wantErr: true,
		},
		{
			name:    "malformed output",
			output:  "only\ntwo\nlines",
			err:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewMockCommandRunner(t)
			runner.EXPECT().
				Run(mock.Anything, "git", "-C", "/tmp/test", "log", "-1", "--pretty=format:%H%n%aN%n%ad%n%s").
				Return(tt.output, tt.err)

			svc := newSvcWithRunner("/tmp/test", runner)
			info, err := svc.GetInfo(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && info != nil {
				if info.Hash != tt.wantHash {
					t.Errorf("Hash = %q, want %q", info.Hash, tt.wantHash)
				}
				if info.Author != tt.wantAuthor {
					t.Errorf("Author = %q, want %q", info.Author, tt.wantAuthor)
				}
				if info.Message != tt.wantMsg {
					t.Errorf("Message = %q, want %q", info.Message, tt.wantMsg)
				}
			}
		})
	}
}

// ============================================================================
// GetRemote tests
// ============================================================================

func TestGetRemote(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		err     error
		wantErr bool
		want    string
	}{
		{
			name:    "success",
			output:  "https://github.com/org/repo.git\n",
			err:     nil,
			wantErr: false,
			want:    "https://github.com/org/repo.git",
		},
		{
			name:    "trims whitespace",
			output:  "  git@github.com:org/repo.git  \n",
			err:     nil,
			wantErr: false,
			want:    "git@github.com:org/repo.git",
		},
		{
			name:    "command fails",
			output:  "error: No such remote 'origin'",
			err:     errors.New("exit status 1"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewMockCommandRunner(t)
			runner.EXPECT().
				Run(mock.Anything, "git", "-C", "/tmp/test", "config", "--get", "remote.origin.url").
				Return(tt.output, tt.err)

			svc := newSvcWithRunner("/tmp/test", runner)
			got, err := svc.GetRemote(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRemote() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetRemote() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ============================================================================
// GetTopLevel tests
// ============================================================================

func TestGetTopLevel(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		err     error
		wantErr bool
		want    string
	}{
		{
			name:    "success",
			output:  "/home/user/project\n",
			err:     nil,
			wantErr: false,
			want:    "/home/user/project",
		},
		{
			name:    "not a git repo",
			output:  "fatal: not a git repository",
			err:     errors.New("exit status 128"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewMockCommandRunner(t)
			runner.EXPECT().
				Run(mock.Anything, "git", "-C", "/tmp/test", "rev-parse", "--show-toplevel").
				Return(tt.output, tt.err)

			svc := newSvcWithRunner("/tmp/test", runner)
			got, err := svc.GetTopLevel(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTopLevel() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetTopLevel() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ============================================================================
// SetLocalExclude tests
// ============================================================================

func TestSetLocalExclude(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		setup    func(t *testing.T, dir string)
		wantErr  bool
		verify   func(t *testing.T, dir string)
	}{
		{
			name:     "writes patterns",
			patterns: []string{"*.log", "node_modules/", ".env"},
			setup: func(t *testing.T, dir string) {
				err := os.MkdirAll(filepath.Join(dir, ".git/info"), 0755)
				if err != nil {
					t.Fatal(err)
				}
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				content, err := os.ReadFile(filepath.Join(dir, ".git/info/exclude"))
				if err != nil {
					t.Fatal(err)
				}
				expected := "*.log\nnode_modules/\n.env\n"
				if string(content) != expected {
					t.Errorf("content = %q, want %q", string(content), expected)
				}
			},
		},
		{
			name:     "empty patterns",
			patterns: []string{},
			setup: func(t *testing.T, dir string) {
				err := os.MkdirAll(filepath.Join(dir, ".git/info"), 0755)
				if err != nil {
					t.Fatal(err)
				}
			},
			wantErr: false,
			verify: func(t *testing.T, dir string) {
				content, err := os.ReadFile(filepath.Join(dir, ".git/info/exclude"))
				if err != nil {
					t.Fatal(err)
				}
				if string(content) != "" {
					t.Errorf("content should be empty, got %q", string(content))
				}
			},
		},
		{
			name:     "directory does not exist",
			patterns: []string{"*.log"},
			setup:    func(t *testing.T, dir string) {},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			svc := New(dir)
			err := svc.SetLocalExclude(tt.patterns)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetLocalExclude() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.verify != nil && !tt.wantErr {
				tt.verify(t, dir)
			}
		})
	}
}

// ============================================================================
// Sync tests
// ============================================================================

func TestSync(t *testing.T) {
	tests := []struct {
		name           string
		setupDir       bool
		setupGit       bool
		sparseCheckout []string
		setupMock      func(m *MockCommandRunner, targetPath string)
		wantErr        bool
		errContain     string
	}{
		{
			name:           "fresh clone without sparse",
			setupDir:       false,
			setupGit:       false,
			sparseCheckout: nil,
			setupMock: func(m *MockCommandRunner, targetPath string) {
				m.EXPECT().RunWithTTY(mock.Anything, "git", "clone", "--no-checkout", "--depth", "1", "https://github.com/org/repo.git", targetPath).Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "sparse-checkout", "disable").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "checkout", "main").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "reset", "--hard").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "clean", "-fd").Return("", nil)
				m.EXPECT().RunWithTTY(mock.Anything, "git", "-C", targetPath, "pull", "--rebase").Return("", nil)
			},
			wantErr: false,
		},
		{
			name:           "fresh clone with sparse",
			setupDir:       false,
			setupGit:       false,
			sparseCheckout: []string{"src", "docs"},
			setupMock: func(m *MockCommandRunner, targetPath string) {
				m.EXPECT().RunWithTTY(mock.Anything, "git", "clone", "--no-checkout", "--depth", "1", "https://github.com/org/repo.git", targetPath).Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "sparse-checkout", "init", "--cone").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "sparse-checkout", "set", "src", "docs").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "checkout", "main").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "reset", "--hard").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "clean", "-fd").Return("", nil)
				m.EXPECT().RunWithTTY(mock.Anything, "git", "-C", targetPath, "pull", "--rebase").Return("", nil)
			},
			wantErr: false,
		},
		{
			name:           "existing repo resets",
			setupDir:       true,
			setupGit:       true,
			sparseCheckout: nil,
			setupMock: func(m *MockCommandRunner, targetPath string) {
				// reset() called first
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "reset", "--hard").Return("", nil).Once()
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "clean", "-fd").Return("", nil).Once()
				// sparse-checkout disable
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "sparse-checkout", "disable").Return("", nil)
				// checkout
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "checkout", "main").Return("", nil)
				// Pull calls reset() again
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "reset", "--hard").Return("", nil).Once()
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "clean", "-fd").Return("", nil).Once()
				m.EXPECT().RunWithTTY(mock.Anything, "git", "-C", targetPath, "pull", "--rebase").Return("", nil)
			},
			wantErr: false,
		},
		{
			name:           "clone fails",
			setupDir:       false,
			setupGit:       false,
			sparseCheckout: nil,
			setupMock: func(m *MockCommandRunner, targetPath string) {
				m.EXPECT().RunWithTTY(mock.Anything, "git", "clone", "--no-checkout", "--depth", "1", "https://github.com/org/repo.git", targetPath).Return("auth failed", errors.New("exit status 128"))
			},
			wantErr:    true,
			errContain: "failed to clone",
		},
		{
			name:           "checkout fails",
			setupDir:       false,
			setupGit:       false,
			sparseCheckout: nil,
			setupMock: func(m *MockCommandRunner, targetPath string) {
				m.EXPECT().RunWithTTY(mock.Anything, "git", "clone", "--no-checkout", "--depth", "1", "https://github.com/org/repo.git", targetPath).Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "sparse-checkout", "disable").Return("", nil)
				m.EXPECT().Run(mock.Anything, "git", "-C", targetPath, "checkout", "main").Return("branch not found", errors.New("exit status 1"))
			},
			wantErr:    true,
			errContain: "failed to checkout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			targetPath := filepath.Join(dir, "repo")

			if tt.setupDir {
				_ = os.MkdirAll(targetPath, 0755)
			}
			if tt.setupGit {
				_ = os.MkdirAll(filepath.Join(targetPath, ".git"), 0755)
			}

			runner := NewMockCommandRunner(t)
			tt.setupMock(runner, targetPath)

			svc := newSvcWithRunner(targetPath, runner)
			err := svc.Sync(context.Background(), "https://github.com/org/repo.git", "main", tt.sparseCheckout)

			if (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errContain != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
				}
			}
		})
	}
}

// ============================================================================
// gitConfigHint tests
// ============================================================================

func TestGitConfigHint(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantContain string
		wantEmpty   bool
	}{
		{
			name:        "HTTPS URL suggests SSH",
			url:         "https://github.com/org/repo.git",
			wantContain: "git@github.com:",
		},
		{
			name:        "HTTP URL suggests SSH",
			url:         "http://github.com/org/repo.git",
			wantContain: "git@github.com:",
		},
		{
			name:        "SSH URL suggests HTTPS",
			url:         "git@github.com:org/repo.git",
			wantContain: "https://github.com/",
		},
		{
			name:      "unknown protocol returns empty",
			url:       "ftp://example.com/repo.git",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gitConfigHint(tt.url)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("gitConfigHint(%q) = %q, want empty", tt.url, got)
				}
				return
			}

			if !strings.Contains(got, tt.wantContain) {
				t.Errorf("gitConfigHint(%q) = %q, should contain %q", tt.url, got, tt.wantContain)
			}
			if !strings.Contains(got, "Tip:") {
				t.Errorf("gitConfigHint(%q) should contain 'Tip:'", tt.url)
			}
		})
	}
}

// ============================================================================
// NormalizeURL tests
// ============================================================================

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Basic formats
		{"HTTPS with .git", "https://github.com/company/repo.git", "github.com/company/repo"},
		{"HTTPS without .git", "https://github.com/company/repo", "github.com/company/repo"},
		{"HTTP", "http://github.com/company/repo.git", "github.com/company/repo"},
		{"SSH standard", "git@github.com:company/repo.git", "github.com/company/repo"},
		{"SSH without .git", "git@github.com:company/repo", "github.com/company/repo"},
		{"SSH explicit", "ssh://git@github.com/company/repo.git", "github.com/company/repo"},
		{"SSH with port", "ssh://git@github.com:22/company/repo.git", "github.com/company/repo"},
		{"git:// protocol", "git://github.com/company/repo.git", "github.com/company/repo"},
		{"git+ssh:// protocol", "git+ssh://git@github.com/company/repo.git", "github.com/company/repo"},

		// Case normalization
		{"uppercase URL", "https://GitHub.com/Company/Repo.git", "github.com/company/repo"},
		{"trailing spaces", "  https://github.com/company/repo.git  ", "github.com/company/repo"},

		// GitLab subgroups
		{"GitLab subgroup HTTPS", "https://gitlab.com/group/subgroup/repo.git", "gitlab.com/group/subgroup/repo"},
		{"GitLab subgroup SSH", "git@gitlab.com:group/subgroup/repo.git", "gitlab.com/group/subgroup/repo"},

		// Bitbucket
		{"Bitbucket HTTPS", "https://bitbucket.org/team/project.git", "bitbucket.org/team/project"},
		{"Bitbucket SSH", "git@bitbucket.org:team/project.git", "bitbucket.org/team/project"},

		// Self-hosted
		{"self-hosted HTTPS", "https://git.company.com/team/repo.git", "git.company.com/team/repo"},
		{"self-hosted SSH", "git@git.company.com:team/repo.git", "git.company.com/team/repo"},

		// Azure DevOps (different URL formats normalize to same value)
		{"Azure SSH", "git@ssh.dev.azure.com:v3/myorg/myproject/myrepo", "dev.azure.com/myorg/myproject/myrepo"},
		{"Azure HTTPS", "https://dev.azure.com/myorg/myproject/_git/myrepo", "dev.azure.com/myorg/myproject/myrepo"},
		{"Azure HTTPS with user", "https://myorg@dev.azure.com/myorg/myproject/_git/myrepo", "dev.azure.com/myorg/myproject/myrepo"},
		{"Azure old format", "https://myorg.visualstudio.com/myproject/_git/myrepo", "dev.azure.com/myorg/myproject/myrepo"},

		// AWS CodeCommit
		{"CodeCommit HTTPS", "https://git-codecommit.us-east-1.amazonaws.com/v1/repos/myrepo", "git-codecommit.us-east-1.amazonaws.com/v1/repos/myrepo"},
		{"CodeCommit SSH", "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/myrepo", "git-codecommit.us-east-1.amazonaws.com/v1/repos/myrepo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeURL(tt.input); got != tt.want {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeURL_MatchesDifferentFormats(t *testing.T) {
	testCases := []struct {
		name     string
		urls     []string
		expected string
	}{
		{
			name: "GitHub formats",
			urls: []string{
				"https://github.com/company/repo.git",
				"http://github.com/company/repo.git",
				"git@github.com:company/repo.git",
				"ssh://git@github.com/company/repo.git",
				"https://github.com/company/repo",
				"git@github.com:company/repo",
			},
			expected: "github.com/company/repo",
		},
		{
			name: "Azure DevOps formats",
			urls: []string{
				"https://dev.azure.com/myorg/myproject/_git/myrepo",
				"https://myorg@dev.azure.com/myorg/myproject/_git/myrepo",
				"git@ssh.dev.azure.com:v3/myorg/myproject/myrepo",
				"https://myorg.visualstudio.com/myproject/_git/myrepo",
			},
			expected: "dev.azure.com/myorg/myproject/myrepo",
		},
		{
			name: "CodeCommit formats",
			urls: []string{
				"https://git-codecommit.us-east-1.amazonaws.com/v1/repos/myrepo",
				"ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/myrepo",
			},
			expected: "git-codecommit.us-east-1.amazonaws.com/v1/repos/myrepo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, u := range tc.urls {
				got := NormalizeURL(u)
				if got != tc.expected {
					t.Errorf("NormalizeURL(%q) = %q, want %q", u, got, tc.expected)
				}
			}
		})
	}
}

func TestFallbackNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"strips https", "https://github.com/owner/repo.git", "github.com/owner/repo"},
		{"strips http", "http://github.com/owner/repo.git", "github.com/owner/repo"},
		{"strips ssh", "ssh://github.com/owner/repo.git", "github.com/owner/repo"},
		{"strips git+ssh", "git+ssh://github.com/owner/repo.git", "github.com/owner/repo"},
		{"strips git://", "git://github.com/owner/repo.git", "github.com/owner/repo"},
		{"strips git@ and converts colon", "git@github.com:owner/repo.git", "github.com/owner/repo"},
		{"lowercases", "HTTPS://GITHUB.COM/OWNER/REPO.GIT", "github.com/owner/repo"},
		{"trims spaces", "  https://github.com/owner/repo.git  ", "github.com/owner/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fallbackNormalize(tt.input)
			if got != tt.want {
				t.Errorf("fallbackNormalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
