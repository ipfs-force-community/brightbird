package main

import (
	"context"
	"github.com/hunjixin/brightbird/env/plugin_utils"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/types"
	"github.com/hunjixin/brightbird/version"
	"go.uber.org/fx"
)

var Info = types.PluginInfo{
	Name:        "check_sync",
	Version:     version.Version(),
	Category:    types.TestExec,
	Description: "check if sync successed",
}

type TestCaseParams struct {
	fx.In

	K8sEnv    *env.K8sEnvDeployer `json:"-"`
	Venus     env.IDeployer       `json:"-" svcname:"Venus"`
	VenusAuth env.IDeployer       `json:"-" svcname:"VenusAuth"`
}

func Exec(ctx context.Context, params TestCaseParams) (env.IExec, error) {
	adminToken, err := params.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	pods, err := params.Venus.Pods(ctx)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		err := plugin_utils.SyncWait(ctx, params.K8sEnv, pod, adminToken.(string))
		if err != nil {
			return nil, err
		}
	}
	return env.NewSimpleExec(), nil
}
