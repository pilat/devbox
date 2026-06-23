package project

import (
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v5/pkg/api"
)

type (
	Networks = types.Networks
	Duration = types.Duration
)

var (
	IncludeDependents  = types.IncludeDependents
	IgnoreDependencies = types.IgnoreDependencies
)

type (
	UpOptions     = api.UpOptions
	CreateOptions = api.CreateOptions
	StartOptions  = api.StartOptions
)

type RunOptions = api.RunOptions

type BuildOptions = api.BuildOptions

type DownOptions = api.DownOptions

type PsOptions = api.PsOptions

type LogOptions = api.LogOptions

var (
	ServiceLabel    = api.ServiceLabel
	ProjectLabel    = api.ProjectLabel
	WorkingDirLabel = api.WorkingDirLabel
)
