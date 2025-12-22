package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pilat/devbox/internal/app"
	"github.com/pilat/devbox/internal/cert"
	"github.com/pilat/devbox/internal/hosts"
	"github.com/pilat/devbox/internal/manager"
	"github.com/pilat/devbox/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	var profiles []string

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start devbox project",
		Long:  "That command will start devbox project",
		ValidArgsFunction: validArgsWrapper(func(ctx context.Context, cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			p, err := manager.AutodetectProject(projectName)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			allProfiles := p.AllServices().GetProfiles()
			if len(profiles) == 0 && len(allProfiles) > 0 {
				return []string{"--profile"}, cobra.ShellCompDirectiveNoFileComp
			}

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

			if err := runBuild(ctx, p); err != nil {
				return fmt.Errorf("failed to build project: %w", err)
			}

			// By default project has all profiles enabled ("*"). Before we up container we need to reload
			// project to have only default and selected profiles enabled.
			if err := p.Reload(ctx, profiles); err != nil {
				return fmt.Errorf("failed to reload project with profiles: %w", err)
			}

			if err := runUp(ctx, p); err != nil {
				return fmt.Errorf("failed to start project: %w", err)
			}

			return nil
		}),
	}

	cmd.PersistentFlags().StringSliceVarP(&profiles, "profile", "p", []string{}, "Profile to use")

	_ = cmd.RegisterFlagCompletionFunc("profile", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		p, err := manager.AutodetectProject(projectName)
		if err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		return getProfileCompletions(p, toComplete)
	})

	root.AddCommand(cmd)
}

func getProfileCompletions(p *project.Project, toComplete string) ([]string, cobra.ShellCompDirective) {
	profiles := p.AllServices().GetProfiles()
	result := []string{}
	for _, profile := range profiles {
		if strings.HasPrefix(profile, toComplete) {
			result = append(result, profile)
		}
	}

	return result, cobra.ShellCompDirectiveNoFileComp
}

func runBuild(ctx context.Context, p *project.Project) error {
	uniqueImages := map[string]bool{}
	services := []string{}
	for _, service := range p.Services {
		if service.Build == nil || service.Image == "" {
			continue
		}

		uniqueImages[service.Image] = true
		services = append(services, service.Name)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := project.BuildOptions{
		Services: services,
		Quiet:    true,
	}

	fmt.Println("[*] Build services...")
	if err := apiService.Build(ctx, p.Project, opts); err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}

	fmt.Println("")

	return nil
}

func runUp(ctx context.Context, p *project.Project) error {
	timeout := 60 * time.Minute
	opts := project.UpOptions{
		Create: project.CreateOptions{
			RemoveOrphans: true,
			QuietPull:     true,
			Timeout:       &timeout,
			Inherit:       false,
		},
		Start: project.StartOptions{
			Project: p.Project,
		},
	}

	fmt.Println("[*] Up services...")
	if err := apiService.Up(ctx, p.Project, opts); err != nil {
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
	entities := p.HostEntities
	if cleanup {
		entities = []string{}
	}

	changed, err := hosts.Save(p.Name, entities)
	if err != nil && firstTime {
		// Permission denied - retry with sudo
		args := []string{"--", "devbox", "update-hosts", "--name", p.Name}
		if cleanup {
			args = append(args, "--cleanup")
		}

		fmt.Println("[*] Update hosts file...")
		cmd := exec.Command("sudo", args...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to save hosts file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to save hosts file: %w", err)
	} else if changed {
		fmt.Println("[*] Updated hosts file")
	}

	return nil
}
