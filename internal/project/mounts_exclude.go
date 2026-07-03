package project

import (
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"

	"github.com/pilat/devbox/internal/app"
)

// SourceCleanExcludes returns, per source name, the repo-relative paths `git clean` must skip
// because Docker mounts a different source into them as a live nested bind mount. Keys match the
// on-disk sources/<name> segment; values carry no leading slash.
func (p *Project) SourceCleanExcludes() map[string][]string {
	sourcesRoot := filepath.Join(p.WorkingDir, app.SourcesDir)

	return nestedForeignMountExcludes(p.Services, sourcesRoot)
}

// nestedForeignMountExcludes derives cross-source nested-mount excludes from the compose volume
// topology: a bind volume (child) of a source is mounted nested inside a same-service parent bind
// whose host source is a *different* source. The different-source discriminator keeps self-nested
// and file/config mounts out, so clean is never weakened where it should still run.
func nestedForeignMountExcludes(services types.Services, sourcesRoot string) map[string][]string {
	type bindVol struct {
		source string
		target string
	}

	result := make(map[string][]string)

	for _, service := range services {
		binds := make([]bindVol, 0, len(service.Volumes))
		for _, v := range service.Volumes {
			if v.Type != "bind" {
				continue
			}
			binds = append(binds, bindVol{source: v.Source, target: path.Clean(v.Target)})
		}

		for _, child := range binds {
			childSrc, ok := sourceSegment(child.source, sourcesRoot)
			if !ok {
				continue
			}

			var parent *bindVol
			for i := range binds {
				cand := &binds[i]
				if !strings.HasPrefix(child.target, cand.target+"/") {
					continue
				}
				if parent == nil || len(cand.target) > len(parent.target) {
					parent = cand
				}
			}
			if parent == nil {
				continue
			}

			parentSrc, ok := sourceSegment(parent.source, sourcesRoot)
			if !ok {
				continue
			}

			if childSrc == parentSrc {
				continue
			}

			childRel := strings.TrimPrefix(strings.TrimPrefix(child.target, parent.target), "/")
			hostMountpoint := filepath.Join(parent.source, childRel)
			repoRel, err := filepath.Rel(filepath.Join(sourcesRoot, parentSrc), hostMountpoint)
			if err != nil {
				continue
			}

			result[parentSrc] = append(result[parentSrc], repoRel)
		}
	}

	for src, excludes := range result {
		slices.Sort(excludes)
		result[src] = slices.Compact(excludes)
	}

	return result
}

// sourceSegment reports the sources/<name> segment of an absolute host path, if it lives under
// sourcesRoot. It returns false for anything outside sourcesRoot (envs/, named volumes, or a
// LocalMounts-rewritten local path).
func sourceSegment(hostPath, sourcesRoot string) (string, bool) {
	prefix := sourcesRoot + string(filepath.Separator)
	if !strings.HasPrefix(hostPath, prefix) {
		return "", false
	}

	seg, _, _ := strings.Cut(hostPath[len(prefix):], string(filepath.Separator))
	if seg == "" {
		return "", false
	}

	return seg, true
}
