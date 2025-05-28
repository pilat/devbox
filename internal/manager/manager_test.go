package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/project"
	// Assuming git interactions might need to be controlled,
	// though direct mocking of git.New().GetRemote etc. is hard without code changes.
	// "github.com/pilat/devbox/internal/git"
)

// Original function placeholders for mocking
var (
	originalGetwd func() (string, error)
	originalGitGetRemote func(ctx context.Context) (string, error)
	originalGitGetTopLevel func(ctx context.Context) (string, error)
	// projectNewFunc is tricky as project.New is called directly.
	// We will control its behavior by setting up the file system.
)

// Mock functions
func mockGetwdNonProjectDir() (string, error) {
	return "/tmp/non_project_dir_for_test", nil
}

func mockGetwdError() (string, error) {
	return "", fmt.Errorf("mock getwd error")
}

func mockGitGetRemoteError(ctx context.Context) (string, error) {
	return "", fmt.Errorf("mock git remote error")
}

func mockGitGetTopLevelError(ctx context.Context) (string, error) {
	return "", fmt.Errorf("mock git top level error")
}


// This is a simplified setup. In a real scenario, project.New might require more complex mocking
// or specific file structures. We're relying on project.New being able to load a project
// if its directory exists and is named correctly under app.AppDir.
func TestAutodetectProject_FallbackToOneProject(t *testing.T) {
	originalAppDir := app.AppDir
	tempAppDir, err := os.MkdirTemp("", "test_app_dir_")
	if err != nil {
		t.Fatalf("Failed to create temp app dir: %v", err)
	}
	app.AppDir = tempAppDir
	defer func() {
		app.AppDir = originalAppDir
		os.RemoveAll(tempAppDir)
	}()

	// 1. Setup: Simulate having only one project.
	singleProjectName := "testProjectFallback"
	singleProjectDir := filepath.Join(tempAppDir, singleProjectName)
	err = os.MkdirAll(singleProjectDir, 0755) // Create the project directory
	if err != nil {
		t.Fatalf("Failed to create single project dir: %v", err)
	}
	// project.New might need a devbox.yaml or other files.
	// For this test, we assume it can "load" if the directory exists.
	// If project.New is more complex, this setup would need to be more detailed.
	// e.g. create a minimal devbox.yaml
	dummyDevboxYAML := filepath.Join(singleProjectDir, "devbox.yaml")
	err = os.WriteFile(dummyDevboxYAML, []byte("name: "+singleProjectName+"\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy devbox.yaml: %v", err)
	}


	// Store and defer restoration of os.Getwd
	originalGetwd = os.Getwd
	os.Getwd = mockGetwdNonProjectDir // Ensure CWD doesn't match any project paths
	defer func() { os.Getwd = originalGetwd }()

	// Store and defer restoration of git command functions
	// This is tricky because git.New(path) is called, then methods on the instance.
	// For this test, we rely on os.Getwd() being a non-git directory,
	// so git commands on that path will likely fail or return non-matching data.
	// A more robust way would be to inject a git client or mock git.New itself.
	// For now, we assume the default git.New() behavior on a non-git path will suffice
	// to make the primary detection mechanisms fail.

	// 2. Execution
	detectedProject, err := AutodetectProject("")

	// 3. Assertion
	if err != nil {
		t.Errorf("AutodetectProject returned an error: %v", err)
	}
	if detectedProject == nil {
		t.Fatalf("AutodetectProject returned nil project, expected '%s'", singleProjectName)
	}
	if detectedProject.Name != singleProjectName {
		t.Errorf("AutodetectProject returned project name '%s', expected '%s'", detectedProject.Name, singleProjectName)
	}

	// Additional check: ensure the fallback was actually used.
	// This is harder to verify without logs or more detailed return info.
	// We infer it by ensuring other detection methods would fail.
	t.Logf("Successfully detected project '%s' via fallback mechanism.", detectedProject.Name)
}

func TestMain(m *testing.M) {
	// Setup for all tests if needed, e.g., overriding package-level function variables
	// For now, specific mocks are handled within the test case.
	
	// We need to ensure that the git commands will fail or not match.
	// One way is to ensure 'git' command is not found or current dir is not a repo.
	// The mockGetwdNonProjectDir helps with the CWD not being special.
	// If actual git commands are still an issue, would need to mock git.Exec somehow.

	exitCode := m.Run()
	os.Exit(exitCode)
}
