package runners

import (
	"context"

	"github.com/pilat/devbox/internal/pkg/depgraph"
)

type ServiceType string

const (
	TypeAction  ServiceType = "action"
	TypeImage   ServiceType = "image"
	TypeNetwork ServiceType = "network"
	TypePull    ServiceType = "pull"
	TypeService ServiceType = "service"
	TypeVolume  ServiceType = "volume"
)

type Runner interface {
	depgraph.Entity

	Type() ServiceType
	Start(context.Context) error
	Stop(context.Context) error
	Destroy(context.Context) error
}
