package e2e

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type E2ESuite struct {
	suite.Suite
	home         string
	projectDir   string
	tempDir      string
	manifestRepo string
	manifestDir  string
	source1Repo  string
	source1Dir   string
	hostsFile    string
	fixturesDir  string
	origDir      string
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}

func (s *E2ESuite) SetupSuite() {
	var err error

	s.origDir, err = os.Getwd()
	s.Require().NoError(err)

	s.home, err = os.UserHomeDir()
	s.Require().NoError(err)
	s.projectDir = filepath.Join(s.home, ".devbox", "test-app")

	s.fixturesDir = filepath.Join(s.origDir, "fixtures")

	// Clean up leftover temp directories from previous runs
	oldTempDirs, _ := filepath.Glob(filepath.Join(os.TempDir(), "devbox-e2e-*"))
	for _, dir := range oldTempDirs {
		_ = os.RemoveAll(dir)
	}

	s.tempDir, err = os.MkdirTemp("", "devbox-e2e-")
	s.Require().NoError(err)

	s.manifestRepo = filepath.Join(s.tempDir, "manifest.git")
	s.manifestDir = filepath.Join(s.tempDir, "manifest")
	s.source1Repo = filepath.Join(s.tempDir, "source-1.git")
	s.source1Dir = filepath.Join(s.tempDir, "source-1")
	s.hostsFile = filepath.Join(s.tempDir, "hosts")

	// Clean previous test-app project if exists (only this specific project)
	if _, err := os.Stat(s.projectDir); err == nil {
		_ = os.RemoveAll(s.projectDir)
	}

	// Setup manifest repo
	s.runGit("", "init", "--bare", s.manifestRepo)
	s.Require().NoError(os.Mkdir(s.manifestDir, 0755))
	s.runGit(s.manifestDir, "init")

	// Copy and configure manifest files
	s.copyDir(filepath.Join(s.fixturesDir, "manifest"), s.manifestDir)

	// Update docker-compose.yml with correct repo path
	composeFile := filepath.Join(s.manifestDir, "docker-compose.yml")
	content, err := os.ReadFile(composeFile)
	s.Require().NoError(err)
	content = []byte(strings.ReplaceAll(string(content), "SOURCE_1_REPO", s.source1Repo))
	s.Require().NoError(os.WriteFile(composeFile, content, 0644))

	// Push manifest
	s.runGit(s.manifestDir, "remote", "add", "origin", s.manifestRepo)
	s.runGit(s.manifestDir, "add", ".")
	s.runGit(s.manifestDir, "commit", "-m", "Initial commit")
	s.runGit(s.manifestDir, "branch", "-M", "main")
	s.runGit(s.manifestDir, "push", "-u", "origin", "main")
	s.runGit(s.manifestRepo, "symbolic-ref", "HEAD", "refs/heads/main")

	// Setup source 1 repo
	s.runGit("", "init", "--bare", s.source1Repo)
	s.Require().NoError(os.Mkdir(s.source1Dir, 0755))

	// Copy service-1 files
	s.copyDir(filepath.Join(s.fixturesDir, "service-1"), s.source1Dir)

	// Push source 1
	s.runGit(s.source1Dir, "init")
	s.runGit(s.source1Dir, "remote", "add", "origin", s.source1Repo)
	s.runGit(s.source1Dir, "add", ".")
	s.runGit(s.source1Dir, "commit", "-m", "Initial commit")
	s.runGit(s.source1Dir, "branch", "-M", "main")
	s.runGit(s.source1Dir, "push", "-u", "origin", "main")
	s.runGit(s.source1Repo, "symbolic-ref", "HEAD", "refs/heads/main")

	// Create hosts file
	f, err := os.Create(s.hostsFile)
	s.Require().NoError(err)
	_ = f.Close()
}

func (s *E2ESuite) TearDownSuite() {
	// Stop and remove containers
	out, _ := exec.Command("docker", "ps", "-aq", "--filter", "name=test-app").Output()
	containers := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, c := range containers {
		if c != "" {
			_ = exec.Command("docker", "stop", "-t0", c).Run()
			_ = exec.Command("docker", "rm", "-f", c).Run()
		}
	}

	// Remove image
	_ = exec.Command("docker", "image", "rm", "local/service-1").Run()

	// Clean directories
	if s.projectDir != "" {
		_ = os.RemoveAll(s.projectDir)
	}
	if s.tempDir != "" {
		_ = os.RemoveAll(s.tempDir)
	}
}

func (s *E2ESuite) runGit(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	s.Require().NoError(err, "git %v failed: %s", args, string(out))
}

