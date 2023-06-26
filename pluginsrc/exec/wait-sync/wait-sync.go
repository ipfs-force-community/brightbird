package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venusutils "github.com/hunjixin/brightbird/env/venus_utils"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
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
	fx.In

	K8sEnv     *env.K8sEnvDeployer `json:"-"`
	Venus      env.IDeployer       `json:"-" svcname:"Venus"`
	SophonAuth env.IDeployer       `json:"-" svcname:"SophonAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	adminToken, err := params.SophonAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	pods, err := params.Venus.Pods(ctx)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		err := venusutils.SyncWait(ctx, params.K8sEnv, pod, 3453, adminToken.MustString())
		if err != nil {
			return nil, err
		}
	}
	return env.NewSimpleExec(), nil
}
