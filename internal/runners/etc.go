package runners

import (
	"fmt"
	"os"
	"strings"

	"github.com/pilat/devbox/internal/docker"
)

func getMounts(volumes []string) ([]docker.Mount, error) {
	mounts := make([]docker.Mount, 0, len(volumes))

	for _, m := range volumes {
		elem := strings.Split(m, ":")

		// one element means ephemeral volume
		if len(elem) == 1 {
			mounts = append(mounts, docker.Mount{
				Type:   docker.MountTypeVolume,
				Target: elem[0],
				VolumeOptions: &docker.VolumeOptions{
					NoCopy: true, // https://docs.docker.com/engine/storage/volumes/#mounting-a-volume-over-existing-data
				},
			})
			continue
		}

		// two elements means bind mount. if starts from / it is bind mount
		if len(elem) == 2 {
			src := elem[0]
			target := elem[1]

			if strings.HasPrefix(src, "/") {
				mounts = append(mounts, docker.Mount{
					Type:   docker.MountTypeBind,
					Source: elem[0],
					Target: elem[1],
				})
				continue
			}

			// if not, it is volume mount, find out volume name and path
			elem2 := strings.Split(src, "/")
			volumeName := elem2[0]
			subpath := ""
			if len(elem2) > 1 {
				subpath = strings.Join(elem2[1:], "/")
			}

			m := docker.Mount{
				Type:   docker.MountTypeVolume,
				Source: volumeName,
				Target: target,
				VolumeOptions: &docker.VolumeOptions{
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

func getEnvs(envs, env_files []string) ([]string, error) {
	env := []string{}
	env = append(env, envs...)
	for _, envFile := range env_files {
		file, err := os.ReadFile(envFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read env file: %v", err)
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

func filterLabels(project, runnerType, entityName, name string) docker.FilterArgs {
	pairs := []docker.ContainerKeyValuePair{}
	pairs = append(pairs, docker.ContainerKeyValuePair{
		Key:   "label",
		Value: "com.devbox=true",
	})

	pairs = append(pairs, docker.ContainerKeyValuePair{
		Key:   "label",
		Value: fmt.Sprintf("com.devbox.project=%s", project),
	})
	pairs = append(pairs, docker.ContainerKeyValuePair{
		Key:   "label",
		Value: fmt.Sprintf("com.devbox.type=%s", runnerType),
	})
	pairs = append(pairs, docker.ContainerKeyValuePair{
		Key:   "label",
		Value: fmt.Sprintf("com.devbox.name=%s", entityName),
	})
	if name != "" {
		pairs = append(pairs, docker.ContainerKeyValuePair{
			Key:   "name",
			Value: name,
		})
	}

	return docker.NewFiltersArgs(pairs...)
}

func makeLabels(project, runnerType, name string) map[string]string {
	return map[string]string{
		"com.devbox":         "true",
		"com.devbox.project": project,
		"com.devbox.type":    runnerType,
		"com.devbox.name":    name,
	}
}