func (s *E2ESuite) copyDir(src, dst string) {
	entries, err := os.ReadDir(src)
	s.Require().NoError(err)

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			s.Require().NoError(os.MkdirAll(dstPath, 0755))
			s.copyDir(srcPath, dstPath)
		} else {
			data, err := os.ReadFile(srcPath)
			s.Require().NoError(err)
			s.Require().NoError(os.WriteFile(dstPath, data, 0644))
		}
	}
}

func (s *E2ESuite) devbox(args ...string) (string, string, error) {
	cmd := exec.Command("devbox", args...)
	cmd.Env = append(os.Environ(), "DEVBOX_TEST_HOSTS_FILE="+s.hostsFile)

	stdout, err := cmd.Output()
	var stderr []byte
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = exitErr.Stderr
	}
	return string(stdout), string(stderr), err
}

func (s *E2ESuite) devboxInDir(dir string, args ...string) (string, string, error) {
	cmd := exec.Command("devbox", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "DEVBOX_TEST_HOSTS_FILE="+s.hostsFile)

	stdout, err := cmd.Output()
	var stderr []byte
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = exitErr.Stderr
	}
	return string(stdout), string(stderr), err
}

// devboxRun executes devbox command ignoring output (for setup/teardown)
func (s *E2ESuite) devboxRun(args ...string) {
	cmd := exec.Command("devbox", args...)
	cmd.Env = append(os.Environ(), "DEVBOX_TEST_HOSTS_FILE="+s.hostsFile)
	_ = cmd.Run()
}

// devboxRunInDir executes devbox command in directory ignoring output
func (s *E2ESuite) devboxRunInDir(dir string, args ...string) {
	cmd := exec.Command("devbox", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "DEVBOX_TEST_HOSTS_FILE="+s.hostsFile)
	_ = cmd.Run()
}

