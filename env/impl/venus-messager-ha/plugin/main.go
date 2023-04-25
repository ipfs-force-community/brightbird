package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_messager_ha "github.com/hunjixin/brightbird/env/impl/venus-messager-ha"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus_messager_ha.PluginInfo

type DepParams struct {
	Params     venus_messager_ha.Config `optional:"true"`
	K8sEnv     *env.K8sEnvDeployer
	VenusAuth  env.IVenusAuthDeployer
	Venus      env.IVenusDeployer
	Gateway    env.IVenusGatewayDeployer
	AdminToken types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusMessageDeployer, error) {
	deployer, err := venus_messager_ha.DeployerFromConfig(depParams.K8sEnv, venus_messager_ha.Config{
		NodeUrl:    depParams.Venus.SvcEndpoint().ToMultiAddr(),
		GatewayUrl: depParams.Gateway.SvcEndpoint().ToMultiAddr(),
		AuthUrl:    depParams.VenusAuth.SvcEndpoint().ToHttp(),
		AuthToken:  string(depParams.AdminToken),
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
