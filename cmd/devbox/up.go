package main

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/cert"
	"github.com/pilat/devbox/internal/hosts"
	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/pilat/devbox/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start devbox project",
		Long:  "That command will start devbox project",
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}),
		RunE: runWrapper(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return err
			}

			if err := runProjectUpdate(ctx, p); err != nil {
				return fmt.Errorf("failed to update project: %w", err)
			}

			if err := runHostsUpdate(p, true, false); err != nil {
				return fmt.Errorf("failed to update hosts file: %w", err)
			}

			if err := runCertUpdate(p, true); err != nil {
				return fmt.Errorf("failed to update certificates: %w", err)
			}

			if err := runSourcesUpdate(ctx, p); err != nil {
				return fmt.Errorf("failed to update sources: %w", err)
			}

			if err := p.Validate(); err != nil {
				return fmt.Errorf("failed to validate project: %w", err)
			}

			if err := runBuild(ctx, p); err != nil {
				return fmt.Errorf("failed to build project: %w", err)
			}

			if err := runUp(ctx, p); err != nil {
				return fmt.Errorf("failed to start project: %w", err)
			}

			return nil
		}),
	}

	root.AddCommand(cmd)
}

func runBuild(ctx context.Context, p *project.Project) error {
	cli, err := service.New()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	fmt.Println("[*] Build services...")
	if err := cli.Build(ctx, p); err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}
	fmt.Println("")

	return nil
}

func runUp(ctx context.Context, p *project.Project) error {
	cli, err := service.New()
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	fmt.Println("[*] Up services...")
	if err := cli.Up(ctx, p); err != nil {
		return fmt.Errorf("failed to start project: %w", err)
	}
	fmt.Println("")

	return nil
}

func runCertUpdate(p *project.Project, firstTime bool) error {
	if len(p.CertConfig.Domains) == 0 {
		return nil
	}

	fmt.Println("[*] Setup CA...")

	err := cert.SetupCA(app.AppDir)
	if err != nil && firstTime {
		args := []string{"--", "devbox", "install-ca", "--name", p.Name}

		cmd := exec.Command("sudo", args...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install CA: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to install CA: %w", err)
	}

	fmt.Println("[*] Generate certificates...")
	err = cert.GeneratePair(app.AppDir, p.CertConfig.CertFile, p.CertConfig.KeyFile, p.CertConfig.Domains)
	if err != nil {
		return fmt.Errorf("failed to generate certificates: %w", err)
	}

	return nil
}

func runHostsUpdate(p *project.Project, firstTime, cleanup bool) error {
	if len(p.HostEntities) == 0 && !p.HasHosts {
		return nil // project has no hosts and there were no hosts before
	}

	entities := p.HostEntities
	if cleanup {
		entities = []string{}
	}

	fmt.Println("[*] Update hosts file...")

	err := hosts.Save(p.Name, entities)
	if err != nil && firstTime {
		args := []string{"--", "devbox", "update-hosts", "--name", p.Name}
		if cleanup {
			args = append(args, "--cleanup")
		}

		cmd := exec.Command("sudo", args...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to save hosts file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to save hosts file: %w", err)
	}

	p.HasHosts = len(entities) == 0
	return p.SaveState()
}
