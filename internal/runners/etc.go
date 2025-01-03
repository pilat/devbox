package runners

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/pkg/container"
	"github.com/pilat/devbox/internal/pkg/utils"
)

func getMounts(cfg *config.Config, volumes []string) ([]container.Mount, error) {
	projectName := cfg.Name
	mounts := make([]container.Mount, 0, len(volumes))

	for _, m := range volumes {
		elem := strings.Split(m, ":")

		// one element means ephemeral volume
		if len(elem) == 1 {
			mounts = append(mounts, container.Mount{
				Type:   container.MountTypeVolume,
				Target: elem[0],
				VolumeOptions: &container.VolumeOptions{
					NoCopy: true, // https://docs.docker.com/engine/storage/volumes/#mounting-a-volume-over-existing-data
				},
			})
			continue
		}

		// two elements means bind mount. if starts from / it is bind mount
		if len(elem) == 2 {
			src := elem[0]
			target := elem[1]

			if strings.HasPrefix(src, "./") { // Relative path for configs or sources
				homedir, err := utils.GetHomeDir()
				if err != nil {
					return nil, fmt.Errorf("failed to get home dir: %w", err)
				}

				projectPath := fmt.Sprintf("%s/.devbox/%s", homedir, projectName)

				src = strings.Replace(src, "./", projectPath+"/", 1)
			}

			if strings.HasPrefix(src, "/") { // Absolute path
				mounts = append(mounts, container.Mount{
					Type:   container.MountTypeBind,
					Source: src,
					Target: target,
				})
				continue
			}

			// if not, it is volume mount, find out volume name and path
			elem2 := strings.Split(src, "/")
			volumeName := fmt.Sprintf("%s-%s", projectName, elem2[0])

			subpath := ""
			if len(elem2) > 1 {
				subpath = strings.Join(elem2[1:], "/")
			}

			m := container.Mount{
				Type:   container.MountTypeVolume,
				Source: volumeName,
				Target: target,
				VolumeOptions: &container.VolumeOptions{
					NoCopy: true,
				},
			}
			if subpath != "" {
				m.VolumeOptions.Subpath = subpath
			}

			mounts = append(mounts, m)
			continue
		}

		return nil, fmt.Errorf("invalid volume format: %s", m)
	}

	return mounts, nil
}

func getEnvs(projectName string, envs, env_files []string) ([]string, error) {
	env := []string{}
	env = append(env, envs...)
	for _, envFile := range env_files {
		if strings.HasPrefix(envFile, "./") { // Relative path for configs or sources
			homedir, err := utils.GetHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home dir: %w", err)
			}

			projectPath := fmt.Sprintf("%s/.devbox/%s", homedir, projectName)
			envFile = strings.Replace(envFile, "./", projectPath+"/", 1)
		}

		file, err := os.ReadFile(envFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read env file: %w", err)
		}

		for _, line := range strings.Split(string(file), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			env = append(env, line)
		}
	}

	return env, nil
}

func filterLabels(project, runnerType, entityName, name string) container.FilterArgs {
	pairs := []container.ContainerKeyValuePair{}
	pairs = append(pairs, container.ContainerKeyValuePair{
		Key:   "label",
		Value: "com.devbox=true",
	})

	pairs = append(pairs, container.ContainerKeyValuePair{
		Key:   "label",
		Value: fmt.Sprintf("com.devbox.project=%s", project),
	})
	pairs = append(pairs, container.ContainerKeyValuePair{
		Key:   "label",
		Value: fmt.Sprintf("com.devbox.type=%s", runnerType),
	})
	pairs = append(pairs, container.ContainerKeyValuePair{
		Key:   "label",
		Value: fmt.Sprintf("com.devbox.name=%s", entityName),
	})
	if name != "" {
		pairs = append(pairs, container.ContainerKeyValuePair{
			Key:   "name",
			Value: name,
		})
	}

	return container.NewFiltersArgs(pairs...)
}

func makeLabels(project, runnerType, name string) map[string]string {
	return map[string]string{
		"com.devbox":         "true",
		"com.devbox.project": project,
		"com.devbox.type":    runnerType,
		"com.devbox.name":    name,
	}
}

func makeNetworkConfig(networkName string, hostAliases ...string) *container.NetworkNetworkingConfig {
	return &container.NetworkNetworkingConfig{
		EndpointsConfig: map[string]*container.NetworkEndpointSettings{
			networkName: {
				NetworkID: networkName,
				Aliases:   hostAliases,
			},
		},
	}
}

func stopContainers(ctx context.Context, cli container.Service, labels container.FilterArgs) error {
	list, err := cli.ContainersList(ctx, container.ContainersListOptions{
		All:     true,
		Filters: labels,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range list {
		timeout := 0
		stopOptions := container.ContainerStopOptions{
			Timeout: &timeout,
		}
		_ = cli.ContainerStop(ctx, c.ID, stopOptions)

		err = cli.ContainerRemove(ctx, c.ID)
		if err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}

	return nil
}
