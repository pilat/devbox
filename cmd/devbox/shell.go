package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "shell <service>",
		Short: "Run interactive shell in one of the services",
		Long:  "That command will run interactive shell in one of the services",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return suggestRunningServices(ctx, cmd, args, toComplete)
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if err := runShell(ctx, p, args[0]); err != nil {
				return fmt.Errorf("failed to run shell: %w", err)
			}

			return nil
		}),
	}

	root.AddCommand(cmd)
}

func runShell(ctx context.Context, p *project.Project, serviceName string) error {
	_, ok := p.Services[serviceName]
	if !ok {
		return fmt.Errorf("service %q not found", serviceName)
	}

	containerID, err := findContainerID(ctx, p.Name, serviceName)
	if err != nil {
		return fmt.Errorf("failed to find container ID: %w", err)
	}

	var lastShell string
	for _, shell := range []string{"/bin/zsh", "/bin/bash", "/bin/sh", "/bin/ash"} {
		stdout, _, err := containerExec(ctx, containerID, []string{shell})
		if err != nil {
			continue
		}

		if strings.Contains(string(stdout), "OCI runtime exec failed") {
			continue
		}

		lastShell = shell
		break
	}

	if lastShell == "" {
		return fmt.Errorf("failed to find a shell")
	}

	opts := project.RunOptions{
		Service:     serviceName,
		Interactive: true,
		Tty:         true,
		Command:     []string{lastShell},
	}

	_, err = apiService.Exec(ctx, p.Name, opts)
	if err != nil {
		return fmt.Errorf("failed to run shell: %w", err)
	}

	return nil
}

func findContainerID(ctx context.Context, projectName, serviceName string) (string, error) {
	list, err := dockerClient.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterLabels(projectName, serviceName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %v", err)
	}

	if len(list) == 0 {
		return "", fmt.Errorf("service %q is not running", serviceName)
	}

	return list[0].ID, nil
}

func containerExec(ctx context.Context, containerID string, cmd []string) ([]byte, []byte, error) {
	execResp, err := dockerClient.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create exec: %w", err)
	}

	execAttachResp, err := dockerClient.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to attach to exec: %w", err)
	}

	defer execAttachResp.Close()

	var stdout, stderr bytes.Buffer
	done := make(chan error)

	go func() {
		_, err = stdcopy.StdCopy(&stdout, &stderr, execAttachResp.Reader)
		done <- err
	}()

	select {
	case <-done:
		break
	case <-ctx.Done():
		return nil, nil, fmt.Errorf("context cancelled")
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

func filterLabels(projectName, serviceName string) filters.Args {
	pairs := []filters.KeyValuePair{}

	pairs = append(pairs, filters.KeyValuePair{
		Key:   "label",
		Value: fmt.Sprintf("com.docker.compose.project=%s", projectName),
	})
	pairs = append(pairs, filters.KeyValuePair{
		Key:   "label",
		Value: fmt.Sprintf("com.docker.compose.service=%s", serviceName),
	})
	pairs = append(pairs, filters.KeyValuePair{
		Key:   "label",
		Value: "com.docker.compose.container-number=1",
	})

	return filters.NewArgs(pairs...)
}
