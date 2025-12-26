package manager

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/git"
	"github.com/pilat/devbox/internal/pkg/fs"
	"github.com/pilat/devbox/internal/project"
)

var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type AutodetectSourceType int

const (
	AutodetectSourceForMount AutodetectSourceType = iota
	AutodetectSourceForUmount
)

// SourceDetectionResult contains the result of source autodetection.
type SourceDetectionResult struct {
	Sources          []string // Detected source paths (e.g., "./sources/backend")
	AffectedServices []string // Services that use these sources
	LocalPath        string   // Path to mount from (empty if using cwd)
}

type gitServiceFactory func(path string) git.Service

type Manager struct {
	gitFactory gitServiceFactory
	fs         fs.FileSystem
	listFn     func(filter string) []string
	loadFn     func(ctx context.Context, name string, profiles []string) (*project.Project, error)
}

// New creates a Manager with default implementations.
func New() *Manager {
	m := &Manager{
		gitFactory: func(path string) git.Service { return git.New(path) },
		fs:         fs.New(),
	}

	m.listFn = m.list
	m.loadFn = m.load

	return m
}

// AutodetectProject validates project name (if provided) and tries to autodetect the project by comparing
// the current directory with the project sources and local mounts. If not successful or ambiguous, it returns an error.
func (m *Manager) AutodetectProject(ctx context.Context, name string) (*project.Project, error) {
	// Load projects
	projects, err := m.loadAllProjects(ctx)
	if err != nil {
		return nil, err
	}

	// If project name was provided we check it and immediately return if found
	if name != "" {
		for _, p := range projects {
			if p.Name == name {
				return p, nil
			}
		}
		return nil, fmt.Errorf("project '%s' not found", name)
	}

	// For autodetection, we need the current directory
	curDir, err := m.fs.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// 1. Check if the current dir is a local mount of one project
	if p, ambiguous := m.detectByLocalMount(curDir, projects); p != nil {
		if ambiguous {
			return nil, fmt.Errorf("ambiguous project, please specify project name")
		}
		return p, nil
	}

	// 2. Check if current dir matches a project source by git remote
	curDirGit := m.gitFactory(curDir)
	remoteURL, err := curDirGit.GetRemote(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot detect project: not a git repository")
	}
	remoteURL = git.NormalizeURL(remoteURL)

	if p, ambiguous := m.detectBySourceRemote(remoteURL, projects); p != nil {
		if ambiguous {
			return nil, fmt.Errorf("ambiguous project, please specify project name")
		}
		return p, nil
	}

	// 3. Check if the current dir is the project's own manifest repository
	if p, ambiguous := m.detectByProjectRepo(ctx, remoteURL, projects); p != nil {
		if ambiguous {
			return nil, fmt.Errorf("ambiguous project, please specify project name")
		}
		return p, nil
	}

	return nil, fmt.Errorf("cannot detect project: git remote does not match any known project")
}

