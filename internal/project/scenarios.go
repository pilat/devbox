package project

import (
	"fmt"
	"strings"
)

func (p *Project) GetScenarios(filter string) []string {
	filter = strings.ToLower(filter)

	results := []string{}
	for name, s := range p.Scenarios {
		if !strings.HasPrefix(strings.ToLower(name), strings.ToLower(filter)) {
			continue
		}

		results = append(results, fmt.Sprintf("%s\t%s", name, s.Description))
	}

	return results
}
