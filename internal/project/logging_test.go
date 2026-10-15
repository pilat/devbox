package project

import (
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyDefaultLogging_NoServiceLogging(t *testing.T) {
	p := &Project{
		Project: &types.Project{
			Services: types.Services{
				"web": {Name: "web"},
				"db":  {Name: "db"},
			},
		},
	}

	require.NoError(t, applyDefaultLogging(p))

	for _, name := range []string{"web", "db"} {
		svc := p.Services[name]
		require.NotNil(t, svc.Logging, "service %s must get logging config", name)
		assert.Equal(t, "json-file", svc.Logging.Driver)
		assert.Equal(t, map[string]string{
			"max-size": "10m",
			"max-file": "1",
		}, map[string]string(svc.Logging.Options))
	}
}

func TestApplyDefaultLogging_ExistingLoggingPreserved(t *testing.T) {
	// assign in a separate statement: ServiceConfig embeds the field via
	// ContainerSpec, so a struct literal like ServiceConfig{Logging: ...}
	// does not compile on compose-go v2.16+
	svc := types.ServiceConfig{Name: "web"}
	svc.Logging = &types.LoggingConfig{
		Driver: "fluentd",
		Options: map[string]string{
			"fluentd-address": "localhost:24224",
		},
	}

	p := &Project{
		Project: &types.Project{
			Services: types.Services{
				"web": svc,
			},
		},
	}

	require.NoError(t, applyDefaultLogging(p))

	logging := p.Services["web"].Logging
	require.NotNil(t, logging)
	assert.Equal(t, "fluentd", logging.Driver)
	assert.Equal(t, map[string]string{"fluentd-address": "localhost:24224"}, map[string]string(logging.Options))
}
