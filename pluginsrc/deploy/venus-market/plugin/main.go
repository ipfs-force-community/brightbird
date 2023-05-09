package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_market "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-market"
)

var Info = venus_market.PluginInfo

type DepParams struct {
	Params venus_market.Config `optional:"true"`

	VenusAuth env.IDeployer `svcname:"VenusAuth"`
	Venus     env.IDeployer `svcname:"Venus"`
	Messager  env.IDeployer `svcname:"VenusMessager"`
	Gateway   env.IDeployer `svcname:"VenusGateway"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}

	deployer, err := venus_market.DeployerFromConfig(depParams.K8sEnv, venus_market.Config{
		NodeUrl:     depParams.Venus.SvcEndpoint().ToMultiAddr(),
		GatewayUrl:  depParams.Gateway.SvcEndpoint().ToMultiAddr(),
		MessagerUrl: depParams.Messager.SvcEndpoint().ToMultiAddr(),
		AuthUrl:     depParams.VenusAuth.SvcEndpoint().ToMultiAddr(),
		AuthToken:   adminToken.(string),
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
