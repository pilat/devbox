package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/pilat/devbox/internal/project"
)

func init() {
	var noTty bool

	cmd := &cobra.Command{
		Use:   "shell <service>",
		Short: "Run interactive shell in one of the services",
		Long:  "That command will run interactive shell in one of the services",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: validArgsWrapper(
			func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				if len(args) > 0 {
					return nil, cobra.ShellCompDirectiveNoFileComp
				}
				return suggestRunningServices(ctx, cmd, args, toComplete)
			},
		),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := mgr.AutodetectProject(ctx, projectName)
			if err != nil {
				return fmt.Errorf("failed to detect project: %w", err)
			}

			if err := runShell(ctx, p, args[0], noTty); err != nil {
				return fmt.Errorf("failed to run shell: %w", err)
			}

			return nil
		}),
	}

	cmd.Flags().BoolVarP(&noTty, "no-tty", "t", false, "Do not allocate a pseudo-TTY")

	root.AddCommand(cmd)
}

func runShell(ctx context.Context, p *project.Project, serviceName string, noTtyFlag bool) error {
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
		return errors.New("failed to find a shell")
	}

	var tty bool
	if noTtyFlag {
		tty = false
	} else {
		tty = isTTYAvailable(os.Stdin)
	}

	opts := project.RunOptions{
		Service:     serviceName,
		Interactive: true,
		Tty:         tty,
		Command:     []string{lastShell},
	}

	_, err = apiService.Exec(ctx, p.Name, opts)
	if err != nil {
		return fmt.Errorf("failed to run shell: %w", err)
	}

	return nil
}

func findContainerID(ctx context.Context, projectName, serviceName string) (string, error) {
	list, err := dockerClient.ContainerList(ctx, client.ContainerListOptions{
		All:     true,
		Filters: filterLabels(projectName, serviceName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	if len(list.Items) == 0 {
		return "", fmt.Errorf("service %q is not running", serviceName)
	}

	return list.Items[0].ID, nil
}

func containerExec(ctx context.Context, containerID string, cmd []string) (stdoutBytes, stderrBytes []byte, err error) {
	execResp, err := dockerClient.ExecCreate(ctx, containerID, client.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create exec: %w", err)
	}

	execAttachResp, err := dockerClient.ExecAttach(ctx, execResp.ID, client.ExecAttachOptions{})
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
		return nil, nil, errors.New("context cancelled")
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

func filterLabels(projectName, serviceName string) client.Filters {
	return make(client.Filters).
		Add("label", "com.docker.compose.project="+projectName).
		Add("label", "com.docker.compose.service="+serviceName).
		Add("label", "com.docker.compose.container-number=1")
}

func isTTYAvailable(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}
