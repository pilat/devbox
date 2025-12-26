package manager

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/pkg/fs"
	"github.com/pilat/devbox/internal/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testProjectName = "myproject"

// mockGitFactory creates a gitServiceFactory that returns preconfigured mocks for specific paths.
func mockGitFactory(t *testing.T, pathMocks map[string]*git.MockService) gitServiceFactory {
	return func(path string) git.Service {
		if m, ok := pathMocks[path]; ok {
			return m
		}
		// Return a default mock that will fail if any method is called unexpectedly
		m := git.NewMockService(t)
		return m
	}
}

// newDirFileInfo creates a MockFileInfo that returns true for IsDir().
func newDirFileInfo(t *testing.T) *fs.MockFileInfo {
	fi := fs.NewMockFileInfo(t)
	fi.EXPECT().IsDir().Return(true).Maybe()
	return fi
}

// ============================================================================
// detectByLocalMount tests
// ============================================================================

func TestDetectByLocalMount(t *testing.T) {
	tests := []struct {
		name        string
		curDir      string
		projects    []*project.Project
		wantProject string
		wantAmbig   bool
	}{
		{
			name:        "no projects",
			curDir:      "/home/user/myproject",
			projects:    []*project.Project{},
			wantProject: "",
			wantAmbig:   false,
		},
		{
			name:   "no match",
			curDir: "/home/user/myproject",
			projects: []*project.Project{
				{
					Project: &types.Project{Name: "proj1"},
					LocalMounts: map[string]string{
						"./sources/backend": "/home/user/other",
					},
				},
			},
			wantProject: "",
			wantAmbig:   false,
		},
		{
			name:   "single match",
			curDir: "/home/user/myproject",
			projects: []*project.Project{
				{
					Project: &types.Project{Name: "proj1"},
					LocalMounts: map[string]string{
						"./sources/backend": "/home/user/myproject",
					},
				},
			},
			wantProject: "proj1",
			wantAmbig:   false,
		},
		{
			name:   "multiple mounts same project",
			curDir: "/home/user/myproject",
			projects: []*project.Project{
				{
					Project: &types.Project{Name: "proj1"},
					LocalMounts: map[string]string{
						"./sources/backend":  "/home/user/myproject",
						"./sources/frontend": "/home/user/myproject",
					},
				},
			},
			wantProject: "proj1",
			wantAmbig:   false,
		},
		{
			name:   "ambiguous - same dir in two projects",
			curDir: "/home/user/myproject",
			projects: []*project.Project{
				{
					Project: &types.Project{Name: "proj1"},
					LocalMounts: map[string]string{
						"./sources/backend": "/home/user/myproject",
					},
				},
				{
					Project: &types.Project{Name: "proj2"},
					LocalMounts: map[string]string{
						"./sources/api": "/home/user/myproject",
					},
				},
			},
			wantProject: "proj2", // last match
			wantAmbig:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			got, ambig := m.detectByLocalMount(tt.curDir, tt.projects)

			if ambig != tt.wantAmbig {
				t.Errorf("detectByLocalMount() ambiguous = %v, want %v", ambig, tt.wantAmbig)
			}

			gotName := ""
			if got != nil {
				gotName = got.Name
			}
			if gotName != tt.wantProject {
				t.Errorf("detectByLocalMount() project = %q, want %q", gotName, tt.wantProject)
			}
		})
	}
}

// ============================================================================
// detectBySourceRemote tests
// ============================================================================

func TestDetectBySourceRemote(t *testing.T) {
	tests := []struct {
		name        string
		remoteURL   string
		projects    []*project.Project
		wantProject string
		wantAmbig   bool
	}{
		{
			name:        "no projects",
			remoteURL:   "github.com/company/repo",
			projects:    []*project.Project{},
			wantProject: "",
			wantAmbig:   false,
		},
		{
			name:      "match by source URL",
			remoteURL: "github.com/company/backend",
			projects: []*project.Project{
				{
					Project: &types.Project{Name: "proj1"},
					Sources: project.SourceConfigs{
						"backend": {URL: "https://github.com/company/backend.git"},
					},
				},
			},
			wantProject: "proj1",
			wantAmbig:   false,
		},
		{
			name:      "no match - different URL",
			remoteURL: "github.com/company/other",
			projects: []*project.Project{
				{
					Project: &types.Project{Name: "proj1"},
					Sources: project.SourceConfigs{
						"backend": {URL: "https://github.com/company/backend.git"},
					},
				},
			},
			wantProject: "",
			wantAmbig:   false,
		},
		{
			name:      "ambiguous - same URL in multiple projects",
			remoteURL: "github.com/company/shared",
			projects: []*project.Project{
				{
					Project: &types.Project{Name: "proj1"},
					Sources: project.SourceConfigs{
						"shared": {URL: "https://github.com/company/shared.git"},
					},
				},
				{
					Project: &types.Project{Name: "proj2"},
					Sources: project.SourceConfigs{
						"shared": {URL: "https://github.com/company/shared.git"},
					},
				},
			},
			wantProject: "proj2",
			wantAmbig:   true,
		},
		{
			name:      "multiple sources - one matches",
			remoteURL: "github.com/company/frontend",
			projects: []*project.Project{
				{
					Project: &types.Project{Name: "proj1"},
					Sources: project.SourceConfigs{
						"backend":  {URL: "https://github.com/company/backend.git"},
						"frontend": {URL: "https://github.com/company/frontend.git"},
					},
				},
			},
			wantProject: "proj1",
			wantAmbig:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			got, ambig := m.detectBySourceRemote(tt.remoteURL, tt.projects)

			if ambig != tt.wantAmbig {
				t.Errorf("detectBySourceRemote() ambiguous = %v, want %v", ambig, tt.wantAmbig)
			}

			gotName := ""
			if got != nil {
				gotName = got.Name
			}
			if gotName != tt.wantProject {
				t.Errorf("detectBySourceRemote() project = %q, want %q", gotName, tt.wantProject)
			}
		})
	}
}

