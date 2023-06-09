package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venusminer "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-miner"
)

func main() {
	plugin.SetupPluginFromStdin(venusminer.PluginInfo, Exec)
}

type DepParams struct {
	Params venusminer.Config `optional:"true"`

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

	deployer, err := venusminer.DeployerFromConfig(depParams.K8sEnv, venusminer.Config{
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
