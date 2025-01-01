package runners

import (
	"context"

	"github.com/pilat/devbox/internal/pkg/depgraph"
)

type Status string

type Runner interface {
	depgraph.Entity

	Start(context.Context) error
	Stop(context.Context) error
	Destroy(context.Context) error
}