// ============================================================================
// detectByProjectRepo tests
// ============================================================================

func TestDetectByProjectRepo(t *testing.T) {
	tests := []struct {
		name        string
		remoteURL   string
		projects    []*project.Project
		setupGit    func(t *testing.T) map[string]*git.MockService
		wantProject string
		wantAmbig   bool
	}{
		{
			name:      "match project repo",
			remoteURL: "github.com/company/devbox-config",
			projects: []*project.Project{
				{
					Project: &types.Project{
						Name:       testProjectName,
						WorkingDir: "/home/user/.devbox/myproject",
					},
				},
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/devbox-config.git", nil)
				return map[string]*git.MockService{
					"/home/user/.devbox/myproject": m,
				}
			},
			wantProject: testProjectName,
			wantAmbig:   false,
		},
		{
			name:      "no match",
			remoteURL: "github.com/company/other-repo",
			projects: []*project.Project{
				{
					Project: &types.Project{
						Name:       testProjectName,
						WorkingDir: "/home/user/.devbox/myproject",
					},
				},
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/devbox-config.git", nil)
				return map[string]*git.MockService{
					"/home/user/.devbox/myproject": m,
				}
			},
			wantProject: "",
			wantAmbig:   false,
		},
		{
			name:      "git error - skip project",
			remoteURL: "github.com/company/devbox-config",
			projects: []*project.Project{
				{
					Project: &types.Project{
						Name:       testProjectName,
						WorkingDir: "/home/user/.devbox/myproject",
					},
				},
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().GetRemote(mock.Anything).Return("", errors.New("not a git repo"))
				return map[string]*git.MockService{
					"/home/user/.devbox/myproject": m,
				}
			},
			wantProject: "",
			wantAmbig:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathMocks := tt.setupGit(t)
			gitFactory := mockGitFactory(t, pathMocks)

			m := New()
			m.gitFactory = gitFactory
			got, ambig := m.detectByProjectRepo(context.Background(), tt.remoteURL, tt.projects)

			if ambig != tt.wantAmbig {
				t.Errorf("detectByProjectRepo() ambiguous = %v, want %v", ambig, tt.wantAmbig)
			}

			gotName := ""
			if got != nil {
				gotName = got.Name
			}
			if gotName != tt.wantProject {
				t.Errorf("detectByProjectRepo() project = %q, want %q", gotName, tt.wantProject)
			}
		})
	}
}

// ============================================================================
// AutodetectProject integration tests
// ============================================================================

func TestAutodetectProject_ByName(t *testing.T) {
	m := New()
	m.listFn = func(filter string) []string {
		return []string{testProjectName}
	}
	m.loadFn = func(ctx context.Context, name string, profiles []string) (*project.Project, error) {
		return &project.Project{Project: &types.Project{Name: testProjectName}}, nil
	}

	got, err := m.AutodetectProject(context.Background(), testProjectName)
	require.NoError(t, err)
	assert.Equal(t, testProjectName, got.Name)
}

func TestAutodetectProject_ByName_NotFound(t *testing.T) {
	m := New()
	m.listFn = func(filter string) []string {
		return []string{"other"}
	}
	m.loadFn = func(ctx context.Context, name string, profiles []string) (*project.Project, error) {
		return &project.Project{Project: &types.Project{Name: "other"}}, nil
	}

	_, err := m.AutodetectProject(context.Background(), testProjectName)
	require.Error(t, err)
	assert.Equal(t, "project '"+testProjectName+"' not found", err.Error())
}

