package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_auth "github.com/hunjixin/brightbird/env/impl/venus-auth"
)

var Info = venus_auth.PluginInfo

type DepParams struct {
	Params venus_auth.Config `optional:"true"`
	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	deployer, err := venus_auth.DeployerFromConfig(depParams.K8sEnv, venus_auth.Config{
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
