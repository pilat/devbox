package planner

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/pkg/depgraph"
	"github.com/pilat/devbox/internal/runners"
)

func Start(ctx context.Context, cli docker.Service, log *slog.Logger, f *config.Config) error {
	plan, err := getPlan(ctx, cli, log, f)
	if err != nil {
		return fmt.Errorf("failed to get plan: %v", err)
	}

	err = depgraph.Exec(ctx, plan, func(ctx context.Context, r runners.Runner) error {
		return r.Start(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to execute steps: %v", err)
	}

	return nil
}

func Stop(ctx context.Context, cli docker.Service, log *slog.Logger, f *config.Config) error {
	plan, err := getPlan(ctx, cli, log, f)
	if err != nil {
		return fmt.Errorf("failed to get plan: %v", err)
	}

	err = depgraph.ExecReverse(ctx, plan, func(ctx context.Context, r runners.Runner) error {
		return r.Stop(ctx)
	})

	if err != nil {
		log.Error("Error occurred while stopping images", "error", err)
	}

	return nil
}

func getPlan(ctx context.Context, cli docker.Service, log *slog.Logger, f *config.Config) ([][]runners.Runner, error) {
	networkName := fmt.Sprintf("devbox-%s", f.Name)
	configsName := "configs"

	f.NetworkName = networkName

	candidates := make([]runners.Runner, 0)

	// Provisioner is a special container that has all necessary tools to work with sources
	f.Containers = append(f.Containers, config.ContainerConfig{
		Image:      provisionerImageName,
		Dockerfile: provisionerDockerfile,
	})

	for _, src := range f.Sources {
		candidates = append(candidates, runners.NewSourceRunner(
			cli,
			log,
			f.Name,
			provisionerImageName,
			src,
			[]string{provisionerImageName},
		))
	}

	candidates = append(candidates, runners.NewConfigsRunner(
		cli,
		log,
		f.Name,
		configsName,
		f.Configs,
		[]string{},
	))

	candidates = append(candidates, runners.NewNetworkRunner(
		cli,
		log,
		f,
		[]string{},
	))

	internalImages := make([]string, 0)
	allDependsOn := make([]string, 0)

	// Add internal images to build
	for _, container := range f.Containers {
		dependsOn := make([]string, 0)

		// Extract all "FROM" from Dockerfile
		images, err := extractImages(container.Dockerfile)
		if err != nil {
			return nil, fmt.Errorf("failed to extract images from Dockerfile: %v", err)
		}

		for _, img := range images {
			dependsOn = append(dependsOn, img)
			allDependsOn = append(allDependsOn, img)
		}

		candidates = append(candidates, runners.NewImageRunner(
			cli,
			log,
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
				log,
				dep,
			))
		}
	}

	// Inspect containers: if image is not among internalImages, then it's external
	for _, cont := range f.Services {
		if !slices.Contains(internalImages, cont.Image) {
			candidates = append(candidates, runners.NewPullRunner(
				cli,
				log,
				cont.Image,
			))
		}
	}

	allActionsAndServices := make([]string, 0)
	for _, svc := range f.Services {
		allActionsAndServices = append(allActionsAndServices, svc.Name)
	}
	for _, act := range f.Actions {
		allActionsAndServices = append(allActionsAndServices, act.Name)
	}

	coveredVolumes := make(map[string]struct{})

	inspectVolumes := func(volumes []string) []string {
		coveredV2 := make(map[string]struct{})
		dependsOn := make([]string, 0)

		for _, vol := range volumes {
			elem := strings.Split(vol, ":")
			if len(elem) != 2 {
				continue
			}

			if strings.HasPrefix(elem[0], "/") {
				continue
			}

			elem2 := strings.Split(elem[0], "/")
			volumeName := elem2[0]

			if _, ok := coveredV2[volumeName]; !ok {
				dependsOn = append(dependsOn, volumeName)
				coveredV2[volumeName] = struct{}{}
			}

			_, ok := coveredVolumes[volumeName]
			if !strings.HasPrefix(volumeName, "source.") && !strings.HasPrefix(volumeName, "configs") && !ok {
				candidates = append(candidates, runners.NewVolumeRunner(
					cli,
					log,
					volumeName,
					[]string{},
				))
				coveredVolumes[volumeName] = struct{}{}
			}
		}

		return dependsOn
	}

	for i := range f.Services {
		svc := &f.Services[i]
		dependsOn := make([]string, 0)
		for _, dep := range svc.DependsOn {
			if !slices.Contains(allActionsAndServices, dep) {
				return nil, fmt.Errorf("action %s depends on unknown action %s", svc.Name, dep)
			}

			dependsOn = append(dependsOn, dep)
		}

		dependsOn = append(dependsOn, svc.Image)
		dependsOn = append(dependsOn, networkName)
		dependsOn = append(dependsOn, configsName)
		dependsOn = append(dependsOn, inspectVolumes(svc.Volumes)...)

		candidates = append(candidates, runners.NewServiceRunner(
			cli,
			log.With("service", svc.Name),
			f,
			svc,
			dependsOn,
		))
	}

	for i := range f.Actions {
		act := &f.Actions[i]
		dependsOn := make([]string, 0)
		for _, dep := range act.DependsOn {
			if !slices.Contains(allActionsAndServices, dep) {
				return nil, fmt.Errorf("action %s depends on unknown action %s", act.Name, dep)
			}

			dependsOn = append(dependsOn, dep)
		}

		dependsOn = append(dependsOn, act.Image)
		dependsOn = append(dependsOn, networkName)
		dependsOn = append(dependsOn, configsName)
		dependsOn = append(dependsOn, inspectVolumes(act.Volumes)...)

		candidates = append(candidates, runners.NewActionRunner(
			cli,
			log.With("action", act.Name),
			f,
			act,
			dependsOn,
		))
	}

	// Build dependency graph and execute rounds
	plan, err := depgraph.BuildDependencyGraph(candidates)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %v", err)
	}

	return plan, nil
}

func extractImages(dockerfile string) ([]string, error) {
	var images []string
	imageRegex := regexp.MustCompile(`(?i)^\s*from\s+([^\s]+)`)

	scanner := bufio.NewScanner(strings.NewReader(dockerfile))
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
