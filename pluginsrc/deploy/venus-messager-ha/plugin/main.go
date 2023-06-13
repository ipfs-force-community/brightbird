package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venusmessager "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-messager-ha"
)

func main() {
	plugin.SetupPluginFromStdin(venusmessager.PluginInfo, Exec)
}

type DepParams struct {
	Params venusmessager.Config `optional:"true"`

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

	deployer, err := venusmessager.DeployerFromConfig(depParams.K8sEnv, venusmessager.Config{
		NodeUrl:    venusEndpoint.ToMultiAddr(),
		GatewayUrl: gatewayEndpoint.ToMultiAddr(),
		AuthUrl:    venusAuthEndpoint.ToHTTP(),
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