// loadAllProjects loads all available projects.
func (m *Manager) loadAllProjects(ctx context.Context) ([]*project.Project, error) {
	projectNames := m.listFn("")
	projects := make([]*project.Project, 0, len(projectNames))

	for _, projectName := range projectNames {
		p, err := m.loadFn(ctx, projectName, []string{"*"})
		if err != nil {
			return nil, fmt.Errorf("failed to load project: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// detectByLocalMount checks if curDir matches any project's local mount.
// Returns the project and whether the match is ambiguous.
func (m *Manager) detectByLocalMount(curDir string, projects []*project.Project) (*project.Project, bool) {
	var found *project.Project
	ambiguous := false

	for _, p := range projects {
		for _, mountPath := range p.LocalMounts {
			if mountPath == curDir {
				if found != nil && found != p {
					ambiguous = true
				}
				found = p
			}
		}
	}

	return found, ambiguous
}

// detectBySourceRemote checks if the current git remote matches any project source URL.
func (m *Manager) detectBySourceRemote(remoteURL string, projects []*project.Project) (*project.Project, bool) {
	var found *project.Project
	ambiguous := false

	for _, p := range projects {
		for _, source := range p.Sources {
			sourceURL := git.NormalizeURL(source.URL)
			if remoteURL != sourceURL {
				continue
			}

			if found != nil && found != p {
				ambiguous = true
			}
			found = p
		}
	}

	return found, ambiguous
}

// detectByProjectRepo checks if the current git remote matches the project's manifest repository.
func (m *Manager) detectByProjectRepo(ctx context.Context, remoteURL string, projects []*project.Project) (*project.Project, bool) {
	var found *project.Project
	ambiguous := false

	for _, p := range projects {
		projectRemoteURL, err := m.gitFactory(p.WorkingDir).GetRemote(ctx)
		if err != nil {
			continue
		}
		projectRemoteURL = git.NormalizeURL(projectRemoteURL)

		if remoteURL != projectRemoteURL {
			continue
		}

		if found != nil && found != p {
			ambiguous = true
		}
		found = p
	}

	return found, ambiguous
}

// AutodetectSource detects sources and affected services based on current directory or explicit selection.
func (m *Manager) AutodetectSource(ctx context.Context, proj *project.Project, sourceNameSel string, purpose AutodetectSourceType) (*SourceDetectionResult, error) {
	if sourceNameSel != "" {
		return m.detectExplicitSource(proj, sourceNameSel, purpose)
	}
	return m.detectSourceByGitRemote(ctx, proj, purpose)
}

// detectExplicitSource handles the case when a source name is explicitly provided.
func (m *Manager) detectExplicitSource(proj *project.Project, sourceNameSel string, purpose AutodetectSourceType) (*SourceDetectionResult, error) {
	if !strings.HasPrefix(sourceNameSel, "./"+app.SourcesDir+"/") {
		return nil, fmt.Errorf("source '%s' not found", sourceNameSel)
	}

	sourcePath := filepath.Join(proj.WorkingDir, sourceNameSel)
	if fstat, err := m.fs.Stat(sourcePath); err != nil || !fstat.IsDir() {
		return nil, fmt.Errorf("source '%s' not found", sourceNameSel)
	}

	if purpose == AutodetectSourceForUmount {
		alt, ok := proj.LocalMounts[sourceNameSel]
		if !ok {
			return nil, fmt.Errorf("source '%s' not found", sourceNameSel)
		}
		sourcePath = alt
	}

	affectedServices := make(map[string]bool)
	for _, service := range proj.Services {
		for _, volume := range service.Volumes {
			if volume.Source != sourcePath {
				continue
			}
			affectedServices[service.Name] = true
		}
	}

	if len(affectedServices) == 0 {
		return nil, fmt.Errorf("no services found using the detected source")
	}

	return &SourceDetectionResult{
		Sources:          []string{sourceNameSel},
		AffectedServices: mapKeys(affectedServices),
	}, nil
}

// detectSourceByGitRemote autodetects sources by matching git remote URLs.
func (m *Manager) detectSourceByGitRemote(ctx context.Context, proj *project.Project, purpose AutodetectSourceType) (*SourceDetectionResult, error) {
	curDir, err := m.fs.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	curDirGit := m.gitFactory(curDir)
	toplevelDir, err := curDirGit.GetTopLevel(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get git top-level directory: %w", err)
	}

	remoteURL, err := curDirGit.GetRemote(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote url: %w", err)
	}
	remoteURL = git.NormalizeURL(remoteURL)

	relativePath, err := filepath.Rel(toplevelDir, curDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	sources := make(map[string]bool)
	affectedServices := make(map[string]bool)
	var localPath string
	var foundButWrongState bool

	for sourceName, source := range proj.Sources {
		sourceURL := git.NormalizeURL(source.URL)
		if remoteURL != sourceURL {
			continue
		}

		sourcePrefix := filepath.Join(proj.WorkingDir, app.SourcesDir, sourceName)

		for _, service := range proj.Services {
			if service.Build != nil {
				if relSource, lp, ok := m.matchSourcePath(sourcePrefix, service.Build.Context, relativePath, toplevelDir, proj, purpose); ok {
					sources[relSource] = true
					affectedServices[service.Name] = true
					localPath = lp
				} else if m.isSourceInWrongState(sourcePrefix, service.Build.Context, relativePath, proj, purpose) {
					foundButWrongState = true
				}
			}

			for _, volume := range service.Volumes {
				if relSource, lp, ok := m.matchSourcePath(sourcePrefix, volume.Source, relativePath, toplevelDir, proj, purpose); ok {
					sources[relSource] = true
					affectedServices[service.Name] = true
					localPath = lp
				} else if m.isSourceInWrongState(sourcePrefix, volume.Source, relativePath, proj, purpose) {
					foundButWrongState = true
				}
			}
		}
	}

	if len(sources) == 0 {
		if foundButWrongState {
			if purpose == AutodetectSourceForMount {
				return nil, fmt.Errorf("source is already mounted")
			}
			return nil, fmt.Errorf("source is not mounted")
		}
		return nil, fmt.Errorf("no services found using the detected source")
	}

	return &SourceDetectionResult{
		Sources:          mapKeys(sources),
		AffectedServices: mapKeys(affectedServices),
		LocalPath:        localPath,
	}, nil
}

// isSourceInWrongState checks if a source path matches but is in the wrong mount state.
func (m *Manager) isSourceInWrongState(sourcePrefix, servicePath, cwdRelPath string, proj *project.Project, purpose AutodetectSourceType) bool {
	if purpose == AutodetectSourceForMount {
		// For mount: check if servicePath is a mounted local path
		for origPath, localPath := range proj.LocalMounts {
			if servicePath == localPath {
				fullOrigPath := filepath.Join(proj.WorkingDir, origPath)
				if strings.HasPrefix(fullOrigPath, sourcePrefix) {
					serviceSubpath := strings.TrimPrefix(fullOrigPath, sourcePrefix)
					serviceSubpath = strings.TrimPrefix(serviceSubpath, string(filepath.Separator))
					if cwdMatchesServiceSubpath(cwdRelPath, serviceSubpath) {
						return true
					}
				}
			}
		}
	} else {
		// For unmount: check if servicePath is an unmounted source path
		if strings.HasPrefix(servicePath, sourcePrefix) {
			serviceSubpath := strings.TrimPrefix(servicePath, sourcePrefix)
			serviceSubpath = strings.TrimPrefix(serviceSubpath, string(filepath.Separator))
			if cwdMatchesServiceSubpath(cwdRelPath, serviceSubpath) {
				relSourcePath, err := filepath.Rel(proj.WorkingDir, servicePath)
				if err != nil {
					return false
				}
				relSourcePath = "./" + relSourcePath
				if _, ok := proj.LocalMounts[relSourcePath]; !ok {
					return true
				}
			}
		}
	}
	return false
}

// matchSourcePath checks if the cwd's relative path is inside the service's source path.
// Returns the relative source path, the local path to mount, and true if matched.
func (m *Manager) matchSourcePath(sourcePrefix, servicePath, cwdRelPath, toplevelDir string, proj *project.Project, purpose AutodetectSourceType) (string, string, bool) {
	var relSourcePath string
	var serviceSubpath string

	if strings.HasPrefix(servicePath, sourcePrefix) {
		// Service path is the original source path (not mounted)
		serviceSubpath = strings.TrimPrefix(servicePath, sourcePrefix)
		serviceSubpath = strings.TrimPrefix(serviceSubpath, string(filepath.Separator))

		var err error
		relSourcePath, err = filepath.Rel(proj.WorkingDir, servicePath)
		if err != nil {
			return "", "", false
		}
		relSourcePath = "./" + relSourcePath
	} else if purpose == AutodetectSourceForUmount {
		// For unmount: servicePath might be a local mount path, reverse lookup original source
		for origPath, localPath := range proj.LocalMounts {
			if servicePath == localPath && strings.HasPrefix(origPath, "./"+app.SourcesDir+"/") {
				// Check if this mount belongs to the current source
				fullOrigPath := filepath.Join(proj.WorkingDir, origPath)
				if strings.HasPrefix(fullOrigPath, sourcePrefix) {
					serviceSubpath = strings.TrimPrefix(fullOrigPath, sourcePrefix)
					serviceSubpath = strings.TrimPrefix(serviceSubpath, string(filepath.Separator))
					relSourcePath = origPath
					break
				}
			}
		}
		if relSourcePath == "" {
			return "", "", false
		}
	} else {
		return "", "", false
	}

	// Check if cwd matches the service's source path:
	// - If service uses root (serviceSubpath=""), cwd must also be at root
	// - If service uses subpath (e.g., "cmd/risk-engine"), cwd must be inside it
	if !cwdMatchesServiceSubpath(cwdRelPath, serviceSubpath) {
		return "", "", false
	}

	if purpose == AutodetectSourceForUmount {
		if _, ok := proj.LocalMounts[relSourcePath]; !ok {
			return "", "", false
		}
	}

	// Calculate the local path: <toplevel>/<serviceSubpath>
	localPath := toplevelDir
	if serviceSubpath != "" {
		localPath = filepath.Join(toplevelDir, serviceSubpath)
	}

	return relSourcePath, localPath, true
}

// cwdMatchesServiceSubpath checks if the current working directory matches the service's subpath.
// If serviceSubpath is empty (service uses root), cwd must also be at root.
// If serviceSubpath is not empty, cwd must be inside it.
func cwdMatchesServiceSubpath(cwdRelPath, serviceSubpath string) bool {
	if serviceSubpath == "" {
		return cwdRelPath == "" || cwdRelPath == "."
	}
	return pathStartsWith(cwdRelPath, serviceSubpath)
}

// pathStartsWith checks if path starts with prefix (using path components, not string prefix).
func pathStartsWith(path, prefix string) bool {
	if prefix == "." || prefix == "" {
		return true
	}
	pathParts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	prefixParts := strings.Split(filepath.Clean(prefix), string(filepath.Separator))

	if len(pathParts) < len(prefixParts) {
		return false
	}

	for i, part := range prefixParts {
		if pathParts[i] != part {
			return false
		}
	}
	return true
}

func (m *Manager) List(filter string) []string {
	return m.listFn(filter)
}

func (m *Manager) Load(ctx context.Context, name string, profiles []string) (*project.Project, error) {
	return m.loadFn(ctx, name, profiles)
}

func (m *Manager) list(filter string) []string {
	folders, err := m.fs.ReadDir(app.AppDir)
	if err != nil {
		return []string{}
	}

	results := make([]string, 0, len(folders))
	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}

		name := folder.Name()
		if filter != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(filter)) {
			continue
		}

		results = append(results, name)
	}

	return results
}

func (m *Manager) load(ctx context.Context, name string, profiles []string) (*project.Project, error) {
	return project.New(ctx, name, profiles)
}

func (m *Manager) Init(ctx context.Context, name, url, branch string) error {
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("invalid project name: %s", name)
	}

	projectFolder := filepath.Join(app.AppDir, name)

	if _, err := m.fs.Stat(projectFolder); err == nil {
		return fmt.Errorf("project already exists")
	}

	cleanup := func() {
		_ = m.fs.RemoveAll(projectFolder)
	}

	g := m.gitFactory(projectFolder)
	if err := g.Clone(ctx, url, branch); err != nil {
		cleanup()
		return fmt.Errorf("failed to clone git repo: %w", err)
	}

	patterns := []string{
		fmt.Sprintf("/%s/", app.SourcesDir),
		fmt.Sprintf("/%s", app.StateFile),
		fmt.Sprintf("/%s", app.EnvFile),
	}

	if err := g.SetLocalExclude(patterns); err != nil {
		cleanup()
		return fmt.Errorf("failed to set local exclude: %w", err)
	}

	for k, content := range map[string]string{
		app.EnvFile:   "",
		app.StateFile: "{}",
	} {
		if err := m.fs.WriteFile(filepath.Join(projectFolder, k), []byte(content), 0644); err != nil {
			cleanup()
			return fmt.Errorf("failed to create file: %w", err)
		}
	}

	return nil
}

func (m *Manager) Destroy(ctx context.Context, proj *project.Project) error {
	if _, err := m.fs.Stat(proj.WorkingDir); err != nil {
		return fmt.Errorf("project %s does not exist", proj.Name)
	}

	if err := m.fs.RemoveAll(proj.WorkingDir); err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}

	return nil
}

