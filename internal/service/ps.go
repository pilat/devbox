package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pilat/devbox/internal/project"
	"github.com/pilat/devbox/internal/table"
)

func (a *Service) Ps(ctx context.Context, p *project.Project) error {
	errCh := make(chan error)

	go func() {
		count := 0
		for {
			processTable := table.New("Age", "Name", "State", "Health")
			processTable.Compact()
			processTable.SortBy([]table.SortBy{
				{Name: "State", Mode: table.Dsc},
				{Name: "Age", Mode: table.AscAlphaNumeric},
			})

			opts := project.PsOptions{
				Project: p.Project,
				All:     true,
			}

			containers, err := a.service.Ps(ctx, p.Name, opts)
			if err != nil {
				errCh <- fmt.Errorf("failed to list services: %w", err)
			}

			if len(containers) == 0 {
				errCh <- nil
			}

			for _, container := range containers {
				uptimeDuration := time.Since(time.Unix(container.Created, 0))

				h := uptimeDuration.Hours()
				days := int(h / 24)
				uptimeDuration -= time.Duration(days) * 24 * time.Hour
				hh := uptimeDuration.Hours()
				mm := uptimeDuration.Minutes()
				ss := uptimeDuration.Seconds()

				uptimeStr := ""
				if days > 0 {
					uptimeStr += fmt.Sprintf("%dd ", days)
				} else {
					uptimeStr += fmt.Sprintf("%02d:%02d:%02d", int(hh), int(mm)%60, int(ss)%60)
				}

				if container.State == "exited" {
					uptimeStr = ""
				}

				name := container.Labels[project.ServiceLabel]
				if name == "" {
					name = container.Name
				}

				processTable.AppendRow(uptimeStr, name, container.State, container.Health)
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

func (a *Service) GetRunningServices(ctx context.Context, p *project.Project, filter string) ([]string, error) {
	opts := project.PsOptions{
		Project: p.Project,
	}

	containers, err := a.service.Ps(ctx, p.Name, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	results := []string{}
	for _, container := range containers {
		containerName := container.Labels[project.ServiceLabel]
		if !strings.HasPrefix(strings.ToLower(containerName), strings.ToLower(filter)) {
			continue
		}

		results = append(results, containerName)
	}

	return results, nil
}

func (a *Service) IsRunning(ctx context.Context, p *project.Project) (bool, error) {
	opts := project.PsOptions{
		Project: p.Project,
	}

	containers, err := a.service.Ps(ctx, p.Name, opts)
	if err != nil {
		return false, fmt.Errorf("failed to get services: %w", err)
	}

	hasAny := false
	for _, container := range containers {
		hasAny = container.Labels[project.ProjectLabel] == p.Name &&
			container.Labels[project.WorkingDirLabel] == p.WorkingDir
		if hasAny {
			break
		}
	}

	return hasAny, nil
}
