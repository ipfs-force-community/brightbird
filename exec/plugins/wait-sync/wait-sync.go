package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/types"
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
	AdminToken types.AdminToken
	K8sEnv     *env.K8sEnvDeployer `json:"-"`
	Venus      env.IVenusDeployer  `json:"-" svcname:"Wallet"`
}

func Exec(ctx context.Context, params TestCaseParams) error {
	pods, err := params.Venus.Pods(ctx)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		err := env.SyncWait(ctx, params.K8sEnv, pod, string(params.AdminToken))
		if err != nil {
			return err
		}
	}
	return nil
}
