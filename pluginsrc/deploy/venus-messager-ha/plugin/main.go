package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venus_messager_ha "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-messager-ha"
)

func main() {
	plugin.SetupPluginFromStdin(venus_messager_ha.PluginInfo, Exec)
}

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

	venusEndpoint, err := depParams.Venus.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	gatewayEndpoint, err := depParams.Gateway.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	venusAuthEndpoint, err := depParams.VenusAuth.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	deployer, err := venus_messager_ha.DeployerFromConfig(depParams.K8sEnv, venus_messager_ha.Config{
		NodeUrl:    venusEndpoint.ToMultiAddr(),
		GatewayUrl: gatewayEndpoint.ToMultiAddr(),
		AuthUrl:    venusAuthEndpoint.ToHttp(),
		AuthToken:  adminToken.MustString(),
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
