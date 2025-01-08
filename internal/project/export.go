package project

import (
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

type Networks = types.Networks
type Duration = types.Duration

var IncludeDependents = types.IncludeDependents
var IgnoreDependencies = types.IgnoreDependencies

type UpOptions = api.UpOptions
type CreateOptions = api.CreateOptions
type StartOptions = api.StartOptions

type RunOptions = api.RunOptions

type BuildOptions = api.BuildOptions

type DownOptions = api.DownOptions

type PsOptions = api.PsOptions

var ServiceLabel = api.ServiceLabel
var ProjectLabel = api.ProjectLabel
var WorkingDirLabel = api.WorkingDirLabel
