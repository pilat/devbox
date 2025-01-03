package app

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/pilat/devbox/internal/pkg/container"
	"github.com/pilat/devbox/internal/pkg/depgraph"
	"github.com/pilat/devbox/internal/runners"
)

func (a *app) getPlan(cli container.Service) ([][]runners.Runner, error) {
	candidates := make([]runners.Runner, 0)

	candidates = append(candidates, runners.NewNetworkRunner(
		cli,
		a.cfg,
		[]string{},
	))

	internalImages := make([]string, 0)
	allDependsOn := make([]string, 0)

	// Add internal images to build
	for _, container := range a.cfg.Containers {
		dependsOn := make([]string, 0)

		// Extract all "FROM" from Dockerfile
		images, err := extractImages(a.projectPath, container.Dockerfile)
		if err != nil {
			return nil, fmt.Errorf("failed to extract images from Dockerfile: %w", err)
		}

		for _, img := range images {
			dependsOn = append(dependsOn, img)
			allDependsOn = append(allDependsOn, img)
		}

		candidates = append(candidates, runners.NewImageRunner(
			cli,
			a.cfg,
			&container,
			dependsOn,
		))
		internalImages = append(internalImages, container.Image)
	}

	// Inspect dependsOn: if image is not among internalImages, then it's external
	for _, dep := range allDependsOn {
		if !slices.Contains(internalImages, dep) {
			candidates = append(candidates, runners.NewPullRunner(
				cli,
				dep,
			))
		}
	}

	// Inspect containers: if image is not among internalImages, then it's external
	for _, cont := range a.cfg.Services {
		if !slices.Contains(internalImages, cont.Image) {
			candidates = append(candidates, runners.NewPullRunner(
				cli,
				cont.Image,
			))
		}
	}

	allActionsAndServices := make([]string, 0)
	for _, svc := range a.cfg.Services {
		allActionsAndServices = append(allActionsAndServices, svc.Name)
	}
	for _, act := range a.cfg.Actions {
		allActionsAndServices = append(allActionsAndServices, act.Name)
	}

	coveredVolumes := make(map[string]struct{})
	inspectVolumes := func(volumes []string) []string {
		dependsOn := make([]string, 0)

		for _, vol := range volumes {
			elem := strings.Split(vol, ":")
			if len(elem) != 2 {
				continue
			}

			if strings.HasPrefix(elem[0], "/") { // Absolute path
				continue
			}

			if strings.HasPrefix(elem[0], "./") { // Relative path for configs or sources
				continue
			}

			elem2 := strings.Split(elem[0], "/")
			volumeName := elem2[0] // Volume name

			dependsOn = append(dependsOn, volumeName)

			_, ok := coveredVolumes[volumeName]
			if !ok {
				candidates = append(candidates, runners.NewVolumeRunner(
					cli,
					a.cfg,
					volumeName,
					[]string{},
				))
				coveredVolumes[volumeName] = struct{}{}
			}
		}

		return dependsOn
	}

	for i := range a.cfg.Services {
		svc := &a.cfg.Services[i]
		dependsOn := make([]string, 0)
		for _, dep := range svc.DependsOn {
			if !slices.Contains(allActionsAndServices, dep) {
				return nil, fmt.Errorf("action %s depends on unknown action %s", svc.Name, dep)
			}

			dependsOn = append(dependsOn, dep)
		}

		dependsOn = append(dependsOn, svc.Image)
		dependsOn = append(dependsOn, a.cfg.NetworkName)
		dependsOn = append(dependsOn, inspectVolumes(svc.Volumes)...)

		candidates = append(candidates, runners.NewServiceRunner(
			cli,
			a.cfg,
			svc,
			dependsOn,
		))
	}

	for i := range a.cfg.Actions {
		act := &a.cfg.Actions[i]
		dependsOn := make([]string, 0)
		for _, dep := range act.DependsOn {
			if !slices.Contains(allActionsAndServices, dep) {
				return nil, fmt.Errorf("action %s depends on unknown action %s", act.Name, dep)
			}

			dependsOn = append(dependsOn, dep)
		}

		dependsOn = append(dependsOn, act.Image)
		dependsOn = append(dependsOn, a.cfg.NetworkName)
		dependsOn = append(dependsOn, inspectVolumes(act.Volumes)...)

		candidates = append(candidates, runners.NewActionRunner(
			cli,
			a.cfg,
			act,
			dependsOn,
		))
	}

	// Build dependency graph and execute rounds
	plan, err := depgraph.BuildDependencyGraph(candidates)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	return plan, nil
}

func extractImages(projectPath, dockerfile string) ([]string, error) {
	dockerfilePath := filepath.Join(projectPath, dockerfile)
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Dockerfile: %w", err)
	}

	var images []string
	imageRegex := regexp.MustCompile(`(?i)^\s*from\s+([^\s]+)`)

	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		matches := imageRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			images = append(images, matches[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return images, nil
}
