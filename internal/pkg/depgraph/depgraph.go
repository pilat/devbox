package depgraph

import (
	"fmt"
)

// Entity represents an interface that every entity must implement.
// It provides an ID and a list of dependencies.
type Entity interface {
	Ref() string
	DependsOn() []string
}

func BuildDependencyGraph[T Entity](entities []T) ([][]T, error) {
	rounds := make([][]T, 0)           // rounds of steps for output
	satisfied := make(map[string]bool) // map of entity ID to whether it was satisfied in previous rounds
	done := make(map[int]bool)         // map of entity index to whether it was already processed

	for {
		steps := make([]int, 0)
		for i := 0; i < len(entities); i++ {
			entity := entities[i]

			if done[i] {
				continue
			}

			// Check if all dependencies were satisfied
			ready := true
			for _, dep := range entity.DependsOn() {
				if !satisfied[dep] {
					ready = false
					break
				}
			}

			if ready {
				done[i] = true
				steps = append(steps, i)
			}
		}

		if len(steps) == 0 {
			return nil, fmt.Errorf("circular dependency detected")
		}

		toRounds := make([]T, 0)
		for _, entityIdx := range steps {
			entity := entities[entityIdx]
			satisfied[entity.Ref()] = true
			toRounds = append(toRounds, entity)
		}
		rounds = append(rounds, toRounds)

		ok := true
		for i := 0; i < len(entities); i++ {
			if !done[i] {
				ok = false
				break
			}
		}

		if ok {
			break
		}
	}

	return rounds, nil
}
