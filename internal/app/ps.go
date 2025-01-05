package app

import (
	"context"
	"fmt"
	"time"

	"github.com/pilat/devbox/internal/composer"
	"github.com/pilat/devbox/internal/table"
)

func (a *app) Ps() error {
	ctx := context.TODO()

	isRunning, err := composer.IsRunning(ctx, a.project)
	if err != nil {
		return fmt.Errorf("failed to check if services are running: %w", err)
	}

	if !isRunning {
		return nil
	}

	errCh := make(chan error)

	go func() {
		count := 0
		for {
			services, err := composer.ListServices(ctx, a.project)
			if err != nil {
				errCh <- fmt.Errorf("failed to list services: %w", err)
			}

			if len(services) == 0 {
				errCh <- nil
			}

			processTable := table.New("Age", "Name", "State", "Health")
			processTable.Compact()
			processTable.SortBy([]table.SortBy{
				{Name: "State", Mode: table.Dsc},
				{Name: "Age", Mode: table.AscAlphaNumeric},
			})

			for _, service := range services {
				processTable.AppendRow(service.Age, service.Name, service.State, service.Health)
			}

			// only return caret to top left corner
			fmt.Print("\033[H")

			// every 5 seconds, clear the screen (to reduce flickering)
			if count%20 == 0 {
				fmt.Print("\033[2J")
			}

			processTable.Render()

			time.Sleep(250 * time.Millisecond)
		}
	}()

	<-errCh

	return nil
}
