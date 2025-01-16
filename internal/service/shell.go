package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pilat/devbox/internal/project"
)

func (a *Service) Shell(ctx context.Context, p *project.Project, serviceName string) error {
	_, ok := p.Services[serviceName]
	if !ok {
		return fmt.Errorf("service %q not found", serviceName)
	}

	containerID, err := a.findContainerID(ctx, p.Name, serviceName)
	if err != nil {
		return fmt.Errorf("failed to find container ID: %w", err)
	}

	var lastShell string
	for _, shell := range []string{"/bin/zsh", "/bin/bash", "/bin/sh", "/bin/ash"} {
		stdout, _, err := a.exec(ctx, containerID, []string{shell})
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

	_, err = a.service.Exec(ctx, p.Name, opts)
	if err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (a *Service) findContainerID(ctx context.Context, projectName, serviceName string) (string, error) {
	list, err := a.apiClient.ContainerList(ctx, container.ListOptions{
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

func (a *Service) exec(ctx context.Context, containerID string, cmd []string) ([]byte, []byte, error) {
	execResp, err := a.apiClient.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create exec: %w", err)
	}

	execAttachResp, err := a.apiClient.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
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