func TestAutodetectProject_ByLocalMount(t *testing.T) {
	mockFS := fs.NewMockFileSystem(t)
	mockFS.EXPECT().Getwd().Return("/home/user/backend", nil)

	m := New()
	m.listFn = func(filter string) []string {
		return []string{testProjectName}
	}
	m.loadFn = func(ctx context.Context, name string, profiles []string) (*project.Project, error) {
		return &project.Project{
			Project:     &types.Project{Name: testProjectName},
			LocalMounts: map[string]string{"./sources/backend": "/home/user/backend"},
		}, nil
	}
	m.fs = mockFS

	got, err := m.AutodetectProject(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, testProjectName, got.Name)
}

func TestAutodetectProject_ByGitRemote(t *testing.T) {
	curDirMock := git.NewMockService(t)
	curDirMock.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/backend.git", nil)

	pathMocks := map[string]*git.MockService{
		"/home/user/code/backend": curDirMock,
	}

	mockFS := fs.NewMockFileSystem(t)
	mockFS.EXPECT().Getwd().Return("/home/user/code/backend", nil)

	m := New()
	m.listFn = func(filter string) []string {
		return []string{testProjectName}
	}
	m.loadFn = func(ctx context.Context, name string, profiles []string) (*project.Project, error) {
		return &project.Project{
			Project: &types.Project{
				Name:       testProjectName,
				WorkingDir: "/home/user/.devbox/myproject",
			},
			Sources: project.SourceConfigs{
				"backend": {URL: "https://github.com/company/backend.git"},
			},
		}, nil
	}
	m.gitFactory = mockGitFactory(t, pathMocks)
	m.fs = mockFS

	got, err := m.AutodetectProject(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, testProjectName, got.Name)
}

func TestAutodetectProject_ByProjectRepo(t *testing.T) {
	// GetRemote is called twice: once for source detection, once for project repo detection
	projectMock := git.NewMockService(t)
	projectMock.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/devbox-config.git", nil).Times(2)

	pathMocks := map[string]*git.MockService{
		"/home/user/.devbox/myproject": projectMock,
	}

	mockFS := fs.NewMockFileSystem(t)
	mockFS.EXPECT().Getwd().Return("/home/user/.devbox/myproject", nil)

	m := New()
	m.listFn = func(filter string) []string {
		return []string{testProjectName}
	}
	m.loadFn = func(ctx context.Context, name string, profiles []string) (*project.Project, error) {
		return &project.Project{
			Project: &types.Project{
				Name:       testProjectName,
				WorkingDir: "/home/user/.devbox/myproject",
			},
		}, nil
	}
	m.gitFactory = mockGitFactory(t, pathMocks)
	m.fs = mockFS

	got, err := m.AutodetectProject(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, testProjectName, got.Name)
}

func TestAutodetectProject_Unknown(t *testing.T) {
	curDirMock := git.NewMockService(t)
	curDirMock.EXPECT().GetRemote(mock.Anything).Return("https://github.com/unknown/repo.git", nil)

	projectMock := git.NewMockService(t)
	projectMock.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/devbox-config.git", nil)

	pathMocks := map[string]*git.MockService{
		"/some/random/dir":             curDirMock,
		"/home/user/.devbox/myproject": projectMock,
	}

	mockFS := fs.NewMockFileSystem(t)
	mockFS.EXPECT().Getwd().Return("/some/random/dir", nil)

	m := New()
	m.listFn = func(filter string) []string {
		return []string{testProjectName}
	}
	m.loadFn = func(ctx context.Context, name string, profiles []string) (*project.Project, error) {
		return &project.Project{
			Project: &types.Project{
				Name:       testProjectName,
				WorkingDir: "/home/user/.devbox/myproject",
			},
		}, nil
	}
	m.gitFactory = mockGitFactory(t, pathMocks)
	m.fs = mockFS

	_, err := m.AutodetectProject(context.Background(), "")
	require.Error(t, err)
	assert.Equal(t, "cannot detect project: git remote does not match any known project", err.Error())
}

func TestAutodetectProject_Ambiguous(t *testing.T) {
	mockFS := fs.NewMockFileSystem(t)
	mockFS.EXPECT().Getwd().Return("/home/user/shared", nil)

	m := New()
	m.listFn = func(filter string) []string {
		return []string{"proj1", "proj2"}
	}
	m.loadFn = func(ctx context.Context, name string, profiles []string) (*project.Project, error) {
		if name == "proj1" {
			return &project.Project{
				Project: &types.Project{Name: "proj1"},
				LocalMounts: map[string]string{
					"./sources/shared": "/home/user/shared",
				},
			}, nil
		}
		return &project.Project{
			Project: &types.Project{Name: "proj2"},
			LocalMounts: map[string]string{
				"./sources/shared": "/home/user/shared",
			},
		}, nil
	}
	m.fs = mockFS

	_, err := m.AutodetectProject(context.Background(), "")
	require.Error(t, err)
	assert.Equal(t, "ambiguous project, please specify project name", err.Error())
}

// ============================================================================
// AutodetectSource tests
// ============================================================================

