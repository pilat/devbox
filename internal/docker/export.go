package docker

import (
	types2 "github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

// We re-export the types we are using in a codebase to avoid direct dependency on the docker/docker package

type ImageBuildOptions = types.ImageBuildOptions

type ContainerConfig = container.Config
type ContainerHostConfig = container.HostConfig

type ContainerAttachOptions = container.AttachOptions
type ContainerStopOptions = container.StopOptions

type ContainerExecOptions = container.ExecOptions
type ContainerExecAttachOptions = container.ExecAttachOptions

type ContainerCopyToContainerOptions = container.CopyToContainerOptions

type ContainersListOptions = container.ListOptions

var NewFiltersArgs = filters.NewArgs

type FilterArgs = filters.Args

type ContainerKeyValuePair = filters.KeyValuePair

type ImagePullOptions = image.PullOptions

type NetworkNetworkingConfig = network.NetworkingConfig
type NetworkEndpointSettings = network.EndpointSettings

type Mount = mount.Mount

const MountTypeBind = mount.TypeBind
const MountTypeVolume = mount.TypeVolume

type VolumeOptions = mount.VolumeOptions

type PortMap = nat.PortMap
type PortBinding = nat.PortBinding
type Port = nat.Port

var StdCopy = stdcopy.StdCopy

type HealthConfig = container.HealthConfig

type ServicePortConfig = types2.ServicePortConfig

var ParsePortConfig = types2.ParsePortConfig

type VolumeCreateOptions = volume.CreateOptions
type VolumeListOptions = volume.ListOptions
type NetworksListOptions = types.NetworkListOptions

type NetworkCreateOptions = network.CreateOptions
