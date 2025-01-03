package container

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type Service interface {
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ImageInspectWithRaw(ctx context.Context, imageID string) (types.ImageInspect, error)

	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (string, error)
	ContainerStart(ctx context.Context, containerID string) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)

	ContainerExecCreate(ctx context.Context, containerID string, options ContainerExecOptions) (string, error)
	ContainerExecAttach(ctx context.Context, execID string, options ContainerExecAttachOptions) (types.HijackedResponse, error)
	CopyToContainer(ctx context.Context, containerID, destPath string, content io.Reader, options ContainerCopyToContainerOptions) error

	ContainersList(ctx context.Context, options ContainersListOptions) ([]types.Container, error)
	ContainerRemove(ctx context.Context, containerID string) error

	CreateNetwork(ctx context.Context, name string, options NetworkCreateOptions) error
	DeleteNetwork(ctx context.Context, name string) error
	ListNetworks(ctx context.Context, options NetworksListOptions) ([]network.Summary, error)

	CreateVolume(ctx context.Context, options VolumeCreateOptions) error
	DeleteVolume(ctx context.Context, name string) error
	ListVolumes(ctx context.Context, options VolumeListOptions) (volume.ListResponse, error)
}

type svc struct {
	cli *client.Client
}

func New() (*svc, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &svc{cli: cli}, nil
}

func (s *svc) Close() error {
	return s.cli.Close()
}

func (s *svc) Ping(ctx context.Context) error {
	_, err := s.cli.Ping(ctx)
	return err
}

func (s *svc) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	return s.cli.ImageBuild(ctx, buildContext, options)
}

func (s *svc) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	return s.cli.ImagePull(ctx, refStr, options)
}

func (s *svc) ImageInspectWithRaw(ctx context.Context, imageID string) (types.ImageInspect, error) {
	resp, _, err := s.cli.ImageInspectWithRaw(ctx, imageID)
	return resp, err
}

func (s *svc) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (string, error) {
	resp, err := s.cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, containerName)
	if err != nil {
		return "", err
	}

	return resp.ID, err
}

func (s *svc) ContainerStart(ctx context.Context, containerID string) error {
	return s.cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (s *svc) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return s.cli.ContainerStop(ctx, containerID, options)
}

func (s *svc) ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
	return s.cli.ContainerAttach(ctx, container, options)
}

func (s *svc) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return s.cli.ContainerInspect(ctx, containerID)
}

func (s *svc) ContainerExecCreate(ctx context.Context, containerID string, options ContainerExecOptions) (string, error) {
	resp, err := s.cli.ContainerExecCreate(ctx, containerID, options)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (s *svc) ContainerExecAttach(ctx context.Context, execID string, options ContainerExecAttachOptions) (types.HijackedResponse, error) {
	return s.cli.ContainerExecAttach(ctx, execID, options)
}

func (s *svc) CopyToContainer(ctx context.Context, containerID, destPath string, content io.Reader, options ContainerCopyToContainerOptions) error {
	return s.cli.CopyToContainer(ctx, containerID, destPath, content, options)
}

func (s *svc) ContainersList(ctx context.Context, options ContainersListOptions) ([]types.Container, error) {
	return s.cli.ContainerList(ctx, options)
}

func (s *svc) ContainerRemove(ctx context.Context, containerID string) error {
	return s.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
}

func (s *svc) CreateNetwork(ctx context.Context, name string, options NetworkCreateOptions) error {
	_, err := s.cli.NetworkCreate(ctx, name, options)
	if err != nil {
		return err
	}

	return nil
}

func (s *svc) DeleteNetwork(ctx context.Context, name string) error {
	return s.cli.NetworkRemove(ctx, name)
}

func (s *svc) ListNetworks(ctx context.Context, options NetworksListOptions) ([]network.Summary, error) {
	return s.cli.NetworkList(ctx, options)
}

func (s *svc) CreateVolume(ctx context.Context, options VolumeCreateOptions) error {
	_, err := s.cli.VolumeCreate(ctx, options)
	if err != nil {
		return err
	}

	return nil
}

func (s *svc) DeleteVolume(ctx context.Context, name string) error {
	return s.cli.VolumeRemove(ctx, name, true)
}

func (s *svc) ListVolumes(ctx context.Context, options VolumeListOptions) (volume.ListResponse, error) {
	return s.cli.VolumeList(ctx, options)
}
