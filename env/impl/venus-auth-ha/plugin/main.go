package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_auth_ha "github.com/hunjixin/brightbird/env/impl/venus-auth-ha"
)

var Info = venus_auth_ha.PluginInfo

type DepParams struct {
	Params venus_auth_ha.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusAuthDeployer, error) {
	deployer, err := venus_auth_ha.DeployerFromConfig(depParams.K8sEnv, venus_auth_ha.Config{
		Replicas: 1,
	}, depParams.Params)
	if err != nil {
		return nil, err
	}
	err = deployer.Deploy(ctx)
	if err != nil {
		return nil, err
	}
	return deployer, nil
}
