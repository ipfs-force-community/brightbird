package main

import (
	"context"

	"github.com/hunjixin/brightbird/env"
	"github.com/hunjixin/brightbird/env/plugin"
	venus_market "github.com/hunjixin/brightbird/pluginsrc/deploy/venus-market"
)

var Info = venus_market.PluginInfo

func main() {
	plugin.SetupPluginFromStdin(venus_market.PluginInfo, Exec)
}

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

	venusEndpoint, err := depParams.Venus.SvcEndpoint()
	if err != nil {
		return nil, err
	}

	messagerEndpoint, err := depParams.Messager.SvcEndpoint()
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

	deployer, err := venus_market.DeployerFromConfig(depParams.K8sEnv, venus_market.Config{
		NodeUrl:     venusEndpoint.ToMultiAddr(),
		GatewayUrl:  gatewayEndpoint.ToMultiAddr(),
		MessagerUrl: messagerEndpoint.ToMultiAddr(),
		AuthUrl:     venusAuthEndpoint.ToMultiAddr(),
		AuthToken:   adminToken.MustString(),
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