func TestAutodetectSource_ExplicitSource(t *testing.T) {
	proj := &project.Project{
		Project: &types.Project{
			Name:       testProjectName,
			WorkingDir: "/home/user/.devbox/myproject",
			Services: types.Services{
				"api": {
					Name: "api",
					Volumes: []types.ServiceVolumeConfig{
						{Source: "/home/user/.devbox/myproject/sources/backend"},
					},
				},
			},
		},
	}

	mockFS := fs.NewMockFileSystem(t)
	mockFS.EXPECT().Stat("/home/user/.devbox/myproject/sources/backend").
		Return(newDirFileInfo(t), nil)

	m := New()
	m.fs = mockFS

	result, err := m.AutodetectSource(context.Background(), proj, "./sources/backend", AutodetectSourceForMount)
	if err != nil {
		t.Fatalf("AutodetectSource() error = %v", err)
	}

	if len(result.Sources) != 1 || result.Sources[0] != "./sources/backend" {
		t.Errorf("AutodetectSource() sources = %v, want [./sources/backend]", result.Sources)
	}
	if len(result.AffectedServices) != 1 || result.AffectedServices[0] != "api" {
		t.Errorf("AutodetectSource() services = %v, want [api]", result.AffectedServices)
	}
	if result.LocalPath != "" {
		t.Errorf("AutodetectSource() localPath = %q, want empty", result.LocalPath)
	}
}

func TestAutodetectSource_ExplicitSource_NotFound(t *testing.T) {
	proj := &project.Project{
		Project: &types.Project{
			Name:       testProjectName,
			WorkingDir: "/home/user/.devbox/myproject",
		},
	}

	mockFS := fs.NewMockFileSystem(t)
	mockFS.EXPECT().Stat("/home/user/.devbox/myproject/sources/nonexistent").
		Return(nil, os.ErrNotExist)

	m := New()
	m.fs = mockFS

	_, err := m.AutodetectSource(context.Background(), proj, "./sources/nonexistent", AutodetectSourceForMount)
	if err == nil {
		t.Fatal("AutodetectSource() expected error, got nil")
	}
}

func TestAutodetectSource_InvalidSourcePath(t *testing.T) {
	proj := &project.Project{
		Project: &types.Project{
			Name:       testProjectName,
			WorkingDir: "/home/user/.devbox/myproject",
		},
	}

	m := New()

	_, err := m.AutodetectSource(context.Background(), proj, "invalid/path", AutodetectSourceForMount)
	if err == nil {
		t.Fatal("AutodetectSource() expected error, got nil")
	}
}

// ============================================================================
// Helper function tests
// ============================================================================