func (m *Manager) GetLocalMountCandidates(proj *project.Project, filter string) []string {
	filter = strings.ToLower(filter)
	results := make(map[string]bool)

	for sourceName := range proj.Sources {
		sourcePath := filepath.Join(proj.WorkingDir, app.SourcesDir, sourceName)

		for _, service := range proj.Services {
			if service.Build != nil && strings.HasPrefix(service.Build.Context, sourcePath) {
				relSourcePath, err := filepath.Rel(proj.WorkingDir, service.Build.Context)
				if err != nil {
					continue
				}
				relSourcePath = "./" + relSourcePath

				if _, alreadyMounted := proj.LocalMounts[relSourcePath]; !alreadyMounted {
					results[relSourcePath] = true
				}
			}

			for _, volume := range service.Volumes {
				if volume.Type == "bind" && strings.HasPrefix(volume.Source, sourcePath) {
					relSourcePath, err := filepath.Rel(proj.WorkingDir, volume.Source)
					if err != nil {
						continue
					}
					relSourcePath = "./" + relSourcePath

					if _, alreadyMounted := proj.LocalMounts[relSourcePath]; !alreadyMounted {
						results[relSourcePath] = true
					}
				}
			}
		}
	}

	filtered := make([]string, 0, len(results))
	for k := range results {
		if filter == "" || strings.Contains(strings.ToLower(k), filter) {
			filtered = append(filtered, k)
		}
	}

	sort.Strings(filtered)
	return filtered
}

func (m *Manager) GetLocalMounts(proj *project.Project, filter string) []string {
	filter = strings.ToLower(filter)

	results := make([]string, 0, len(proj.LocalMounts))
	for k := range proj.LocalMounts {
		if filter == "" || strings.Contains(strings.ToLower(k), filter) {
			results = append(results, k)
		}
	}

	sort.Strings(results)
	return results
}

func mapKeys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
