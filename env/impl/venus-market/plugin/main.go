package main

import (
	"context"
	"github.com/hunjixin/brightbird/env"
	venus_market "github.com/hunjixin/brightbird/env/impl/venus-market"
	"github.com/hunjixin/brightbird/types"
)

var Info = venus_market.PluginInfo

type DepParams struct {
	Params     venus_market.Config `optional:"true"`
	K8sEnv     *env.K8sEnvDeployer
	VenusAuth  env.IVenusAuthDeployer
	Venus      env.IVenusDeployer
	Messager   env.IVenusMessageDeployer
	Gateway    env.IVenusGatewayDeployer
	AdminToken types.AdminToken
}

func Exec(ctx context.Context, depParams DepParams) (env.IVenusMarketDeployer, error) {
	deployer, err := venus_market.DeployerFromConfig(depParams.K8sEnv, venus_market.Config{
		NodeUrl:     depParams.Venus.SvcEndpoint().ToMultiAddr(),
		GatewayUrl:  depParams.Gateway.SvcEndpoint().ToMultiAddr(),
		MessagerUrl: depParams.Messager.SvcEndpoint().ToMultiAddr(),
		AuthUrl:     depParams.VenusAuth.SvcEndpoint().ToMultiAddr(),
		AuthToken:   string(depParams.AdminToken),
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