func (s *E2ESuite) waitFor(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

func (s *E2ESuite) checkContainersUp(count int) bool {
	out, _ := exec.Command("docker", "ps", "--filter", "name=test-app").Output()
	return strings.Count(string(out), "Up") == count
}

func (s *E2ESuite) checkContainersDown() bool {
	out, _ := exec.Command("docker", "ps", "--filter", "name=test-app").Output()
	return !strings.Contains(string(out), "Up")
}

func (s *E2ESuite) checkServiceResponse(url, expected string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return strings.Contains(string(body), expected)
}

func (s *E2ESuite) removeVolumeFromDockerCompose() {
	composeFile := filepath.Join(s.projectDir, "docker-compose.yml")
	content, err := os.ReadFile(composeFile)
	s.Require().NoError(err)

	newContent := strings.ReplaceAll(string(content), `    volumes:
      - ./sources/service-1:/app`, "")
	s.Require().NoError(os.WriteFile(composeFile, []byte(newContent), 0644))
}

func (s *E2ESuite) removeBuildContextFromDockerCompose() {
	composeFile := filepath.Join(s.projectDir, "docker-compose.yml")
	content, err := os.ReadFile(composeFile)
	s.Require().NoError(err)

	newContent := strings.ReplaceAll(string(content), `    build:
      context: ./sources/service-1
      dockerfile: Dockerfile`, "")
	newContent = strings.ReplaceAll(newContent, `image: local/service-1:latest`, `image: golang:1.21-alpine`)
	s.Require().NoError(os.WriteFile(composeFile, []byte(newContent), 0644))
}

func (s *E2ESuite) cleanupProject() {
	s.devboxRun("destroy", "--name", "test-app")
	_ = os.RemoveAll(s.projectDir)
}

func (s *E2ESuite) resetSourceFile() {
	// Reset main.go to original content
	mainGo := filepath.Join(s.source1Dir, "cmd", "service-1", "main.go")
	originalContent := `package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World from service 1")
	})

	http.ListenAndServe(":80", nil)
}
`
	_ = os.WriteFile(mainGo, []byte(originalContent), 0644)
}

// ============================================================================
// Test 10: Project Initialization
// ============================================================================

func (s *E2ESuite) Test10_ProjectInitialization() {
	s.cleanupProject()

	// Test initialization
	stdout, _, err := s.devbox("init", s.manifestRepo, "--name", "test-app", "--branch", "main")
	s.Require().NoError(err)
	s.Contains(stdout, "Project has been successfully initialized!")

	// Test project structure
	for _, file := range []string{"docker-compose.yml", ".devboxstate", ".env"} {
		_, err := os.Stat(filepath.Join(s.projectDir, file))
		s.NoError(err, "File %s should exist", file)
	}

	// Test git exclude configuration
	excludeFile := filepath.Join(s.projectDir, ".git", "info", "exclude")
	content, err := os.ReadFile(excludeFile)
	s.Require().NoError(err)
	expectedContent := "/sources/\n/.devboxstate\n/.env"
	s.Equal(expectedContent, strings.TrimSpace(string(content)))

	// Test project list
	stdout, _, err = s.devbox("list")
	s.Require().NoError(err)
	s.Contains(stdout, "test-app")
}

// ============================================================================
// Test 30: Info Operations
// ============================================================================

func (s *E2ESuite) Test30_InfoOperations() {
	s.cleanupProject()

	// Initialize project
	s.devboxRun("init", s.manifestRepo, "--name", "test-app", "--branch", "main")

	// Test info after init should fail (no sources synced yet)
	stdout, stderr, _ := s.devbox("info", "--name", "test-app")
	s.Contains(stdout+stderr, "No such file or directory")

	// Test updating sources
	stdout, _, err := s.devbox("update", "--name", "test-app")
	s.Require().NoError(err)
	s.Contains(stdout, "Source service-1  Synced")

	// Test info after update
	stdout, _, err = s.devbox("info", "--name", "test-app")
	s.Require().NoError(err)
	s.Contains(stdout, "Initial commit")

	// Test services are not started
	out, _ := exec.Command("docker", "ps", "--filter", "name=test-app").Output()
	s.NotContains(string(out), "Up")
}

// ============================================================================
// Test 40: Project Context Detection
// ============================================================================

func (s *E2ESuite) Test40_ProjectContextDetection() {
	s.cleanupProject()

	// Initialize and update project
	s.devboxRun("init", s.manifestRepo, "--name", "test-app", "--branch", "main")
	s.devboxRun("update", "--name", "test-app")

	// Test info in random directory not working
	stdout, stderr, _ := s.devboxInDir("/tmp", "info")
	s.Contains(stdout+stderr, "Error has occurred")

	// Test info in random directory works if we mention project name
	stdout, _, err := s.devboxInDir("/tmp", "info", "--name", "test-app")
	s.Require().NoError(err)
	s.Contains(stdout, "Initial commit")

	// Test info in source directory
	stdout, _, err = s.devboxInDir(s.source1Dir, "info")
	s.Require().NoError(err)
	s.Contains(stdout, "Initial commit")

	// Test info in subdirectory
	subDir := filepath.Join(s.source1Dir, "cmd", "service-1")
	stdout, _, err = s.devboxInDir(subDir, "info")
	s.Require().NoError(err)
	s.Contains(stdout, "Initial commit")

	// Test that we can detect project just because it mentioned in project's sources
	s.removeVolumeFromDockerCompose()
	stdout, _, err = s.devboxInDir(subDir, "info")
	s.Require().NoError(err)
	s.Contains(stdout, "Initial commit")
}

// ============================================================================
// Test 50: Mount Context Detection
// ============================================================================

func (s *E2ESuite) Test50_MountContextDetection() {
	s.cleanupProject()

	// Initialize and update project
	s.devboxRun("init", s.manifestRepo, "--name", "test-app", "--branch", "main")
	s.devboxRun("update", "--name", "test-app")

	// Test mount in random directory not working even if we mention project name
	stdout, stderr, _ := s.devboxInDir("/tmp", "mount", "--name", "test-app")
	s.Contains(stdout+stderr, "Error has occurred")

	// Test mount in project directory
	stdout, _, err := s.devboxInDir(s.source1Dir, "mount")
	s.Require().NoError(err)
	s.Contains(stdout, "./sources/service-1")

	// We have to unmount
	s.devboxRunInDir(s.source1Dir, "umount")

	// It is expected that mountpoint won't be found for subdirectory
	subDir := filepath.Join(s.source1Dir, "cmd", "service-1")
	stdout, _, _ = s.devboxInDir(subDir, "mount")
	s.NotContains(stdout, "./sources/service-1")

	// Test that it will work if not mentioned in volumes and mentioned in build.context
	s.removeVolumeFromDockerCompose()
	stdout, _, err = s.devboxInDir(s.source1Dir, "mount")
	s.Require().NoError(err)
	s.Contains(stdout, "./sources/service-1")
}

// ============================================================================
// Test 55: Mount for Build Operations
// ============================================================================

func (s *E2ESuite) Test55_MountForBuildOperations() {
	s.cleanupProject()
	s.resetSourceFile()

	// Initialize and update project
	s.devboxRun("init", s.manifestRepo, "--name", "test-app", "--branch", "main")
	s.devboxRun("update", "--name", "test-app")

	s.removeVolumeFromDockerCompose()

	// Check services are not started
	out, _ := exec.Command("docker", "ps", "--filter", "name=test-app").Output()
	s.NotContains(string(out), "Up", "Services should not be started")

	// Test mounting source
	stdout, _, err := s.devboxInDir(s.source1Dir, "mount")
	s.Require().NoError(err)
	s.Contains(stdout, "LOCAL PATH")
	s.Contains(stdout, "./sources/service-1")

	// Start project
	_, _, err = s.devbox("up", "--name", "test-app")
	s.Require().NoError(err)

	// Wait for containers to be up
	s.True(s.waitFor(func() bool { return s.checkContainersUp(2) }, 30*time.Second),
		"Project containers should be running")

	// Test info shows commit
	stdout, _, err = s.devboxInDir(s.source1Dir, "info")
	s.Require().NoError(err)
	s.Contains(stdout, "Initial commit")

	// Test code changes
	mainGo := filepath.Join(s.source1Dir, "cmd", "service-1", "main.go")
	content, err := os.ReadFile(mainGo)
	s.Require().NoError(err)
	newContent := strings.ReplaceAll(string(content),
		"Hello, World from service 1",
		"Hello, World from service 1 updated")
	s.Require().NoError(os.WriteFile(mainGo, []byte(newContent), 0644))

	// Restart service
	s.devboxRun("restart", "--name", "test-app", "service-1")

	// Verify changes
	s.True(s.waitFor(func() bool {
		return s.checkServiceResponse("http://localhost:8081", "Hello, World from service 1 updated")
	}, 30*time.Second), "Service should return updated response")
}

// ============================================================================
// Test 57: Mount for Volume Operations
// ============================================================================

func (s *E2ESuite) Test57_MountForVolumeOperations() {
	s.cleanupProject()
	s.resetSourceFile()

	// Initialize and update project
	s.devboxRun("init", s.manifestRepo, "--name", "test-app", "--branch", "main")
	s.devboxRun("update", "--name", "test-app")

	s.removeBuildContextFromDockerCompose()

	// Test mounting source
	stdout, _, err := s.devboxInDir(s.source1Dir, "mount")
	s.Require().NoError(err)
	s.Contains(stdout, "LOCAL PATH")
	s.Contains(stdout, "./sources/service-1")

	// Start project
	_, _, err = s.devbox("up", "--name", "test-app")
	s.Require().NoError(err)

	// Wait for containers to be up
	s.True(s.waitFor(func() bool { return s.checkContainersUp(2) }, 30*time.Second),
		"Project containers should be running")

	// Test info shows commit
	stdout, _, err = s.devboxInDir(s.source1Dir, "info")
	s.Require().NoError(err)
	s.Contains(stdout, "Initial commit")

	// Test code changes
	mainGo := filepath.Join(s.source1Dir, "cmd", "service-1", "main.go")
	content, err := os.ReadFile(mainGo)
	s.Require().NoError(err)
	newContent := strings.ReplaceAll(string(content),
		"Hello, World from service 1",
		"Hello, World from service 1 updated")
	s.Require().NoError(os.WriteFile(mainGo, []byte(newContent), 0644))

	// Restart service
	s.devboxRun("restart", "--name", "test-app", "service-1")

	// Verify changes
	s.True(s.waitFor(func() bool {
		return s.checkServiceResponse("http://localhost:8081", "Hello, World from service 1 updated")
	}, 30*time.Second), "Service should return updated response")
}

// ============================================================================
// Test 60: Project Operations
// ============================================================================

func (s *E2ESuite) Test60_ProjectOperations() {
	s.cleanupProject()

	// Initialize project
	s.devboxRun("init", s.manifestRepo, "--name", "test-app", "--branch", "main")

	// Start project
	_, _, err := s.devbox("up", "--name", "test-app")
	s.Require().NoError(err)

	// Wait for containers to be up
	s.True(s.waitFor(func() bool { return s.checkContainersUp(2) }, 30*time.Second),
		"Project containers should be running")

	// Stop project
	s.devboxRun("down", "--name", "test-app")

	// Wait for containers to stop
	s.True(s.waitFor(s.checkContainersDown, 30*time.Second),
		"Project containers should be stopped")
}

// ============================================================================
// Test 90: Project Cleanup
// ============================================================================

func (s *E2ESuite) Test90_ProjectCleanup() {
	s.cleanupProject()

	// Initialize and start project
	s.devboxRun("init", s.manifestRepo, "--name", "test-app", "--branch", "main")
	s.devboxRun("up", "--name", "test-app")

	// Test project destroy
	s.devboxRun("destroy", "--name", "test-app")

	_, err := os.Stat(s.projectDir)
	s.True(os.IsNotExist(err), "Project directory should be removed")

	// Verify no containers
	out, _ := exec.Command("docker", "ps", "-q", "--filter", "name=test-app").Output()
	s.Empty(strings.TrimSpace(string(out)), "No containers should be running")

	// Make sure files were removed
	_, err = os.Stat(s.projectDir)
	s.True(os.IsNotExist(err), "Project files should be removed")
}
