package main

import (
	"context"
	"time"

	"github.com/ipfs-force-community/brightbird/env"
	"github.com/ipfs-force-community/brightbird/env/plugin"
	venusutils "github.com/ipfs-force-community/brightbird/env/venus_utils"
	sophonauth "github.com/ipfs-force-community/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/ipfs-force-community/brightbird/pluginsrc/deploy/venus"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/version"
)

func main() {
	plugin.SetupPluginFromStdin(Info, Exec)
}

var Info = types.PluginInfo{
	Name:        "check_sync",
	Version:     version.Version(),
	PluginType:  types.TestExec,
	Description: "check if sync successed",
}

type TestCaseParams struct {
	Venus   venus.VenusDeployReturn           `json:"Venus" jsonschema:"Venus"  title:"Venus Daemon" require:"true" description:"venus deploy return"`
	Auth    sophonauth.SophonAuthDeployReturn `json:"SophonAuth" jsonschema:"SophonAuth" title:"Sophon Auth" require:"true" description:"sophon auth return"`
	Timeout string                            `json:"timeout" jsonschema:"timeout" title:"Timeout" default:"20m" require:"true" description:"time to wait power default to 1h"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	pods, err := venus.GetPods(ctx, k8sEnv, params.Venus.InstanceName)
	if err != nil {
		return err
	}

	dur, err := time.ParseDuration(params.Timeout)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, dur)
	defer cancel()

	for _, pod := range pods {
		err := venusutils.SyncWait(ctx, k8sEnv, pod, params.Venus.SvcEndpoint.Port(), params.Auth.AdminToken)
		if err != nil {
			return err
		}
	}
	return nil
}
