package depgraph_test

import (
	"context"
	"testing"

	"github.com/pilat/devbox/internal/pkg/depgraph"
	"github.com/stretchr/testify/assert"
)

// MockEntity is a mock implementation of the depgraph.Entity interface for testing.
type MockEntity struct {
	ref  string
	deps []string
}

func (e MockEntity) Ref() string {
	return e.ref
}

func (e MockEntity) DependsOn() []string {
	return e.deps
}

func (e MockEntity) Run(context.Context) error {
	return nil
}

func TestBuildDependencyGraph(t *testing.T) {
	tests := []struct {
		name      string
		entities  []depgraph.Entity
		expected  [][]string // Expected output, using entity IDs for simplicity
		expectErr bool       // Whether an error is expected
	}{
		{
			name: "Case 2",
			entities: []depgraph.Entity{
				MockEntity{ref: "localhost/base:latest", deps: []string{}},
				MockEntity{ref: "localhost/provisioner5:latest", deps: []string{"localhost/base:latest"}},
				MockEntity{ref: "localhost/bnpl5:latest", deps: []string{}},
				MockEntity{ref: "localhost/risk-engine5:latest", deps: []string{}},
			},
			expected: [][]string{
				{"localhost/base:latest", "localhost/bnpl5:latest", "localhost/risk-engine5:latest"},
				{"localhost/provisioner5:latest"},
			},
		},
		{
			name: "Case 1",
			entities: []depgraph.Entity{
				MockEntity{ref: "infra", deps: []string{}},
				MockEntity{ref: "repo.update", deps: []string{}},
				MockEntity{ref: "bnpl.install_dependencies", deps: []string{"repo.update"}},
				MockEntity{ref: "bnpl.migrate", deps: []string{"bnpl.install_dependencies", "infra"}},
			},
			expected: [][]string{
				{"infra", "repo.update"},
				{"bnpl.install_dependencies"},
				{"bnpl.migrate"},
			},
			expectErr: false,
		},
		{
			name: "Simple linear dependency",
			entities: []depgraph.Entity{
				MockEntity{ref: "A", deps: []string{}},
				MockEntity{ref: "B", deps: []string{"A"}},
				MockEntity{ref: "C", deps: []string{"B"}},
			},
			expected: [][]string{
				{"A"},
				{"B"},
				{"C"},
			},
			expectErr: false,
		},
		{
			name: "Multi-ID",
			entities: []depgraph.Entity{
				MockEntity{ref: "A", deps: []string{}},
				MockEntity{ref: "A", deps: []string{}},
				MockEntity{ref: "B", deps: []string{"A"}},
			},
			expected: [][]string{
				{"A", "A"},
				{"B"},
			},
			expectErr: false,
		},
		{
			name: "Multi-ID deps",
			entities: []depgraph.Entity{
				MockEntity{ref: "A", deps: []string{}},
				MockEntity{ref: "A", deps: []string{}},
				MockEntity{ref: "B", deps: []string{"A"}},
				MockEntity{ref: "C", deps: []string{"A", "A"}},
				MockEntity{ref: "D", deps: []string{"C"}},
			},
			expected: [][]string{
				{"A", "A"},
				{"B", "C"},
				{"D"},
			},
			expectErr: false,
		},
		{
			name: "Parallel dependencies",
			entities: []depgraph.Entity{
				MockEntity{ref: "A", deps: []string{}},
				MockEntity{ref: "B", deps: []string{"A"}},
				MockEntity{ref: "C", deps: []string{"A"}},
				MockEntity{ref: "D", deps: []string{"B", "C"}},
			},
			expected: [][]string{
				{"A"},
				{"B", "C"},
				{"D"},
			},
			expectErr: false,
		},
		{
			name: "Independent entities",
			entities: []depgraph.Entity{
				MockEntity{ref: "A", deps: []string{}},
				MockEntity{ref: "B", deps: []string{}},
				MockEntity{ref: "C", deps: []string{}},
			},
			expected: [][]string{
				{"A", "B", "C"},
			},
			expectErr: false,
		},
		{
			name: "Complex graph",
			entities: []depgraph.Entity{
				MockEntity{ref: "A", deps: []string{}},
				MockEntity{ref: "B", deps: []string{"A"}},
				MockEntity{ref: "C", deps: []string{"A"}},
				MockEntity{ref: "D", deps: []string{"B"}},
				MockEntity{ref: "E", deps: []string{"B", "C"}},
				MockEntity{ref: "F", deps: []string{"D", "E"}},
			},
			expected: [][]string{
				{"A"},
				{"B", "C"},
				{"D", "E"},
				{"F"},
			},
			expectErr: false,
		},
		{
			name: "Circular dependency",
			entities: []depgraph.Entity{
				MockEntity{ref: "A", deps: []string{"B"}},
				MockEntity{ref: "B", deps: []string{"C"}},
				MockEntity{ref: "C", deps: []string{"A"}},
			},
			expected:  nil,
			expectErr: true,
		},
		{
			name: "Missing dependency",
			entities: []depgraph.Entity{
				MockEntity{ref: "A", deps: []string{"B"}}, // B is missing
			},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			steps, err := depgraph.BuildDependencyGraph(test.entities)

			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var result [][]string
				for _, step := range steps {
					var stepIDs []string
					for _, entity := range step {
						stepIDs = append(stepIDs, entity.Ref())
					}
					result = append(result, stepIDs)
				}
				assert.Equal(t, test.expected, result)
			}
		})
	}
}
