package runners

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
)

type configsRunner struct {
	cli docker.Service
	log *slog.Logger

	appName   string
	name      string
	configs   map[string]config.ConfigFile
	dependsOn []string
}

var _ Runner = (*configsRunner)(nil)

func NewConfigsRunner(cli docker.Service, log *slog.Logger, appName, name string, configs map[string]config.ConfigFile, dependsOn []string) Runner {
	return &configsRunner{
		cli: cli,
		log: log,

		appName:   appName,
		name:      name,
		configs:   configs,
		dependsOn: dependsOn,
	}
}

func (s *configsRunner) Ref() string {
	return s.name
}

func (s *configsRunner) DependsOn() []string {
	return s.dependsOn
}

func (s *configsRunner) Start(ctx context.Context) error {
	err := s.start(ctx)
	if err != nil {
		s.log.Error("Failed to start configs", "error", err)
		return err
	}

	return nil
}

func (s *configsRunner) Stop(ctx context.Context) error {
	err := s.stop(ctx)
	if err != nil {
		s.log.Error("Failed to stop configs", "error", err)
		return err
	}

	return nil
}

func (s *configsRunner) start(ctx context.Context) error {
	s.log.Info("Creating volume", "name", s.name)

	err := s.cli.CreateVolume(ctx, docker.VolumeCreateOptions{
		Name: s.name,
		Labels: map[string]string{
			"devbox":      "true",
			"devbox.type": "configs",
			"devbox.name": s.appName,
		},
	})
	if err != nil {
		return err
	}

	// 2. Run temp container with volume mounted
	networkConfig := &docker.NetworkNetworkingConfig{}

	mounts := []docker.Mount{
		{
			Type:   docker.MountTypeVolume,
			Source: s.name,
			Target: "/configs",
			VolumeOptions: &docker.VolumeOptions{
				NoCopy: true,
			},
		},
	}

	hostConfig := &docker.ContainerHostConfig{
		Mounts:     mounts,
		AutoRemove: true,
	}

	containerConfig := &docker.ContainerConfig{
		Cmd:        []string{"sleep", "infinity"},
		Entrypoint: []string{""},
		Image:      "docker.io/library/alpine:3.20.3",
	}

	containerID, err := s.cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, "")
	if err != nil {
		s.log.Warn("Probably already running", "error", err) // TODO
		return nil
		// return fmt.Errorf("failed to create container: %w", err)
	}

	defer func() {
		// 4. Stop container
		timeout := 0
		stopOptions := docker.ContainerStopOptions{
			Timeout: &timeout,
		}
		_ = s.cli.ContainerStop(context.Background(), containerID, stopOptions)
	}()

	err = s.cli.ContainerStart(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	buf := &bytes.Buffer{}
	tarWriter := tar.NewWriter(buf)
	defer tarWriter.Close()

	for filename, file := range s.configs {
		mode := uint32(0o444)
		if file.Mode != nil {
			mode = *file.Mode
		}

		var uid, gid int
		if file.UID != "" {
			v, err := strconv.Atoi(file.UID)
			if err != nil {
				return err
			}
			uid = v
		}
		if file.GID != "" {
			v, err := strconv.Atoi(file.GID)
			if err != nil {
				return err
			}
			gid = v
		}

		header := &tar.Header{
			Name:    fmt.Sprintf("/configs/%s", filename),
			Size:    int64(len(file.Content)),
			Mode:    int64(mode),
			ModTime: time.Now(),
			Uid:     uid,
			Gid:     gid,
		}

		err := tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("failed to write Dockerfile header to tar: %v", err)
		}
		_, err = tarWriter.Write([]byte(file.Content))
		if err != nil {
			return fmt.Errorf("failed to write Dockerfile content to tar: %v", err)
		}
	}

	// 3. Copy files to volume
	err = s.cli.CopyToContainer(ctx, containerID, "/", buf, docker.ContainerCopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
		CopyUIDGID:                true,
	})
	if err != nil {
		return fmt.Errorf("failed to copy config to container: %v", err)
	}

	return nil
}

func (s *configsRunner) stop(ctx context.Context) error {
	items, err := s.cli.ListVolumes(ctx, docker.VolumeListOptions{
		Filters: docker.NewFiltersArgs(
			docker.ContainerKeyValuePair{
				Key:   "label",
				Value: "devbox=true",
			},
			docker.ContainerKeyValuePair{
				Key:   "label",
				Value: "devbox.type=configs",
			},
			docker.ContainerKeyValuePair{
				Key:   "label",
				Value: fmt.Sprintf("devbox.name=%s", s.appName),
			},
		),
	})
	if err != nil {
		return err
	}

	for _, item := range items.Volumes {
		err := s.cli.DeleteVolume(ctx, item.Name)
		if err != nil {
			return err
		}
	}

	return nil
}
