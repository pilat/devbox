package composer

import (
	"fmt"

	"github.com/compose-spec/compose-go/v2/types"
)

type Networks = types.Networks
type Duration = types.Duration

var IncludeDependents = types.IncludeDependents
var IgnoreDependencies = types.IgnoreDependencies

type Project struct {
	*types.Project

	Sources   SourceConfigs
	Scenarios ScenarioConfigs

	envFiles []string
}

func (p *Project) WithSelectedServices(names []string, options ...types.DependencyOption) (*Project, error) {
	p2, err := p.Project.WithSelectedServices(names, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to select services: %w", err)
	}

	return &Project{
		Project: p2,
		Sources: p.Sources,
	}, nil
}