func TestMapKeys(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]bool
		want  int
	}{
		{"empty", map[string]bool{}, 0},
		{"single", map[string]bool{"a": true}, 1},
		{"multiple", map[string]bool{"a": true, "b": true, "c": true}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapKeys(tt.input)
			if len(got) != tt.want {
				t.Errorf("mapKeys() len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid alphanumeric", testProjectName, true},
		{"valid with numbers", "project123", true},
		{"valid with hyphen", "my-project", true},
		{"valid with underscore", "my_project", true},
		{"invalid with space", "my project", false},
		{"invalid with dot", "my.project", false},
		{"invalid with slash", "my/project", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validNameRegex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("validNameRegex.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================================
// Error handling tests
// ============================================================================

func TestAutodetectProject_GetCwdError(t *testing.T) {
	mockFS := fs.NewMockFileSystem(t)
	mockFS.EXPECT().Getwd().Return("", errors.New("permission denied"))

	m := New()
	m.listFn = func(filter string) []string {
		return []string{}
	}
	m.fs = mockFS

	_, err := m.AutodetectProject(context.Background(), "")
	require.Error(t, err)
	assert.Equal(t, "failed to get current directory: permission denied", err.Error())
}

func TestAutodetectProject_LoadError(t *testing.T) {
	m := New()
	m.listFn = func(filter string) []string {
		return []string{"broken"}
	}
	m.loadFn = func(ctx context.Context, name string, profiles []string) (*project.Project, error) {
		return nil, errors.New("invalid compose file")
	}

	_, err := m.AutodetectProject(context.Background(), "")
	require.Error(t, err)
	assert.Equal(t, "failed to load project: invalid compose file", err.Error())
}

// ============================================================================
// pathStartsWith tests
// ============================================================================

func TestPathStartsWith(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix string
		want   bool
	}{
		{"empty prefix", "foo/bar/baz", "", true},
		{"dot prefix", "foo/bar/baz", ".", true},
		{"exact match", "foo/bar", "foo/bar", true},
		{"prefix match", "foo/bar/baz", "foo/bar", true},
		{"single component", "foo/bar/baz", "foo", true},
		{"no match", "foo/bar/baz", "other", false},
		{"partial component no match", "foo/bar/baz", "fo", false},
		{"longer prefix than path", "foo", "foo/bar", false},
		{"case sensitive", "Foo/Bar/Baz", "foo/bar", false},
		{"trailing separator", "foo/bar/", "foo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathStartsWith(tt.path, tt.prefix)
			if got != tt.want {
				t.Errorf("pathStartsWith(%q, %q) = %v, want %v", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}

// ============================================================================
// detectSourceByGitRemote tests
// ============================================================================

func TestDetectSourceByGitRemote(t *testing.T) {
	tests := []struct {
		name         string
		proj         *project.Project
		purpose      AutodetectSourceType
		setupFS      func(t *testing.T) *fs.MockFileSystem
		setupGit     func(t *testing.T) map[string]*git.MockService
		wantSources  []string
		wantServices []string
		wantErr      string
	}{
		{
			name: "mount - source found via volume",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
					Services: types.Services{
						"api": {
							Name: "api",
							Volumes: []types.ServiceVolumeConfig{
								{Source: "/home/user/.devbox/myproject/sources/backend"},
							},
						},
					},
				},
				Sources: project.SourceConfigs{
					"backend": {URL: "https://github.com/company/backend.git"},
				},
			},
			purpose: AutodetectSourceForMount,
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Getwd().Return("/home/user/code/backend", nil)
				return mockFS
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().GetTopLevel(mock.Anything).Return("/home/user/code/backend", nil)
				m.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/backend.git", nil)
				return map[string]*git.MockService{"/home/user/code/backend": m}
			},
			wantSources:  []string{"./sources/backend"},
			wantServices: []string{"api"},
		},
		{
			name: "mount - source found via build context",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
					Services: types.Services{
						"api": {
							Name:  "api",
							Build: &types.BuildConfig{Context: "/home/user/.devbox/myproject/sources/backend"},
						},
					},
				},
				Sources: project.SourceConfigs{
					"backend": {URL: "https://github.com/company/backend.git"},
				},
			},
			purpose: AutodetectSourceForMount,
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Getwd().Return("/home/user/code/backend", nil)
				return mockFS
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().GetTopLevel(mock.Anything).Return("/home/user/code/backend", nil)
				m.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/backend.git", nil)
				return map[string]*git.MockService{"/home/user/code/backend": m}
			},
			wantSources:  []string{"./sources/backend"},
			wantServices: []string{"api"},
		},
		{
			name: "getwd error",
			proj: &project.Project{
				Project: &types.Project{Name: testProjectName},
			},
			purpose: AutodetectSourceForMount,
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Getwd().Return("", errors.New("permission denied"))
				return mockFS
			},
			setupGit: func(t *testing.T) map[string]*git.MockService { return nil },
			wantErr:  "failed to get current directory",
		},
		{
			name: "git toplevel error",
			proj: &project.Project{
				Project: &types.Project{Name: testProjectName},
			},
			purpose: AutodetectSourceForMount,
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Getwd().Return("/some/dir", nil)
				return mockFS
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().GetTopLevel(mock.Anything).Return("", errors.New("not a git repo"))
				return map[string]*git.MockService{"/some/dir": m}
			},
			wantErr: "failed to get git top-level directory",
		},
		{
			name: "no matching source",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
				Sources: project.SourceConfigs{
					"backend": {URL: "https://github.com/company/backend.git"},
				},
			},
			purpose: AutodetectSourceForMount,
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Getwd().Return("/home/user/code/other", nil)
				return mockFS
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().GetTopLevel(mock.Anything).Return("/home/user/code/other", nil)
				m.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/other.git", nil)
				return map[string]*git.MockService{"/home/user/code/other": m}
			},
			wantErr: "no services found using the detected source",
		},
		{
			name: "source already mounted",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
					Services: types.Services{
						"api": {
							Name: "api",
							Volumes: []types.ServiceVolumeConfig{
								{Source: "/home/user/code/backend"},
							},
						},
					},
				},
				Sources: project.SourceConfigs{
					"backend": {URL: "https://github.com/company/backend.git"},
				},
				LocalMounts: map[string]string{
					"./sources/backend": "/home/user/code/backend",
				},
			},
			purpose: AutodetectSourceForMount,
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Getwd().Return("/home/user/code/backend", nil)
				return mockFS
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().GetTopLevel(mock.Anything).Return("/home/user/code/backend", nil)
				m.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/backend.git", nil)
				return map[string]*git.MockService{"/home/user/code/backend": m}
			},
			wantErr: "source is already mounted",
		},
		{
			name: "umount - source found",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
					Services: types.Services{
						"api": {
							Name: "api",
							Volumes: []types.ServiceVolumeConfig{
								{Source: "/home/user/code/backend"},
							},
						},
					},
				},
				Sources: project.SourceConfigs{
					"backend": {URL: "https://github.com/company/backend.git"},
				},
				LocalMounts: map[string]string{
					"./sources/backend": "/home/user/code/backend",
				},
			},
			purpose: AutodetectSourceForUmount,
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Getwd().Return("/home/user/code/backend", nil)
				return mockFS
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().GetTopLevel(mock.Anything).Return("/home/user/code/backend", nil)
				m.EXPECT().GetRemote(mock.Anything).Return("https://github.com/company/backend.git", nil)
				return map[string]*git.MockService{"/home/user/code/backend": m}
			},
			wantSources:  []string{"./sources/backend"},
			wantServices: []string{"api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := tt.setupFS(t)
			pathMocks := tt.setupGit(t)

			m := New()
			m.gitFactory = mockGitFactory(t, pathMocks)
			m.fs = mockFS

			result, err := m.detectSourceByGitRemote(context.Background(), tt.proj, tt.purpose)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("detectSourceByGitRemote() expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Errorf("detectSourceByGitRemote() error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("detectSourceByGitRemote() unexpected error = %v", err)
			}

			if !slicesEqual(result.Sources, tt.wantSources) {
				t.Errorf("detectSourceByGitRemote() sources = %v, want %v", result.Sources, tt.wantSources)
			}
			if !slicesEqual(result.AffectedServices, tt.wantServices) {
				t.Errorf("detectSourceByGitRemote() services = %v, want %v", result.AffectedServices, tt.wantServices)
			}
		})
	}
}

