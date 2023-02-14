package main

import (
	"context"
	"github.com/hunjixin/brightbird/env"
	venus_auth "github.com/hunjixin/brightbird/env/impl/venus-auth"
)

var Info = venus_auth.PluginInfo

type DepParams struct {
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusAuthDeployer, error) {
	deployer := venus_auth.NewVenusAuthDeployer(depParams.K8sEnv)
	err := deployer.Deploy(ctx)
	if err != nil {
		return nil, err
	}
	return deployer, nil
}
