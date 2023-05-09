package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	venus_messager_ha "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-messager-ha"
)

var Info = venus_messager_ha.PluginInfo

type DepParams struct {
	Params venus_messager_ha.Config `optional:"true"`

	VenusAuth env.IDeployer `svcname:"VenusAuth"`
	Venus     env.IDeployer `svcname:"Venus"`
	Gateway   env.IDeployer `svcname:"VenusGateway"`

	K8sEnv *env.K8sEnvDeployer
}

func Exec(ctx context.Context, depParams DepParams) (env.IDeployer, error) {
	adminToken, err := depParams.VenusAuth.Param("AdminToken")
	if err != nil {
		return nil, err
	}
	deployer, err := venus_messager_ha.DeployerFromConfig(depParams.K8sEnv, venus_messager_ha.Config{
		NodeUrl:    depParams.Venus.SvcEndpoint().ToMultiAddr(),
		GatewayUrl: depParams.Gateway.SvcEndpoint().ToMultiAddr(),
		AuthUrl:    depParams.VenusAuth.SvcEndpoint().ToHttp(),
		AuthToken:  adminToken.(string),
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