// ============================================================================
// GetLocalMountCandidates tests
// ============================================================================

func TestGetLocalMountCandidates(t *testing.T) {
	tests := []struct {
		name   string
		proj   *project.Project
		filter string
		want   []string
	}{
		{
			name: "no sources",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
			},
			filter: "",
			want:   []string{},
		},
		{
			name: "source with volume",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
					Services: types.Services{
						"api": {
							Name: "api",
							Volumes: []types.ServiceVolumeConfig{
								{Type: "bind", Source: "/home/user/.devbox/myproject/sources/backend"},
							},
						},
					},
				},
				Sources: project.SourceConfigs{
					"backend": {URL: "https://github.com/company/backend.git"},
				},
			},
			filter: "",
			want:   []string{"./sources/backend"},
		},
		{
			name: "source with build context",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
					Services: types.Services{
						"api": {
							Name:  "api",
							Build: &types.BuildConfig{Context: "/home/user/.devbox/myproject/sources/backend"},
						},
					},
				},
				Sources: project.SourceConfigs{
					"backend": {URL: "https://github.com/company/backend.git"},
				},
			},
			filter: "",
			want:   []string{"./sources/backend"},
		},
		{
			name: "already mounted - excluded",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
					Services: types.Services{
						"api": {
							Name: "api",
							Volumes: []types.ServiceVolumeConfig{
								{Type: "bind", Source: "/home/user/.devbox/myproject/sources/backend"},
							},
						},
					},
				},
				Sources: project.SourceConfigs{
					"backend": {URL: "https://github.com/company/backend.git"},
				},
				LocalMounts: map[string]string{
					"./sources/backend": "/home/user/code/backend",
				},
			},
			filter: "",
			want:   []string{},
		},
		{
			name: "filter matches",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
					Services: types.Services{
						"api": {
							Name: "api",
							Volumes: []types.ServiceVolumeConfig{
								{Type: "bind", Source: "/home/user/.devbox/myproject/sources/backend"},
								{Type: "bind", Source: "/home/user/.devbox/myproject/sources/frontend"},
							},
						},
					},
				},
				Sources: project.SourceConfigs{
					"backend":  {URL: "https://github.com/company/backend.git"},
					"frontend": {URL: "https://github.com/company/frontend.git"},
				},
			},
			filter: "back",
			want:   []string{"./sources/backend"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			got := m.GetLocalMountCandidates(tt.proj, tt.filter)

			if !slicesEqual(got, tt.want) {
				t.Errorf("GetLocalMountCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// GetLocalMounts tests
// ============================================================================

func TestGetLocalMounts(t *testing.T) {
	tests := []struct {
		name   string
		proj   *project.Project
		filter string
		want   []string
	}{
		{
			name: "no mounts",
			proj: &project.Project{
				Project:     &types.Project{Name: testProjectName},
				LocalMounts: map[string]string{},
			},
			filter: "",
			want:   []string{},
		},
		{
			name: "single mount",
			proj: &project.Project{
				Project: &types.Project{Name: testProjectName},
				LocalMounts: map[string]string{
					"./sources/backend": "/home/user/code/backend",
				},
			},
			filter: "",
			want:   []string{"./sources/backend"},
		},
		{
			name: "multiple mounts",
			proj: &project.Project{
				Project: &types.Project{Name: testProjectName},
				LocalMounts: map[string]string{
					"./sources/backend":  "/home/user/code/backend",
					"./sources/frontend": "/home/user/code/frontend",
				},
			},
			filter: "",
			want:   []string{"./sources/backend", "./sources/frontend"},
		},
		{
			name: "filter matches",
			proj: &project.Project{
				Project: &types.Project{Name: testProjectName},
				LocalMounts: map[string]string{
					"./sources/backend":  "/home/user/code/backend",
					"./sources/frontend": "/home/user/code/frontend",
				},
			},
			filter: "front",
			want:   []string{"./sources/frontend"},
		},
		{
			name: "filter no match",
			proj: &project.Project{
				Project: &types.Project{Name: testProjectName},
				LocalMounts: map[string]string{
					"./sources/backend": "/home/user/code/backend",
				},
			},
			filter: "other",
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			got := m.GetLocalMounts(tt.proj, tt.filter)

			if !slicesEqualUnordered(got, tt.want) {
				t.Errorf("GetLocalMounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// Init tests
// ============================================================================

func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		projName string
		url      string
		branch   string
		setupFS  func(t *testing.T) *fs.MockFileSystem
		setupGit func(t *testing.T) map[string]*git.MockService
		wantErr  string
	}{
		{
			name:     "invalid project name",
			projName: "my.project",
			url:      "https://github.com/company/devbox.git",
			branch:   "main",
			setupFS:  func(t *testing.T) *fs.MockFileSystem { return nil },
			setupGit: func(t *testing.T) map[string]*git.MockService { return nil },
			wantErr:  "invalid project name",
		},
		{
			name:     "project already exists",
			projName: "myproject",
			url:      "https://github.com/company/devbox.git",
			branch:   "main",
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Stat(mock.Anything).Return(newDirFileInfo(t), nil)
				return mockFS
			},
			setupGit: func(t *testing.T) map[string]*git.MockService { return nil },
			wantErr:  "project already exists",
		},
		{
			name:     "clone error",
			projName: "myproject",
			url:      "https://github.com/company/devbox.git",
			branch:   "main",
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Stat(mock.Anything).Return(nil, os.ErrNotExist)
				mockFS.EXPECT().RemoveAll(mock.Anything).Return(nil)
				return mockFS
			},
			setupGit: func(t *testing.T) map[string]*git.MockService {
				m := git.NewMockService(t)
				m.EXPECT().Clone(mock.Anything, "https://github.com/company/devbox.git", "main").
					Return(errors.New("clone failed"))
				return map[string]*git.MockService{mock.Anything: m}
			},
			wantErr: "failed to clone git repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := tt.setupFS(t)
			pathMocks := tt.setupGit(t)

			var gitFactory gitServiceFactory
			if pathMocks != nil {
				gitFactory = func(path string) git.Service {
					for _, m := range pathMocks {
						return m
					}
					return git.NewMockService(t)
				}
			}

			m := New()
			m.gitFactory = gitFactory
			m.fs = mockFS

			err := m.Init(context.Background(), tt.projName, tt.url, tt.branch)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Init() expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Errorf("Init() error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Init() unexpected error = %v", err)
			}
		})
	}
}

// ============================================================================
// Destroy tests
// ============================================================================

func TestDestroy(t *testing.T) {
	tests := []struct {
		name    string
		proj    *project.Project
		setupFS func(t *testing.T) *fs.MockFileSystem
		wantErr string
	}{
		{
			name: "project does not exist",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
			},
			setupFS: func(t *testing.T) *fs.MockFileSystem {
				mockFS := fs.NewMockFileSystem(t)
				mockFS.EXPECT().Stat("/home/user/.devbox/myproject").Return(nil, os.ErrNotExist)
				return mockFS
			},
			wantErr: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := tt.setupFS(t)
			m := New()
			m.fs = mockFS

			err := m.Destroy(context.Background(), tt.proj)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Destroy() expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Errorf("Destroy() error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Destroy() unexpected error = %v", err)
			}
		})
	}
}

