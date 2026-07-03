package project

import (
	"path/filepath"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestNestedForeignMountExcludes(t *testing.T) {
	root := "/work"
	sourcesRoot := filepath.Join(root, "sources")

	bind := func(source, target string) types.ServiceVolumeConfig {
		return types.ServiceVolumeConfig{Type: "bind", Source: source, Target: target}
	}

	tests := []struct {
		name     string
		services types.Services
		want     map[string][]string
	}{
		{
			name: "cross-source nested dir mount is excluded, .env file mount is not",
			services: types.Services{
				"service-a": {
					Volumes: []types.ServiceVolumeConfig{
						bind(filepath.Join(sourcesRoot, "service-a", "cmd", "service-a"), "/app"),
						bind(filepath.Join(root, "envs", "service-a", ".env"), "/app/.env"),
						bind(filepath.Join(sourcesRoot, "shared"), "/app/shared"),
					},
				},
			},
			want: map[string][]string{
				"service-a": {"cmd/service-a/shared"},
			},
		},
		{
			name: "self-nested same source is not excluded",
			services: types.Services{
				"service-b": {
					Volumes: []types.ServiceVolumeConfig{
						bind(filepath.Join(sourcesRoot, "service-b"), "/app"),
						bind(filepath.Join(sourcesRoot, "service-b", "cmd", "service-b"), "/app/cmd/service-b"),
					},
				},
			},
			want: map[string][]string{},
		},
		{
			name: "file/config mount only is not excluded",
			services: types.Services{
				"svc": {
					Volumes: []types.ServiceVolumeConfig{
						bind(filepath.Join(sourcesRoot, "svc"), "/app"),
						bind(filepath.Join(root, "envs", "svc", ".env"), "/app/.env"),
					},
				},
			},
			want: map[string][]string{},
		},
		{
			name: "locally-mounted parent (source outside sources/) is not excluded",
			services: types.Services{
				"service-a": {
					Volumes: []types.ServiceVolumeConfig{
						bind("/Users/me/local/service-a", "/app"),
						bind(filepath.Join(sourcesRoot, "shared"), "/app/shared"),
					},
				},
			},
			want: map[string][]string{},
		},
		{
			name: "multiple cross-source children under one source are sorted",
			services: types.Services{
				"foo": {
					Volumes: []types.ServiceVolumeConfig{
						bind(filepath.Join(sourcesRoot, "foo"), "/app"),
						bind(filepath.Join(sourcesRoot, "zebra"), "/app/zebra"),
						bind(filepath.Join(sourcesRoot, "alpha"), "/app/alpha"),
					},
				},
			},
			want: map[string][]string{
				"foo": {"alpha", "zebra"},
			},
		},
		{
			name: "two services feeding one source key are merged, sorted and deduped",
			services: types.Services{
				"svc-a": {
					Volumes: []types.ServiceVolumeConfig{
						bind(filepath.Join(sourcesRoot, "foo"), "/app"),
						bind(filepath.Join(sourcesRoot, "shared"), "/app/shared"),
						bind(filepath.Join(sourcesRoot, "extra"), "/app/extra"),
					},
				},
				"svc-b": {
					Volumes: []types.ServiceVolumeConfig{
						bind(filepath.Join(sourcesRoot, "foo"), "/app"),
						bind(filepath.Join(sourcesRoot, "shared"), "/app/shared"),
					},
				},
			},
			want: map[string][]string{
				"foo": {"extra", "shared"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nestedForeignMountExcludes(tt.services, sourcesRoot)
			assert.Equal(t, tt.want, got)
		})
	}
}
