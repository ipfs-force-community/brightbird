package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venusutils "github.com/hunjixin/brightbird/env/venus_utils"
	sophonauth "github.com/hunjixin/brightbird/pluginsrc/deploy/sophon-auth"
	"github.com/hunjixin/brightbird/pluginsrc/deploy/venus"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
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
	Venus      venus.VenusDeployReturn           `json:"Venus" description:"venus return"`
	SophonAuth sophonauth.SophonAuthDeployReturn `json:"SophonAuth" description:"sophon auth return"`
}

func Exec(ctx context.Context, k8sEnv *env.K8sEnvDeployer, params TestCaseParams) error {
	pods, err := venus.GetPods(ctx, k8sEnv, params.Venus.InstanceName)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		err := venusutils.SyncWait(ctx, k8sEnv, pod, params.Venus.SvcEndpoint.Port(), params.SophonAuth.AdminToken)
		if err != nil {
			return err
		}
	}
	return nil
}