// ============================================================================
// isSourceInWrongState tests
// ============================================================================

func TestIsSourceInWrongState(t *testing.T) {
	tests := []struct {
		name         string
		sourcePrefix string
		servicePath  string
		cwdRelPath   string
		proj         *project.Project
		purpose      AutodetectSourceType
		want         bool
	}{
		{
			name:         "mount - service path is mounted local path",
			sourcePrefix: "/home/user/.devbox/myproject/sources/backend",
			servicePath:  "/home/user/code/backend",
			cwdRelPath:   ".",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
				LocalMounts: map[string]string{
					"./sources/backend": "/home/user/code/backend",
				},
			},
			purpose: AutodetectSourceForMount,
			want:    true,
		},
		{
			name:         "mount - service path is not mounted",
			sourcePrefix: "/home/user/.devbox/myproject/sources/backend",
			servicePath:  "/home/user/.devbox/myproject/sources/backend",
			cwdRelPath:   ".",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
				LocalMounts: map[string]string{},
			},
			purpose: AutodetectSourceForMount,
			want:    false,
		},
		{
			name:         "umount - source is not mounted",
			sourcePrefix: "/home/user/.devbox/myproject/sources/backend",
			servicePath:  "/home/user/.devbox/myproject/sources/backend",
			cwdRelPath:   ".",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
				LocalMounts: map[string]string{},
			},
			purpose: AutodetectSourceForUmount,
			want:    true,
		},
		{
			name:         "umount - source is mounted",
			sourcePrefix: "/home/user/.devbox/myproject/sources/backend",
			servicePath:  "/home/user/code/backend",
			cwdRelPath:   ".",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
				LocalMounts: map[string]string{
					"./sources/backend": "/home/user/code/backend",
				},
			},
			purpose: AutodetectSourceForUmount,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			got := m.isSourceInWrongState(tt.sourcePrefix, tt.servicePath, tt.cwdRelPath, tt.proj, tt.purpose)

			if got != tt.want {
				t.Errorf("isSourceInWrongState() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// matchSourcePath tests
// ============================================================================

func TestMatchSourcePath(t *testing.T) {
	tests := []struct {
		name          string
		sourcePrefix  string
		servicePath   string
		cwdRelPath    string
		toplevelDir   string
		proj          *project.Project
		purpose       AutodetectSourceType
		wantRelSource string
		wantLocalPath string
		wantOk        bool
	}{
		{
			name:         "mount - direct match",
			sourcePrefix: "/home/user/.devbox/myproject/sources/backend",
			servicePath:  "/home/user/.devbox/myproject/sources/backend",
			cwdRelPath:   ".",
			toplevelDir:  "/home/user/code/backend",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
			},
			purpose:       AutodetectSourceForMount,
			wantRelSource: "./sources/backend",
			wantLocalPath: "/home/user/code/backend",
			wantOk:        true,
		},
		{
			name:         "mount - subpath match",
			sourcePrefix: "/home/user/.devbox/myproject/sources/backend",
			servicePath:  "/home/user/.devbox/myproject/sources/backend/cmd/api",
			cwdRelPath:   "cmd/api",
			toplevelDir:  "/home/user/code/backend",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
			},
			purpose:       AutodetectSourceForMount,
			wantRelSource: "./sources/backend/cmd/api",
			wantLocalPath: "/home/user/code/backend/cmd/api",
			wantOk:        true,
		},
		{
			name:         "mount - service path does not match source",
			sourcePrefix: "/home/user/.devbox/myproject/sources/backend",
			servicePath:  "/home/user/.devbox/myproject/sources/frontend",
			cwdRelPath:   ".",
			toplevelDir:  "/home/user/code/backend",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
			},
			purpose:       AutodetectSourceForMount,
			wantRelSource: "",
			wantLocalPath: "",
			wantOk:        false,
		},
		{
			name:         "umount - mounted source found",
			sourcePrefix: "/home/user/.devbox/myproject/sources/backend",
			servicePath:  "/home/user/code/backend",
			cwdRelPath:   ".",
			toplevelDir:  "/home/user/code/backend",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
				LocalMounts: map[string]string{
					"./sources/backend": "/home/user/code/backend",
				},
			},
			purpose:       AutodetectSourceForUmount,
			wantRelSource: "./sources/backend",
			wantLocalPath: "/home/user/code/backend",
			wantOk:        true,
		},
		{
			name:         "umount - source not mounted",
			sourcePrefix: "/home/user/.devbox/myproject/sources/backend",
			servicePath:  "/home/user/.devbox/myproject/sources/backend",
			cwdRelPath:   ".",
			toplevelDir:  "/home/user/code/backend",
			proj: &project.Project{
				Project: &types.Project{
					Name:       testProjectName,
					WorkingDir: "/home/user/.devbox/myproject",
				},
				LocalMounts: map[string]string{},
			},
			purpose:       AutodetectSourceForUmount,
			wantRelSource: "",
			wantLocalPath: "",
			wantOk:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			relSource, localPath, ok := m.matchSourcePath(tt.sourcePrefix, tt.servicePath, tt.cwdRelPath, tt.toplevelDir, tt.proj, tt.purpose)

			if ok != tt.wantOk {
				t.Errorf("matchSourcePath() ok = %v, want %v", ok, tt.wantOk)
			}
			if relSource != tt.wantRelSource {
				t.Errorf("matchSourcePath() relSource = %q, want %q", relSource, tt.wantRelSource)
			}
			if localPath != tt.wantLocalPath {
				t.Errorf("matchSourcePath() localPath = %q, want %q", localPath, tt.wantLocalPath)
			}
		})
	}
}

// ============================================================================
// Test helpers
// ============================================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func slicesEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]int)
	for _, v := range a {
		aMap[v]++
	}
	for _, v := range b {
		if aMap[v] == 0 {
			return false
		}
		aMap[v]--
	}
	return true
}
